package goroutine

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/background"
)

var GracefulShutdownTimeout = env.MustGetDuration("SRC_GRACEFUL_SHUTDOWN_TIMEOUT", 10*time.Second, "Graceful shutdown timeout")

// BackgroundRoutine represents a component of a binary that consists of a long
// running process with a graceful shutdown mechanism.
//
// See
// https://docs-legacy.sourcegraph.com/dev/background-information/backgroundroutine
// for more information and a step-by-step guide on how to implement a
// BackgroundRoutine.
type BackgroundRoutine = background.Routine

// WaitableBackgroundRoutine enhances BackgroundRoutine with a Wait method that
// blocks until the value's Start method has returned.
type WaitableBackgroundRoutine interface {
	BackgroundRoutine
	Wait()
}

// MonitorBackgroundRoutines will start the given background routines in their own
// goroutine. If the given context is canceled or a signal is received, the Stop
// method of each routine will be called. This method blocks until the Stop methods
// of each routine have returned. Two signals will cause the app to shutdown
// immediately.
var MonitorBackgroundRoutines = background.Monitor

// CombinedRoutine is a list of routines which are started and stopped in unison.
type CombinedRoutine = background.CombinedRoutine

// NoopRoutine return a background routine that does nothing for start or stop.
// If the name is empty, it will default to "noop".
var NoopRoutine = background.NoopRoutine
