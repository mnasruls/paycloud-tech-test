package models

import "time"

// ReportRequest represents a request to generate a report
type ReportRequest struct {
	ID         string            `json:"id"`
	ReportType string            `json:"report_type"` // e.g., "sales", "inventory"
	Parameters map[string]string `json:"parameters"`
	CreatedAt  time.Time         `json:"created_at"`
}

// ReportResult represents the final result of a processed report
type ReportResult struct {
	RequestID   string       `json:"request_id"`
	Status      ReportStatus `json:"status"`
	GeneratedAt time.Time    `json:"generated_at"`
	ReportData  string       `json:"report_data,omitempty"` // Example report data (could be complex struct)
	Error       string       `json:"error,omitempty"`
}
