terraform {
  backend "s3" {
    bucket         = "my-url-shortener-terraform-state"  # Change to your unique bucket name
    key            = "state/url-shortener/terraform.tfstate"
    region         = "eu-central-1"
    dynamodb_table = "url-shortener-terraform-lock"
    encrypt        = true
  }
}

