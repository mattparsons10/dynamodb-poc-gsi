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
	http.HandleFunc("/Budget", handler.GetBudget)
	http.HandleFunc("/Budget/Rollon", handler.GetBudgetRollOn)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
