package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dakotalillie/webping-lambda/internal/ping"
)

func main() {
	lambda.Start(ping.HandleRequest)
}
