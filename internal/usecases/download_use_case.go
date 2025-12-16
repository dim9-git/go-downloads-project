package usecases

import (
	"context"
	"fmt"
	"gin-quickstart/internal/domain/entity"
	"gin-quickstart/internal/domain/ports"
	repository "gin-quickstart/internal/infra/repository/memory"
	"gin-quickstart/pkg/json"
	"io"
	"net/http"
)

type DownloadUseCase struct {
	DownloadJobRepository ports.DownloadJobRepository
	FileRepository        ports.FileRepository
}

func NewDownloadUseCase() *DownloadUseCase {
	return &DownloadUseCase{
		DownloadJobRepository: repository.NewDownloadJobMemoryRepository(),
		FileRepository:        repository.NewFileMemoryRepository(),
	}
}

func (u *DownloadUseCase) RunJob(job entity.DownloadJob) (entity.DownloadJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	createdJob, err := u.DownloadJobRepository.Create(job)
	if err != nil {
		return entity.DownloadJob{}, err
	}

	json.PrettyPrint(createdJob)

	// job.Status = entity.Running
	// if err := u.DownloadJobRepository.Update(job); err != nil {
	// 	return err
	// }

	for _, url := range job.RequestedURLs {
		if err := ctx.Err(); err != nil {
			return entity.DownloadJob{}, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return entity.DownloadJob{}, err
		}

		req.Header.Set("User-Agent", "go-school-downloader/1.0 (contact: tarek.fakhfakh@gmail.com)")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return entity.DownloadJob{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return entity.DownloadJob{}, fmt.Errorf("upstream returned %s", resp.Status)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return entity.DownloadJob{}, err
		}

		file := entity.File{
			Metadata: entity.FileMetadata{
				MimeType: resp.Header.Get("Content-Type"),
				Size:     resp.ContentLength,
			},
			Data: data,
		}

		fileID, err := u.FileRepository.Put(file)
		if err != nil {
			return entity.DownloadJob{}, err
		}

		fmt.Println("File ID: ", fileID)

		createdJob.FileIDs = append(createdJob.FileIDs, fileID)

		if err := u.DownloadJobRepository.Update(createdJob); err != nil {
			return entity.DownloadJob{}, err
		}

	}

	createdJob.Status = entity.Done
	if err := u.DownloadJobRepository.Update(createdJob); err != nil {
		return entity.DownloadJob{}, err
	}

	return createdJob, nil
}

func (u *DownloadUseCase) GetJob(jobID string) (entity.DownloadJob, error) {
	return u.DownloadJobRepository.Get(jobID)
}

func (u *DownloadUseCase) GetFile(jobID, fileID string) (entity.File, error) {
	return u.FileRepository.Get(fileID)
}
