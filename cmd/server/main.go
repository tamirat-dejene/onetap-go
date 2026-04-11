package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	delivery "github.com/onetap/salary-advance/internal/delivery/http"
	"github.com/onetap/salary-advance/internal/delivery/http/handler"
	"github.com/onetap/salary-advance/internal/domain"
	"github.com/onetap/salary-advance/internal/repository/memory"
	"github.com/onetap/salary-advance/internal/usecase"
	"github.com/onetap/salary-advance/pkg/config"
	pkgjwt "github.com/onetap/salary-advance/pkg/jwt"
	"github.com/onetap/salary-advance/pkg/loader"
)

func main() {
	// ── 1. Load configuration ─────────────────────────────────────────────────
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	// ── 2. Initialize in-memory repositories ─────────────────────────────────
	customerRepo := memory.NewCustomerRepo()
	txRepo := memory.NewTransactionRepo()
	userRepo := memory.NewUserRepo()

	// ── 3. Seed data from disk ────────────────────────────────────────────────
	if err := loader.SeedAll(cfg.CustomersFile, cfg.TransactionsFile, customerRepo, txRepo); err != nil {
		log.Fatalf("seeding data: %v", err)
	}
	log.Printf("Seeded %d canonical customers and %d transactions", len(customerRepo.GetAll()), len(txRepo.GetAll()))

	// ── 4. Bootstrap JWT manager ──────────────────────────────────────────────
	jwtManager := pkgjwt.NewManager(cfg.JWTSecret, cfg.JWTExpiryHours)

	// ── 5. Bootstrap usecases ─────────────────────────────────────────────────
	authUsecase := usecase.NewAuthUsecase(
		userRepo,
		jwtManager,
		cfg.BcryptCost,
		cfg.RateLimitRequests,
		cfg.RateLimitWindowSeconds,
	)

	// Seed default users (admin + uploader) at startup
	if err := authUsecase.SeedUser(uuid.New().String(), cfg.AdminUsername, cfg.AdminPassword, domain.RoleAdmin); err != nil {
		log.Fatalf("seeding admin user: %v", err)
	}
	if err := authUsecase.SeedUser(uuid.New().String(), cfg.UploaderUsername, cfg.UploaderPassword, domain.RoleUploader); err != nil {
		log.Fatalf("seeding uploader user: %v", err)
	}
	log.Printf("Default users created: %s (admin), %s (uploader)", cfg.AdminUsername, cfg.UploaderUsername)

	customerUsecase := usecase.NewCustomerUsecase(customerRepo, cfg.SampleCustomersFile)
	txUsecase := usecase.NewTransactionUsecase(txRepo, customerRepo)

	// ── 6. Bootstrap handlers ─────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUsecase)
	customerHandler := handler.NewCustomerHandler(customerUsecase)
	txHandler := handler.NewTransactionHandler(txUsecase)

	// ── 7. Build router ────────────────────────────────────────────────────────
	router := delivery.NewRouter(jwtManager, authHandler, customerHandler, txHandler)

	// ── 8. Start server ────────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
