package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gin-quickstart/internal/transport/http/handlers"
)

func NewRouter(
	httpHandlers *handlers.HTTPHandlers,
) http.Handler {
	r := chi.NewRouter()

	r.Route("/downloads", func(r chi.Router) {
		r.Post("/", httpHandlers.CreateDownloadJob)
		r.Get("/{jobID}", httpHandlers.GetDownloadJob)
		r.Get("/{jobID}/files/{fileID}", httpHandlers.GetFile)
	})

	return r
}
