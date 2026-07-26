package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	v1 "github.com/negeek/solar-sphere/solar-auth/api/v1"
	repo "github.com/negeek/solar-sphere/solar-auth/repository/v1"
	service "github.com/negeek/solar-sphere/solar-auth/service/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/env"
	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
	"github.com/negeek/solar-sphere/solar-spectrum/logging"
	"github.com/negeek/solar-sphere/solar-spectrum/mongoutil"
)

func main() {
	// .env is optional — present for local dev, absent in Docker (where
	// Compose injects env vars directly). Only a malformed file is an error.
	if err := env.Load(".env"); err != nil && !os.IsNotExist(err) {
		panic("solar-auth: loading .env: " + err.Error())
	}

	log := logging.New("solar-auth")

	ctx := context.Background()
	client, db, err := mongoutil.Connect(ctx, os.Getenv("DATABASE_URL"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Error("connect to mongo", "error", err)
		os.Exit(1)
	}
	log.Info("connected to mongo")

	repository := repo.NewRepository(db)
	authService := service.NewAuthService(repository, os.Getenv("SIGNING_KEY"), os.Getenv("VERIFICATION_KEY"))
	handler := v1.NewHandler(authService)

	router := mux.NewRouter()
	router.Use(httpapi.CORS)
	v1.Routes(router.StrictSlash(true), handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
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

	if err := mongoutil.Disconnect(shutdownCtx, client, 15*time.Second); err != nil {
		log.Error("disconnect mongo", "error", err)
	}
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown", "error", err)
	}

	log.Info("shut down")
}
