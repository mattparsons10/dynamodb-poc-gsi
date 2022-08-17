package dynamoclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client DynamoDBInterface
}

type DynamoDBInterface interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

func NewDynamoDBClient(region string) (di DynamoDBInterface, err error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return
	}
	cfg.Region = region
	di = dynamodb.NewFromConfig(cfg)

	return
}
