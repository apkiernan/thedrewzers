# Phase 1: Database and Infrastructure Setup

## Overview
This phase establishes the foundational AWS infrastructure for the RSVP system, including DynamoDB tables, Lambda configuration, and CloudFront distributions.

## Prerequisites
- AWS account with appropriate permissions
- Terraform installed (v1.2.0+)
- Go 1.23+ installed
- Domain configured in Route53

## Step 1: DynamoDB Tables

### 1.1 Create Terraform Variables
Create `terraform/rsvp-variables.tf`:

```hcl
variable "jwt_secret" {
  description = "Secret key for JWT token signing"
  type        = string
  sensitive   = true
}

variable "admin_email_whitelist" {
  description = "List of allowed admin emails"
  type        = list(string)
  default     = ["apkiernan@gmail.com"]
}
```

### 1.2 Add DynamoDB Tables
Add to `terraform/rsvp-database.tf`:

```hcl
# Guests table - stores invitation information
resource "aws_dynamodb_table" "wedding_guests" {
  name           = "${var.project_name}-guests"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "guest_id"

  attribute {
    name = "guest_id"
    type = "S"
  }

  attribute {
    name = "invitation_code"
    type = "S"
  }

  global_secondary_index {
    name            = "invitation_code_index"
    hash_key        = "invitation_code"
    projection_type = "ALL"
  }

  tags = {
    Name        = "${var.project_name}-guests"
    Environment = var.environment
  }
}

# RSVPs table - stores guest responses
resource "aws_dynamodb_table" "wedding_rsvps" {
  name           = "${var.project_name}-rsvps"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "rsvp_id"
  range_key      = "guest_id"

  attribute {
    name = "rsvp_id"
    type = "S"
  }

  attribute {
    name = "guest_id"
    type = "S"
  }

  tags = {
    Name        = "${var.project_name}-rsvps"
    Environment = var.environment
  }
}

# Admin users table
resource "aws_dynamodb_table" "wedding_admins" {
  name           = "${var.project_name}-admins"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "email"

  attribute {
    name = "email"
    type = "S"
  }

  tags = {
    Name        = "${var.project_name}-admins"
    Environment = var.environment
  }
}
```

## Step 2: Lambda Function Updates

### 2.1 Update Lambda Environment Variables
Modify `terraform/lambda.tf`:

```hcl
resource "aws_lambda_function" "main" {
  # ... existing configuration ...
  
  environment {
    variables = {
      # Existing variables
      STATIC_BUCKET = aws_s3_bucket.static_assets.id
      STATIC_URL    = "https://${aws_cloudfront_distribution.main.domain_name}"
      
      # New RSVP variables
      GUESTS_TABLE  = aws_dynamodb_table.wedding_guests.name
      RSVPS_TABLE   = aws_dynamodb_table.wedding_rsvps.name
      ADMINS_TABLE  = aws_dynamodb_table.wedding_admins.name
      JWT_SECRET    = var.jwt_secret
    }
  }
}
```

### 2.2 Add Lambda IAM Permissions
Add to `terraform/iam.tf`:

```hcl
# DynamoDB permissions for Lambda
resource "aws_iam_role_policy" "lambda_dynamodb" {
  name = "${var.project_name}-lambda-dynamodb"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:DeleteItem"
        ]
        Resource = [
          aws_dynamodb_table.wedding_guests.arn,
          "${aws_dynamodb_table.wedding_guests.arn}/index/*",
          aws_dynamodb_table.wedding_rsvps.arn,
          aws_dynamodb_table.wedding_admins.arn
        ]
      }
    ]
  })
}
```

## Step 3: Admin Subdomain Infrastructure

### 3.1 Create Admin CloudFront Distribution
Create `terraform/rsvp-cloudfront.tf`:

```hcl
# Certificate for admin subdomain
resource "aws_acm_certificate" "admin" {
  provider          = aws.us_east_1
  domain_name       = "admin.${var.domain_name}"
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# Certificate validation
resource "aws_route53_record" "admin_cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.admin.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.main.zone_id
}

resource "aws_acm_certificate_validation" "admin" {
  provider                = aws.us_east_1
  certificate_arn         = aws_acm_certificate.admin.arn
  validation_record_fqdns = [for record in aws_route53_record.admin_cert_validation : record.fqdn]
}

# Admin CloudFront distribution
resource "aws_cloudfront_distribution" "admin" {
  enabled = true
  aliases = ["admin.${var.domain_name}"]
  
  origin {
    domain_name = replace(aws_apigatewayv2_api.main.api_endpoint, "https://", "")
    origin_id   = "admin-api-gateway"
    
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }
  
  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "admin-api-gateway"
    
    forwarded_values {
      query_string = true
      cookies {
        forward = "all"
      }
      headers = ["*"]
    }
    
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
    compress               = true
  }
  
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  
  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate_validation.admin.certificate_arn
    ssl_support_method  = "sni-only"
  }
  
  web_acl_id = aws_wafv2_web_acl.admin_protection.arn
  
  tags = {
    Name        = "${var.project_name}-admin-cdn"
    Environment = var.environment
  }
}

# Route53 record for admin subdomain
resource "aws_route53_record" "admin" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = "admin.${var.domain_name}"
  type    = "A"
  
  alias {
    name                   = aws_cloudfront_distribution.admin.domain_name
    zone_id                = aws_cloudfront_distribution.admin.hosted_zone_id
    evaluate_target_health = false
  }
}
```

### 3.2 Add WAF Protection
Create `terraform/rsvp-waf.tf`:

```hcl
# WAF for admin protection
resource "aws_wafv2_web_acl" "admin_protection" {
  provider = aws.us_east_1
  name     = "${var.project_name}-admin-waf"
  scope    = "CLOUDFRONT"

  default_action {
    allow {}
  }

  # Rate limiting rule
  rule {
    name     = "RateLimitRule"
    priority = 1

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 100
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "RateLimitRule"
      sampled_requests_enabled   = true
    }
  }

  # Block common attack patterns
  rule {
    name     = "CommonAttackProtection"
    priority = 2

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        vendor_name = "AWS"
        name        = "AWSManagedRulesCommonRuleSet"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "CommonAttackProtection"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "${var.project_name}-admin-waf"
    sampled_requests_enabled   = true
  }

  tags = {
    Name        = "${var.project_name}-admin-waf"
    Environment = var.environment
  }
}
```

## Step 4: Deploy Infrastructure

### 4.1 Create terraform.tfvars
```hcl
jwt_secret = "your-secret-key-here"  # Generate with: openssl rand -base64 32
```

### 4.2 Deploy
```bash
cd terraform
terraform init
terraform plan -out=rsvp-phase1.plan
terraform apply rsvp-phase1.plan
```

## Step 5: Verify Infrastructure

### 5.1 Check DynamoDB Tables
```bash
aws dynamodb list-tables --region us-east-1 | grep wedding
```

### 5.2 Verify CloudFront Distributions
```bash
aws cloudfront list-distributions --query "DistributionList.Items[?Comment=='${var.project_name}-admin-cdn'].DomainName" --output text
```

### 5.3 Test Admin Subdomain
```bash
curl -I https://admin.thedrewzers.com
# Should return 200 or redirect
```

## Next Steps
- Phase 2: Guest Data Model and QR Generation
- Create initial admin user
- Set up local development environment

## Rollback Plan
If issues occur:
```bash
terraform destroy -target=aws_dynamodb_table.wedding_guests
terraform destroy -target=aws_dynamodb_table.wedding_rsvps
terraform destroy -target=aws_dynamodb_table.wedding_admins
terraform destroy -target=aws_cloudfront_distribution.admin
```