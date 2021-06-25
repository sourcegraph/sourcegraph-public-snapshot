// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package output

import (
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCapabilityWatcher(t *testing.T) {
	// Let's set up two capability watcher channels and ensure they both get
	// triggered on a single SIGWINCH and that they receive the same value.
	var (
		capA        capabilities
		capB        capabilities
		invocations int32
		wg          sync.WaitGroup
	)

	createWatcher := func(out *capabilities) {
		c := newCapabilityWatcher()
		if c == nil {
			t.Error("unexpected nil watcher channel")
		}

		wg.Add(1)
		go func() {
			*out = <-c
			atomic.AddInt32(&invocations, 1)
			wg.Done()
		}()
	}
	createWatcher(&capA)
	createWatcher(&capB)

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.Signal(syscall.SIGWINCH); err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	if invocations != 2 {
		t.Errorf("unexpected number of invocations: have=%d want=2", invocations)
	}
	if diff := cmp.Diff(capA, capB); diff != "" {
		t.Errorf("unexpected difference between capabilities:\n%s", diff)
	}
}
