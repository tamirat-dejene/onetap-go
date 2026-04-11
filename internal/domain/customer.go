package domain

// Customer represents a canonical bank customer.
type Customer struct {
	ID              int     `json:"id"`
	CustomerName    string  `json:"customerName"`
	Mobile          string  `json:"mobile"`
	AccountNo       string  `json:"accountNo"`
	BranchName      string  `json:"branchName"`
	ProductName     string  `json:"productName"`
	CustomerID      string  `json:"customerId"`
	BranchCode      string  `json:"branchCode"`
	CustomerBalance float64 `json:"customerBalance"`
}
