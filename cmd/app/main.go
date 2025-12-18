package app

import (
	"context"
	router "gin-quickstart/internal/transport/http"
	"gin-quickstart/internal/transport/http/handlers"
	"gin-quickstart/pkg/graceful_shutdown"
	httpserver "gin-quickstart/pkg/http_server"
	"gin-quickstart/pkg/http_server/mw"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpHandlers := handlers.NewHTTPHandlers()

	router := router.NewRouter(httpHandlers)

	server := httpserver.NewHTTPServer(router,
		httpserver.WithAddress(":8080"),
		httpserver.WithMiddleware(mw.RequestMetadata))

	gfl := graceful_shutdown.NewGracefulShutdown(ctx)

	gfl.Go(server.Start)
	gfl.MustClose(server.Stop)

	gfl.Wait()
}
