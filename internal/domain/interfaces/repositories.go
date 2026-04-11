package interfaces

import "github.com/onetap/salary-advance/internal/domain"

// CustomerRepository defines the contract for customer data persistence.
type CustomerRepository interface {
	// GetAll returns all canonical customers loaded from the data file.
	GetAll() []*domain.Customer
	// GetByAccountNo finds a canonical customer by account number.
	GetByAccountNo(accountNo string) (*domain.Customer, bool)
	// SaveVerified persists a verified customer record.
	SaveVerified(c *domain.Customer)
	// GetAllVerified returns all successfully validated and persisted customers.
	GetAllVerified() []*domain.Customer
	// SaveValidationRecord saves the processed sample customer.
	SaveValidationRecord(r domain.ValidationRecord)
	// GetAllValidationRecords returns all processed samples.
	GetAllValidationRecords() []domain.ValidationRecord
}

// TransactionRepository defines the contract for transaction data persistence.
type TransactionRepository interface {
	// GetAll returns all transactions.
	GetAll() []*domain.Transaction
	// GetByAccountNo returns all transactions for a given account number.
	GetByAccountNo(accountNo string) []*domain.Transaction
	// Save persists a transaction.
	Save(t *domain.Transaction)
}

// UserRepository defines the contract for user data persistence.
type UserRepository interface {
	// GetByUsername finds a user by their username.
	GetByUsername(username string) (*domain.User, bool)
	// Save persists a user record.
	Save(u *domain.User)
}
