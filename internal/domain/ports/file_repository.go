package ports

import "gin-quickstart/internal/domain/entity"

type FileRepository interface {
	Put(f entity.File) (string, error)
	Get(fileID string) (entity.File, error)
	Metadata(fileID string) (entity.FileMetadata, error)
}
