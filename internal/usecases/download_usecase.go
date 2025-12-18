package usecases

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"gin-quickstart/internal/domain/entity"
	"gin-quickstart/internal/domain/ports"
	repository "gin-quickstart/internal/infra/repository/memory"
	"gin-quickstart/pkg/json"
	"io"
	"net"
	"net/http"
	"time"
)

type DownloadUseCase struct {
	DownloadJobRepository ports.DownloadJobRepository
	FileRepository        ports.FileRepository
	httpClient            *http.Client
}

func NewDownloadUseCase() *DownloadUseCase {
	return &DownloadUseCase{
		DownloadJobRepository: repository.NewDownloadJobMemoryRepository(),
		FileRepository:        repository.NewFileMemoryRepository(),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (u *DownloadUseCase) createJobEntity(duration time.Duration) entity.DownloadJob {
	return entity.DownloadJob{
		Status:    entity.Process,
		Timeout:   duration,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u *DownloadUseCase) fetchFile(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "go-school-downloader/1.0 (contact: tarek.fakhfakh@gmail.com)")

	client := u.httpClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type upstreamError struct {
	Status     string
	StatusCode int
}

func (e *upstreamError) Error() string {
	return fmt.Sprintf("upstream error: %s %d", e.Status, e.StatusCode)
}

func getErrorCode(err error) entity.DownloadItemErrorCode {
	var netErr net.Error
	var upstreamErr *upstreamError
	if errors.Is(err, context.DeadlineExceeded) {
		return entity.ErrorTimeout
	} else if errors.As(err, &netErr) {
		return entity.ErrorHTTP
	} else if errors.As(err, &upstreamErr) {
		return entity.ErrorHTTP
	}
	return entity.ErrorUnknown
}

func (u *DownloadUseCase) runJob(ctx context.Context, job entity.DownloadJob, urls []string) entity.DownloadJob {
	sem := make(chan struct{}, 10)
	type res struct {
		url    string
		fileID string
		err    error
	}
	resCh := make(chan res, len(urls))

	for _, url := range urls {

		go func(goURL string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			// if timeout, return error
			if err := ctx.Err(); err != nil {
				resCh <- res{url: goURL, err: err}
				return
			}

			resp, err := u.fetchFile(ctx, goURL)
			if err != nil {
				resCh <- res{url: goURL, err: err}
				return
			}

			defer resp.Body.Close()

			json.PrettyPrint(resp.StatusCode)

			if resp.StatusCode != http.StatusOK {
				resCh <- res{url: goURL, err: &upstreamError{Status: resp.Status, StatusCode: resp.StatusCode}}
				return
			}

			const maxSize = int64(10 << 20) // 10mb
			lr := &io.LimitedReader{R: resp.Body, N: maxSize}

			var buf bytes.Buffer
			n, err := buf.ReadFrom(lr)
			if err != nil {
				resCh <- res{url: goURL, err: err}
				return
			}
			if n > maxSize {
				resCh <- res{url: goURL, err: &upstreamError{Status: resp.Status, StatusCode: http.StatusRequestEntityTooLarge}}
				return
			}

			data := buf.Bytes()

			file := entity.File{
				Metadata: entity.FileMetadata{
					MimeType: resp.Header.Get("Content-Type"),
					Size:     resp.ContentLength,
				},
				Data: data,
			}

			fileID, err := u.FileRepository.Create(ctx, file)
			if err != nil {
				resCh <- res{url: goURL, err: err}
				return
			}

			resCh <- res{url: goURL, fileID: fileID, err: nil}

		}(url)

	}

	for i := 0; i < len(urls); i++ {
		select {
		case <-ctx.Done():
			i = len(urls)
		case res := <-resCh:
			if res.err != nil {
				job.Items = append(job.Items, entity.DownloadItem{
					URL:   res.url,
					Error: &entity.DownloadItemError{Code: getErrorCode(res.err)},
				})
			} else if res.fileID != "" {
				job.Items = append(job.Items, entity.DownloadItem{
					URL:    res.url,
					FileID: res.fileID,
				})
			}
		}
	}

	job.Status = entity.Done
	_ = u.DownloadJobRepository.Update(ctx, job)

	return job
}

func (u *DownloadUseCase) StartJob(rCtx context.Context, duration time.Duration, urls []string) (entity.DownloadJob, error) {
	parentCtx := context.WithoutCancel(rCtx)
	jobCtx, cancel := context.WithTimeout(parentCtx, duration)

	jobEntity := u.createJobEntity(duration)
	job, err := u.DownloadJobRepository.Create(jobCtx, jobEntity)
	if err != nil {
		cancel()
		return entity.DownloadJob{}, err
	}

	go func() {
		defer cancel()
		_ = u.runJob(jobCtx, job, urls)
	}()

	return job, nil
}

func (u *DownloadUseCase) GetJob(rCtx context.Context, jobID string) (entity.DownloadJob, error) {
	return u.DownloadJobRepository.Get(rCtx, jobID)
}

func (u *DownloadUseCase) GetFile(rCtx context.Context, jobID, fileID string) (entity.File, error) {
	return u.FileRepository.Get(rCtx, fileID)
}
