package main

import (
	"dispatcher"
	"log"
	"net/http"

	"github.com/justicezyx/serverless/dispatcher"
)

func main() {
	// Set up the HTTP server
	http.HandleFunc("/forward", dispatcher.ForwardRequest)
	port := ":8080"
	log.Printf("Server is listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
