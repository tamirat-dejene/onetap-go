package memory

import (
	"sync"

	"github.com/onetap/salary-advance/internal/domain"
)

// TransactionRepo is a thread-safe in-memory implementation of TransactionRepository.
type TransactionRepo struct {
	mu   sync.RWMutex
	byAccount map[string][]*domain.Transaction
	all       []*domain.Transaction
}

// NewTransactionRepo creates an empty TransactionRepo.
func NewTransactionRepo() *TransactionRepo {
	return &TransactionRepo{
		byAccount: make(map[string][]*domain.Transaction),
	}
}

// Seed bulk-loads transactions from the data file.
func (r *TransactionRepo) Seed(txs []*domain.Transaction) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range txs {
		r.byAccount[t.FromAccount] = append(r.byAccount[t.FromAccount], t)
		r.all = append(r.all, t)
	}
}

// GetAll returns all transactions.
func (r *TransactionRepo) GetAll() []*domain.Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Transaction, len(r.all))
	copy(result, r.all)
	return result
}

// GetByAccountNo returns all transactions whose fromAccount matches.
func (r *TransactionRepo) GetByAccountNo(accountNo string) []*domain.Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	txs := r.byAccount[accountNo]
	result := make([]*domain.Transaction, len(txs))
	copy(result, txs)
	return result
}

// Save persists a single transaction.
func (r *TransactionRepo) Save(t *domain.Transaction) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byAccount[t.FromAccount] = append(r.byAccount[t.FromAccount], t)
	r.all = append(r.all, t)
}
