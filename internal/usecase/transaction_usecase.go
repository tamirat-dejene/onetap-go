package usecase

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/domain/interfaces"
)

// ratingWeights defines the contribution of each factor to the overall score.
const (
	weightTxCount   = 0.25
	weightVolume    = 0.30
	weightTimeSpan  = 0.25
	weightBalance   = 0.20
	maxTxCount      = 50.0
	maxTimeSpanDays = 365.0
)

// TransactionUsecase implements the loan rating calculation logic.
type TransactionUsecase struct {
	txRepo       interfaces.TransactionRepository
	customerRepo interfaces.CustomerRepository
}

// NewTransactionUsecase creates a TransactionUsecase with injected dependencies.
func NewTransactionUsecase(
	txRepo interfaces.TransactionRepository,
	customerRepo interfaces.CustomerRepository,
) *TransactionUsecase {
	return &TransactionUsecase{
		txRepo:       txRepo,
		customerRepo: customerRepo,
	}
}

// GetRatingForCustomer calculates the rating for a single verified customer account.
func (u *TransactionUsecase) GetRatingForCustomer(accountNo string) (*domain.RatingResult, error) {
	customer, ok := u.customerRepo.GetByAccountNo(accountNo)
	if !ok {
		// Also check verified store
		verified := u.customerRepo.GetAllVerified()
		for _, c := range verified {
			if c.AccountNo == accountNo {
				customer = c
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil, fmt.Errorf("customer with account %q not found", accountNo)
	}

	txs := u.txRepo.GetByAccountNo(accountNo)
	if len(txs) == 0 {
		txs = u.generateSyntheticTransactions(accountNo, customer.CustomerBalance)
	}

	return u.computeRating(customer, txs), nil
}

// GetAllRatings computes ratings for every verified customer.
func (u *TransactionUsecase) GetAllRatings() ([]*domain.RatingResult, error) {
	customers := u.customerRepo.GetAllVerified()
	if len(customers) == 0 {
		return nil, fmt.Errorf("no verified customers found; run validation first")
	}
	results := make([]*domain.RatingResult, 0, len(customers))
	for _, c := range customers {
		r, err := u.GetRatingForCustomer(c.AccountNo)
		if err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

// computeRating calculates the 1-10 risk rating based on four weighted factors.
func (u *TransactionUsecase) computeRating(customer *domain.Customer, txs []*domain.Transaction) *domain.RatingResult {
	// ---- Factor 1: Transaction Count (normalized 0-1, max=50) ----
	txCount := len(txs)
	countScore := math.Min(float64(txCount)/maxTxCount, 1.0)

	// ---- Factor 2: Total Volume (log-normalized) ----
	totalVolume := 0.0
	for _, t := range txs {
		totalVolume += t.Amount
	}
	// log10(volume+1) normalized against log10(10_000_000)
	volumeScore := math.Min(math.Log10(totalVolume+1)/math.Log10(10_000_000), 1.0)

	// ---- Factor 3: Time Span (days between first and last tx, max=365) ----
	minMs, maxMs := int64(math.MaxInt64), int64(0)
	for _, t := range txs {
		if t.TransactionDateMs < minMs {
			minMs = t.TransactionDateMs
		}
		if t.TransactionDateMs > maxMs {
			maxMs = t.TransactionDateMs
		}
	}
	timeSpanDays := 0
	var timeScore float64
	if len(txs) > 1 {
		timeSpanDays = int((maxMs-minMs)/1000) / 86400
		timeScore = math.Min(float64(timeSpanDays)/maxTimeSpanDays, 1.0)
	}

	// ---- Factor 4: Balance Stability (avg cleared balance / customer balance) ----
	balanceSum := 0.0
	balanceCount := 0
	for _, t := range txs {
		if t.ClearedBalance != nil {
			if v, err := strconv.ParseFloat(*t.ClearedBalance, 64); err == nil {
				balanceSum += v
				balanceCount++
			}
		}
	}
	var balanceStability float64
	if balanceCount > 0 && customer.CustomerBalance > 0 {
		avgBalance := balanceSum / float64(balanceCount)
		balanceStability = math.Min(avgBalance/customer.CustomerBalance, 1.0)
		if balanceStability < 0 {
			balanceStability = 0
		}
	}

	// ---- Weighted composite score -> 1-10 ----
	composite := countScore*weightTxCount + volumeScore*weightVolume + timeScore*weightTimeSpan + balanceStability*weightBalance

	rating := min(max(int(math.Round(composite*9))+1, 1), 10)

	return &domain.RatingResult{
		AccountNo:        customer.AccountNo,
		CustomerName:     customer.CustomerName,
		Rating:           rating,
		Description:      ratingDescription(rating),
		TransactionCount: txCount,
		TotalVolume:      totalVolume,
		TimeSpanDays:     timeSpanDays,
		BalanceStability: math.Round(balanceStability*100) / 100,
	}
}

// ratingDescription maps a numeric score to a human-readable risk label.
func ratingDescription(score int) string {
	switch {
	case score <= 2:
		return "Very High Risk"
	case score <= 4:
		return "High Risk"
	case score <= 6:
		return "Moderate Risk"
	case score <= 8:
		return "Low Risk"
	default:
		return "Very Low Risk / Excellent"
	}
}

// generateSyntheticTransactions generates 1-3 small synthetic transactions for
// customers who have no real transaction history.
func (u *TransactionUsecase) generateSyntheticTransactions(accountNo string, balance float64) []*domain.Transaction {
	rng := rand.New(rand.NewSource(hashAccountNo(accountNo)))
	count := rng.Intn(3) + 1
	txs := make([]*domain.Transaction, 0, count)

	maxAmount := balance * 0.05
	if maxAmount < 100 {
		maxAmount = 100
	}

	baseTime := time.Now().Add(-30 * 24 * time.Hour).UnixMilli()

	for i := 0; i < count; i++ {
		amount := math.Round((rng.Float64()*maxAmount+50)*100) / 100
		offsetMs := int64(rng.Intn(2592000)) * 1000 // up to 30 days offset
		dateMs := baseTime + offsetMs
		amountStr := strconv.FormatFloat(amount, 'f', 2, 64)
		txs = append(txs, &domain.Transaction{
			ID:                 uuid.New().String(),
			FromAccount:        accountNo,
			ToAccount:          "",
			Amount:             amount,
			AmountRaw:          amountStr,
			TransactionType:    "Synthetic",
			RequestID:          fmt.Sprintf("SYN-%d", i),
			Reference:          fmt.Sprintf("synthetic-%s-%d", accountNo, i),
			TransactionDateMs:  dateMs,
			TransactionDateRaw: strconv.FormatInt(dateMs, 10),
		})
	}
	return txs
}

// hashAccountNo produces a deterministic int64 seed from an account number string.
func hashAccountNo(s string) int64 {
	var h int64 = 5381
	for _, c := range s {
		h = h*33 + int64(c)
	}
	return h
}
