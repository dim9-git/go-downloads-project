package repository

import (
	"context"
	"fmt"
	"gin-quickstart/internal/domain/entity"
	"sync"

	"github.com/google/uuid"
)

type FileMemoryRepository struct {
	mu    sync.RWMutex
	files map[string]entity.File
}

func NewFileMemoryRepository() *FileMemoryRepository {
	return &FileMemoryRepository{files: make(map[string]entity.File)}
}

func cloneFile(f entity.File) entity.File {
	f.Data = append([]byte(nil), f.Data...)
	return f
}

func (m *FileMemoryRepository) Create(ctx context.Context, file entity.File) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	id := uuid.New().String()

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.files[id]; exists {
		return "", fmt.Errorf("CREATE: File with ID %s already exists", id)
	}

	file.Metadata.ID = id

	m.files[id] = cloneFile(file)
	return id, nil
}

func (m *FileMemoryRepository) Get(ctx context.Context, fileID string) (entity.File, error) {
	if err := ctx.Err(); err != nil {
		return entity.File{}, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	file, exists := m.files[fileID]
	if !exists {
		return entity.File{}, fmt.Errorf("file not found for id: %s", fileID)
	}
	return cloneFile(file), nil
}

func (m *FileMemoryRepository) Metadata(ctx context.Context, fileID string) (entity.FileMetadata, error) {
	if err := ctx.Err(); err != nil {
		return entity.FileMetadata{}, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	files, exists := m.files[fileID]
	if !exists {
		return entity.FileMetadata{}, fmt.Errorf("download job with id %s not found", fileID)
	}
	return files.Metadata, nil
}
