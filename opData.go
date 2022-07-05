package muSupervisor

import "time"

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
	DONERWAIT // Read unlocked but waiting for all other read locks to be gone
)

type routineNum uint64

//TODO: split opData with reqData. See the TODO in supervisedMutex.go
// The distinction is
//	Operation: a cycle of LOCK, OBTAIN, UNLOCK operations for a mutex
//  Request: the single LOCK/UNLOCK request

// opData represents the specific lock/unlock operation for a request.
type opData struct {
	t          opType
	numRoutine routineNum
	mutexPtr   *supervisedMutex
	reqTime    time.Time
	state      opState
	stackTrace *string

	alreadyLogged bool
}

// reqData represents one of the requests for a mutex (which may have more than one op)
type reqData struct {
}
