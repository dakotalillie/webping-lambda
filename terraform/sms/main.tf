terraform {
  required_version = "~> 1.3"

  required_providers {
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.2"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.48"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_iam_role" "this" {
  name = "sms-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

data "archive_file" "this" {
  type        = "zip"
  source_file = "../../../bin/sms"
  output_path = "../../../bin/sms.zip"
}

resource "aws_lambda_function" "this" {
  filename         = data.archive_file.this.output_path
  function_name    = "sms"
  handler          = "sms"
  role             = aws_iam_role.this.arn
  runtime          = "go1.x"
  source_code_hash = data.archive_file.this.output_base64sha256

  environment {
    variables = {
      SMS_ENDPOINT = var.sms_endpoint
    }
  }
}

resource "aws_ssm_parameter" "twilio_account_sid" {
  name  = "/Twilio/AccountSID"
  type  = "String"
  value = var.twilio_account_sid
}

resource "aws_ssm_parameter" "twilio_auth_token" {
  name  = "/Twilio/AuthToken"
  type  = "SecureString"
  value = var.twilio_auth_token
}

resource "aws_ssm_parameter" "twilio_phone_number" {
  name  = "/Twilio/PhoneNumber"
  type  = "String"
  value = var.twilio_phone_number
}

resource "aws_ssm_parameter" "personal_phone_number" {
  name  = "/Personal/PhoneNumber"
  type  = "String"
  value = var.personal_phone_number
}

resource "aws_sns_topic" "this" {
  name = "sms"
}

resource "aws_sns_topic_subscription" "this" {
  endpoint  = aws_lambda_function.this.arn
  protocol  = "lambda"
  topic_arn = aws_sns_topic.this.arn
}
