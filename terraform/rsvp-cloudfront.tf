# Single-distribution setup: admin subdomains are aliases on app_distribution.
# Output for DNS configuration on Porkbun.
output "admin_cloudfront_domain" {
  description = "CloudFront domains for admin subdomains - create CNAME records in Porkbun"
  value = {
    "admin.thekiernan.wedding" = {
      record_type  = "CNAME"
      record_value = aws_cloudfront_distribution.app_distribution.domain_name
    }
    "admin.thekiernanwedding.com" = {
      record_type  = "CNAME"
      record_value = aws_cloudfront_distribution.app_distribution.domain_name
    }
  }
}
