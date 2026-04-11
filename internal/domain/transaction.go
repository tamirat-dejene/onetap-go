package domain

// Transaction represents a financial transaction record.
type Transaction struct {
	ID                 string   `json:"id"`
	FromAccount        string   `json:"fromAccount"`
	ToAccount          string   `json:"toAccount"`
	Amount             float64  `json:"-"` // parsed from string
	AmountRaw          string   `json:"amount"`
	Remark             *string  `json:"remark"`
	TransactionType    string   `json:"transactionType"`
	RequestID          string   `json:"requestId"`
	Reference          string   `json:"reference"`
	ThirdPartyRef      *string  `json:"thirdPartyReference"`
	InstitutionID      *string  `json:"institutionId"`
	ClearedBalance     *string  `json:"clearedBalance"`
	TransactionDateRaw string   `json:"transactionDate"`
	TransactionDateMs  int64    `json:"-"` // parsed from string millis
	BillerID           *string  `json:"billerId"`
}
