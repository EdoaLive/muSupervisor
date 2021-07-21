package muSupervisor

import "container/list"

type mutexOpMap map[*supervisedMutex]opMap

func (m mutexOpMap) mapRequest(d *opData) {
	if _, exists := m[d.mutexPtr]; !exists {
		m[d.mutexPtr] =
			opMap{
				make(map[*opData]struct{}),
				list.New(),
				list.New(),
			}
	}

	d.state = PENDING

	m[d.mutexPtr].allOps[d] = struct{}{}

	switch d.t {
	case LOCK:
		m[d.mutexPtr].pendingQueue.PushFront(d)
	case RLOCK:
		m[d.mutexPtr].pendingRQueue.PushFront(d)
	}
}

func (m mutexOpMap) mapUnlock(d *opData) {
	var e *list.Element
	switch d.t {
	case UNLOCK:
		e = m[d.mutexPtr].pendingQueue.Front()
		m[d.mutexPtr].pendingQueue.Remove(e)

	case RUNLOCK:
		e = m[d.mutexPtr].pendingRQueue.Front()
		m[d.mutexPtr].pendingRQueue.Remove(e)
	}

	if e != nil {
		e.Value.(*opData).state = DONE
		e.Value.(*opData).stackTrace = nil
	}
}

func (m mutexOpMap) doCheck() {
	for _, om := range m {
		om.checkPending()

		om.cleanup()
	}
}
