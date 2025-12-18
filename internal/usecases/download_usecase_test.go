package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gin-quickstart/internal/domain/entity"
	"gin-quickstart/internal/domain/ports/mocks"
	"gin-quickstart/internal/usecases"

	"github.com/golang/mock/gomock"
)

func TestDownloadUseCase_GetJob_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobRepo := mocks.NewMockDownloadJobRepository(ctrl)
	fileRepo := mocks.NewMockFileRepository(ctrl)

	u := usecases.NewDownloadUseCase()
	u.DownloadJobRepository = jobRepo
	u.FileRepository = fileRepo

	want := entity.DownloadJob{ID: "job-1"}

	jobRepo.EXPECT().
		Get(gomock.Any(), "job-1").
		Return(want, nil).
		Times(1)

	got, err := u.GetJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if got.ID != "job-1" {
		t.Fatalf("expected job-1, got %q", got.ID)
	}
}

func TestDownloadUseCase_GetJob_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobRepo := mocks.NewMockDownloadJobRepository(ctrl)
	fileRepo := mocks.NewMockFileRepository(ctrl)

	u := usecases.NewDownloadUseCase()
	u.DownloadJobRepository = jobRepo
	u.FileRepository = fileRepo

	jobRepo.EXPECT().
		Get(gomock.Any(), "job-1").
		Return(entity.DownloadJob{}, errors.New("db down")).
		Times(1)

	_, err := u.GetJob(context.Background(), "job-1")
	if err == nil {
		t.Fatalf("expected err, got nil")
	}
}

func TestDownloadUseCase_GetFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobRepo := mocks.NewMockDownloadJobRepository(ctrl)
	fileRepo := mocks.NewMockFileRepository(ctrl)

	u := usecases.NewDownloadUseCase()
	u.DownloadJobRepository = jobRepo
	u.FileRepository = fileRepo

	want := entity.File{
		Metadata: entity.FileMetadata{ID: "file-1", MimeType: "text/plain", Size: 3},
		Data:     []byte("abc"),
	}

	fileRepo.EXPECT().
		Get(gomock.Any(), "file-1").
		Return(want, nil).
		Times(1)

	got, err := u.GetFile(context.Background(), "job-ignored", "file-1")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if string(got.Data) != "abc" {
		t.Fatalf("expected data abc, got %q", string(got.Data))
	}
}

func TestDownloadUseCase_GetFile_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobRepo := mocks.NewMockDownloadJobRepository(ctrl)
	fileRepo := mocks.NewMockFileRepository(ctrl)

	u := usecases.NewDownloadUseCase()
	u.DownloadJobRepository = jobRepo
	u.FileRepository = fileRepo

	fileRepo.EXPECT().
		Get(gomock.Any(), "file-1").
		Return(entity.File{}, errors.New("not found")).
		Times(1)

	_, err := u.GetFile(context.Background(), "job-ignored", "file-1")
	if err == nil {
		t.Fatalf("expected err, got nil")
	}
}

func TestDownloadUseCase_StartJob_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobRepo := mocks.NewMockDownloadJobRepository(ctrl)
	fileRepo := mocks.NewMockFileRepository(ctrl)

	u := usecases.NewDownloadUseCase()
	u.DownloadJobRepository = jobRepo
	u.FileRepository = fileRepo

	created := entity.DownloadJob{
		ID:     "job-1",
		Status: entity.Process,
	}

	done := make(chan struct{})

	// StartJob must create a job...
	jobRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, job entity.DownloadJob) (entity.DownloadJob, error) {
			if job.Status != entity.Process {
				t.Fatalf("expected Process, got %v", job.Status)
			}
			if job.Timeout != 50*time.Millisecond {
				t.Fatalf("expected 50ms timeout, got %v", job.Timeout)
			}
			return created, nil
		}).
		Times(1)

	// ...and since we pass urls = empty, runJob finishes quickly and calls Update(Done)
	jobRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, job entity.DownloadJob) error {
			if job.ID != "job-1" {
				t.Fatalf("expected id job-1, got %q", job.ID)
			}
			if job.Status != entity.Done {
				t.Fatalf("expected Done, got %v", job.Status)
			}
			close(done)
			return nil
		}).
		Times(1)

	got, err := u.StartJob(context.Background(), 50*time.Millisecond, nil) // nil/empty => no HTTP
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if got.ID != "job-1" {
		t.Fatalf("expected job-1, got %q", got.ID)
	}

	select {
	case <-done:
		// ok, goroutine finished
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for Update(Done)")
	}
}

func TestDownloadUseCase_StartJob_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobRepo := mocks.NewMockDownloadJobRepository(ctrl)
	fileRepo := mocks.NewMockFileRepository(ctrl)

	u := usecases.NewDownloadUseCase()
	u.DownloadJobRepository = jobRepo
	u.FileRepository = fileRepo

	jobRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(entity.DownloadJob{}, errors.New("create failed")).
		Times(1)

	_, err := u.StartJob(context.Background(), 50*time.Millisecond, nil)
	if err == nil {
		t.Fatalf("expected err, got nil")
	}
}
