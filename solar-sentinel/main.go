package main

import (
	"os"
	"log"
	"time"
	"context"
	"syscall"
	"net/http"
    "os/signal"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	v1routes "github.com/negeek/solar-sphere/solar-sentinel/api/v1"
	"github.com/negeek/solar-sphere/solar-sentinel/db"
	v1middlewares "github.com/negeek/solar-sphere/solar-sentinel/middlewares/v1"
)

func main() {
	appEnv:=os.Getenv("APP_ENV")
	if appEnv=="dev"{
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

    //custom servermutiplexer
	router := mux.NewRouter()
	router.Use(v1middlewares.CORS)
	v1routes.Routes(router.StrictSlash(true))

    //custom server
	server:=&http.Server{
		Addr: ":5000",
		Handler: router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 *  time.Second,
	}

    // DB connection
	dbUrl:= os.Getenv("DATABASE_URL")
	dbName:=os.Getenv("DB_NAME")
	log.Println("Connecting to db")
	dbctx, dbcancel, err:= db.Connect(dbUrl,dbName)
	if err != nil {
		log.Fatal(err)
	}
	
    go func() {
		log.Println("Start server")
		if err:= server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

    c := make(chan os.Signal, 1)
	// accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL will not be caught.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// disconnect db
	db.Disconnect(dbctx,dbcancel)
    server.Shutdown(ctx)
	os.Exit(0)
}

