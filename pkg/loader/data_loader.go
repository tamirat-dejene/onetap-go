package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/repository/memory"
)

// rawCustomer mirrors customers.json for JSON unmarshalling.
type rawCustomer struct {
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

// rawTransaction mirrors transactions.json for JSON unmarshalling.
type rawTransaction struct {
	ID                string  `json:"id"`
	FromAccount       string  `json:"fromAccount"`
	ToAccount         string  `json:"toAccount"`
	Amount            string  `json:"amount"`
	Remark            *string `json:"remark"`
	TransactionType   string  `json:"transactionType"`
	RequestID         string  `json:"requestId"`
	Reference         string  `json:"reference"`
	ThirdPartyRef     *string `json:"thirdPartyReference"`
	InstitutionID     *string `json:"institutionId"`
	ClearedBalance    *string `json:"clearedBalance"`
	TransactionDate   string  `json:"transactionDate"`
	BillerID          *string `json:"billerId"`
}

// SeedAll loads customers and transactions from disk and seeds the in-memory stores.
func SeedAll(
	customersFile, transactionsFile string,
	custRepo *memory.CustomerRepo,
	txRepo *memory.TransactionRepo,
) error {
	if err := seedCustomers(customersFile, custRepo); err != nil {
		return fmt.Errorf("seeding customers: %w", err)
	}
	if err := seedTransactions(transactionsFile, txRepo); err != nil {
		return fmt.Errorf("seeding transactions: %w", err)
	}
	return nil
}

func seedCustomers(path string, repo *memory.CustomerRepo) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raws []rawCustomer
	if err := json.Unmarshal(data, &raws); err != nil {
		return err
	}
	customers := make([]*domain.Customer, 0, len(raws))
	for _, r := range raws {
		customers = append(customers, &domain.Customer{
			ID:              r.ID,
			CustomerName:    r.CustomerName,
			Mobile:          r.Mobile,
			AccountNo:       r.AccountNo,
			BranchName:      r.BranchName,
			ProductName:     r.ProductName,
			CustomerID:      r.CustomerID,
			BranchCode:      r.BranchCode,
			CustomerBalance: r.CustomerBalance,
		})
	}
	repo.Seed(customers)
	return nil
}

func seedTransactions(path string, repo *memory.TransactionRepo) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raws []rawTransaction
	if err := json.Unmarshal(data, &raws); err != nil {
		return err
	}
	var txs []*domain.Transaction
	for _, r := range raws {
		amount, _ := strconv.ParseFloat(r.Amount, 64)
		dateMs, _ := strconv.ParseInt(r.TransactionDate, 10, 64)
		txs = append(txs, &domain.Transaction{
			ID:                 r.ID,
			FromAccount:        r.FromAccount,
			ToAccount:          r.ToAccount,
			Amount:             amount,
			AmountRaw:          r.Amount,
			Remark:             r.Remark,
			TransactionType:    r.TransactionType,
			RequestID:          r.RequestID,
			Reference:          r.Reference,
			ThirdPartyRef:      r.ThirdPartyRef,
			InstitutionID:      r.InstitutionID,
			ClearedBalance:     r.ClearedBalance,
			TransactionDateRaw: r.TransactionDate,
			TransactionDateMs:  dateMs,
			BillerID:           r.BillerID,
		})
	}
	repo.Seed(txs)
	return nil
}
