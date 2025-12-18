package entity

import "time"

type DownloadJobStatus int

const (
	Process DownloadJobStatus = iota
	Done
	Failed
	Canceled
)

func (s *DownloadJobStatus) String() string {
	switch *s {
	case Process:
		return "PROCESS"
	case Done:
		return "DONE"
	case Failed:
		return "FAILED"
	case Canceled:
		return "CANCELED"
	default:
		return "UNKNOWN"
	}
}

type DownloadItemErrorCode string

const (
	ErrorTimeout DownloadItemErrorCode = "TIMEOUT"
	ErrorHTTP    DownloadItemErrorCode = "HTTP_ERROR"
	ErrorUnknown DownloadItemErrorCode = "UNKNOWN"
)

type DownloadItemError struct {
	Code DownloadItemErrorCode
}

type DownloadItem struct {
	URL    string
	FileID string
	Error  *DownloadItemError
}

type DownloadJob struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Timeout   time.Duration
	Status    DownloadJobStatus
	Items     []DownloadItem
}
