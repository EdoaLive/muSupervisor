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

func (d *opData) trackObtained() {
	d.state = ACTIVE
}
