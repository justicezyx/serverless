package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
)

type Request struct {
	Prompt string `json:"prompt"`
}

type Response struct {
	Answer string `json:"answer"`
}

var alphaCount, betaCount int
var mu sync.Mutex

const maxInstances = 3

func startContainer(runtime string) error {
	cmd := exec.Command("docker", "run", "-d", "--rm", "--name", runtime, "your-container-image", runtime)
	return cmd.Run()
}

func stopContainer(runtime string) error {
	cmd := exec.Command("docker", "stop", runtime)
	return cmd.Run()
}

func dispatchHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	runtime := r.URL.Path[1:]

	mu.Lock()
	if runtime == "alpha" && alphaCount >= maxInstances {
		http.Error(w, "Max Alpha instances reached", http.StatusTooManyRequests)
		mu.Unlock()
		return
	} else if runtime == "beta" && betaCount >= maxInstances {
		http.Error(w, "Max Beta instances reached", http.StatusTooManyRequests)
		mu.Unlock()
		return
	}

	if runtime == "alpha" {
		alphaCount++
	} else if runtime == "beta" {
		betaCount++
	}
	mu.Unlock()

	err := startContainer(runtime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := invokeRuntime(runtime, req.Prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu.Lock()
	if runtime == "alpha" {
		alphaCount--
	} else if runtime == "beta" {
		betaCount--
	}
	mu.Unlock()

	stopContainer(runtime)

	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/alpha", dispatchHandler)
	http.HandleFunc("/beta", dispatchHandler)
	http.ListenAndServe(":8080", nil)
}

func invokeRuntime(runtime, prompt string) (*Response, error) {
	reqBody, err := json.Marshal(Request{Prompt: prompt})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("http://localhost:8080/%s", runtime), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

