package repository

import (
	"fmt"
	"gin-quickstart/internal/domain/entity"

	"github.com/google/uuid"
)

type FileMemoryRepository struct {
	files map[string]entity.File
}

func NewFileMemoryRepository() *FileMemoryRepository {
	return &FileMemoryRepository{files: make(map[string]entity.File)}
}

func cloneFile(f entity.File) entity.File {
	f.Data = append([]byte(nil), f.Data...)
	return f
}

func (m *FileMemoryRepository) Put(file entity.File) (string, error) {
	id := uuid.New().String()

	if _, exists := m.files[id]; exists {
		return "", fmt.Errorf("file with id %s already exists", id)
	}

	file.Metadata.ID = id

	m.files[id] = cloneFile(file)
	return id, nil
}

func (m *FileMemoryRepository) Get(fileID string) (entity.File, error) {
	file, ok := m.files[fileID]
	if !ok {
		return entity.File{}, fmt.Errorf("file not found for id: %s", fileID)
	}
	return cloneFile(file), nil
}

func (m *FileMemoryRepository) Metadata(fileID string) (entity.FileMetadata, error) {
	files, ok := m.files[fileID]
	if !ok {
		return entity.FileMetadata{}, fmt.Errorf("download job with id %s not found", fileID)
	}
	return files.Metadata, nil
}
