package models

// ReportStatus represents the processing status of a report
type ReportStatus string

const (
	StatusPending    ReportStatus = "PENDING"
	StatusInProgress ReportStatus = "IN_PROGRESS"
	StatusCompleted  ReportStatus = "COMPLETED"
	StatusFailed     ReportStatus = "FAILED"
)
