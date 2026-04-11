package memory

import (
	"sync"

	"github.com/onetap/salary-advance/internal/domain"
)

// CustomerRepo is a thread-safe in-memory implementation of CustomerRepository.
type CustomerRepo struct {
	mu       sync.RWMutex
	canonical   map[string]*domain.Customer // keyed by accountNo
	verified    map[string]*domain.Customer // keyed by accountNo
	validations []domain.ValidationRecord
}

// NewCustomerRepo creates an empty CustomerRepo.
func NewCustomerRepo() *CustomerRepo {
	return &CustomerRepo{
		canonical:   make(map[string]*domain.Customer),
		verified:    make(map[string]*domain.Customer),
		validations: make([]domain.ValidationRecord, 0),
	}
}

// Seed loads the canonical customer list into the store.
func (r *CustomerRepo) Seed(customers []*domain.Customer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range customers {
		r.canonical[c.AccountNo] = c
	}
}

// GetAll returns all canonical customers.
func (r *CustomerRepo) GetAll() []*domain.Customer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Customer, 0, len(r.canonical))
	for _, c := range r.canonical {
		result = append(result, c)
	}
	return result
}

// GetByAccountNo looks up a canonical customer by account number.
func (r *CustomerRepo) GetByAccountNo(accountNo string) (*domain.Customer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.canonical[accountNo]
	return c, ok
}

// SaveVerified stores a verified customer record.
func (r *CustomerRepo) SaveVerified(c *domain.Customer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.verified[c.AccountNo] = c
}

// GetAllVerified returns all verified customers.
func (r *CustomerRepo) GetAllVerified() []*domain.Customer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Customer, 0, len(r.verified))
	for _, c := range r.verified {
		result = append(result, c)
	}
	return result
}

// SaveValidationRecord appends a processed validation record.
func (r *CustomerRepo) SaveValidationRecord(rec domain.ValidationRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.validations = append(r.validations, rec)
}

// GetAllValidationRecords returns all stored validation records.
func (r *CustomerRepo) GetAllValidationRecords() []domain.ValidationRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.ValidationRecord, len(r.validations))
	copy(result, r.validations)
	return result
}
