package muSupervisor

import (
	"fmt"
	"time"
)

// opMap keep track of all operations of a single mutex instance
type opMap struct {
	pending   reqQueue
	active    reqQueue
	done      map[routineNum]*opData
	doneRWait map[routineNum]*opData // this contains all read ops that have been unlocked until the last is unlocked
}

func (om *opMap) checkPending() {

	for _, pendingReq := range om.pending.read {
		om.logIfTimeout(pendingReq)
	}

	for _, pendingReq := range om.pending.rw {
		om.logIfTimeout(pendingReq)
	}

	// TODO: add a flag and a different timeout for logging active mutex timeouts
	for _, pendingReq := range om.active.read {
		om.logIfTimeout(pendingReq)
	}
	for _, pendingReq := range om.active.rw {
		om.logIfTimeout(pendingReq)
	}
}

func (om *opMap) logIfTimeout(pendingReq *opData) {
	now := time.Now()
	pendingTime := now.Sub(pendingReq.reqTime)
	if pendingTime >= Opts.DeadlockTimeout &&
		!pendingReq.alreadyLogged {
		om.logPending(pendingReq)
	}
}

func (om *opMap) logPending(pendingReq *opData) {
	pendingReq.alreadyLogged = true
	fmt.Println("\n=======================================================================")
	fmt.Println("Mutex request timeout for request:")
	logOpDetails(pendingReq)
	loggedOps := 0
	for _, op := range om.active.read {
		if op != pendingReq {
			fmt.Print("Active read lock: ")
			logOpDetails(op)
			loggedOps++
		}
	}

	for _, op := range om.active.rw {
		if op != pendingReq {
			fmt.Print("Active lock: ")
			logOpDetails(op)
			loggedOps++
		}
	}
	for _, op := range om.pending.read {
		if op.t == LOCK && op != pendingReq {
			// The documentation for sync's RWMutex.Lock says:
			// To ensure that the lock eventually becomes available, a blocked Lock call excludes new readers from acquiring the lock.
			fmt.Println("Pending RLock request (see documentation):")
			logOpDetails(op)
			loggedOps++
		}
	}

	for _, op := range om.doneRWait {
		fmt.Println("Additionally, the following read locks might still be held:")
		logOpDetails(op)
	}

	if loggedOps == 0 {
		println("Could not find other pending requests.")
	}
	fmt.Println("\n=======================================================================")
}

func (om *opMap) cleanup() bool {
	now := time.Now()

	for k, op := range om.done {
		elapsed := now.Sub(op.reqTime)
		if op.state == DONE && elapsed >= Opts.CleanTimeout {
			delete(om.done, k)
		}
	}
	if len(om.done) == 0 &&
		len(om.doneRWait) == 0 &&
		len(om.active.read) == 0 &&
		len(om.active.rw) == 0 &&
		len(om.pending.read) == 0 &&
		len(om.pending.rw) == 0 {
		return true
	}
	return false
}
