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
		)


func main(){
	//custom servermutiplexer
	router := mux.NewRouter()
	
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
	
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)

	log.Println("Shutting down server")
	os.Exit(0)

}