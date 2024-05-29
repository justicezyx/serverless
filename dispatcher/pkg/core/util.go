package core

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Send HTTP GET request to the input url, every checkInterval, until timeout.
// Returns OK if get OK status within timeout.
func WaitForHTTPGetOK(url string, checkInterval, timeout time.Duration) error {
	now := time.Now()
	log.Println("now:", now)
	deadline := now.Add(timeout)
	log.Println("deadline:", deadline)
	for time.Now().Before(deadline) {
		client := http.Client{
			Timeout: checkInterval,
		}
		resp, err := client.Get(url)
		log.Println("resp:", resp, "err:", err)
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
