package dispatcher

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type RuntimeInstance struct {
	ID string
	Url string

	// A group of goroutines of the concurrent requests to this instance.
	// Key is endpoint address
	Exec errgroup.Group
}

// Return true if the invocation succeeded.
func (i RuntimeInstance) Invoke(prompt string) (string, error) {
	jsonData := []byte(fmt.Sprintf(`{"prompt": "%s"}`, prompt))
	req, err := http.NewRequest("POST", i.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Error creating HTTP request, error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending HTTP request, error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading HTTP response, error: %v", err)
	}
	return string(body), nil
}
