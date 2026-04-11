package usecase

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/domain/interfaces"
	pkgjwt "github.com/onetap/salary-advance/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// rateLimitEntry tracks login attempts per IP.
type rateLimitEntry struct {
	mu        sync.Mutex
	attempts  int
	windowStart time.Time
}

// AuthUsecase implements the authentication business logic.
type AuthUsecase struct {
	userRepo   interfaces.UserRepository
	jwtManager *pkgjwt.Manager
	bcryptCost int

	// Rate limiting
	rlMu            sync.Mutex
	rateLimitStore  map[string]*rateLimitEntry
	maxRequests     int
	windowDuration  time.Duration
}

// NewAuthUsecase creates a new AuthUsecase with its dependencies injected.
func NewAuthUsecase(
	userRepo interfaces.UserRepository,
	jwtManager *pkgjwt.Manager,
	bcryptCost, rateLimitRequests, rateLimitWindowSeconds int,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:       userRepo,
		jwtManager:     jwtManager,
		bcryptCost:     bcryptCost,
		rateLimitStore: make(map[string]*rateLimitEntry),
		maxRequests:    rateLimitRequests,
		windowDuration: time.Duration(rateLimitWindowSeconds) * time.Second,
	}
}

// SeedUser hashes a plain password and stores the user. Called at startup.
func (a *AuthUsecase) SeedUser(id, username, plainPassword string, role domain.Role) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), a.bcryptCost)
	if err != nil {
		return fmt.Errorf("hashing password for %q: %w", username, err)
	}
	a.userRepo.Save(&domain.User{
		ID:           id,
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	})
	return nil
}

// Login validates credentials and returns a signed JWT. Enforces rate limiting per IP.
func (a *AuthUsecase) Login(ip, username, password string) (string, error) {
	if err := a.checkRateLimit(ip); err != nil {
		return "", err
	}

	user, ok := a.userRepo.GetByUsername(username)
	if !ok {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err := a.jwtManager.Generate(user.ID, user.Username, string(user.Role))
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	return token, nil
}

// checkRateLimit enforces a sliding window rate limit per IP address.
func (a *AuthUsecase) checkRateLimit(ip string) error {
	a.rlMu.Lock()
	entry, exists := a.rateLimitStore[ip]
	if !exists {
		entry = &rateLimitEntry{windowStart: time.Now()}
		a.rateLimitStore[ip] = entry
	}
	a.rlMu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()
	if now.Sub(entry.windowStart) > a.windowDuration {
		// Reset the window
		entry.attempts = 0
		entry.windowStart = now
	}
	entry.attempts++
	if entry.attempts > a.maxRequests {
		return fmt.Errorf("rate limit exceeded: max %d login attempts per %s", a.maxRequests, a.windowDuration)
	}
	return nil
}
