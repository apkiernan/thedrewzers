# Guests table - stores invitation information
resource "aws_dynamodb_table" "wedding_guests" {
  name         = "${var.project_name}-guests"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "guest_id"

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
  name         = "${var.project_name}-rsvps"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "rsvp_id"
  range_key    = "guest_id"

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
  name         = "${var.project_name}-admins"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "email"

  attribute {
    name = "email"
    type = "S"
  }

  tags = {
    Name        = "${var.project_name}-admins"
    Environment = var.environment
  }
}
