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

	var newBudgetItem dataobjects.BudgetItem

	path := path.Base(r.URL.Path)
	budget.AccountID, _ = strconv.Atoi(path)

	err := json.NewDecoder(r.Body).Decode(&newBudgetItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	budget.BudgetItems = append(budget.BudgetItems, newBudgetItem)
	budget.CreatedDateTime = time.Now().UnixMilli()

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
	fmt.Println(budget)

	h.Repository.SaveBudgetItem(budget)

	_ = json.NewEncoder(w).Encode(&budget)
}

func (h Handler) GetBudget(w http.ResponseWriter, r *http.Request) {
	accountId, _ := strconv.Atoi(path.Base(r.URL.Path))
	budget, err := h.Repository.GetBudgetItemsByAccount(accountId)
	if err != nil {
		fmt.Println("Error getting budget items for account", err)
	}
	resp, err := json.Marshal(budget)
	if err != nil {
		fmt.Println("Error marshalling budget items", err)
	}
	w.Write(resp)
}

func (h Handler) GetAccountsRollOn(w http.ResponseWriter, r *http.Request) {
	date := time.Now().Format(DateFormat)

	accounts := h.Repository.GetAccountsToRollOn(date)
	resp, err := json.Marshal(accounts)
	if err != nil {
		fmt.Println("Error marshalling budget items", err)
	}

	w.Write(resp)
}

func (h Handler) AddAccountRollOn(w http.ResponseWriter, r *http.Request) {
	var rollOnData dataobjects.RollOnItem

	err := json.NewDecoder(r.Body).Decode(&rollOnData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.Repository.SaveRollOnDate(rollOnData.RollOnDate, rollOnData.AccountID)

	_ = json.NewEncoder(w).Encode(&rollOnData)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h Handler) AddBudgetItem(w http.ResponseWriter, r *http.Request) {
	path := path.Base(r.URL.Path)
	accountID, _ := strconv.Atoi(path)

	var newBudgetItem dataobjects.BudgetItem
	err := json.NewDecoder(r.Body).Decode(&newBudgetItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	budget, err := h.Repository.GetAccountBudget(accountID)
	if err != nil {
		fmt.Println("Error fetching existing budget")
	}

	if len(budget.BudgetItems) == 0 {
		//There is no budget already so create one with the new budget item
		budget.AccountID = accountID
		createMutableBudgetRecord(&budget, newBudgetItem)
		err = h.Repository.UpdateAccountBudget(budget)
		if err != nil {
			fmt.Println("Error adding new budgeting for account", err)
		}

		// Insert a roll on record for the account
		rollOnDate := calculateNextRollOnDate(budget.BudgetItems, "")

		h.Repository.SaveRollOnDate(rollOnDate, budget.AccountID)
		_ = json.NewEncoder(w).Encode(&budget)
		return
	}

	// There is already a budget so append this budget item to the existing record

	createMutableBudgetRecord(&budget, newBudgetItem)
	err = h.Repository.UpdateAccountBudget(budget)

	if err != nil {
		fmt.Println("Error updating existing budget", err)
	}
	rollOnDate := calculateNextRollOnDate(budget.BudgetItems, "")

	h.Repository.SaveRollOnDate(rollOnDate, budget.AccountID)
	_ = json.NewEncoder(w).Encode(&budget)
}

func (h Handler) FetchAccountBudget(w http.ResponseWriter, r *http.Request) {
	path := path.Base(r.URL.Path)
	accountID, _ := strconv.Atoi(path)

	budget, err := h.Repository.GetAccountBudget(accountID)
	if err != nil {
		fmt.Println("Error fetching existing budget")
	}

	_ = json.NewEncoder(w).Encode(&budget)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func (h Handler) AddBudgetItemInTransaction(w http.ResponseWriter, r *http.Request) {
	path := path.Base(r.URL.Path)
	accountID, _ := strconv.Atoi(path)

	var newBudgetItem dataobjects.BudgetItem
	err := json.NewDecoder(r.Body).Decode(&newBudgetItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	budget, err := h.Repository.GetAccountBudget(accountID)
	if err != nil {
		fmt.Println("Error fetching existing budget")
	}

	if len(budget.BudgetItems) == 0 {
		//There is no budget already so create one with the new budget item
		budget.AccountID = accountID
		createMutableBudgetRecord(&budget, newBudgetItem)
		rollOnDate := calculateNextRollOnDate(budget.BudgetItems, "")

		err = h.Repository.UpdateBudgetingInTransaction(budget, rollOnDate)
		if err != nil {
			fmt.Println("Error adding new budgeting for account", err)
		}

		_ = json.NewEncoder(w).Encode(&budget)
		return
	}

	// There is already a budget so append this budget item to the existing record
	createMutableBudgetRecord(&budget, newBudgetItem)
	rollOnDate := calculateNextRollOnDate(budget.BudgetItems, "")

	err = h.Repository.UpdateBudgetingInTransaction(budget, rollOnDate)

	if err != nil {
		fmt.Println("Error updating existing budget", err)
	}
	_ = json.NewEncoder(w).Encode(&budget)
}

func createMutableBudgetRecord(singleBudget *dataobjects.SingleBudget, item dataobjects.BudgetItem) {
	if singleBudget.CreatedDateTime == 0 {
		singleBudget.CreatedDateTime = time.Now().UnixMilli()
	}

	item.ID = uuid.New().String()
	item.CreatedDateTime = singleBudget.CreatedDateTime

	singleBudget.BudgetItems = append(singleBudget.BudgetItems, item)

}

func calculateNextRollOnDate(budgetItems []dataobjects.BudgetItem, currentRollOn string) (rollOnDate string) {
	rollOn, _ := time.Parse(DateFormat, currentRollOn)

	for _, item := range budgetItems {
		dueDate, _ := time.Parse(DateFormat, item.NextDueDate)
		if rollOn.IsZero() || dueDate.Before(rollOn) {
			rollOn = dueDate.AddDate(0, 0, -3)
		}
	}

	rollOnDate = rollOn.Format(DateFormat)
	return
}
