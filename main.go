package main

import (
	"fmt"
	"log"
	"net/http"

	"dynamodb-poc-gsi/dynamoclient"
	"dynamodb-poc-gsi/handler"
	"dynamodb-poc-gsi/repository"
)

func main() {

	dc, err := dynamoclient.NewDynamoDBClient("eu-west-1")
	if err != nil {
		panic(err)
	}
	repo := repository.NewRepository(dc)
	handler := handler.NewHandler(repo)

	http.HandleFunc("/BudgetItem/", handler.CreateBudget)
	http.HandleFunc("/Budget/Account/", handler.GetBudget)
	http.HandleFunc("/Budget/Rollon", handler.GetAccountsRollOn)
	http.HandleFunc("/Budget/Rollon/Add", handler.AddAccountRollOn)

	http.HandleFunc("/BudgetItem/Add/", handler.AddBudgetItem)
	http.HandleFunc("/BudgetItem/Fetch/", handler.FetchAccountBudget)
	http.HandleFunc("/BudgetItem/TransactionAdd/", handler.AddBudgetItemInTransaction)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
