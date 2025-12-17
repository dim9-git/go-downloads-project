package repository

import (
	"fmt"
	"gin-quickstart/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type DownloadJobMemoryRepository struct {
	jobs map[string]entity.DownloadJob
}

func NewDownloadJobMemoryRepository() *DownloadJobMemoryRepository {
	return &DownloadJobMemoryRepository{jobs: make(map[string]entity.DownloadJob)}
}

func cloneJob(j entity.DownloadJob) entity.DownloadJob {
	j.Items = append([]entity.DownlaodItem(nil), j.Items...)
	return j
}

func (m *DownloadJobMemoryRepository) Create(job entity.DownloadJob) (entity.DownloadJob, error) {
	id := uuid.New().String()

	if _, exists := m.jobs[id]; exists {
		return entity.DownloadJob{}, fmt.Errorf("CREATE: download job with id %s already exists", id)
	}

	now := time.Now()
	job.ID = id
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now

	m.jobs[job.ID] = cloneJob(job)
	return job, nil
}

func (m *DownloadJobMemoryRepository) Get(id string) (entity.DownloadJob, error) {
	job, ok := m.jobs[id]
	if !ok {
		return entity.DownloadJob{}, fmt.Errorf("download job not found for id: %s", id)
	}
	return cloneJob(job), nil
}

func (m *DownloadJobMemoryRepository) Update(job entity.DownloadJob) error {
	if job.ID == "" {
		return fmt.Errorf("UPDATE: job ID cannot be empty")
	}
	if _, exists := m.jobs[job.ID]; !exists {
		return fmt.Errorf("UPDATE: download job with id %s not found", job.ID)
	}

	job.UpdatedAt = time.Now()
	m.jobs[job.ID] = cloneJob(job)
	return nil
}

func (m *DownloadJobMemoryRepository) Delete(id string) error {
	if _, exists := m.jobs[id]; !exists {
		return fmt.Errorf("download job with id %s not found", id)
	}
	delete(m.jobs, id)
	return nil
}
