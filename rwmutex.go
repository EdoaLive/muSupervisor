package muSupervisor

import "sync"

type RWMutex struct {
	supervisedMutex
	mu sync.RWMutex
}

func (m *RWMutex) Lock() {
	m.mutexOp(LOCK, m.mu.Lock)
}
func (m *RWMutex) Unlock() {
	m.mutexOp(UNLOCK, m.mu.Unlock)
}
func (m *RWMutex) RLock() {
	m.mutexOp(RLOCK, m.mu.RLock)
}
func (m *RWMutex) RUnlock() {
	m.mutexOp(RUNLOCK, m.mu.RUnlock)
}
