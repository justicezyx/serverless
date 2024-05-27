package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"serverless/dispatcher/pkg/core"
	"syscall"

	"github.com/gorilla/mux"
)

func cleanup() {
	// Perform any necessary cleanup here
	log.Println("Performing cleanup tasks...")
	core.Shutdown()
	log.Println("Cleanup completed.")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/alpha", func(w http.ResponseWriter, r *http.Request) {
		core.Dispatch("alpha", w, r)
	})
	r.HandleFunc("/beta", func(w http.ResponseWriter, r *http.Request) {
		core.Dispatch("beta", w, r)
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

	// Perform cleanup tasks
	cleanup()

	log.Println("Server gracefully stopped.")
	os.Exit(0)
}
