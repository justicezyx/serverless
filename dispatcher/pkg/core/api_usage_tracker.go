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

	// TODO/Req: Add tracking of the number of concurrent API calls for each instance.
	// The goal is to track each RunningContainer's backup calls. Use map[string]*int64, the key is containerID, value is
	// the busy time (aka. the actual time used for serving function requests), the busy-time/running-time is the
	// concurrency level, assume the maximal concurrency level to be C, and desired concurrency level fraction's upper
	// limit to be F1, and lower limit to be F2
	// 1. Whenever average busy-time/running-time among all running containers, is > F1*C, we should add new instances.
	// 2. Whenever average busy-time/running-time among all running containers, is < F2*C, we should reduce instances.
	//
	// The above check should happen periodically in accordance to the fluctuation of utilization ratio.
	//
	// I.e., if the usage
	// fraction changed C in time T, we should check C'/(C/T), C' is the desired change in utilization ratio we want each
	// observation cycle to be.
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
