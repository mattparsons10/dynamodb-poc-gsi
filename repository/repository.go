package repository

import (
	"context"
	"dynamodb-poc-gsi/dataobjects"
	"dynamodb-poc-gsi/dynamoclient"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Repository struct {
	DynamoClient dynamoclient.DynamoDBInterface
}

type IRepository interface {
	SaveBudgetItem(item dataobjects.Budget)
	GetBudgetItemsByAccount(AccountId int) ([]dataobjects.BudgetItem, error)
}

func NewRepository(dc dynamoclient.DynamoDBInterface) Repository {
	return Repository{
		DynamoClient: dc,
	}
}

func (r Repository) SaveBudgetItem(item dataobjects.Budget) {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		fmt.Println("Error mapping item")
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("poc-account-budget"),
	}

	_, err = r.DynamoClient.PutItem(context.Background(), input)

	if err != nil {
		fmt.Println("Error putting item")
	}
}

func (r Repository) GetBudgetItemsByAccount(AccountId int) ([]dataobjects.BudgetItem, error) {
	return nil, nil
}
