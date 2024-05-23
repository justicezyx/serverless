package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Request struct {
	Prompt string `json:"prompt"`
}

type Response struct {
	Answer string `json:"answer"`
}

func invokeRuntime(runtime string, prompt string) (*Response, error) {
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

func main() {
	runtime := "alpha" // or "beta"
	prompt := "What should I eat today?"

	response, err := invokeRuntime(runtime, prompt)
	if err != nil {
		fmt.Printf("Error invoking %s: %v\n", runtime, err)
		return
	}

	fmt.Printf("Response from %s: %s\n", runtime, response.Answer)
}

