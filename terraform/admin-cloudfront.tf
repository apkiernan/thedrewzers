# Separate CloudFront distribution for admin subdomains.
# Routes all traffic to API Gateway/Lambda (except static assets).

resource "aws_cloudfront_distribution" "admin_distribution" {
  enabled         = true
  is_ipv6_enabled = true
  comment         = "CloudFront distribution for ${var.app_name} admin"
  price_class     = "PriceClass_100"

  # API Gateway origin for all admin routes
  origin {
    domain_name = replace(aws_apigatewayv2_api.lambda.api_endpoint, "https://", "")
    origin_id   = "api-gateway"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }

    # Custom header so Lambda knows this is an admin request
    # (CloudFront replaces the Host header with the origin domain)
    custom_header {
      name  = "X-Admin-Request"
      value = "true"
    }
  }

  # S3 origin for static assets (CSS, JS, fonts, images)
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

  # Default behavior - route everything to Lambda via API Gateway
  default_cache_behavior {
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

  # Static assets behavior - serve from S3
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
    default_ttl            = 604800
    max_ttl                = 31536000
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

  aliases = [
    "admin.thekiernan.wedding",
    "admin.thekiernanwedding.com"
  ]

  depends_on = [aws_acm_certificate_validation.wedding_cert]
}
