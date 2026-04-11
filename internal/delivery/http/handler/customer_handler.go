package handler

import (
	"net/http"

	"github.com/onetap/salary-advance/internal/domain/interfaces"
)

// CustomerHandler handles customer-related endpoints.
type CustomerHandler struct {
	customerUsecase interfaces.CustomerUsecase
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(customerUsecase interfaces.CustomerUsecase) *CustomerHandler {
	return &CustomerHandler{customerUsecase: customerUsecase}
}

// ValidateCustomers godoc
// @Summary      Validate sample customers
// @Description  Reads the sample CSV, validates each record against the canonical list, persists verified records, and returns a detailed validation report
// @Tags         customers
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  domain.ValidationResult
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /customers/validate [get]
func (h *CustomerHandler) ValidateCustomers(w http.ResponseWriter, r *http.Request) {
	result, err := h.customerUsecase.ValidateAndPersist()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "validation failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetVerifiedCustomers godoc
// @Summary      Get verified customers
// @Description  Returns all customer records that passed validation
// @Tags         customers
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.Customer
// @Failure      401  {object}  errorResponse
// @Router       /customers/verified [get]
func (h *CustomerHandler) GetVerifiedCustomers(w http.ResponseWriter, r *http.Request) {
	customers := h.customerUsecase.GetVerifiedCustomers()
	writeJSON(w, http.StatusOK, map[string]any{
		"total":     len(customers),
		"customers": customers,
	})
}
