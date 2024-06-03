package interrupt

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

const (
	InterruptGeneral = iota
	InterruptCleanup
)

var (
	hooksInit sync.Once
	hooks     map[int][]func()
	mux       sync.Mutex
	closed    chan struct{}
)

// Register adds a hook to be executed before program exit. The most recently added hooks
// are called first.
func Register(hook func()) {
	register(hook, InterruptGeneral)
}

func register(hook func(), interrupt int) {
	mux.Lock()

	hooksInit.Do(func() {
		hooks = map[int][]func(){}
	})

	var hookValues []func()
	if v, ok := hooks[interrupt]; ok {
		hookValues = append(v, hook)
	} else {
		hookValues = []func(){hook}
	}
	hooks[interrupt] = hookValues
	mux.Unlock()
}

func RegisterCleanup(hook func()) {
	register(hook, InterruptCleanup)
}

// Listen starts a goroutine that listens for interrupts and executes registered hooks
// before exiting with status 1.
func Listen() {
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		<-interrupt
		std.Out.WriteWarningf("Interrupt received, executing hook groups for graceful shutdown...")
		closed = make(chan struct{})

		// prevent additional hooks from registering once we've received an interrupt
		mux.Lock()

		go func() {
			// If we receive a second interrupt, forcibly exit.
			<-interrupt
			os.Exit(1)
		}()

		// Concurrently execute all hooks
		var wg sync.WaitGroup
		std.Out.WriteWarningf("Executing %d 'cleanup' hooks for graceful shutdown...", len(hooks[InterruptCleanup]))
		for i, h := range hooks[InterruptCleanup] {
			wg.Add(1)
			fn := h
			idx := i
			go func() {
				println("+", idx)
				fn()
				println("-", idx)
				wg.Done()
			}()
		}
		wg.Wait()
		println("--------")

		std.Out.WriteWarningf("Executing %d 'general' hooks for for graceful shutdown...", len(hooks[InterruptGeneral]))
		for _, h := range hooks[InterruptGeneral] {
			fn := h
			fn()
		}
		close(closed)

	}()
}

func Wait() {
	if closed != nil {
		<-closed
	}
}
