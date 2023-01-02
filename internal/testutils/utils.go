package testutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAWSConfigForLocalStack loads an AWS config for communicating with AWS resources hosted in LocalStack. This
// config is meant to be used when running tests, outside the LocalStack docker containers themselves, which use a
// separate endpoint.
func LoadAWSConfigForLocalStack(ctx context.Context) (aws.Config, error) {
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

	return config.LoadDefaultConfig(
		ctx,
		config.WithEndpointResolverWithOptions(customResolver),
	)
}
