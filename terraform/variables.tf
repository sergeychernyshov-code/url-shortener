variable "image_uri" {
  description = "The ECR image URI for the Lambda function"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Deployment environment (e.g. dev, prod)"
  type        = string
  default     = "prod"
}

variable "lambda_function_name" {
  description = "Name of the Lambda function"
  type        = string
  default     = "url-shortener"
}

variable "sns_email" {
  description = "Email address for SNS alarm notifications"
  type        = string
  default     = "sergiy.chernyshow@gmail.com"  # Replace or override in tfvars or via CLI
}
