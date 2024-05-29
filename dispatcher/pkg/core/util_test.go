package core

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestWaitForHTTPGetOK(t *testing.T) {
	// Define the /ready handler
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Listen on a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	// Get the actual port assigned
	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Server is listening on port %d\n", port)

	url := fmt.Sprintf("http://localhost:%d/ready", port)
	expErrMsg := fmt.Sprintf("request to %s did not succeed within the timeout period", url)
	err = WaitForHTTPGetOK(url, 100*time.Millisecond, time.Second)
	if err == nil || err.Error() != expErrMsg {
		t.Errorf("Should have error, got %v", err)
	}

	go func() {
		// Start the server
		if err := http.Serve(listener, nil); err != nil {
			log.Fatalf("Error serving: %v", err)
		}
	}()

	err = WaitForHTTPGetOK(url, 100*time.Millisecond, time.Second)
	if err != nil {
		t.Errorf("Should have succeeded, got %v", err)
	}
}
