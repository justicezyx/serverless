package core

import (
	"testing"
	"time"
)

// TestTickerStart verifies that the Ticker calls the provided function at the specified interval
func TestTickerStart(t *testing.T) {
	interval := time.Second
	ticker := NewTicker(interval)

	count := 0
	fn := func() {
		count++
	}

	// Start the ticker
	ticker.Start(fn)

	// Allow some time for the ticker to tick a few times
	time.Sleep(3 * time.Second)

	// Stop the ticker
	ticker.Stop()

	if count < 2 || count > 4 {
		t.Errorf("expected function to be called between 2 and 4 times, but got %d", count)
	}
}
