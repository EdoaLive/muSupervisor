package muSupervisor

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Opts = struct {
	// Disable supervisor and most of the wrapping to reduce overhead (eg. in production)
	Disable bool
	// Waiting for a lock for longer than DeadlockTimeout is considered a deadlock.
	DeadlockTimeout time.Duration
	// The frequency at which timeout checks are performed
	CheckFrequency time.Duration
	// After this time already satisfied requests will be cleaned up
	CleanTimeout time.Duration
}{
	Disable:         false,
	DeadlockTimeout: time.Second * 5,
	CheckFrequency:  time.Second * 1,
	CleanTimeout:    time.Second * 20,
}

type supervisedMutex struct{}

type Mutex struct {
	supervisedMutex
	mu sync.Mutex
}

type RWMutex struct {
	supervisedMutex
	mu sync.RWMutex
}

func (m *Mutex) Lock() {
	m.mutexOp(LOCK, m.mu.Lock)
}
func (m *Mutex) Unlock() {
	m.mutexOp(UNLOCK, m.mu.Unlock)
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

func (m *supervisedMutex) mutexOp(t opType, f func()) {

	if Opts.Disable == true {
		f()
		return
	}

	op := opData{
		t:          t,
		numRoutine: routineNum(getGID()),
		mutexPtr:   m,
	}

	switch t {
	case LOCK, RLOCK:

		st := getStackTrace()
		op.stackTrace = &st

		opReq <- &op
		f()
		opObtained <- &op

	case UNLOCK, RUNLOCK:
		f()
		unlockChan <- &op
	}
}

// Taken from https://blog.sgmansfield.com/2015/12/goroutine-ids/
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// getStackTraceClassic is the original function used to get the stack trace.
// I'll leave it here for reference/testing
func getStackTraceClassic() string {
	st := make([]byte, 2048)
	num := runtime.Stack(st, false)
	rst := st[0:num]
	return string(rst)
}

func getStackTrace() string {
	const maxStackLength = 50
	stackBuf := make([]uintptr, maxStackLength)
	length := runtime.Callers(4, stackBuf[:])
	stack := stackBuf[:length]

	trace := ""
	frames := runtime.CallersFrames(stack)
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			trace = trace + fmt.Sprintf("\n%s\n\t%s:%d", frame.Function, frame.File, frame.Line)
		}
		if !more {
			break
		}
	}
	return trace
}
