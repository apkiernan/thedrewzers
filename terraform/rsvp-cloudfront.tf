# Certificate for admin subdomain
resource "aws_acm_certificate" "admin" {
  provider          = aws.us_east_1
  domain_name       = "admin.${var.domain_name}"
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# Since DNS is managed externally, we'll output the validation records
output "admin_acm_validation_records" {
  description = "DNS validation records for admin subdomain - add these to Porkbun"
  value = {
    for dvo in aws_acm_certificate.admin.domain_validation_options : dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  }
}

# Note: Certificate validation must be done manually
# After adding DNS records, uncomment the following resource:
# resource "aws_acm_certificate_validation" "admin" {
#   provider                = aws.us_east_1
#   certificate_arn         = aws_acm_certificate.admin.arn
# }

# Admin CloudFront distribution
resource "aws_cloudfront_distribution" "admin" {
  enabled = true
  # aliases = ["admin.${var.domain_name}"]  # Uncomment after certificate validation

  origin {
    domain_name = replace(aws_apigatewayv2_api.lambda.api_endpoint, "https://", "")
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
    cloudfront_default_certificate = true
  }

  web_acl_id = aws_wafv2_web_acl.admin_protection.arn

  tags = {
    Name        = "${var.app_name}-admin-cdn"
    Environment = var.environment
  }
}

# Route53 record for admin subdomain - Not needed since DNS is managed on Porkbun
# resource "aws_route53_record" "admin" {
#   zone_id = data.aws_route53_zone.main.zone_id
#   name    = "admin.${var.domain_name}"
#   type    = "A"
#
#   alias {
#     name                   = aws_cloudfront_distribution.admin.domain_name
#     zone_id                = aws_cloudfront_distribution.admin.hosted_zone_id
#     evaluate_target_health = false
#   }
# }

# Output for DNS configuration on Porkbun
output "admin_cloudfront_domain" {
  description = "CloudFront domain for admin subdomain - create CNAME record in Porkbun"
  value = {
    record_name  = "admin.${var.domain_name}"
    record_type  = "CNAME"
    record_value = aws_cloudfront_distribution.admin.domain_name
  }
}
