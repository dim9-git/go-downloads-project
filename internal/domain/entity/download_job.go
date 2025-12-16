package entity

import "time"

type DownloadJobStatus int

const (
	Pending DownloadJobStatus = iota
	Running
	Done
	Failed
	Canceled
)

func (s *DownloadJobStatus) String() string {
	switch *s {
	case Pending:
		return "pending"
	case Running:
		return "running"
	case Done:
		return "done"
	case Failed:
		return "failed"
	case Canceled:
		return "canceled"
	default:
		return "unknown"
	}
}

type DownloadJob struct {
	ID            string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Timeout       time.Duration
	Status        DownloadJobStatus
	RequestedURLs []string
	FileIDs       []string
}
