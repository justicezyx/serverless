package core

import (
	"sync"
	"sync/atomic"
	"time"
)

// TODO/Enhancement: Might create an APIContext object to hold all information related to invoking function.
// And create a APIMgr interface:
// type APIMgr interface {
//	 // Called before serving a function call request, so that we can add various checks with the same interface.
//	 StartAPICall(ctx APIContext) error
//
//	 // Called after serving a function call request, serves as campanion to StartAPICall().
//   FinishAPICall(ctx APIContext) error
// }

type APILimitMgr struct {
	limit      int64
	callCount  map[string]*int64
	semaphores map[string]chan struct{}
	mu         sync.Mutex
}

func NewAPILimitMgr(limit int64) APILimitMgr {
	return APILimitMgr{
		limit:      limit,
		callCount:  make(map[string]*int64),
		semaphores: make(map[string]chan struct{}),
	}
}

func (m *APILimitMgr) StartAPICall(api string, timeout time.Duration) bool {
	var semaphore chan struct{}

	m.mu.Lock()
	if _, exists := m.semaphores[api]; !exists {
		m.semaphores[api] = make(chan struct{}, m.limit)
		var count int64
		m.callCount[api] = &count
	}
	semaphore = m.semaphores[api]
	m.mu.Unlock()

	select {
	case semaphore <- struct{}{}:
		atomic.AddInt64(m.callCount[api], 1)
		return true
	case <-time.After(timeout):
		return false
	}
}

func (m *APILimitMgr) FinishAPICall(api string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.semaphores[api]; exists {
		select {
		case <-m.semaphores[api]:
			atomic.AddInt64(m.callCount[api], -1)
		default:
			panic("This should never happen if calls are correctly paired")
		}
	}
}

func (m *APILimitMgr) GetConcurrentCallCount(api string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if count, exists := m.callCount[api]; exists {
		return atomic.LoadInt64(count)
	}
	return 0
}

func (m *APILimitMgr) SetLimit(limit int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.limit = limit
}
