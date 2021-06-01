package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// SetupHook creates configuration struct and background routine instances
// to be run as part of the worker process.
type SetupHook interface {
	// Config returns a set of configuration struct pointers that should
	// be loaded and validated as part of application startup.
	Config() []env.Config

	// Routines constructs and returns the set of background routines that
	// should run as part of the worker process. Service initialization should
	// be shared between setup hooks when possible (e.g. sync.Once initialization).
	Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error)
}

var setupHooks = map[string]SetupHook{
	// Empty for now
}
