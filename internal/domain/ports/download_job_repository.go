package ports

import "gin-quickstart/internal/domain/entity"

type DownloadJobRepository interface {
	Create(job entity.DownloadJob) (entity.DownloadJob, error)
	Get(id string) (entity.DownloadJob, error)
	Update(job entity.DownloadJob) error
	Delete(id string) error
}
