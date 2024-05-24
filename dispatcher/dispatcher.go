package dispatcher

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type RuntimeLauncher struct {
	ContainerImage string

	DockerCmd []string
}

// Launch runtime instance
func (l RuntimeLauncher) Launch() RuntimeInstance {
	return RuntimeInstance{}
}

type Runtime struct {
	ID string

	// The maximal number of instances can run concurrently.
	MaxInstances int

	// The launcher to launch instance
	Launcher RuntimeLauncher

	// Map from endpoint to runtime instance
	Instances map[string]RuntimeInstance
}

type Dispatcher struct {
	// Maximal number of concurrent calls for each serverless runtime.
	// Beyond this number, new instances will be launched.
	maxConCallsPerRuntime int

	runtimeInstanceGroups map[string]

	// A group of goroutines can execute concurrently sending requests to different endpoint.
	// Key is endpoint address
	exec map[string]errgroup.Group
}

func forwardRequest(ctx context.Context, r *http.Request) error {
	// Define the target server URL
	targetURL := "http://target-server:port/target-endpoint"

	// Read the body of the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Create a new request to the target server
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Copy headers from the original request to the new request
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Log the response status for debugging purposes
	log.Printf("Response from target server: %s", resp.Status)
	return nil
}

func ForwardRequest(w http.ResponseWriter, r *http.Request) {
	var g errgroup.Group
	g.SetLimit(3)

	g.Go(func() error {
		return forwardRequest(context.Background(), r)
	})

	if err := g.Wait(); err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}

	// Send a response back to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request forwarded successfully"))
}

func main() {
	// Define the max-instances flag
	maxInstances := flag.Int("max-instances", 5, "Maximum number of concurrent instances")
	targetURL := flag.String("target-url", "http://target-server:port/target-endpoint", "URL of the target server")
	flag.Parse()

	// Set up the HTTP server
	http.HandleFunc("/forward", ForwardRequest)
	port := ":8080"
	log.Printf("Server is listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

