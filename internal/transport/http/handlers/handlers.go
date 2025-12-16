package handlers

import "gin-quickstart/internal/usecases"

type HTTPHandlers struct {
	DownloadUseCase *usecases.DownloadUseCase
}

func NewHTTPHandlers() *HTTPHandlers {
	return &HTTPHandlers{
		DownloadUseCase: usecases.NewDownloadUseCase(),
	}
}
