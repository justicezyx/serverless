package core

import (
	"sync"
)

type PermMgr struct {
	mu       sync.RWMutex
	userAPIs map[string]map[string]bool
}

func NewPermMgr() PermMgr {
	return PermMgr{
		userAPIs: make(map[string]map[string]bool),
	}
}

func (m *PermMgr) AllowUserAPI(user, api string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.userAPIs[user] == nil {
		m.userAPIs[user] = make(map[string]bool)
	}
	m.userAPIs[user][api] = true
}

func (m *PermMgr) IsUserAllowed(user, api string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.userAPIs[user][api]
}
