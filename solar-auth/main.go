package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"context"
	"syscall"
	"net/http"
    "os/signal"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/negeek/solar-sphere/solar-auth/db"
	v1routes "github.com/negeek/solar-sphere/solar-auth/api/v1"
		)


func main(){
	appEnv:=os.Getenv("APP_ENV")
	if appEnv=="dev"{
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	//custom servermutiplexer
	router := mux.NewRouter()
	v1routes.Routes(router.StrictSlash(true))
	
	// DB connection
	dbURL:= os.Getenv("DATABASE_URL")
	fmt.Println("Connecting to db")
	dbctx, dbcancel, err:= db.Connect(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	
	
	//custom server
	server:=&http.Server{
		Addr: ":3000",
		Handler: router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 *  time.Second,
	}

	// Run server in a goroutine so that it doesn't block.
	go func() {
		fmt.Println("Start solar-auth server")
		if err:= server.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL will not be caught.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c
	// disconnect db
	db.Disconnect(dbctx,dbcancel)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)

	fmt.Println("Shutting down solar-auth server")
	os.Exit(0)

}