package main

import (
	"log"
	"net/http"
	"time"
	"os"
    "os/signal"
	"context"
	"syscall"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	api"github.com/negeek/solar-sphere/solar-galaxy/api/v1"
		)


func main(){
	//custom servermutiplexer
	router := mux.NewRouter()
	router.HandleFunc("/", api.Gateway).Methods("GET")
	
	//custom server
	server:=&http.Server{
		Addr: ":8080",
		Handler: router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 *  time.Second,
	}

	// Run server in a goroutine so that it doesn't block.
	go func() {
		log.Println("Start server")
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

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)

}