package main

import (
	"log"
	"net/http"
	"serverless/dispatcher/pkg/core"
)

func main() {
	// Set up the HTTP server
	http.HandleFunc("/forward", core.ForwardRequest)
	port := ":8080"
	log.Printf("Server is listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
