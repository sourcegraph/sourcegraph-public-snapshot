package memcmd

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/bytesize"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Observer is an interface for observing and tracking the memory usage of a process.
//
// Implementations of this interface should provide methods to start and stop the observation,
// as well as retrieve the maximum memory usage of the observed process.
//
// Callers must call Stop when they are done with the observer to release any associated resources.
type Observer interface {
	// Start starts the observer. It should be called before any other method.
	//
	// After Start is called, callers must call Stop when they are done with the
	// observer to release any resources.
	//
	// Calling Start() multiple times is safe and has no effect after the first invocation.
	Start()

	// Stop stops the observer and releases any associated resources. For accurate measurement,
	// Stop must be called _after_ Wait has been called on the *exec.Cmd.
	// Stop stops the observer and releases any associated resources.
	//
	// Calling Stop() multiple times is safe and has no effect after the first invocation.
	Stop()

	// MaxMemoryUsage returns the maximum memory usage in bytes of the process since
	// the observer was started.
	//
	// Calling this method will also stop the observer
	//
	// It is only valid to call this method after:
	// 1) Start() has been called and
	// 2) the underlying process has stopped.
	//
	// See the individual observer implementations for more details on how memory
	// usage is calculated.
	MaxMemoryUsage() (bytes bytesize.Size, err error)
}

type noopObserver struct {
	startOnce sync.Once
	started   chan struct{}

	stopOnce sync.Once
	stopped  chan struct{}
}

func (o *noopObserver) Start() {
	o.startOnce.Do(func() {
		close(o.started)
	})
}

func (o *noopObserver) Stop() {
	o.stopOnce.Do(func() {
		close(o.stopped)
	})
}

func (o *noopObserver) MaxMemoryUsage() (bytesize.Size, error) {
	select {
	case <-o.started:
	default:
		return 0, errObserverNotStarted
	}

	o.Stop()

	return 0, nil
}

// NewNoOpObserver returns an observer that does nothing. It is useful for
// testing or when you want to disable memory usage tracking.
func NewNoOpObserver() Observer {
	return &noopObserver{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

var _ Observer = &noopObserver{}

var errProcessNotStopped = errors.New("command has not stopped yet")
var errObserverNotStarted = errors.New("observer has not started yet")
