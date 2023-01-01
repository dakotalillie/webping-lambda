package main_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/joho/godotenv"
	"github.com/teris-io/shortid"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func TestSMS(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		// This should not fail the test, because in CI, these values aren't derived from .env
		t.Log("failed to load environment variables from .env:", err)
	}

	ctx := context.Background()

	id, err := shortid.Generate()
	if err != nil {
		t.Fatal("failed to generate id:", err)
	}
	snsMsg := fmt.Sprintf("Testing SMS lambda function: %s", id)

	snsClient, err := initializeSNSClient(ctx)
	if err != nil {
		t.Fatal("failed to initialize SNS client:", err)
	}

	if _, err = snsClient.Publish(ctx, &sns.PublishInput{
		Message:  aws.String(snsMsg),
		TopicArn: aws.String("arn:aws:sns:us-east-1:000000000000:sms"),
	}); err != nil {
		t.Fatal("failed to publish to SNS:", err)
	}

	var targetMsg *twilioApi.ApiV2010Message

	// Give Twilio API time to ingest the message before we query for it
	time.Sleep(5 * time.Second)

	client := twilio.NewRestClient()
	params := twilioApi.ListMessageParams{}
	params.SetDateSentAfter(time.Now().Add(-10 * time.Minute))
	params.SetLimit(5)

	for attempt := 0; attempt < 3; attempt++ {
		messages, err := client.Api.ListMessage(&params)
		if err != nil {
			t.Fatal("failed to list messages from Twilio:", err)
		}

		for i, msg := range messages {
			if msg.Body != nil && strings.Contains(*msg.Body, snsMsg) {
				targetMsg = &messages[i]
				break
			}
		}

		if targetMsg != nil {
			break
		}
		time.Sleep(10 * time.Second)
	}

	if targetMsg == nil {
		t.Fatal("could not find target message")
	} else if targetMsg.Status == nil {
		t.Fatal("target message has no status")
	} else if *targetMsg.Status != "delivered" {
		t.Fatal("target message is not marked as delivered. got:", *targetMsg.Status)
	}
}

func initializeSNSClient(ctx context.Context) (*sns.Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(
		service, region string,
		options ...interface{},
	) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:4566",
			SigningRegion: "us-east-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	return sns.NewFromConfig(cfg), nil
}
