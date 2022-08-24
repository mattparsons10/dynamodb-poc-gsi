package repository

import (
	"context"
	"dynamodb-poc-gsi/dataobjects"
	"dynamodb-poc-gsi/dynamoclient"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Repository struct {
	DynamoClient dynamoclient.DynamoDBInterface
}

type IRepository interface {
	SaveBudgetItem(item dataobjects.Budget)
	GetBudgetItemsByAccount(AccountId int) ([]dataobjects.Budget, error)
	SaveRollOnDate(date string, accountID int)
	GetAccountsToRollOn(date string) (rollOnAccs []dataobjects.RollOnAccResponse)
	GetAccountBudget(accountID int) (budget dataobjects.SingleBudget, err error)
	UpdateAccountBudget(budget dataobjects.SingleBudget) error
	UpdateBudgetingInTransaction(budget dataobjects.SingleBudget, rollOnDate string) (err error)
}

func NewRepository(dc dynamoclient.DynamoDBInterface) Repository {
	return Repository{
		DynamoClient: dc,
	}
}

func (r Repository) SaveBudgetItem(item dataobjects.Budget) {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		fmt.Println("Error marshalling budget", err)
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String("poc-account-budget"),
		Item:      av,
	}

	_, err = r.DynamoClient.PutItem(context.Background(), input)

	if err != nil {
		fmt.Println("Error putting item", err)
	}
}

func (r Repository) SaveRollOnDate(date string, accountID int) {
	var account dataobjects.Account
	account.AccountID = accountID

	var accs []dataobjects.Account
	accs = append(accs, account)

	var emptyList []dataobjects.Account

	mappedAccs := updateItemMapper(accs)

	emptyMap := updateItemMapper(emptyList)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("poc-budgeting-rollon"),
		Key: map[string]types.AttributeValue{
			"roll_on_date": &types.AttributeValueMemberS{Value: date},
		},
		ExpressionAttributeNames: map[string]string{
			"#accs": "accounts",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accs": &types.AttributeValueMemberL{Value: mappedAccs},
			":el":   &types.AttributeValueMemberL{Value: emptyMap},
		},
		UpdateExpression: aws.String("SET #accs = list_append(if_not_exists(#accs, :el), :accs)"),
	}
	_, err := r.DynamoClient.UpdateItem(context.Background(), input)

	if err != nil {
		fmt.Println("Error putting item", err)
	}
}

func (r Repository) GetBudgetItemsByAccount(accountID int) (budget []dataobjects.Budget, err error) {
	ascOrder := false

	input := &dynamodb.QueryInput{
		TableName: aws.String("poc-account-budget"),
		ExpressionAttributeNames: map[string]string{
			"#pkn": "AccountID",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pkv": &types.AttributeValueMemberN{Value: strconv.Itoa(accountID)},
		},
		KeyConditionExpression: aws.String("#pkn = :pkv"),
		ScanIndexForward:       &ascOrder,
	}

	result, err := r.DynamoClient.Query(context.Background(), input)
	if err != nil {
		fmt.Println(err, "Error sending DynamoDB Query")
		return
	}

	if result.Items != nil {
		err = attributevalue.UnmarshalListOfMaps(result.Items, &budget)
		if err != nil {
			fmt.Println(err, "Error when unmarshalling db item into response")
			return
		}
	}
	return
}

func (r Repository) GetAccountsToRollOn(date string) (rollOnAccs []dataobjects.RollOnAccResponse) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("poc-budgeting-rollon"),
		ExpressionAttributeNames: map[string]string{
			"#pkn": "roll_on_date",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pkv": &types.AttributeValueMemberS{Value: date},
		},
		KeyConditionExpression: aws.String("#pkn = :pkv"),
	}

	result, err := r.DynamoClient.Query(context.Background(), input)
	if err != nil {
		fmt.Println(err, "Error sending DynamoDB Query")
		return
	}

	if result.Items != nil {
		err = attributevalue.UnmarshalListOfMaps(result.Items, &rollOnAccs)
		if err != nil {
			fmt.Println(err, "Error when unmarshalling db item into response")
			return
		}
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r Repository) GetAccountBudget(accountID int) (budget dataobjects.SingleBudget, err error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("poc-mutable-budgeting"),
		ExpressionAttributeNames: map[string]string{
			"#pkn": "accountID",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pkv": &types.AttributeValueMemberN{Value: strconv.Itoa(accountID)},
		},
		KeyConditionExpression: aws.String("#pkn = :pkv"),
	}

	result, err := r.DynamoClient.Query(context.Background(), input)
	if err != nil {
		fmt.Println(err, "Error sending DynamoDB Query")
		return
	}

	if result.Items != nil {
		var budgets []dataobjects.SingleBudget
		err = attributevalue.UnmarshalListOfMaps(result.Items, &budgets)
		if err != nil {
			fmt.Println(err, "Error when unmarshalling db item into response")
			return
		}
		if len(budgets) > 0 {
			budget = budgets[0]
		}
		return
	}
	return
}

func (r Repository) UpdateAccountBudget(budget dataobjects.SingleBudget) error {

	var items []dataobjects.BudgetItem
	items = append(items, budget.BudgetItems...)

	var emptyList []dataobjects.BudgetItem

	mappedItems := updateItemMapper(items)

	emptyMap := updateItemMapper(emptyList)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("poc-mutable-budgeting"),
		Key: map[string]types.AttributeValue{
			"accountID": &types.AttributeValueMemberN{Value: strconv.Itoa(budget.AccountID)},
		},
		ExpressionAttributeNames: map[string]string{
			"#bis": "budgetItems",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":bis": &types.AttributeValueMemberL{Value: mappedItems},
			":el":  &types.AttributeValueMemberL{Value: emptyMap},
		},
		UpdateExpression: aws.String("SET #bis = list_append(if_not_exists(#bis, :el), :bis)"),
	}
	_, err := r.DynamoClient.UpdateItem(context.Background(), input)

	if err != nil {
		fmt.Println("Error putting item", err)
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////
func (r Repository) UpdateBudgetingInTransaction(budget dataobjects.SingleBudget, rollOnDate string) (err error) {
	var account dataobjects.Account
	account.AccountID = budget.AccountID

	var accs []dataobjects.Account
	accs = append(accs, account)

	var emptyAccList []dataobjects.Account

	mappedAccs := updateItemMapper(accs)

	emptyAccMap := updateItemMapper(emptyAccList)

	var items []dataobjects.BudgetItem
	items = append(items, budget.BudgetItems...)

	var emptyBudgetList []dataobjects.BudgetItem

	mappedItems := updateItemMapper(items)

	emptyBudgetMap := updateItemMapper(emptyBudgetList)

	twii := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String("poc-mutable-budgeting"),
					Key: map[string]types.AttributeValue{
						"accountID": &types.AttributeValueMemberN{Value: strconv.Itoa(budget.AccountID)},
					},
					ExpressionAttributeNames: map[string]string{
						"#bis": "budgetItems",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":bis": &types.AttributeValueMemberL{Value: mappedItems},
						":el":  &types.AttributeValueMemberL{Value: emptyBudgetMap},
					},
					UpdateExpression: aws.String("SET #bis = list_append(if_not_exists(#bis, :el), :bis)"),
				},
			},
			{
				Update: &types.Update{
					TableName: aws.String("poc-budgeting-rollon"),
					Key: map[string]types.AttributeValue{
						"roll_on_date": &types.AttributeValueMemberS{Value: rollOnDate},
					},
					ExpressionAttributeNames: map[string]string{
						"#accs": "accounts",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":accs": &types.AttributeValueMemberL{Value: mappedAccs},
						":el":   &types.AttributeValueMemberL{Value: emptyAccMap},
					},
					UpdateExpression: aws.String("SET #accs = list_append(if_not_exists(#accs, :el), :accs)"),
				},
			},
		},
	}
	dynoResp, err := r.DynamoClient.TransactWriteItems(context.Background(), twii)
	if err != nil {
		fmt.Println("Error putting item", err)
		return err
	}
	fmt.Println(dynoResp.ResultMetadata)

	return
}

func updateItemMapper(in interface{}) (resp []types.AttributeValue) {
	resp, err := attributevalue.MarshalList(in)
	if err != nil {
		fmt.Println("could not marshal the accounts")
	}
	return
}
