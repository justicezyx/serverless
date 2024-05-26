package core

import (
	"time"
)

// Timer struct to hold start and end times
type Timer struct {
	start time.Time
}

// Start method to record the start time
func NewTimer() *Timer {
	return &Timer{
		start: time.Now(),
	}
}

// Elapsed method to get the elapsed time duration
func (t *Timer) Elapsed() time.Duration {
	end := time.Now()
	return end.Sub(t.start)
}
