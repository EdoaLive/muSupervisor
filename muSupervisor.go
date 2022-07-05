package muSupervisor

import "time"

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
	CleanTimeout:    time.Second * 5,
}
