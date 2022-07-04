package main

import (
	sync "github.com/EdoaLive/muSupervisor"
	"math/rand"
	sync2 "sync"
	"time"
)

var wg sync2.WaitGroup

func main() {
	sync.Opts.DeadlockTimeout = 500 * time.Millisecond
	sync.Opts.CheckFrequency = 200 * time.Millisecond

	var mu1 sync.RWMutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go testLock(&mu1)
		time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	}

	wg.Wait()
}

func testLock(mu1 *sync.RWMutex) {
	mu1.Lock()
	time.Sleep(time.Duration(rand.Intn(400)) * time.Millisecond)
	mu1.Unlock()
	wg.Done()
}
