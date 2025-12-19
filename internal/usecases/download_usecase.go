package usecases

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"gin-quickstart/internal/domain/entity"
	"gin-quickstart/internal/domain/ports"
	repository "gin-quickstart/internal/infra/repository/memory"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	fileMaxSize = int64(10 << 20) // 10mb
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

func (u *DownloadUseCase) fetchFile(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "go-school-downloader/1.0 (contact: dim.i@gmail.com)")

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
	return fmt.Sprintf("Upstream error: %s %d", e.Status, e.StatusCode)
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

func isFatalErr(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}

type jobCollector struct {
	mu  sync.Mutex
	job *entity.DownloadJob
}

func NewJobCollector(job *entity.DownloadJob) *jobCollector {
	return &jobCollector{
		job: job,
	}
}

func (jc *jobCollector) addItemError(url string, err error) {
	jc.mu.Lock()
	defer jc.mu.Unlock()
	jc.job.Items = append(jc.job.Items, entity.DownloadItem{
		URL:   url,
		Error: &entity.DownloadItemError{Code: getErrorCode(err)},
	})
}

func (jc *jobCollector) addItemSuccess(url, fileID string) {
	jc.mu.Lock()
	defer jc.mu.Unlock()
	jc.job.Items = append(jc.job.Items, entity.DownloadItem{
		URL:    url,
		FileID: fileID,
	})
}

func handleErr(jc *jobCollector, url string, err error) error {
	jc.addItemError(url, err)

	if isFatalErr(err) {
		return err
	}
	return nil
}

func (u *DownloadUseCase) runJob(ctx context.Context, job entity.DownloadJob, urls []string) entity.DownloadJob {
	var (
		g errgroup.Group
	)

	g.SetLimit(10)

	gCtx, gCancel := context.WithCancel(ctx)
	defer gCancel()

	jc := NewJobCollector(&job)

	for _, url := range urls {

		url := url

		g.Go(func() error {
			if err := gCtx.Err(); err != nil {
				return err
			}

			resp, err := u.fetchFile(ctx, url)
			if err != nil {
				slog.Warn("download %s failed: %v", url, err)

				return handleErr(jc, url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				slog.Warn("download %s upstream error: %v", url, err)

				return handleErr(jc, url, &upstreamError{Status: resp.Status, StatusCode: resp.StatusCode})

			}

			lr := &io.LimitedReader{R: resp.Body, N: fileMaxSize}
			var buf bytes.Buffer
			n, err := buf.ReadFrom(lr)
			if err != nil {
				return handleErr(jc, url, err)
			}
			if n > fileMaxSize {
				return handleErr(jc, url, &upstreamError{Status: resp.Status, StatusCode: http.StatusRequestEntityTooLarge})
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
				return handleErr(jc, url, err)
			}

			jc.addItemSuccess(url, fileID)
			return nil

		})
	}

	err := g.Wait()

	if err != nil && isFatalErr(err) {
		job.Status = entity.Failed
	} else {
		job.Status = entity.Done
	}

	_ = u.DownloadJobRepository.Update(ctx, job)

	return job
}

func (u *DownloadUseCase) StartJob(rCtx context.Context, duration time.Duration, urls []string) (entity.DownloadJob, error) {
	parentCtx := context.WithoutCancel(rCtx) // detach from parent request context
	ctx, cancel := context.WithTimeout(parentCtx, duration)

	jobEntity := entity.DownloadJob{
		Status:    entity.Process,
		Timeout:   duration,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	createdJob, err := u.DownloadJobRepository.Create(ctx, jobEntity)
	if err != nil {
		cancel()
		return entity.DownloadJob{}, err
	}

	go func() {
		defer cancel()
		_ = u.runJob(ctx, createdJob, urls)
	}()

	return createdJob, nil
}

func (u *DownloadUseCase) GetJob(rCtx context.Context, jobID string) (entity.DownloadJob, error) {
	return u.DownloadJobRepository.Get(rCtx, jobID)
}

func (u *DownloadUseCase) GetFile(rCtx context.Context, jobID, fileID string) (entity.File, error) {
	return u.FileRepository.Get(rCtx, fileID)
}
