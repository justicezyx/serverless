package core

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type Runtime struct {
	ID string

	// The maximal number of instances can run concurrently.
	MaxInstances int

	// Map from endpoint group to runtime instances
	InstanceGroups map[string][]*Docker

	// Map from endpoint to runtime instance
	Instances map[string]*Docker
}

type Dispatcher struct {
}

func forwardRequest(ctx context.Context, targetURL string, r *http.Request) error {
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
		targetURL := "http://target-server:port/target-endpoint"
		return forwardRequest(context.Background(), targetURL, r)
	})

	if err := g.Wait(); err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}

	// Send a response back to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request forwarded successfully"))
}
