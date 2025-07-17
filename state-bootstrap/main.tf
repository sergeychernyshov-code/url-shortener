provider "aws" {
  region = "eu-central-1"
}

resource "aws_s3_bucket" "tf_state" {
  bucket = "sergeys-url-shortener-terraform-state" # 🔧 Replace with a unique name!

  versioning {
    enabled = true
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Name = "Terraform State Bucket"
  }
}

resource "aws_dynamodb_table" "tf_lock" {
  name         = "url-shortener-terraform-lock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    Name = "Terraform Lock Table"
  }
}
