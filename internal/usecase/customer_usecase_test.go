package usecase_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/repository/memory"
	"github.com/onetap/salary-advance/internal/usecase"
)

func makeCustomerRepo() *memory.CustomerRepo {
	repo := memory.NewCustomerRepo()
	repo.Seed([]*domain.Customer{
		{ID: 1, CustomerName: "MEHADI ALIYE MOHAMMED", AccountNo: "1050001035901", CustomerBalance: 282180.89},
		{ID: 2, CustomerName: "USMAN AHMED UMER", AccountNo: "1050001032701", CustomerBalance: 45837.54},
		{ID: 3, CustomerName: "ABDULEFETAH ABDURAHAMAN ALFEK", AccountNo: "1010001046601", CustomerBalance: 282625.04},
		{ID: 4, CustomerName: "ADNAN MOHAMMED YUSUF", AccountNo: "1050017670001", CustomerBalance: 375612.61},
	})
	return repo
}

func makeSampleCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.csv")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	return path
}

func TestValidateAndPersist_AllValid(t *testing.T) {
	repo := makeCustomerRepo()
	csv := "customerName,accountNo\nMEHADI ALIYE MOHAMMED,1050001035901\nUSMAN AHMED UMER,1050001032701\n"
	uc := usecase.NewCustomerUsecase(repo, makeSampleCSV(t, csv))

	result, err := uc.ValidateAndPersist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total=2, got %d", result.Total)
	}
	if result.Verified != 2 {
		t.Errorf("expected verified=2, got %d", result.Verified)
	}
	if result.Failed != 0 {
		t.Errorf("expected failed=0, got %d", result.Failed)
	}
}

func TestValidateAndPersist_AccountNotFound(t *testing.T) {
	repo := makeCustomerRepo()
	csv := "customerName,accountNo\nJOHN DOE,9999999999999\n"
	uc := usecase.NewCustomerUsecase(repo, makeSampleCSV(t, csv))

	result, err := uc.ValidateAndPersist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified != 0 {
		t.Errorf("expected verified=0, got %d", result.Verified)
	}
	if result.Failed != 1 {
		t.Errorf("expected failed=1, got %d", result.Failed)
	}
	
	// Find the failed record
	var failedRec *domain.ValidationRecord
	for _, b := range result.Batches {
		for _, r := range b.Records {
			if !r.Verified {
				failedRec = &r
			}
		}
	}
	if failedRec == nil || len(failedRec.Errors) == 0 {
		t.Error("expected at least one validation error")
	}
}

func TestValidateAndPersist_NameMismatch(t *testing.T) {
	repo := makeCustomerRepo()
	csv := "customerName,accountNo\nWRONG NAME,1050001035901\n"
	uc := usecase.NewCustomerUsecase(repo, makeSampleCSV(t, csv))

	result, err := uc.ValidateAndPersist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified != 0 {
		t.Errorf("expected verified=0, got %d", result.Verified)
	}
	
	foundNameError := false
	for _, b := range result.Batches {
		for _, r := range b.Records {
			for _, e := range r.Errors {
				if e.Field == "customer_name" {
					foundNameError = true
				}
			}
		}
	}
	if !foundNameError {
		t.Error("expected customer_name validation error")
	}
}

func TestValidateAndPersist_CaseInsensitiveMatch(t *testing.T) {
	repo := makeCustomerRepo()
	// lowercase should still match
	csv := "customerName,accountNo\nmehadi aliye mohammed,1050001035901\n"
	uc := usecase.NewCustomerUsecase(repo, makeSampleCSV(t, csv))

	result, err := uc.ValidateAndPersist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified != 1 {
		t.Errorf("expected verified=1, got %d (case-insensitive match should pass)", result.Verified)
	}
}

func TestValidateAndPersist_EmptyAccountNo(t *testing.T) {
	repo := makeCustomerRepo()
	csv := "customerName,accountNo\nSOME NAME,\n"
	uc := usecase.NewCustomerUsecase(repo, makeSampleCSV(t, csv))

	result, err := uc.ValidateAndPersist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified != 0 {
		t.Errorf("expected verified=0, got %d", result.Verified)
	}
}

func TestValidateAndPersist_PersistsOnlyVerified(t *testing.T) {
	repo := makeCustomerRepo()
	csv := "customerName,accountNo\nMEHADI ALIYE MOHAMMED,1050001035901\nNOBODY,9999999999999\n"
	uc := usecase.NewCustomerUsecase(repo, makeSampleCSV(t, csv))

	_, err := uc.ValidateAndPersist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	verified := uc.GetVerifiedCustomers()
	if len(verified) != 1 {
		t.Errorf("expected 1 verified customer persisted, got %d", len(verified))
	}
	if verified[0].AccountNo != "1050001035901" {
		t.Errorf("expected account 1050001035901, got %q", verified[0].AccountNo)
	}
}

func TestValidateAndPersist_RecordIndex(t *testing.T) {
	repo := makeCustomerRepo()
	csv := "customerName,accountNo\nMEHADI ALIYE MOHAMMED,1050001035901\nUSMAN AHMED UMER,1050001032701\n"
	uc := usecase.NewCustomerUsecase(repo, makeSampleCSV(t, csv))

	result, _ := uc.ValidateAndPersist()
	
	// collect all records across batches to check ordering
	var all []domain.ValidationRecord
	for _, b := range result.Batches {
		all = append(all, b.Records...)
	}
	
	if len(all) >= 2 {
		if all[0].RecordIndex != 1 {
			t.Errorf("expected RecordIndex=1, got %d", all[0].RecordIndex)
		}
		if all[1].RecordIndex != 2 {
			t.Errorf("expected RecordIndex=2, got %d", all[1].RecordIndex)
		}
	}
}
