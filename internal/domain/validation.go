package domain

// ValidationError describes a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationRecord holds the result for one sample customer record.
type ValidationRecord struct {
	RecordIndex      int               `json:"record_index"`
	CustomerName     string            `json:"customer_name"`
	AccountNo        string            `json:"account_no"`
	Verified         bool              `json:"verified"`
	Errors           []ValidationError `json:"errors,omitempty"`
	NormalizedRecord *Customer         `json:"normalized_record,omitempty"`
}

// BatchResult represents a grouped chunk of validation records.
type BatchResult struct {
	BatchNumber int                `json:"batch_number"`
	Total       int                `json:"total"`
	Verified    int                `json:"verified"`
	Failed      int                `json:"failed"`
	Records     []ValidationRecord `json:"records"`
}

// ValidationResult is the aggregate result of the full validation run.
type ValidationResult struct {
	Total    int           `json:"total"`
	Verified int           `json:"verified"`
	Failed   int           `json:"failed"`
	Batches  []BatchResult `json:"batches"`
}
