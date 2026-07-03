package relay

import (
	"sync"
)

type SubscriptionManager struct {
	mu    sync.RWMutex
	subs  map[string][]Filter
	conns map[string]map[string]struct{}
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subs:  make(map[string][]Filter),
		conns: make(map[string]map[string]struct{}),
	}
}

func (sm *SubscriptionManager) Add(connID, subID string, filters []Filter) bool {
	if !ValidateSubscriptionID(subID) {
		return false
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, ok := sm.subs[subID]; ok {
		return false
	}
	sm.subs[subID] = filters
	if sm.conns[connID] == nil {
		sm.conns[connID] = make(map[string]struct{})
	}
	sm.conns[connID][subID] = struct{}{}
	return true
}

func (sm *SubscriptionManager) Remove(subID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.subs, subID)
	for connID := range sm.conns {
		delete(sm.conns[connID], subID)
		if len(sm.conns[connID]) == 0 {
			delete(sm.conns, connID)
		}
	}
}

func (sm *SubscriptionManager) GetFilters(subID string) []Filter {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.subs[subID]
}

func (sm *SubscriptionManager) Exists(subID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, ok := sm.subs[subID]
	return ok
}

func (sm *SubscriptionManager) RemoveAllForConn(connID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	subs, ok := sm.conns[connID]
	if !ok {
		return
	}
	for subID := range subs {
		delete(sm.subs, subID)
	}
	delete(sm.conns, connID)
}

func (sm *SubscriptionManager) ConnSubscriptionCount(connID string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.conns[connID])
}
