# DNS Setup Guide for Porkbun

This guide will help you configure your domains (thekiernan.wedding and thekiernanwedding.com) to work with your AWS CloudFront distribution.

## Step 1: Deploy Infrastructure First

Before setting up DNS, you need to deploy your infrastructure to get the necessary values:

```bash
# Deploy the infrastructure
make deploy
```

This will output:
1. ACM validation records needed for SSL certificate
2. CloudFront distribution domain name

## Step 2: Validate SSL Certificate

After running `terraform apply`, you'll see output like this:

```
acm_validation_records = {
  "thekiernan.wedding" = {
    name  = "_abc123.thekiernan.wedding."
    type  = "CNAME"
    value = "_def456.acm-validations.aws."
  }
  ...
}
```

### Add these validation records in Porkbun:

1. Log in to Porkbun
2. Go to Domain Management
3. For each domain, click "DNS"
4. Add the validation records:
   - Type: CNAME
   - Host: Copy the subdomain part (e.g., `_abc123`)
   - Answer: Copy the full value
   - TTL: 600 (default)

**Important**: You need to add validation records for ALL domains listed in the output.

## Step 3: Wait for Certificate Validation

The certificate validation can take 5-30 minutes. You can check the status:

```bash
cd terraform
terraform refresh
terraform output acm_validation_records
```

## Step 4: Add DNS Records for Your Domains

Once the certificate is validated and CloudFront is deployed, add these records in Porkbun:

### For thekiernan.wedding:

| Type | Host | Answer | TTL |
|------|------|--------|-----|
| A | @ | CloudFront IPv4 addresses (see below) | 600 |
| AAAA | @ | CloudFront IPv6 addresses (see below) | 600 |
| A | www | CloudFront IPv4 addresses (see below) | 600 |
| AAAA | www | CloudFront IPv6 addresses (see below) | 600 |

### For thekiernanwedding.com:

| Type | Host | Answer | TTL |
|------|------|--------|-----|
| A | @ | CloudFront IPv4 addresses (see below) | 600 |
| AAAA | @ | CloudFront IPv6 addresses (see below) | 600 |
| A | www | CloudFront IPv4 addresses (see below) | 600 |
| AAAA | www | CloudFront IPv6 addresses (see below) | 600 |

### CloudFront IP Addresses:

You'll need to use A/AAAA records pointing to CloudFront's IP addresses. Get the current list by running:

```bash
dig +short d123456abcdef.cloudfront.net
```

Replace `d123456abcdef.cloudfront.net` with your actual CloudFront distribution domain from the Terraform output.

## Alternative: Using CNAME Records (for www subdomains only)

For the www subdomains, you can use CNAME records instead:

| Type | Host | Answer | TTL |
|------|------|--------|-----|
| CNAME | www | your-distribution.cloudfront.net | 600 |

**Note**: You cannot use CNAME for the root domain (@) - you must use A/AAAA records.

## Step 5: Update API Configuration

After DNS propagation (can take up to 48 hours, usually much faster), update your API configuration:

1. Edit `static/js/config.js`
2. Update the API endpoint to use your custom domain:

```javascript
window.API_CONFIG = {
    apiEndpoint: 'https://thekiernan.wedding/api',
};
```

3. Rebuild and redeploy:

```bash
make static-deploy
```

## Troubleshooting

### Certificate Not Validating
- Make sure you added ALL validation records exactly as shown
- Check for typos in the DNS records
- Wait at least 30 minutes before troubleshooting

### Site Not Loading
- DNS propagation can take up to 48 hours
- Use `dig` or `nslookup` to verify DNS records:
  ```bash
  dig thekiernan.wedding
  dig www.thekiernan.wedding
  ```

### SSL Errors
- Ensure the certificate is fully validated in ACM
- Check that CloudFront is using the correct certificate
- Verify all domain aliases are included in the certificate

## Monitoring

You can monitor your domains at:
- https://dnschecker.org/
- https://www.whatsmydns.net/

Enter your domain to see global DNS propagation status.