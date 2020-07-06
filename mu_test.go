package muSupervisor

import (
	"fmt"
	"testing"
	"time"
)

func Test_mu(t *testing.T) {

	var mu1, mu2 Mutex

	go testLock(&mu1)
	time.Sleep(1 * time.Second)
	mu1.Lock()

	mu2.Lock()
	mu2.Unlock()

	mu1.Unlock()

	mu1.Lock()
	fmt.Printf("Done.\n")
	mu1.Unlock()

}

func testLock(mu1 *Mutex) {
	mu1.Lock()
	time.Sleep(6 * time.Second)
	mu1.Unlock()
}
