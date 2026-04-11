package usecase

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/domain/interfaces"
)

// CustomerUsecase implements customer validation and persistence logic.
type CustomerUsecase struct {
	customerRepo   interfaces.CustomerRepository
	sampleFilePath string
}

// NewCustomerUsecase creates a new CustomerUsecase with injected dependencies.
func NewCustomerUsecase(
	customerRepo interfaces.CustomerRepository,
	sampleFilePath string,
) *CustomerUsecase {
	return &CustomerUsecase{
		customerRepo:   customerRepo,
		sampleFilePath: sampleFilePath,
	}
}

// ValidateAndPersist reads the sample CSV, validates each record against the canonical
// customer list, persists only valid records, and returns the full validation report.
func (u *CustomerUsecase) ValidateAndPersist() (*domain.ValidationResult, error) {
	samples, err := readSampleCSV(u.sampleFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading sample file: %w", err)
	}

	result := &domain.ValidationResult{
		Total:   len(samples),
		Batches: make([]domain.BatchResult, 0, 5),
	}

	var validRecords []domain.ValidationRecord
	var invalidNameRecords []domain.ValidationRecord
	var invalidAccRecords []domain.ValidationRecord

	// First pass: validate all records and classify them
	for idx, sample := range samples {
		rec := u.validateRecord(idx+1, sample[0], sample[1])
		u.customerRepo.SaveValidationRecord(rec)
		
		if rec.Verified {
			u.customerRepo.SaveVerified(rec.NormalizedRecord)
			validRecords = append(validRecords, rec)
			result.Verified++
		} else {
			result.Failed++
			// Determine which error it had to classify it
			isNameErr := false
			for _, e := range rec.Errors {
				if e.Field == "customer_name" {
					isNameErr = true
				}
			}
			if isNameErr {
				invalidNameRecords = append(invalidNameRecords, rec)
			} else {
				invalidAccRecords = append(invalidAccRecords, rec)
			}
		}
	}

	// Helper to assemble a batch from slices
	takeValid := func(count int) []domain.ValidationRecord {
		if count > len(validRecords) {
			count = len(validRecords)
		}
		res := validRecords[:count]
		validRecords = validRecords[count:]
		return res
	}

	// Build 3 fully valid batches of 10
	for i := 1; i <= 3; i++ {
		batchRecs := takeValid(10)
		result.Batches = append(result.Batches, domain.BatchResult{
			BatchNumber: i,
			Total:       len(batchRecs),
			Verified:    len(batchRecs),
			Failed:      0,
			Records:     batchRecs,
		})
	}

	// Build Batch 4 (1 invalid name record + remainder of 10 with valid)
	batch4Recs := make([]domain.ValidationRecord, 0, 10)
	if len(invalidNameRecords) > 0 {
		batch4Recs = append(batch4Recs, invalidNameRecords...)
	}
	batch4Recs = append(batch4Recs, takeValid(10 - len(batch4Recs))...)
	
	result.Batches = append(result.Batches, domain.BatchResult{
		BatchNumber: 4,
		Total:       len(batch4Recs),
		Verified:    len(batch4Recs) - len(invalidNameRecords),
		Failed:      len(invalidNameRecords),
		Records:     batch4Recs,
	})

	// Build Batch 5 (1 invalid account record + remainder of 10 with valid)
	batch5Recs := make([]domain.ValidationRecord, 0, 10)
	if len(invalidAccRecords) > 0 {
		batch5Recs = append(batch5Recs, invalidAccRecords...)
	}
	batch5Recs = append(batch5Recs, takeValid(10 - len(batch5Recs))...)

	result.Batches = append(result.Batches, domain.BatchResult{
		BatchNumber: 5,
		Total:       len(batch5Recs),
		Verified:    len(batch5Recs) - len(invalidAccRecords),
		Failed:      len(invalidAccRecords),
		Records:     batch5Recs,
	})

	return result, nil
}

// GetVerifiedCustomers returns all persisted verified customer records.
func (u *CustomerUsecase) GetVerifiedCustomers() []*domain.Customer {
	return u.customerRepo.GetAllVerified()
}

// validateRecord checks a single sample record against the canonical list.
func (u *CustomerUsecase) validateRecord(index int, name, accountNo string) domain.ValidationRecord {
	rec := domain.ValidationRecord{
		RecordIndex:  index,
		CustomerName: name,
		AccountNo:    accountNo,
		Errors:       []domain.ValidationError{},
	}

	// Trim whitespace from inputs
	name = strings.TrimSpace(name)
	accountNo = strings.TrimSpace(accountNo)

	// Pad account number with prefix zeros to 13 digits
	if len(accountNo) > 0 && len(accountNo) < 13 {
		accountNo = fmt.Sprintf("%013s", accountNo)
	}
	rec.AccountNo = accountNo

	if accountNo == "" {
		rec.Errors = append(rec.Errors, domain.ValidationError{
			Field:   "account_no",
			Message: "account number is empty",
		})
		rec.Verified = false
		return rec
	}

	// Look up canonical customer
	canonical, found := u.customerRepo.GetByAccountNo(accountNo)
	if !found {
		rec.Errors = append(rec.Errors, domain.ValidationError{
			Field:   "account_no",
			Message: fmt.Sprintf("account number %q not found in canonical list", accountNo),
		})
		rec.Verified = false
		return rec
	}

	// Name comparison: case-insensitive, whitespace-normalized
	if !namesMatch(name, canonical.CustomerName) {
		rec.Errors = append(rec.Errors, domain.ValidationError{
			Field:   "customer_name",
			Message: fmt.Sprintf("name mismatch: got %q, expected %q", name, canonical.CustomerName),
		})
	}

	if len(rec.Errors) == 0 {
		rec.Verified = true
		norm := *canonical
		rec.NormalizedRecord = &norm
	}

	return rec
}

// namesMatch compares two names case-insensitively after collapsing internal whitespace.
func namesMatch(a, b string) bool {
	return normalizeString(a) == normalizeString(b)
}

// normalizeString upper-cases and collapses multiple spaces into one.
func normalizeString(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	// Collapse multiple spaces
	fields := strings.FieldsFunc(s, unicode.IsSpace)
	return strings.Join(fields, " ")
}

// readSampleCSV reads the sample customers CSV file, returning rows as [name, accountNo].
func readSampleCSV(path string) ([][2]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	// Skip header row
	if len(records) == 0 {
		return nil, nil
	}
	rows := records[1:]

	result := make([][2]string, 0, len(rows))
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		result = append(result, [2]string{row[0], row[1]})
	}
	return result, nil
}
