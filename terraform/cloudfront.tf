# CloudFront distribution for caching and HTTPS
resource "aws_cloudfront_distribution" "app_distribution" {
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "CloudFront distribution for ${var.app_name}"
  default_root_object = "index.html"
  price_class         = "PriceClass_100" # Use only North America and Europe edge locations

  # S3 bucket origin for static website
  origin {
    domain_name = aws_s3_bucket_website_configuration.static_assets.website_endpoint
    origin_id   = "s3-website"
    
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  # API Gateway origin for API routes
  origin {
    domain_name = replace(aws_apigatewayv2_api.lambda.api_endpoint, "https://", "")
    origin_id   = "api-gateway"
    
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  # Default behavior - serve static site from S3
  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD", "OPTIONS"]
    target_origin_id = "s3-website"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 86400
    max_ttl                = 31536000
    compress               = true
  }

  # API behavior - route to API Gateway
  ordered_cache_behavior {
    path_pattern     = "/api/*"
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD", "OPTIONS"]
    target_origin_id = "api-gateway"

    forwarded_values {
      query_string = true
      headers      = ["Authorization", "Content-Type", "X-Requested-With"]
      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
    compress               = true
  }

  # Static assets behavior - cached longer
  ordered_cache_behavior {
    path_pattern     = "/static/*"
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD", "OPTIONS"]
    target_origin_id = "s3-website"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 604800  # 1 week
    max_ttl                = 31536000 # 1 year
    compress               = true
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate_validation.wedding_cert.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  # Custom domain aliases
  aliases = [
    "thekiernan.wedding",
    "www.thekiernan.wedding",
    "thekiernanwedding.com",
    "www.thekiernanwedding.com"
  ]
  
  depends_on = [aws_acm_certificate_validation.wedding_cert]
}

# Output the CloudFront distribution URL
output "cloudfront_url" {
  description = "CloudFront distribution URL"
  value       = "https://${aws_cloudfront_distribution.app_distribution.domain_name}"
}