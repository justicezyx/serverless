package core

import (
	"sync"
	"time"
)

// APIUsageTracker tracks the running time of APIs called by different users.
type APIUsageTracker struct {
	mu     sync.Mutex
	timing map[string]time.Duration
	count  map[string]int
}

// NewAPIUsageTracker initializes a new APIUsageTracker.
func NewAPIUsageTracker() APIUsageTracker {
	return APIUsageTracker{
		timing: make(map[string]time.Duration),
		count:  make(map[string]int),
	}
}

// StartAPICall records the start time of an API call.
func (tracker *APIUsageTracker) StartAPICall(user string) time.Time {
	return time.Now()
}

// EndAPICall records the end time of an API call and updates the total running time.
func (tracker *APIUsageTracker) EndAPICall(user string, startTime time.Time) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	duration := time.Since(startTime)
	tracker.timing[user] += duration
	tracker.count[user]++
}

// GetTotalTime returns the total running time of the specified API.
func (tracker *APIUsageTracker) GetTotalTime(user string) time.Duration {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	return tracker.timing[user]
}
