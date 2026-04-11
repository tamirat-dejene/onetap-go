package interfaces

import "github.com/onetap/salary-advance/internal/domain"

// AuthUsecase defines authentication business logic.
type AuthUsecase interface {
	// Login validates credentials and returns a signed JWT token.
	Login(ip, username, password string) (string, error)
}

// CustomerUsecase defines customer validation business logic.
type CustomerUsecase interface {
	// ValidateAndPersist reads the sample CSV, validates against the canonical list,
	// persists verified records, and returns the full validation report.
	ValidateAndPersist() (*domain.ValidationResult, error)
	// GetVerifiedCustomers returns all previously validated customers.
	GetVerifiedCustomers() []*domain.Customer
}

// TransactionUsecase defines transaction and rating business logic.
type TransactionUsecase interface {
	// GetRatingForCustomer calculates the loan risk rating for a given account number.
	GetRatingForCustomer(accountNo string) (*domain.RatingResult, error)
	// GetAllRatings calculates ratings for all verified customers.
	GetAllRatings() ([]*domain.RatingResult, error)
}
