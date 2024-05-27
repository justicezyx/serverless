package core

import (
	"testing"
	"time"
)

func TestAPIUsageTracker(t *testing.T) {
	tracker := NewAPIUsageTracker()

	// Test API call tracking
	startTime := tracker.StartAPICall("API1")
	time.Sleep(100 * time.Millisecond)
	tracker.EndAPICall("API1", startTime)

	if count := tracker.count["API1"]; count != 1 {
		t.Errorf("Expected call count 1, got %d", count)
	}

	if tracker.GetTotalTime("API1") < 100*time.Millisecond {
		t.Errorf("Expected total time at least 100ms, got %d", tracker.GetTotalTime("API1"))
	}

	startTime = tracker.StartAPICall("API1")
	time.Sleep(200 * time.Millisecond)
	tracker.EndAPICall("API1", startTime)

	if count := tracker.count["API1"]; count != 2 {
		t.Errorf("Expected call count 2, got %d", count)
	}
}
