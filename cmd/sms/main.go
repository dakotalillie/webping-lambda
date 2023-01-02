package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dakotalillie/webping-lambda/internal/sms"
)

func main() {
	lambda.Start(sms.HandleRequest)
}
