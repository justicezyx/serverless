package core

import (
	"sync"
)

// TODO: Add black list, and default: allow/deny.
type PermMgr struct {
	mu        sync.RWMutex
	whiteList map[string]map[string]bool
}

func NewPermMgr() PermMgr {
	return PermMgr{
		whiteList: make(map[string]map[string]bool),
	}
}

func (m *PermMgr) AllowUserAPI(user, api string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.whiteList[user] == nil {
		m.whiteList[user] = make(map[string]bool)
	}
	m.whiteList[user][api] = true
}

func (m *PermMgr) IsUserAllowed(user, api string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.whiteList[user][api]
}
