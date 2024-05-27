package core

import (
	"io"
	"net/http"
	"net/url"
)

func ProxyRequest(target string, w http.ResponseWriter, r *http.Request) {
	proxyURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, "Bad target URL", http.StatusBadRequest)
		return
	}

	proxyReq, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	proxyReq.Header = r.Header

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to get response from proxy URL", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, value := range resp.Header {
		w.Header().Set(key, value[0])
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
