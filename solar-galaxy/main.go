package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	gateway "github.com/negeek/solar-sphere/solar-galaxy/api/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/env"
	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
	"github.com/negeek/solar-sphere/solar-spectrum/logging"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	// .env is optional — present for local dev, absent in Docker (where
	// Compose injects env vars directly). Only a malformed file is an error.
	if err := env.Load(".env"); err != nil && !os.IsNotExist(err) {
		panic("solar-galaxy: loading .env: " + err.Error())
	}

	log := logging.New("solar-galaxy")

	// Base URLs are env-driven so the same binary resolves backends by
	// localhost in local dev and by Docker service name in docker-compose.
	serviceBaseURLs := map[string]string{
		"auth":     envOr("AUTH_BASE_URL", "http://localhost:3000"),
		"sentinel": envOr("SENTINEL_BASE_URL", "http://localhost:5000"),
	}
	gw := gateway.NewGateway(serviceBaseURLs)

	router := mux.NewRouter()
	router.Use(httpapi.CORS)
	router.HandleFunc("/{path:.*}", gw.Handle).Methods("POST", "GET", "OPTIONS", "PUT", "DELETE", "PATCH")

	port := envOr("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("starting server", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown", "error", err)
	}

	log.Info("shut down")
}
