package interrupt

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var hooks []func()
var mux sync.Mutex

// Register adds a hook to be executed before program exit. The most recently added hooks
// are called first.
func Register(hook func()) {
	mux.Lock()
	hooks = append([]func(){hook}, hooks...)
	mux.Unlock()
}

// Listen starts a goroutine that listens for interrupts and executes registered hooks
// before exiting with status 1.
func Listen() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		mux.Lock()
		for _, h := range hooks {
			h()
		}
		os.Exit(1)
	}()
}
