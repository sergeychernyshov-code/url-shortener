terraform {
  backend "s3" {
    bucket         = "sergeys-url-shortener-terraform-state"
    key            = "state/url-shortener/terraform.tfstate"
    region         = "eu-central-1"
    dynamodb_table = "url-shortener-terraform-lock"
    encrypt        = true
  }
}

