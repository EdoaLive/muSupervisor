package muSupervisor

import (
	"container/list"
	"fmt"
	"runtime"
	"time"
)

type opType int
type opState int

//go:generate stringer -type=opType
const (
	LOCK opType = iota
	UNLOCK
	RLOCK
	RUNLOCK
)

//go:generate stringer -type=opState
const (
	PENDING opState = iota
	ACTIVE
	DONE
)

type mutexPointer interface{}
type routineNum uint64

type opData struct {
	t          opType
	numRoutine routineNum
	mutexPtr   *supervisedMutex
	reqTime    time.Time
	state      opState
	stackTrace *string

	alreadyLogged bool
}
type opMap struct {
	allOps        map[*opData]struct{}
	pendingQueue  *list.List
	pendingRQueue *list.List
}

type mutexOpMap map[mutexPointer]opMap

type mutexMap struct {
	mutexOpMap
}

var opReq chan *opData
var opObtained chan *opData
var unlockChan chan *opData

func init() {
	opReq = make(chan *opData)
	opObtained = make(chan *opData)
	unlockChan = make(chan *opData)

	go supervisor()
}

func supervisor() {
	if Opts.Disable == true {
		fmt.Println("supervisor disabled.")
		return
	}
	fmt.Println("supervisor enabled.")

	var mutexes mutexMap
	mutexes.mutexOpMap = make(mutexOpMap)

	t := time.NewTicker(Opts.CheckFrequency)

	for {
		select {
		case opD := <-opReq:
			opD.reqTime = time.Now()
			mapRequested(opD, &mutexes)

		case opD := <-opObtained:
			mapObtained(opD, &mutexes)

		case opD := <-unlockChan:
			mapUnlock(opD, &mutexes)

		case <-t.C:
			doCheck(&mutexes)
		}

	}
}

func doCheck(m *mutexMap) {
	for _, om := range m.mutexOpMap {
		checkPending(om.pendingQueue, om.allOps)
		checkPending(om.pendingRQueue, om.allOps)

		cleanOld(om.allOps)
	}

}

func cleanOld(ops map[*opData]struct{}) {
	now := time.Now()

	for op := range ops {
		elapsed := now.Sub(op.reqTime)
		if op.state == DONE && elapsed >= Opts.CleanTimeout {
			delete(ops, op)
		}
	}
}

func checkPending(l *list.List, ops map[*opData]struct{}) {
	now := time.Now()

	for e := l.Front(); e != nil; e = e.Next() {
		pendingReq := e.Value.(*opData)
		pendingTime := now.Sub(pendingReq.reqTime)
		if pendingReq.state == PENDING && pendingTime >= Opts.DeadlockTimeout && !pendingReq.alreadyLogged {
			pendingReq.alreadyLogged = true
			fmt.Println("\n=======================================================================")
			fmt.Println("Mutex request timeout for request:")
			logOpDetails(pendingReq)
			loggedOps := 0
			for op := range ops {
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

func logOpDetails(op *opData) {
	now := time.Now()
	fmt.Printf("\tGoRoutine: %v, state: %v, type: %v, time: %v, request: %p\n",
		op.numRoutine,
		op.state,
		op.t,
		now.Sub(op.reqTime),
		op,
	)
	if op.stackTrace != nil {
		fmt.Printf("%s\n\n", *op.stackTrace)
	}
}

// When the lock is requested, add it to pending map
func mapRequested(d *opData, m *mutexMap) {

	if _, exists := m.mutexOpMap[d.mutexPtr]; !exists {
		m.mutexOpMap[d.mutexPtr] =
			opMap{
				make(map[*opData]struct{}),
				list.New(),
				list.New(),
			}
	}

	d.state = PENDING

	m.mutexOpMap[d.mutexPtr].allOps[d] = struct{}{}

	switch d.t {
	case LOCK:
		m.mutexOpMap[d.mutexPtr].pendingQueue.PushFront(d)
	case RLOCK:
		m.mutexOpMap[d.mutexPtr].pendingRQueue.PushFront(d)
	}
}

// When is obtained, update routine status accordingly and remove pending
func mapObtained(d *opData, _ *mutexMap) {
	switch d.t {
	case LOCK:
		d.state = ACTIVE

	case RLOCK:
		d.state = ACTIVE
	}
}

func mapUnlock(d *opData, m *mutexMap) {
	var e *list.Element
	switch d.t {
	case UNLOCK:
		e = m.mutexOpMap[d.mutexPtr].pendingQueue.Front()
		m.mutexOpMap[d.mutexPtr].pendingQueue.Remove(e)

	case RUNLOCK:
		e = m.mutexOpMap[d.mutexPtr].pendingRQueue.Front()
		m.mutexOpMap[d.mutexPtr].pendingRQueue.Remove(e)
	}

	if e != nil {
		e.Value.(*opData).state = DONE
		e.Value.(*opData).stackTrace = nil
	}
}
