package muSupervisor

import (
	"container/list"
	"fmt"
	"runtime"
	"time"
)

// opMap keep track of all operations of a single mutex instance
type opMap struct {
	allOps        map[*opData]struct{}
	pendingQueue  *list.List
	pendingRQueue *list.List
}

func (om *opMap) checkPending() {
	now := time.Now()

	for _, l := range []*list.List{om.pendingQueue, om.pendingRQueue} {
		for e := l.Front(); e != nil; e = e.Next() {
			pendingReq := e.Value.(*opData)
			pendingTime := now.Sub(pendingReq.reqTime)
			if pendingReq.state == PENDING && pendingTime >= Opts.DeadlockTimeout && !pendingReq.alreadyLogged {
				pendingReq.alreadyLogged = true
				fmt.Println("\n=======================================================================")
				fmt.Println("Mutex request timeout for request:")
				logOpDetails(pendingReq)
				loggedOps := 0
				for op := range om.allOps {
					if op.state == ACTIVE && op != pendingReq {
						fmt.Println("Active mutex: ")
						logOpDetails(op)
						loggedOps++
					} else if pendingReq.t == RLOCK && op.t == LOCK && op.state == PENDING && op != pendingReq {
						// The documentation for sync's RWMutex.Lock says:
						// To ensure that the lock eventually becomes available, a blocked Lock call excludes new readers from acquiring the lock.
						fmt.Println("Pending RLock request (see documentation):")
						logOpDetails(op)
						loggedOps++
					}
				}
				if loggedOps == 0 {
					println("Could not find other pending requests. Calling the debugger.")
					runtime.Breakpoint()
				}
				fmt.Println("\n=======================================================================")
			}
		}
	}
}

func (om *opMap) cleanup() {
	now := time.Now()

	for op := range om.allOps {
		elapsed := now.Sub(op.reqTime)
		if op.state == DONE && elapsed >= Opts.CleanTimeout {
			delete(om.allOps, op)
		}
	}
}
