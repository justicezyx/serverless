package core

import (
	"testing"
	"time"
)

func TestAPIUsageTracker(t *testing.T) {
	tracker := NewAPIUsageTracker()

	// Test API call tracking
	startTime := tracker.StartAPICall("user")
	time.Sleep(100 * time.Millisecond)
	tracker.EndAPICall("user", startTime)

	if count := tracker.count["user"]; count != 1 {
		t.Errorf("Expected call count 1, got %d", count)
	}

	if tracker.GetTotalTime("user") < 100*time.Millisecond {
		t.Errorf("Expected total time at least 100ms, got %d", tracker.GetTotalTime("user"))
	}

	startTime = tracker.StartAPICall("user")
	time.Sleep(200 * time.Millisecond)
	tracker.EndAPICall("user", startTime)

	if count := tracker.count["user"]; count != 2 {
		t.Errorf("Expected call count 2, got %d", count)
	}
}
