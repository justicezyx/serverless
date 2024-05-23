package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Request struct {
	Prompt string `json:"prompt"`
}

type Response struct {
	Answer string `json:"answer"`
}

type RuntimeAlpha struct{}

func (r *RuntimeAlpha) Load() {
	time.Sleep(4 * time.Second)
}

func (r *RuntimeAlpha) Generate(prompt string) string {
	time.Sleep(750 * time.Millisecond)
	return fmt.Sprintf("Given your question: %s. I think the best answer is to buy ice cream.", prompt)
}

type RuntimeBeta struct{}

func (r *RuntimeBeta) Load() {
	time.Sleep(2.4 * time.Second)
}

func (r *RuntimeBeta) Generate(prompt string) string {
	time.Sleep(1.75 * time.Second)
	return fmt.Sprintf("Given your question: %s. I think the best answer is to get a hamburger.", prompt)
}

func alphaHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	alpha := &RuntimeAlpha{}
	alpha.Load()
	answer := alpha.Generate(req.Prompt)

	resp := Response{Answer: answer}
	json.NewEncoder(w).Encode(resp)
}

func betaHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	beta := &RuntimeBeta{}
	beta.Load()
	answer := beta.Generate(req.Prompt)

	resp := Response{Answer: answer}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/alpha", alphaHandler)
	http.HandleFunc("/beta", betaHandler)
	http.ListenAndServe(":8080", nil)
}
