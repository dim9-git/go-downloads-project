package main

import (
	router "gin-quickstart/internal/transport/http"
	"gin-quickstart/internal/transport/http/handlers"
	"log"
	"net/http"
)

func main() {
	httpHandlers := handlers.NewHTTPHandlers()

	router := router.NewRouter(httpHandlers)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
