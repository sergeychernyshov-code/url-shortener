locals {
  env_suffix       = var.environment == "prod" ? "" : "-${var.environment}"
  lambda_name      = "${var.lambda_function_name}${local.env_suffix}"
  sns_topic_name   = "url-shortener-alerts${local.env_suffix}"
  api_name         = "url-shortener-api${local.env_suffix}"
  lambda_role_name = "url-shortener-lambda-exec-role${local.env_suffix}"

  dynamodb_table_name = "url-shortener-table${local.env_suffix}"
  ecr_repo_name       = "url-shortener-lambda${local.env_suffix}"
}

resource "aws_dynamodb_table" "url_shortener" {
  name         = local.dynamodb_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "code"

  attribute {
    name = "code"
    type = "S"
  }
}

resource "aws_ecr_repository" "url_shortener" {
  name = local.ecr_repo_name
}

resource "aws_ecr_repository_policy" "lambda_repo_policy" {
  repository = aws_ecr_repository.url_shortener.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = "AllowLambdaPull"
      Effect    = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      Action = [
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:BatchCheckLayerAvailability"
      ]
    }]
  })
}

resource "aws_iam_role" "lambda_exec" {
  name = local.lambda_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy" "dynamo_access" {
  name = "url-shortener-dynamo-access${local.env_suffix}"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem"
      ]
      Resource = aws_dynamodb_table.url_shortener.arn
    }]
  })
}

resource "aws_lambda_function" "url_shortener" {
  function_name = local.lambda_name
  package_type  = "Image"
  image_uri     = var.image_uri
  role          = aws_iam_role.lambda_exec.arn
  timeout       = 10
  memory_size   = 256

  environment {
    variables = {
      DYNAMO_TABLE   = local.dynamodb_table_name
      API_AUTH_TOKEN = var.api_auth_token
    }
  }
}

resource "aws_apigatewayv2_api" "http_api" {
  name          = local.api_name
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id                 = aws_apigatewayv2_api.http_api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.url_shortener.arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_stage" "default_stage" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_lambda_permission" "apigw_invoke" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.url_shortener.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

resource "aws_sns_topic" "alerts" {
  name = local.sns_topic_name
}

resource "aws_sns_topic_subscription" "email_sub" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.sns_email
}

resource "aws_cloudwatch_metric_alarm" "lambda_error_alarm" {
  alarm_name          = "url-shortener-lambda-errors${local.env_suffix}"
  alarm_description   = "Alarm when the Lambda function experiences errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_actions       = [aws_sns_topic.alerts.arn]
  dimensions = {
    FunctionName = aws_lambda_function.url_shortener.function_name
  }
}

resource "aws_cloudwatch_metric_alarm" "lambda_throttle_alarm" {
  alarm_name          = "url-shortener-lambda-throttles${local.env_suffix}"
  alarm_description   = "Alarm when the Lambda function is throttled"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Throttles"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_actions       = [aws_sns_topic.alerts.arn]
  dimensions = {
    FunctionName = aws_lambda_function.url_shortener.function_name
  }
}