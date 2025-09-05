variable "jwt_secret" {
  description = "Secret key for JWT token signing"
  type        = string
  sensitive   = true
}

variable "admin_email_whitelist" {
  description = "List of allowed admin emails"
  type        = list(string)
  default     = ["apkiernan@gmail.com"]
}
