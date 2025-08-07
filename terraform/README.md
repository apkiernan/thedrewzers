# AWS Lambda Deployment for The Drewzers Wedding Site

This directory contains Terraform configuration for deploying The Drewzers wedding site to AWS Lambda and S3.

## Architecture

The deployment consists of:

- **AWS Lambda**: Hosts the Go web application
- **S3 Bucket**: Stores static assets (images, CSS, JS)
- **API Gateway**: Provides HTTP endpoint for the Lambda function
- **IAM Roles**: Grants necessary permissions

## Prerequisites

1. [AWS CLI](https://aws.amazon.com/cli/) installed and configured
2. [Terraform](https://www.terraform.io/downloads.html) installed (v1.2.0+)
3. [Go](https://golang.org/dl/) installed (v1.18+)

## Deployment Steps

### 1. Build the Lambda Package

From the project root:

```bash
# Use the Lambda Makefile to build the Lambda package
make -f Makefile.lambda build
```

This creates a `build/lambda.zip` file containing the Lambda function.

### 2. Deploy with Terraform

```bash
# Initialize Terraform
cd terraform
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply
```

### 3. Upload Static Assets to S3

After the infrastructure is deployed:

```bash
# Upload static assets to S3
make -f Makefile.lambda upload-static
```

## Customization

Edit `variables.tf` to customize:

- AWS region
- Application name
- S3 bucket name
- Lambda configuration

## Cleanup

To remove all resources:

```bash
cd terraform
terraform destroy
```

## Notes

- The Lambda function uses environment variables to locate the S3 bucket
- Static assets are served directly from S3 via redirects
- API Gateway provides the main HTTP endpoint for the application
