package core

import (
	"fmt"
	"net/http"
	"time"
)

// Send HTTP GET request to the input url, every checkInterval, until timeout.
// Returns OK if get OK status within timeout.
func WaitForHTTPGetOK(url string, checkInterval, timeout time.Duration) error {
	now := time.Now()
	deadline := now.Add(timeout)
	for time.Now().Before(deadline) {
		client := http.Client{
			Timeout: checkInterval,
		}
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(checkInterval)
	}
	return fmt.Errorf("request to %s did not succeed within the timeout period", url)
}
