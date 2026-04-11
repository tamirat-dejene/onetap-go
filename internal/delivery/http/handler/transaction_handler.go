package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onetap/salary-advance/internal/domain/interfaces"
)

// TransactionHandler handles transaction rating endpoints.
type TransactionHandler struct {
	txUsecase interfaces.TransactionUsecase
}

// NewTransactionHandler creates a new TransactionHandler.
func NewTransactionHandler(txUsecase interfaces.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{txUsecase: txUsecase}
}

// GetAllRatings godoc
// @Summary      Get all customer ratings
// @Description  Calculates and returns loan risk ratings for all verified customers
// @Tags         ratings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.RatingResult
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /ratings [get]
func (h *TransactionHandler) GetAllRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := h.txUsecase.GetAllRatings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":   len(ratings),
		"ratings": ratings,
	})
}

// GetRating godoc
// @Summary      Get rating for a specific account
// @Description  Calculates and returns the loan risk rating for a single customer account
// @Tags         ratings
// @Produce      json
// @Security     BearerAuth
// @Param        accountNo  path      string  true  "Account Number"
// @Success      200        {object}  domain.RatingResult
// @Failure      401        {object}  errorResponse
// @Failure      404        {object}  errorResponse
// @Router       /ratings/{accountNo} [get]
func (h *TransactionHandler) GetRating(w http.ResponseWriter, r *http.Request) {
	accountNo := chi.URLParam(r, "accountNo")
	rating, err := h.txUsecase.GetRatingForCustomer(accountNo)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rating)
}
