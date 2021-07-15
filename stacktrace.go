package muSupervisor

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

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
