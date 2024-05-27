package core

import (
	"net/http"
)

func Dispatch(w http.ResponseWriter, r *http.Request) {
	// TODO: Launching instances and return the URL.
	target := ""
	ProxyRequest(target, w, r)
}
