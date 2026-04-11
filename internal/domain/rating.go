package domain

// RatingResult holds the loan risk rating for a customer.
type RatingResult struct {
	AccountNo        string  `json:"account_no"`
	CustomerName     string  `json:"customer_name"`
	Rating           int     `json:"rating"`
	Description      string  `json:"description"`
	TransactionCount int     `json:"transaction_count"`
	TotalVolume      float64 `json:"total_volume"`
	TimeSpanDays     int     `json:"time_span_days"`
	BalanceStability float64 `json:"balance_stability"`
}
