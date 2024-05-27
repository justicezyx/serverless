package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Proxy the request to the input target URL.
func ProxyRequest(target string, w http.ResponseWriter, r *http.Request) {
	proxyURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, fmt.Sprintf("The input target URL '%s' is invalid", target), http.StatusBadRequest)
		return
	}

	proxyReq, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request, error: %v", err), http.StatusInternalServerError)
		return
	}

	proxyReq.Header = r.Header

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get response from proxy URL '%s', error: %v", target, err),
			http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, value := range resp.Header {
		w.Header().Set(key, value[0])
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
