package core

import (
	"sync"
	"testing"
	"time"
)

// TestNewAPIManager tests the creation of a new APIManager
func TestNewAPIManager(t *testing.T) {
	limit := int64(3)
	manager := NewAPIManager(limit)
	if manager == nil {
		t.Fatalf("Expected new APIManager instance, got nil")
	}
	if manager.limit != limit {
		t.Fatalf("Expected limit to be %d, got %d", limit, manager.limit)
	}
	if manager.callCount == nil {
		t.Fatalf("Expected initialized callCount map, got nil")
	}
	if manager.semaphores == nil {
		t.Fatalf("Expected initialized semaphores map, got nil")
	}
}

// TestStartAPICall tests the StartAPICall method
func TestStartAPICall(t *testing.T) {
	limit := int64(3)
	manager := NewAPIManager(limit)

	api := "exampleAPI"
	if !manager.StartAPICall(api, 1*time.Second) {
		t.Fatalf("Expected StartAPICall to succeed, got failure")
	}
	if manager.GetConcurrentCallCount(api) != 1 {
		t.Fatalf("Expected concurrent call count to be 1, got %d", manager.GetConcurrentCallCount(api))
	}

	manager.StartAPICall(api, 1*time.Second)
	manager.StartAPICall(api, 1*time.Second)
	if manager.StartAPICall(api, 1*time.Second) {
		t.Fatalf("Expected StartAPICall to fail when limit is reached, got success")
	}
	if manager.GetConcurrentCallCount(api) != 3 {
		t.Fatalf("Expected concurrent call count to be 3, got %d", manager.GetConcurrentCallCount(api))
	}
}

// TestFinishAPICall tests the FinishAPICall method
func TestFinishAPICall(t *testing.T) {
	limit := int64(3)
	manager := NewAPIManager(limit)

	api := "exampleAPI"
	manager.StartAPICall(api, 1*time.Second)
	manager.StartAPICall(api, 1*time.Second)

	manager.FinishAPICall(api)
	if manager.GetConcurrentCallCount(api) != 1 {
		t.Fatalf("Expected concurrent call count to be 1 after finishing one call, got %d", manager.GetConcurrentCallCount(api))
	}

	manager.FinishAPICall(api)
	if manager.GetConcurrentCallCount(api) != 0 {
		t.Fatalf("Expected concurrent call count to be 0 after finishing all calls, got %d", manager.GetConcurrentCallCount(api))
	}
}

// TestConcurrentCalls tests concurrent API calls
func TestConcurrentCalls(t *testing.T) {
	limit := int64(3)
	manager := NewAPIManager(limit)

	api := "exampleAPI"
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if manager.StartAPICall(api, 1*time.Second) {
				time.Sleep(100 * time.Millisecond) // Simulate API call duration
				manager.FinishAPICall(api)
			}
		}()
	}
	wg.Wait()

	if manager.GetConcurrentCallCount(api) != 0 {
		t.Fatalf("Expected concurrent call count to be 0 after all calls finished, got %d", manager.GetConcurrentCallCount(api))
	}
}

// TestStartAPICallTimeout tests the timeout when starting an API call
func TestStartAPICallTimeout(t *testing.T) {
	limit := int64(1)
	manager := NewAPIManager(limit)

	api := "exampleAPI"
	if !manager.StartAPICall(api, 1*time.Second) {
		t.Fatalf("Expected StartAPICall to succeed, got failure")
	}
	if manager.GetConcurrentCallCount(api) != 1 {
		t.Fatalf("Expected concurrent call count to be 1, got %d", manager.GetConcurrentCallCount(api))
	}

	// Attempt to start another call with a short timeout
	if manager.StartAPICall(api, 100*time.Millisecond) {
		t.Fatalf("Expected StartAPICall to fail due to timeout, got success")
	}
	if manager.GetConcurrentCallCount(api) != 1 {
		t.Fatalf("Expected concurrent call count to still be 1, got %d", manager.GetConcurrentCallCount(api))
	}

	manager.FinishAPICall(api)
	if manager.GetConcurrentCallCount(api) != 0 {
		t.Fatalf("Expected concurrent call count to be 0 after finishing the call, got %d", manager.GetConcurrentCallCount(api))
	}
}
