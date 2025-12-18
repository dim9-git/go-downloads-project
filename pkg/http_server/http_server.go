package httpserver

import (
	"context"
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
}

type Option func(*HTTPServer)

func NewHTTPServer(handler http.Handler, options ...Option) *HTTPServer {
	srv := &HTTPServer{
		server: &http.Server{
			Handler: handler,
		},
	}

	for _, opt := range options {
		opt(srv)
	}

	return srv
}

func WithAddress(address string) Option {
	return func(srv *HTTPServer) {
		srv.server.Addr = address
	}
}

func WithMiddleware(middlewares ...func(http.Handler) http.Handler) Option {
	return func(srv *HTTPServer) {
		for _, middleware := range middlewares {
			srv.server.Handler = middleware(srv.server.Handler)
		}
	}
}

func (s *HTTPServer) Start() error {
	slog.Info("Starting HTTP server", "address", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	slog.Info("Stopping HTTP server", "address", s.server.Addr)
	return s.server.Shutdown(ctx)
}
