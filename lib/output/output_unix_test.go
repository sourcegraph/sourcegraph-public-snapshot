//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package output

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestCapabilityWatcher(t *testing.T) {
	// Let's set up two capability watcher channels and ensure they both get
	// triggered on a single SIGWINCH and that they receive the same value.
	//
	// We'll have them both send the capabilities they receive into this channel.
	received := make(chan capabilities)

	createWatcher := func(opts OutputOpts) {
		c := newCapabilityWatcher(opts)
		if c == nil {
			t.Error("unexpected nil watcher channel")
		}

		go func() {
			// We only want to receive one capabilities struct on the channel;
			// if we get more and the test hasn't terminated, that means that
			// the capabilities aren't being fanned out correctly to each
			// watcher.
			caps := <-c
			received <- caps
		}()
	}
	createWatcher(OutputOpts{})
	createWatcher(OutputOpts{})

	// Now we set up the main test. To be able to raise signals on the current
	// process, we need the current process.
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	// We need to track the capabilities we've seen, since we expect both
	// watchers to receive a capabilities struct.
	seen := []capabilities{}

	// We're going to raise the signal on a ticker. The reason for this is that
	// signal handler installation is asynchronous: Go starts a goroutine the
	// first time a signal handler is installed, and there's no guarantee that
	// the goroutine has even installed the OS-level signal handler at the point
	// execution returns from signal.Notify(). The quickest, dirtiest solution
	// is therefore to keep raising SIGWINCH until it's handled, which we can do
	// with a ticker.
	ticker := time.NewTicker(3 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		// Raise SIGWINCH when the ticker ticks.
		case <-ticker.C:
			if err := proc.Signal(syscall.SIGWINCH); err != nil {
				t.Fatal(err)
			}

		// Handle the capabilities we see in the watchers, and test the results
		// once we have capabilities from both watchers and terminate.
		case caps := <-received:
			seen = append(seen, caps)
			if len(seen) > 2 {
				t.Fatalf("too many capabilities")
			} else if len(seen) == 2 {
				if diff := cmp.Diff(seen[0], seen[1]); diff != "" {
					t.Errorf("unexpected difference between capabilities:\n%s", diff)
				}
				return
			}
		}
	}
}
