package muSupervisor

import "sync"

type Mutex struct {
	supervisedMutex
	mu sync.Mutex
}

func (m *Mutex) Lock() {
	m.mutexOp(LOCK, m.mu.Lock)
}
func (m *Mutex) Unlock() {
	m.mutexOp(UNLOCK, m.mu.Unlock)
}
