package ping_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/dakotalillie/webping-lambda/internal/ping"
	"github.com/dakotalillie/webping-lambda/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	cfg, err := testutils.LoadAWSConfigForLocalStack(context.TODO())
	if err != nil {
		t.Fatal("failed to load AWS config:", err)
	}

	t.Run("success", func(t *testing.T) {
		lambdaClient := lambda.NewFromConfig(cfg)
		res, err := lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
			FunctionName: aws.String("ping"),
			LogType:      lambdaTypes.LogTypeTail,
			Payload:      []byte("{}"),
		})
		if err != nil {
			t.Fatal("failed to invoke Lambda function:", err)
		}

		lambdaLogs, err := base64.StdEncoding.DecodeString(*res.LogResult)
		if err != nil {
			t.Fatal("failed to parse Lambda logs:", err)
		}

		var parsed []ping.QueryRecord
		if err = json.Unmarshal(res.Payload, &parsed); err != nil {
			t.Logf("lambda payload: %s", string(res.Payload))
			t.Logf("lambda logs: %s", string(lambdaLogs))
			t.Fatal("failed to parse Lambda response:", err)
		}

		assert.Len(t, parsed, 1)
		if parsed[0].Result != ping.QueryResultPass {
			t.Logf("lambda payload: %s", string(res.Payload))
			t.Logf("lambda logs: %s", string(lambdaLogs))
			t.Fatal("lambda function unexpectedly returned failed result")
		}
	})

	t.Run("failure", func(t *testing.T) {
		lambdaClient := lambda.NewFromConfig(cfg)
		res, err := lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
			FunctionName: aws.String("ping"),
			LogType:      lambdaTypes.LogTypeTail,
			Payload:      []byte(`{"endpoints": ["https://www.google.com/404"]}`),
		})
		if err != nil {
			t.Fatal("failed to invoke Lambda function:", err)
		}

		lambdaLogs, err := base64.StdEncoding.DecodeString(*res.LogResult)
		if err != nil {
			t.Fatal("failed to parse Lambda logs:", err)
		}

		var parsed []ping.QueryRecord
		if err = json.Unmarshal(res.Payload, &parsed); err != nil {
			t.Logf("lambda payload: %s", string(res.Payload))
			t.Logf("lambda logs: %s", string(lambdaLogs))
			t.Fatal("failed to parse Lambda response:", err)
		}

		assert.Len(t, parsed, 1)
		if parsed[0].Result != ping.QueryResultFail {
			t.Logf("lambda payload: %s", string(res.Payload))
			t.Logf("lambda logs: %s", string(lambdaLogs))
			t.Fatal("lambda function unexpectedly returned successful result")
		}
	})

	t.Run("notification", func(t *testing.T) {
		dynamodbClient := dynamodb.NewFromConfig(cfg)
		_, err = dynamodbClient.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]dynamodbTypes.WriteRequest{
				"ping": {
					{
						PutRequest: &dynamodbTypes.PutRequest{
							Item: map[string]dynamodbTypes.AttributeValue{
								"Endpoint":       &dynamodbTypes.AttributeValueMemberS{Value: "https://www.google.com/4042"},
								"ExpirationTime": &dynamodbTypes.AttributeValueMemberN{Value: fmt.Sprint(0)},
								"Result":         &dynamodbTypes.AttributeValueMemberS{Value: string(ping.QueryResultPass)},
								"Timestamp":      &dynamodbTypes.AttributeValueMemberN{Value: fmt.Sprint(0)},
							},
						},
					},
					{
						PutRequest: &dynamodbTypes.PutRequest{
							Item: map[string]dynamodbTypes.AttributeValue{
								"Endpoint":       &dynamodbTypes.AttributeValueMemberS{Value: "https://www.google.com/4042"},
								"ExpirationTime": &dynamodbTypes.AttributeValueMemberN{Value: fmt.Sprint(0)},
								"Result":         &dynamodbTypes.AttributeValueMemberS{Value: string(ping.QueryResultFail)},
								"Timestamp":      &dynamodbTypes.AttributeValueMemberN{Value: fmt.Sprint(10)},
							},
						},
					},
					{
						PutRequest: &dynamodbTypes.PutRequest{
							Item: map[string]dynamodbTypes.AttributeValue{
								"Endpoint":       &dynamodbTypes.AttributeValueMemberS{Value: "https://www.google.com/4042"},
								"ExpirationTime": &dynamodbTypes.AttributeValueMemberN{Value: fmt.Sprint(0)},
								"Result":         &dynamodbTypes.AttributeValueMemberS{Value: string(ping.QueryResultFail)},
								"Timestamp":      &dynamodbTypes.AttributeValueMemberN{Value: fmt.Sprint(20)},
							},
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatal("failed to insert records into database:", err)
		}

		lambdaClient := lambda.NewFromConfig(cfg)
		res, err := lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
			FunctionName: aws.String("ping"),
			LogType:      lambdaTypes.LogTypeTail,
			Payload:      []byte(`{"endpoints": ["https://www.google.com/4042"]}`),
		})
		if err != nil {
			t.Fatal("failed to invoke Lambda function:", err)
		}

		lambdaLogs, err := base64.StdEncoding.DecodeString(*res.LogResult)
		if err != nil {
			t.Fatal("failed to parse Lambda logs:", err)
		}

		sqsClient := sqs.NewFromConfig(cfg)
		out, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String("http://localhost:4566/000000000000/ping"),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     20,
		})
		if err != nil {
			t.Logf("lambda logs: %s", string(lambdaLogs))
			t.Fatal("failed to receive SQS message from queue:", err)
		}

		assert.Len(t, out.Messages, 1)
	})
}
