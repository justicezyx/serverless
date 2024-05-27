package core

import "time"

// Ticker wraps the functionality of calling a function periodically
type Ticker struct {
	interval time.Duration
	stop     chan bool
}

// NewTicker creates a new Ticker
func NewTicker(interval time.Duration) Ticker {
	return Ticker{
		interval: interval,
		stop:     make(chan bool),
	}
}

// Start begins calling the provided function at the specified interval
func (t *Ticker) Start(fn func()) {
	ticker := time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-t.stop:
				ticker.Stop()
				return
			case <-ticker.C:
				fn()
			}
		}
	}()
}

// Stop stops the ticker
func (t *Ticker) Stop() {
	t.stop <- true
}
