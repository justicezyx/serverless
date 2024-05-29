package core

import (
	"fmt"
	"net/http"
	"time"
)

// Send HTTP GET request to the input url, every checkInterval, until timeout.
// Returns OK if get OK status within timeout.
func WaitForHTTPGetOK(url string, checkInterval, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := http.Client{
		Timeout: 100 * time.Millisecond,
	}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
		}
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(checkInterval)
	}
	return fmt.Errorf("request to %s did not succeed within the timeout period", url)
}
