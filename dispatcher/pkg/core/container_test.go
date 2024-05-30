package core

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestContainerRun tests the Run method of the Docker struct.
func TestContainerRun(t *testing.T) {
	// Go to $ToT/runtime for instructions of building this image.
	// This image has to be built locally, we don't do docker pull.
	image := "runtime:latest"
	cmd := []string{"python", "runtime.py", "--file=runtime_alpha.py", "--class_name=RuntimeAlpha"}

	timer := NewTimer()
	container := NewContainer(image, cmd)
	fmt.Println("NewContainer time duration:", timer.Elapsed())

	timer = NewTimer()
	rc, err := container.Run("test-container")
	fmt.Println("RunContainer time duration:", timer.Elapsed())

	assert.Nil(t, err, "Expected no error, got %v", err)
	assert.NotEmpty(t, rc.Url, "Expect non-empty URL to the container instance")

	timer = NewTimer()
	assert.Nil(t, rc.Stop(), "Expected no error stopping container")
	fmt.Println("StopContainer time duration:", timer.Elapsed())

	timer = NewTimer()
	assert.Nil(t, rc.Remove(), "Expected no error removing container")
	fmt.Println("RemoveContainer time duration:", timer.Elapsed())
}

func TestWaitForReady(t *testing.T) {
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

	go func() {
		// Start the server
		if err := http.Serve(listener, nil); err != nil {
			log.Fatalf("Error serving: %v", err)
		}
	}()

	url := fmt.Sprintf("http://localhost:%d/ready", port)
	c := RunningContainer{
		readyUrl: url,
	}
	assert.False(t, c.IsReady())
	err = c.WaitForReady(time.Second)
	assert.Nil(t, err)
	assert.True(t, c.IsReady())
}
