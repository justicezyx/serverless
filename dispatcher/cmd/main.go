package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"serverless/dispatcher/pkg/core"
)

func main() {
	log.SetOutput(os.Stderr)

	var concurLimit int64
	flag.Int64Var(&concurLimit, "concur_limit", 3, "Set the concurrency limit")

	flag.Parse()

	dispatcher := core.NewDispatcher()

	dispatcher.SetAPIConcurLimit(concurLimit)
	log.Println("API limit is set to", concurLimit)

	r := mux.NewRouter()
	r.HandleFunc("/alpha", func(w http.ResponseWriter, r *http.Request) {
		ctx := core.CallContext{
			Fn:             "alpha",
			InstRdyTimeout: 12 * time.Second,
		}
		dispatcher.Dispatch(ctx, w, r)
	})
	r.HandleFunc("/beta", func(w http.ResponseWriter, r *http.Request) {
		ctx := core.CallContext{
			Fn:             "beta",
			InstRdyTimeout: 12 * time.Second,
		}
		dispatcher.Dispatch(ctx, w, r)
	})

	// Channel to listen for interrupt signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting server on :8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatalf("Could not start server: %s\n", err.Error())
		}
	}()

	// Block until an interrupt signal is received
	<-stopChan
	log.Println("Interrupt signal received. Shutting down...")

	dispatcher.StopLaunchMonitor()

	// Perform cleanup tasks
	dispatcher.Shutdown()

	log.Println("Server gracefully stopped.")
	os.Exit(0)
}
