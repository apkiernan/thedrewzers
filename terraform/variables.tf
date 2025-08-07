variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "app_name" {
  description = "Name of the application"
  type        = string
  default     = "thedrewzers-wedding"
}

variable "static_bucket_name" {
  description = "Name of the S3 bucket for static assets"
  type        = string
  default     = "thedrewzers-wedding-static"
}

variable "lambda_zip_path" {
  description = "Path to the Lambda function zip file"
  type        = string
  default     = "../build/lambda.zip"
}
