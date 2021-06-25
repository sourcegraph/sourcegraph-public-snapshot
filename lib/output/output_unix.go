// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package output

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func init() {
	chans := []chan capabilities{}
	var mu sync.RWMutex

	// On these platforms, we have the SIGWINCH signal available, which
	// indicates that the terminal has been resized. When received, we can use
	// this to re-detect the terminal capabilities.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGWINCH)
		for {
			<-c
			caps, err := detectCapabilities()
			// We won't bother reporting an error here; there's no harm in the
			// previous capabilities being used besides possibly being ugly.
			if err == nil {
				mu.RLock()
				for _, out := range chans {
					go func(caps capabilities, out chan capabilities) {
						out <- caps
					}(caps, out)
				}
				mu.RUnlock()
			}
		}
	}()

	newCapabilityWatcher = func() chan capabilities {
		out := make(chan capabilities)
		mu.Lock()
		defer mu.Unlock()
		chans = append(chans, out)

		return out
	}
}
