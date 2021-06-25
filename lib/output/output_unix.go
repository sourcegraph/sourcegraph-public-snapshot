// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package output

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func init() {
	// The platforms this file builds on support the SIGWINCH signal, which
	// indicates that the terminal has been resized. When we receive that
	// signal, we can use this to re-detect the terminal capabilities.
	//
	// We won't do any setup until the first time newCapabilityWatcher is
	// invoked, but we do need some shared state to be ready.
	var (
		// chans contains the listening channels that should be notified when
		// capabilities are updated.
		chans []chan capabilities

		// mu guards the chans variable.
		mu sync.RWMutex

		// once guards the lazy initialisation, including installing the signal
		// handler.
		once sync.Once
	)

	newCapabilityWatcher = func() chan capabilities {
		// Lazily initialise the required global state if we haven't already.
		once.Do(func() {
			chans = make([]chan capabilities, 0, 1)

			// Install the signal handler. To avoid race conditions, we should
			// do this synchronously before spawning the goroutine that will
			// actually listen to the channel.
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)

			go func() {
				for {
					<-c
					caps, err := detectCapabilities()
					// We won't bother reporting an error here; there's no harm
					// in the previous capabilities being used besides possibly
					// being ugly.
					if err == nil {
						mu.RLock()
						for _, out := range chans {
							// Technically, if the listener of this channel is
							// no longer listening, sending caps to out will
							// hang forever. To mitigate this, we'll run in a
							// goroutine.
							//
							// Practically, there's only ever one Output
							// instance in all our uses of this package, and it
							// lives for the lifetime of the process. So the
							// potential for a goroutine leak here is
							// practically zero.
							go func(caps capabilities, out chan capabilities) {
								out <- caps
							}(caps, out)
						}
						mu.RUnlock()
					}
				}
			}()
		})

		// Now we can create and return the actual output channel.
		out := make(chan capabilities)
		mu.Lock()
		defer mu.Unlock()
		chans = append(chans, out)

		return out
	}
}
