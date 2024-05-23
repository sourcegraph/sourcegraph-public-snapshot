package interrupt

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

var (
	hooks  []func()
	mux    sync.Mutex
	closed chan struct{}
)

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
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		<-interrupt
		std.Out.WriteWarningf("Interrupt received, executing %d hooks for graceful shutdown...", len(hooks))
		closed = make(chan struct{})

		// prevent additional hooks from registering once we've received an interrupt
		mux.Lock()

		go func() {
			// If we receive a second interrupt, forcibly exit.
			<-interrupt
			os.Exit(1)
		}()

		// Execute all hooks
		for _, h := range hooks {
			h()
		}
		close(closed)
	}()
}

func Wait() {
	if closed != nil {
		<-closed
	}
}
