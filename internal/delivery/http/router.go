package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/onetap/salary-advance/internal/delivery/http/handler"
	"github.com/onetap/salary-advance/internal/delivery/http/middleware"
	pkgjwt "github.com/onetap/salary-advance/pkg/jwt"
)

// NewRouter builds and returns the application router with all routes registered.
func NewRouter(
	jwtManager *pkgjwt.Manager,
	authHandler *handler.AuthHandler,
	customerHandler *handler.CustomerHandler,
	txHandler *handler.TransactionHandler,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(setJSONContentType)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/login", authHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))

			// Customer endpoints
			r.Get("/customers/validate", customerHandler.ValidateCustomers)
			r.Get("/customers/verified", customerHandler.GetVerifiedCustomers)

			// Rating endpoints
			r.Get("/ratings", txHandler.GetAllRatings)
			r.Get("/ratings/{accountNo}", txHandler.GetRating)
		})
	})

	return r
}

// setJSONContentType sets Content-Type to application/json for all responses.
func setJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
