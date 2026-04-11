package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/onetap/salary-advance/internal/domain/interfaces"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authUsecase interfaces.AuthUsecase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authUsecase interfaces.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// loginRequest is the expected request body for POST /auth/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginResponse is the successful login response body.
type loginResponse struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with username and password to receive a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      loginRequest   true  "Login credentials"
// @Success      200   {object}  loginResponse
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Failure      429   {object}  errorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	ip := clientIP(r)
	token, err := h.authUsecase.Login(ip, req.Username, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if isRateLimitError(err) {
			status = http.StatusTooManyRequests
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token, Type: "Bearer"})
}

// clientIP extracts the real client IP from the request.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// isRateLimitError returns true if the error looks like a rate limit error.
func isRateLimitError(err error) bool {
	return len(err.Error()) > 11 && err.Error()[:11] == "rate limit "
}
