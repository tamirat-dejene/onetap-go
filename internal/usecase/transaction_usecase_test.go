package usecase_test

import (
	"testing"

	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/repository/memory"
	"github.com/onetap/salary-advance/internal/usecase"
)

func setupRatingTest(t *testing.T) (*usecase.TransactionUsecase, *memory.CustomerRepo, *memory.TransactionRepo) {
	t.Helper()
	custRepo := memory.NewCustomerRepo()
	txRepo := memory.NewTransactionRepo()

	custRepo.Seed([]*domain.Customer{
		{ID: 1, CustomerName: "ALICE", AccountNo: "ACC001", CustomerBalance: 200000.0},
		{ID: 2, CustomerName: "BOB", AccountNo: "ACC002", CustomerBalance: 50000.0},
	})
	custRepo.SaveVerified(&domain.Customer{ID: 1, CustomerName: "ALICE", AccountNo: "ACC001", CustomerBalance: 200000.0})
	custRepo.SaveVerified(&domain.Customer{ID: 2, CustomerName: "BOB", AccountNo: "ACC002", CustomerBalance: 50000.0})

	cleared1 := "150000.00"
	cleared2 := "148000.00"
	txRepo.Seed([]*domain.Transaction{
		{ID: "t1", FromAccount: "ACC001", Amount: 5000.0, TransactionDateMs: 1700000000000, ClearedBalance: &cleared1},
		{ID: "t2", FromAccount: "ACC001", Amount: 3000.0, TransactionDateMs: 1710000000000, ClearedBalance: &cleared2},
		{ID: "t3", FromAccount: "ACC001", Amount: 2000.0, TransactionDateMs: 1720000000000, ClearedBalance: nil},
	})
	// ACC002 has no transactions → synthetic

	uc := usecase.NewTransactionUsecase(txRepo, custRepo)
	return uc, custRepo, txRepo
}

func TestGetRatingForCustomer_WithTransactions(t *testing.T) {
	uc, _, _ := setupRatingTest(t)

	result, err := uc.GetRatingForCustomer("ACC001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccountNo != "ACC001" {
		t.Errorf("expected account ACC001, got %q", result.AccountNo)
	}
	if result.Rating < 1 || result.Rating > 10 {
		t.Errorf("rating %d is out of range 1-10", result.Rating)
	}
	if result.TransactionCount != 3 {
		t.Errorf("expected 3 transactions, got %d", result.TransactionCount)
	}
	if result.TotalVolume != 10000.0 {
		t.Errorf("expected total volume 10000.00, got %.2f", result.TotalVolume)
	}
	if result.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestGetRatingForCustomer_SyntheticTransactions(t *testing.T) {
	uc, _, _ := setupRatingTest(t)

	// ACC002 has no real transactions
	result, err := uc.GetRatingForCustomer("ACC002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Rating < 1 || result.Rating > 10 {
		t.Errorf("synthetic rating %d is out of range 1-10", result.Rating)
	}
	if result.TransactionCount == 0 {
		t.Error("expected synthetic transactions to be generated")
	}
}

func TestGetRatingForCustomer_NotFound(t *testing.T) {
	uc, _, _ := setupRatingTest(t)

	_, err := uc.GetRatingForCustomer("NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestGetAllRatings_ReturnsAllVerified(t *testing.T) {
	uc, _, _ := setupRatingTest(t)

	ratings, err := uc.GetAllRatings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ratings) != 2 {
		t.Errorf("expected 2 ratings, got %d", len(ratings))
	}
}

func TestGetAllRatings_NoVerifiedCustomers(t *testing.T) {
	custRepo := memory.NewCustomerRepo()
	txRepo := memory.NewTransactionRepo()
	uc := usecase.NewTransactionUsecase(txRepo, custRepo)

	_, err := uc.GetAllRatings()
	if err == nil {
		t.Fatal("expected error when no verified customers exist")
	}
}

func TestRatingDescription_Coverage(t *testing.T) {
	uc, _, _ := setupRatingTest(t)

	// We test descriptions indirectly by getting ratings and checking they are valid strings
	r, _ := uc.GetRatingForCustomer("ACC001")
	validDescriptions := map[string]bool{
		"Very High Risk":            true,
		"High Risk":                 true,
		"Moderate Risk":             true,
		"Low Risk":                  true,
		"Very Low Risk / Excellent": true,
	}
	if !validDescriptions[r.Description] {
		t.Errorf("unexpected description: %q", r.Description)
	}
}

func TestTimeSpanDays_MultiTx(t *testing.T) {
	uc, _, _ := setupRatingTest(t)
	r, _ := uc.GetRatingForCustomer("ACC001")
	// t1=1700000000000ms, t3=1720000000000ms → span = 20000000s / 86400 ≈ 231 days
	if r.TimeSpanDays <= 0 {
		t.Errorf("expected positive time span, got %d", r.TimeSpanDays)
	}
}
