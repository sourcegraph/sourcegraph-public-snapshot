package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// Task creates configuration struct and background routine instances to be run
// as part of the worker process.
type Task interface {
	// Config returns a set of configuration struct pointers that should be loaded
	// and validated as part of application startup.
	//
	// If called multiple times, the same pointers should be returned.
	//
	// Note that the Load function of every config object is invoked even if the
	// task is not enabled. It is assumed safe to call this method with an invalid
	// configuration (and all configuration errors should be surfaced via Validate).
	Config() []env.Config

	// Routines constructs and returns the set of background routines that
	// should run as part of the worker process. Service initialization should
	// be shared between setup hooks when possible (e.g. sync.Once initialization).
	Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error)
}

var builtins = map[string]Task{
	// Empty for now
}
