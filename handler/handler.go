package handler

import (
	"dynamodb-poc-gsi/dataobjects"
	"dynamodb-poc-gsi/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"
)

//DateFormat is the date format used throughout the handler
const DateFormat = "2006-01-02"

type Handler struct {
	Repository repository.IRepository
}

func NewHandler(repo repository.IRepository) Handler {
	return Handler{
		Repository: repo,
	}
}

func (h Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var budget dataobjects.Budget

	path := path.Base(r.URL.Path)
	budget.AccountID, _ = strconv.Atoi(path)

	err := json.NewDecoder(r.Body).Decode(&budget.BudgetItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	budget.CreatedDateTime = time.Now().Unix()

	for i, item := range budget.BudgetItems {
		budget.BudgetItems[i].ID = uuid.New().String()
		budget.BudgetItems[i].CreatedDateTime = budget.CreatedDateTime
		nextDueDate, _ := time.Parse(DateFormat, item.NextDueDate)
		rollOnDate, _ := time.Parse(DateFormat, budget.RollOnDate)

		if rollOnDate.IsZero() || nextDueDate.Before(rollOnDate) {
			budget.RollOnDate = nextDueDate.AddDate(0, 0, -3).Format(DateFormat)
		}

		fmt.Println(item.NextDueDate, budget.RollOnDate, budget.BudgetItems[0].ID)
	}
	_ = json.NewEncoder(w).Encode(&budget)

	fmt.Println(budget)

	h.Repository.SaveBudgetItem(budget)
}

func (h Handler) GetBudget(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("budget here"))
}

func (h Handler) GetBudgetRollOn(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Roll on records here"))
}
