# Mutex Supervisor
Have you ever been stuck on a mutex deadlock without knowing what caused it? Well, I have.

This simple tool can wrap any sync.Mutex or sync.RWmutex with the same behavior, but it will verbosely log if any lock request doesn't get satisfied in a specified amount of time. The output log will contain stack trace of the timed-out request as well as the stack trace of any other active or pending RLock request on that mutex.

# How to use
If you are just using sync.Mutex and sync.RWMutex, you can import this module instead of the sync module (aliasing it):
```
import sync "github.com/EdoaLive/muSupervisor"
``` 
If you prefer to not replace the sync module or use both differently, you can use it separately:
```
import "github.com/EdoaLive/muSupervisor"

var (
	mutex1 muSupervisor.Mutex
	mutex2 muSupervisor.RWMutex
)
``` 

Please, please, please. Use this tool only in debugging and testing. Not in production.

# TODO
* Write a more exhaustive test (current is merely an example)
* Fix a condition in which the stack trace is not saved
* Improve comments and documentation