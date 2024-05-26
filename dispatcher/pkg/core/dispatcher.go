package core

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
)

type Dispatcher struct {
	// Map from the function to the Container template.
	fnContainerMap map[string]ContainerInterface

	// A map from the exposed url of docker and the running docker container.
	// This is used to route request to corresponding container.
	urlInstanceMap map[string]RunningContainer

	// A map from the function to the corresponding container.
	fnInstanceMap map[string][]RunningContainer
}

func NewDispatcher() Dispatcher {
	return Dispatcher{
		fnContainerMap: make(map[string]ContainerInterface),
		urlInstanceMap: make(map[string]RunningContainer),
		fnInstanceMap:  make(map[string][]RunningContainer),
	}
}

func (d *Dispatcher) RegisterContainer(fn string, c ContainerInterface) {
	d.fnContainerMap[fn] = c
}

func (d *Dispatcher) Launch(fn string) error {
	c, ok := d.fnContainerMap[fn]
	if !ok {
		return fmt.Errorf("Could not find Container for serverless function %s", fn)
	}
	rc, err := c.Run()
	if err != nil {
		return fmt.Errorf("Could not run container for function: %s, error: %v", fn, err)
	}
	d.urlInstanceMap[rc.Url] = rc
	if _, ok := d.fnInstanceMap[fn]; !ok {
		d.fnInstanceMap[fn] = make([]RunningContainer, 0)
	}
	d.fnInstanceMap[fn] = append(d.fnInstanceMap[fn], rc)
	return nil
}

// Returns the URL for the particular function.
func (d Dispatcher) PickUrl(fn string) (string, error) {
	rcs, ok := d.fnInstanceMap[fn]
	if !ok || len(rcs) == 0 {
		return "", fmt.Errorf("No running container for function %s", fn)
	}
	idx := rand.Intn(len(rcs))
	return rcs[idx].Url, nil
}

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
