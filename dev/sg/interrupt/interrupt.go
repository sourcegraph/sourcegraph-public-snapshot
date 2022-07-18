package interrupt

import (
	"os"
	"os/signal"
	"syscall"
)

var hooks []func()

// Register adds a hook to be executed before program exit. The most recently added hooks
// are added first.
func Register(hook func()) {
	hooks = append([]func(){hook}, hooks...)
}

// Listen starts a goroutine that listens for interrupts and executes registered hooks
// before exiting with status 1.
func Listen() {
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-interrupt
		go func() {
			// If we receive a second interrupt, forcibly exit.
			<-interrupt
			os.Exit(1)
		}()

		for _, h := range hooks {
			h()
		}
		os.Exit(1)
	}()
}
