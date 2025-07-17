output "api_url" {
  value = "${aws_api_gateway_rest_api.shortener_api.execution_arn}/prod"
}

output "api_key" {
  value = aws_api_gateway_api_key.shortener_key.value
}

