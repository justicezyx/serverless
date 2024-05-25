package main

import (
	"flag"
	"log"
	"net/http"

	"zyx.com/serverless/dispatcher"
)

func main() {
	// Define the max-instances flag
	maxInstances := flag.Int("max-instances", 5, "Maximum number of concurrent instances")
	targetURL := flag.String("target-url", "http://target-server:port/target-endpoint", "URL of the target server")
	flag.Parse()

	// Set up the HTTP server
	http.HandleFunc("/forward", dispatcher.ForwardRequest)
	port := ":8080"
	log.Printf("Server is listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
