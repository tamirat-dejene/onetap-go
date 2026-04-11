package usecase_test

import (
	"testing"
	"time"

	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/repository/memory"
	"github.com/onetap/salary-advance/internal/usecase"
	pkgjwt "github.com/onetap/salary-advance/pkg/jwt"
)

func newTestAuthUsecase(t *testing.T) (*usecase.AuthUsecase, *memory.UserRepo) {
	t.Helper()
	userRepo := memory.NewUserRepo()
	jwtMgr := pkgjwt.NewManager("test-secret", 1)
	uc := usecase.NewAuthUsecase(userRepo, jwtMgr, 4, 3, 60)
	return uc, userRepo
}

func TestLogin_Success(t *testing.T) {
	uc, _ := newTestAuthUsecase(t)
	if err := uc.SeedUser("u1", "admin", "secret", domain.RoleAdmin); err != nil {
		t.Fatalf("seed: %v", err)
	}

	token, err := uc.Login("127.0.0.1", "admin", "secret")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	uc, _ := newTestAuthUsecase(t)
	_ = uc.SeedUser("u1", "admin", "secret", domain.RoleAdmin)

	_, err := uc.Login("127.0.0.1", "admin", "wrongpassword")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	uc, _ := newTestAuthUsecase(t)

	_, err := uc.Login("127.0.0.1", "nobody", "secret")
	if err == nil {
		t.Fatal("expected error for unknown user")
	}
}

func TestLogin_RateLimit(t *testing.T) {
	uc, _ := newTestAuthUsecase(t)
	_ = uc.SeedUser("u1", "admin", "secret", domain.RoleAdmin)

	// maxRequests = 3; exceed the limit
	for i := 0; i < 3; i++ {
		_, _ = uc.Login("10.0.0.1", "admin", "secret")
	}

	_, err := uc.Login("10.0.0.1", "admin", "secret")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}

func TestLogin_RateLimitPerIP(t *testing.T) {
	uc, _ := newTestAuthUsecase(t)
	_ = uc.SeedUser("u1", "admin", "secret", domain.RoleAdmin)

	// Exhaust rate limit for IP A
	for i := 0; i < 3; i++ {
		_, _ = uc.Login("1.1.1.1", "admin", "secret")
	}

	// IP B should still work
	_, err := uc.Login("2.2.2.2", "admin", "secret")
	if err != nil {
		t.Fatalf("different IP should not be rate limited: %v", err)
	}
}

func TestLogin_TokenIsValid(t *testing.T) {
	uc, _ := newTestAuthUsecase(t)
	_ = uc.SeedUser("u1", "admin", "secret", domain.RoleAdmin)

	jwtMgr := pkgjwt.NewManager("test-secret", 1)
	token, _ := uc.Login("127.0.0.1", "admin", "secret")

	claims, err := jwtMgr.Validate(token)
	if err != nil {
		t.Fatalf("token should be valid: %v", err)
	}
	if claims.Username != "admin" {
		t.Errorf("expected username=admin, got %q", claims.Username)
	}
	if claims.Role != string(domain.RoleAdmin) {
		t.Errorf("expected role=admin, got %q", claims.Role)
	}
}

func TestLogin_ExpiredToken(t *testing.T) {
	userRepo := memory.NewUserRepo()
	jwtMgr := pkgjwt.NewManager("test-secret", 0) // 0 hours -> already expired
	uc := usecase.NewAuthUsecase(userRepo, jwtMgr, 4, 5, 60)
	_ = uc.SeedUser("u1", "admin", "secret", domain.RoleAdmin)

	token, _ := uc.Login("127.0.0.1", "admin", "secret")
	time.Sleep(10 * time.Millisecond)

	_, err := jwtMgr.Validate(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}
