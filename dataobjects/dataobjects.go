package dataobjects

type Budget struct {
	AccountID       int          `json:"accountID"`
	CreatedDateTime int64        `json:"createdDateTime"`
	RollOnDate      string       `json:"rollOnDate"`
	BudgetItems     []BudgetItem `json:"budgetItems"`
}

type BudgetItem struct {
	ID                         string `json:"ID"`
	CreatedDateTime            int64  `json:"createdDateTime"`
	Amount                     int    `json:"amount"`
	BudgetItemType             int    `json:"budgetItemTypeID"`
	Narrative                  string `json:"narrative"`
	NextDueDate                string `json:"nextDueDate"`
	FrequencyID                int    `json:"frequencyID"`
	LastMatchDate              string `json:"lastMatchedDate"`
	PaymentMethodID            int    `json:"paymentMethodID"`
	PaymentGroupID             int    `json:"paymentGroupID"`
	PaysBills                  bool   `json:"paysBills"`
	BudgetAmount               int    `json:"budgetAmount"`
	ConsecutiveUnmatchedRollOn int    `json:"consecutiveUnmatchedRollOn"`
	ThirdPartyReferenceID      string `json:"thirdPartyReferenceID"`
	CreditorMappingID          int    `json:"creditorMappingID"`
	PaymentStartMonth          int    `json:"paymentStartMonth"`
	NumberOfMonthlyPayments    int    `json:"numberOfMonthlyPayments"`
	SUNNumber                  int    `json:"sunNumber"`
}
