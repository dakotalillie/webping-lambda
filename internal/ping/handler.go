package ping

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var (
	endpoints      []string
	dbTable        string
	dynamodbClient *dynamodb.Client
	snsClient      *sns.Client
	snsTopic       string
)

func HandleRequest(ctx context.Context, req Request) ([]QueryRecord, error) {
	if len(req.Endpoints) > 0 {
		endpoints = req.Endpoints
	} else {
		endpoints = strings.Split(os.Getenv("ENDPOINTS"), ",")
	}

	dbTable = os.Getenv("DB_TABLE")
	snsTopic = os.Getenv("SNS_TOPIC")

	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	dynamodbClient = dynamodb.NewFromConfig(cfg)
	snsClient = sns.NewFromConfig(cfg)

	records := SendRequestsToAllEndpoints(ctx, endpoints)
	for _, record := range records {
		if record.Result == QueryResultFail {
			prevRecords, err := GetPreviousRecords(ctx, record.Endpoint)
			if err != nil {
				return nil, fmt.Errorf("failed to get previous records: %w", err)
			}
			log.Printf("previous records for %s: %+v\n", record.Endpoint, prevRecords)
			if HasTransitionedIntoErrorState(prevRecords) {
				err = PublishToSNS(ctx, record)
				if err != nil {
					return nil, fmt.Errorf("failed to publish to SNS: %w", err)
				}
			} else {
				log.Println("skipping publish to sns")
			}
		}
	}

	if err = InsertRecordsIntoDynamoDB(ctx, records); err != nil {
		return nil, fmt.Errorf("failed to insert records into DynamoDB: %w", err)
	}

	return records, nil
}

func loadAWSConfig(ctx context.Context) (aws.Config, error) {
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

	return config.LoadDefaultConfig(
		ctx,
		config.WithEndpointResolverWithOptions(customResolver),
	)
}
