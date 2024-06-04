package interrupt

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

const (
	// MaxInterruptCount is the maximum number of interrupts we will handle "catch" before exiting immediately.
	MaxInterruptCount = 5
	// InterruptSequential is a value for the interrupt type that indicates the hook should be executed sequentially.
	InterruptSequential = iota
	// InterruptConcurrent is a value for the interrupt type that indicates the hook should be executed concurrently.
	InterruptConcurrent
)

var (
	hookTimeout = 2 * time.Second
	hooksInit   sync.Once
	hooks       map[int][]func()
	mux         sync.Mutex
	closed      chan struct{}
)

// Register adds a hook to be executed before program exit. The most recently added hooks
// are called first.
// The hook added is executed sequentially. If you want the hook to be executed concurrently
// see RegisterConcurrent.
func Register(hook func()) {
	register(hook, InterruptSequential)
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

// RegisterConcurrent adds a hook to be executed concurrently before program exit.
//
// The hook added is executed concurrently. If you want the hook to be executed sequentially
// see Register.
func RegisterConcurrent(hook func()) {
	register(hook, InterruptConcurrent)
}

func doWithContext(ctx context.Context, fn func()) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
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
			// Count the interrupts and exit after 5
			count := 0
			for count < 5 {
				select {
				case <-interrupt:
					count++
				case <-closed:
					return
				}
			}
			std.Out.WriteWarningf("Ok. Loads of interrupts received - exiting immediately.")
			os.Exit(1)
		}()

		// Execute all registered hooks
		ctx, cancel := context.WithTimeout(context.Background(), hookTimeout)
		err := doWithContext(ctx, func() {
			executeHooks()
		})
		cancel()
		if err != nil {
			std.Out.WriteWarningf("context failure executing hooks: %s", err)
		}
		close(closed)

	}()
}

// executeHooks executes all registered hooks using different strategies according to the group of the hook.
//
// Cleanup hooks are executed concurrently.
// General hooks are executed sequentially.
func executeHooks() {
	var wg sync.WaitGroup
	std.Out.WriteWarningf("Executing %d 'cleanup' hooks for graceful shutdown...", len(hooks[InterruptConcurrent]))
	for _, h := range hooks[InterruptConcurrent] {
		wg.Add(1)
		fn := h
		go func() {
			defer wg.Done()
			fn()
		}()
	}
	wg.Wait()

	std.Out.WriteWarningf("Executing %d 'general' hooks for for graceful shutdown...", len(hooks[InterruptSequential]))
	for _, h := range hooks[InterruptSequential] {
		fn := h
		fn()
	}
}

func Wait() {
	if closed != nil {
		<-closed
	}
}
