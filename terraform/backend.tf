terraform {
  # Uncomment this block to use S3 as a remote backend
  # backend "s3" {
  #   bucket         = "thedrewzers-terraform-state"
  #   key            = "thedrewzers/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "terraform-state-lock"
  # }
}

# Note: To use S3 backend, you need to:
# 1. Create an S3 bucket for state storage
# 2. Create a DynamoDB table for state locking
# 3. Uncomment the backend block above and run 'terraform init'
#
# AWS CLI commands to set up backend:
# aws s3 mb s3://thedrewzers-terraform-state --region us-east-1
# aws dynamodb create-table \
#   --table-name terraform-state-lock \
#   --attribute-definitions AttributeName=LockID,AttributeType=S \
#   --key-schema AttributeName=LockID,KeyType=HASH \
#   --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
#   --region us-east-1
