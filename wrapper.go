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
	// Waiting for a lock for longer than DeadlockTimeout is considered a deadlock.
	DeadlockTimeout time.Duration
	// The frequency at which timeout checks are performed
	CheckFrequency time.Duration
	// After this time already satisfied requests will be cleaned up
	CleanTimeout time.Duration
}{
	DeadlockTimeout: time.Second * 5,
	CheckFrequency:  time.Second * 1,
	CleanTimeout:    time.Second * 20,
}

type Mutex struct {
	mu sync.Mutex
}

type RWMutex struct {
	mu sync.RWMutex
}

func (m *Mutex) Lock() {
	mutexOp(LOCK, m, m.mu.Lock)
}
func (m *Mutex) Unlock() {
	mutexOp(UNLOCK, m, m.mu.Unlock)
}

func (m *RWMutex) Lock() {
	mutexOp(LOCK, m, m.mu.Lock)
}
func (m *RWMutex) Unlock() {
	mutexOp(UNLOCK, m, m.mu.Unlock)
}
func (m *RWMutex) RLock() {
	mutexOp(RLOCK, m, m.mu.RLock)
}
func (m *RWMutex) RUnlock() {
	mutexOp(RUNLOCK, m, m.mu.RUnlock)
}

func mutexOp(t opType, m interface{}, f func()) {
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
