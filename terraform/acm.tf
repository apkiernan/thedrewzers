# ACM Certificate for CloudFront (must be in us-east-1)
resource "aws_acm_certificate" "wedding_cert" {
  provider = aws.us_east_1

  domain_name = "thekiernan.wedding"
  subject_alternative_names = [
    "*.thekiernan.wedding",
    "thekiernanwedding.com",
    "*.thekiernanwedding.com"
  ]

  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name = "${var.app_name}-certificate"
  }
}

# Certificate validation records
resource "aws_acm_certificate_validation" "wedding_cert" {
  provider = aws.us_east_1

  certificate_arn         = aws_acm_certificate.wedding_cert.arn
  validation_record_fqdns = [for record in aws_acm_certificate.wedding_cert.domain_validation_options : record.resource_record_name]
}

# Output the DNS validation records that need to be added to Porkbun
output "acm_validation_records" {
  description = "DNS validation records to add to Porkbun"
  value = {
    for dvo in aws_acm_certificate.wedding_cert.domain_validation_options : dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  }
}