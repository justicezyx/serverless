package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"serverless/dispatcher/pkg/core"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func cleanup() {
	// Perform any necessary cleanup here
	log.Println("Performing cleanup tasks...")
	core.Shutdown()
	log.Println("Cleanup completed.")
}

func main() {
	var concurLimit int64
	flag.Int64Var(&concurLimit, "concur_limit", 3, "Set the concurrency limit")

	flag.Parse()

	core.SetAPIConcurLimit(concurLimit)
	log.Println("API limit is set to", concurLimit)

	r := mux.NewRouter()
	r.HandleFunc("/alpha", func(w http.ResponseWriter, r *http.Request) {
		core.Dispatch("alpha", w, r)
	})
	r.HandleFunc("/beta", func(w http.ResponseWriter, r *http.Request) {
		core.Dispatch("beta", w, r)
	})

	ticker := core.NewTicker(time.Second)
	ticker.Start(func() {
		log.Println("apiMgr count alpha", core.GetAPIMgr().GetConcurrentCallCount("alpha"))
		log.Println("apiMgr count beta", core.GetAPIMgr().GetConcurrentCallCount("beta"))
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

	ticker.Stop()
	// Perform cleanup tasks
	cleanup()

	log.Println("Server gracefully stopped.")
	os.Exit(0)
}
