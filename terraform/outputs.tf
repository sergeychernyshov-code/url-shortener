output "lambda_function_name" {
  description = "The name of the Lambda function"
  value       = aws_lambda_function.url_shortener.function_name
}

output "lambda_function_arn" {
  description = "The ARN of the Lambda function"
  value       = aws_lambda_function.url_shortener.arn
}

output "api_gateway_endpoint" {
  description = "The invoke URL of the API Gateway HTTP API"
  value       = aws_apigatewayv2_api.http_api.api_endpoint
}

output "ecr_repository_url" {
  description = "URL of the ECR repository for the Lambda image"
  value       = aws_ecr_repository.url_shortener.repository_url
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic for alerts"
  value       = aws_sns_topic.alerts.arn
}

