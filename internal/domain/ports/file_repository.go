package ports

import (
	"context"
	"gin-quickstart/internal/domain/entity"
)

type FileRepository interface {
	Create(ctx context.Context, file entity.File) (string, error)
	Get(ctx context.Context, fileID string) (entity.File, error)
	Metadata(ctx context.Context, fileID string) (entity.FileMetadata, error)
}
