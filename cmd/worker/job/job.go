package job

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Job creates configuration struct and background routine instances to be run
// as part of the worker process.
type Job interface {
	// Description renders a brief overview of what this job does and handles.
	Description() string

	// Config returns a set of configuration struct pointers that should be loaded
	// and validated as part of application startup.
	//
	// If called multiple times, the same pointers should be returned.
	//
	// Note that the Load function of every config object is invoked even if the
	// job is not enabled. It is assumed safe to call this method with an invalid
	// configuration (and all configuration errors should be surfaced via Validate).
	Config() []env.Config

	// Routines constructs and returns the set of background routines that
	// should run as part of the worker process. Service initialization should
	// be shared between setup hooks when possible (e.g. sync.Once initialization).
	//
	// Note that the given context is meant to be used _only_ for setup. A context
	// passed to a periodic routine should be a fresh context unattached to this,
	// as the argument to this function will be canceled after all Routine invocations
	// have exited after application startup.
	Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error)
}
