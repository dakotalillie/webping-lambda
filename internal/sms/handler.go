package sms

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioParams struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	ToNumber   string
}

func HandleRequest(ctx context.Context, req events.SNSEvent) (string, error) {
	record := req.Records[0].SNS

	ssmClient, err := initializeSSMClient(ctx)
	if err != nil {
		return "Failed to initialize SSM client: " + err.Error(), err
	}

	twilioParams, err := getTwilioParamsFromSSM(ctx, ssmClient)
	if err != nil {
		return "Failed to get twilio params from SSM: " + err.Error(), err
	}

	// Twilio requires that credentials be provided via environment variables
	if err = os.Setenv("TWILIO_ACCOUNT_SID", twilioParams.AccountSID); err != nil {
		return "Failed to set TWILIO_ACCOUNT_SID env var: " + err.Error(), err
	}
	if err = os.Setenv("TWILIO_AUTH_TOKEN", twilioParams.AuthToken); err != nil {
		return "Failed to set TWILIO_AUTH_TOKEN env var: " + err.Error(), err
	}

	client := twilio.NewRestClient()
	messageParams := &twilioApi.CreateMessageParams{}
	messageParams.SetBody(record.Message)
	messageParams.SetFrom(twilioParams.FromNumber)
	messageParams.SetTo(twilioParams.ToNumber)
	resp, err := client.Api.CreateMessage(messageParams)
	if err != nil {
		return "Failed to send SMS: " + err.Error(), err
	}

	var messageSID string
	if resp.Sid != nil {
		messageSID = *resp.Sid
	}

	return messageSID, nil
}

func initializeSSMClient(ctx context.Context) (*ssm.Client, error) {
	region := os.Getenv("AWS_REGION")
	localstackHostname := os.Getenv("LOCALSTACK_HOSTNAME")

	customResolver := aws.EndpointResolverWithOptionsFunc(func(
		service, region string,
		options ...interface{},
	) (aws.Endpoint, error) {
		if localstackHostname != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           fmt.Sprintf("http://%s:4566", localstackHostname),
				SigningRegion: region,
			}, nil
		}

		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	return ssm.NewFromConfig(cfg), nil
}

func getTwilioParamsFromSSM(ctx context.Context, ssmClient *ssm.Client) (TwilioParams, error) {
	out, err := ssmClient.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          []string{"/Twilio/AccountSID", "/Twilio/AuthToken", "/Twilio/PhoneNumber", "/Personal/PhoneNumber"},
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return TwilioParams{}, fmt.Errorf("failed to get parameters from ssm: %w", err)
	}

	var params TwilioParams
	for _, p := range out.Parameters {
		switch *p.Name {
		case "/Twilio/AccountSID":
			params.AccountSID = *p.Value
		case "/Twilio/AuthToken":
			params.AuthToken = *p.Value
		case "/Twilio/PhoneNumber":
			params.FromNumber = *p.Value
		case "/Personal/PhoneNumber":
			params.ToNumber = *p.Value
		}
	}

	return params, nil
}
