package ports

import (
	"context"
	"gin-quickstart/internal/domain/entity"
)

type DownloadJobRepository interface {
	Create(ctx context.Context, job entity.DownloadJob) (entity.DownloadJob, error)
	Get(ctx context.Context, id string) (entity.DownloadJob, error)
	Update(ctx context.Context, job entity.DownloadJob) error
	Delete(ctx context.Context, id string) error
}
