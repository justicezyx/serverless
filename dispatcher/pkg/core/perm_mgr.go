package core

import (
	"sync"
)

type PermMgr struct {
	mu       sync.RWMutex
	userAPIs map[string]map[string]bool
}

func NewUserAPIManager() *PermMgr {
	return &PermMgr{
		userAPIs: make(map[string]map[string]bool),
	}
}

func (uam *PermMgr) AllowUserAPI(user, api string) {
	uam.mu.Lock()
	defer uam.mu.Unlock()

	if uam.userAPIs[user] == nil {
		uam.userAPIs[user] = make(map[string]bool)
	}
	uam.userAPIs[user][api] = true
}

func (uam *PermMgr) IsUserAllowed(user, api string) bool {
	uam.mu.RLock()
	defer uam.mu.RUnlock()

	return uam.userAPIs[user][api]
}
