# Webping Lambda

This project contains the source code for the AWS Lambda functions used as part
of my Webping project, which alerts me if any of my websites go down.

Currently, there are two lambdas maintained via this repo:

1. The `ping` lambda will make requests to the specified endpoints, and 
   store the outcomes of these requests in DynamoDB. When a given number of 
   requests fail sequentially for a given endpoint, it will publish a 
   message to an SNS topic.
2. The `sms` lambda subscribes to this SNS topic, and sends an SMS message 
   via [Twilio](https://www.twilio.com/).

## Getting Started

1. Clone this repo.
2. Copy `.env.sample` to `.env`. The requisite values can be found via the 
   Twilio UI.
3. Run `make` to build the Lambda binaries and output them into the `bin`
   directory.
4. Run `make render-all` to render the Terraform variables for each of the 
   functions.
5. Ensure you have Docker running locally, as it is needed to run
   [LocalStack](https://localstack.cloud/).
6. Install the [tflocal](https://github.com/localstack/terraform-local) and 
   [awslocal](https://github.com/localstack/awscli-local) helpers via pip. 

   ```shell
   pip install terraform-local awscli-local
   ```
   
7. Run `make start-all` to provision the local Dockerized AWS infrastructure.

Each of the Lambdas has their own invocation Make target. Running `make 
invoke-ping` will invoke the lambda directly, outputting the logs to stdout 
and the return value of the Lambda to `ping.out`. Running `make invoke-sms` 
will run the SMS function indirectly, by publishing to the SNS topic. If it 
works, you should receive an SMS message shortly after.

## Contributing

Generally speaking, the typical workflow for updating these Lambdas is as 
follows:

1. Make some code changes to the function code in the `internal` directory
2. Run `make build-<name>` to rebuild the binary
3. Run `make start-<name>` to redeploy the binary into the local Lambda
4. Run `make invoke-<name>` to run the intended Lambda

Tests are run using `go test -v ./internal/ping` or `go test -v .
/internal/sms`. The tests for the SMS Lambda function actually send an SMS.

To deploy your changes, raise a PR to the `main` branch. Make sure it passes 
CI, then squash and merge it. A GitHub Actions workflow should kick off 
which will rebuild any updated binaries and push them to S3, where they can be
picked up by the workflows for
[webping-infra](https://github.com/dakotalillie/webping-infra).
