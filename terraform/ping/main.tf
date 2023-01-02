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
  name = "ping-lambda"

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
  source_file = "../../bin/ping"
  output_path = "../../bin/ping.zip"
}

resource "aws_lambda_function" "this" {
  filename         = data.archive_file.this.output_path
  function_name    = "ping"
  handler          = "ping"
  role             = aws_iam_role.this.arn
  runtime          = "go1.x"
  source_code_hash = data.archive_file.this.output_base64sha256

  environment {
    variables = {
      DB_TABLE  = aws_dynamodb_table.this.name
      ENDPOINTS = join(",", var.endpoints)
      SNS_TOPIC = aws_sns_topic.this.arn
    }
  }
}

resource "aws_dynamodb_table" "this" {
  billing_mode   = "PROVISIONED"
  hash_key       = "Endpoint"
  name           = "ping"
  range_key      = "Timestamp"
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "Endpoint"
    type = "S"
  }

  attribute {
    name = "Timestamp"
    type = "N"
  }
}

resource "aws_sqs_queue" "this" {
  name = "ping"
}

resource "aws_sns_topic" "this" {
  name = "ping"
}

resource "aws_sns_topic_subscription" "this" {
  topic_arn = aws_sns_topic.this.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.this.arn
}
