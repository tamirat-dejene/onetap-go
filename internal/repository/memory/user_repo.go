package memory

import (
	"sync"

	"github.com/onetap/salary-advance/internal/domain"
)

// UserRepo is a thread-safe in-memory implementation of UserRepository.
type UserRepo struct {
	mu    sync.RWMutex
	users map[string]*domain.User // keyed by username
}

// NewUserRepo creates an empty UserRepo.
func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[string]*domain.User),
	}
}

// GetByUsername looks up a user by username.
func (r *UserRepo) GetByUsername(username string) (*domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[username]
	return u, ok
}

// Save persists a user.
func (r *UserRepo) Save(u *domain.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.Username] = u
}
