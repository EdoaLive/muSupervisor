package muSupervisor

import (
	"fmt"
	"log"
)

type mutexOpMap map[*supervisedMutex]opMap

type reqQueue struct {
	read map[routineNum]*opData
	rw   map[routineNum]*opData
}

func newReqQueue() reqQueue {
	return reqQueue{
		read: make(map[routineNum]*opData),
		rw:   make(map[routineNum]*opData),
	}
}

func newOpMap() opMap {
	return opMap{
		pending:   newReqQueue(),
		active:    newReqQueue(),
		done:      make(map[routineNum]*opData),
		doneRWait: make(map[routineNum]*opData),
	}
}

func (m mutexOpMap) mapRequest(d *opData) {

	var pendingQueue map[routineNum]*opData

	_, mutexExists := m[d.mutexPtr]
	if !mutexExists {
		m[d.mutexPtr] = newOpMap()
	}

	switch d.t {
	case RLOCK:
		pendingQueue = m[d.mutexPtr].pending.read
	case LOCK:
		pendingQueue = m[d.mutexPtr].pending.rw
	}

	if previous, exists := pendingQueue[d.numRoutine]; exists {
		fmt.Printf("FATAL: muSupervisor: lock already requested on the same routine %d.\n", d.numRoutine)
		logOpDetails(previous)
		logOpDetails(d)
		log.Fatal("Exiting")
	}

	d.state = PENDING

	pendingQueue[d.numRoutine] = d
}

func (m mutexOpMap) mapObtained(d *opData) {
	// We always expect to enter here on the same routine and same *opData pointer.

	var pendingQueue, activeQueue map[routineNum]*opData

	switch d.t {
	case RLOCK:
		if len(m[d.mutexPtr].active.rw) != 0 {
			log.Fatal("ERROR: active rw lock when obtaining read lock.\n")
		}

		pendingQueue = m[d.mutexPtr].pending.read
		activeQueue = m[d.mutexPtr].active.read

	case LOCK:
		if len(m[d.mutexPtr].active.rw) != 0 {
			log.Fatal("ERROR: active ops not empty when obtaining lock.\n")
		}
		if len(m[d.mutexPtr].active.read) != 0 {
			log.Fatal("ERROR: read active ops not empty when obtaining lock.\n")
		}

		pendingQueue = m[d.mutexPtr].pending.rw
		activeQueue = m[d.mutexPtr].active.rw

	}

	pendOp, exists := pendingQueue[d.numRoutine]
	if !exists {
		log.Fatalf("ERROR: no pending op found for routine %d when obtaining lock.\n", d.numRoutine)
	}

	pendOp.state = ACTIVE
	activeQueue[d.numRoutine] = pendOp
	delete(pendingQueue, d.numRoutine)
}

func (m mutexOpMap) mapUnlock(d *opData) {

	var lockRt routineNum
	//TODO: see supervisedMutex.go, track the unlock better if needed.
	//  unlockRt := d.numRoutine

	var activeQueue, doneQueue map[routineNum]*opData
	var destState opState

	switch d.t {
	case RUNLOCK:
		if len(m[d.mutexPtr].active.read) < 1 {
			log.Fatalf("ERROR: no read op active was found when runlocking.\n")
		}

		activeQueue = m[d.mutexPtr].active.read

		// if we find the same routine that locked, we're sure (?) is the one being unlocked
		if _, found := activeQueue[d.numRoutine]; found {
			lockRt = d.numRoutine
			doneQueue = m[d.mutexPtr].done // In this case, we're positive that this operation is done
			destState = DONE
		} else {
			// otherwise, we currently can't track which routine did the lock. We randomly pick the first one.
			for lockRtNum := range activeQueue {
				lockRt = lockRtNum
				break
			}
			doneQueue = m[d.mutexPtr].doneRWait
			destState = DONERWAIT
		}

	case UNLOCK:
		if len(m[d.mutexPtr].active.rw) != 1 {
			log.Fatalf("ERROR: no single rw ops active was found when unlocking.\n")
		}

		activeQueue = m[d.mutexPtr].active.rw
		doneQueue = m[d.mutexPtr].done

		for lockRtNum := range activeQueue {
			lockRt = lockRtNum
		}
		destState = DONE
	}

	activeOp := activeQueue[lockRt]

	activeOp.state = destState

	delete(activeQueue, lockRt)
	doneQueue[lockRt] = activeOp

	if (d.t == RUNLOCK) && (len(activeQueue) == 0) {
		// All read locks are done, so we can move to the done queue.
		for rtNum, op := range m[d.mutexPtr].doneRWait {
			delete(m[d.mutexPtr].doneRWait, rtNum)
			op.state = DONE
			m[d.mutexPtr].done[rtNum] = op
		}
	}
}

func (m mutexOpMap) doCheck() {
	for k, om := range m {
		om.checkPending()

		isEmpty := om.cleanup()
		if isEmpty {
			delete(m, k)
		}
	}
}
