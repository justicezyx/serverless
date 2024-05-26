package main

import (
	"log"
	"net/http"
	"serverless/dispatcher/pkg/core"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/alpha", func(w http.ResponseWriter, r *http.Request) {
		core.ProxyRequest("http://127.0.0.1:8001/invoke", w, r)
	})
	r.HandleFunc("/beta", func(w http.ResponseWriter, r *http.Request) {
		core.ProxyRequest("http://127.0.0.1:8002/invoke", w, r)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
