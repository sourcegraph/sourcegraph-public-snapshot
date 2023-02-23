// Package service defines a service that runs as part of the Sourcegraph application. Examples
// include frontend, gitserver, and repo-updater.
package service

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// A Service provides independent functionality in the Sourcegraph application. Examples include
// frontend, gitserver, and repo-updater. A service may run in the same process as any other
// service, in a separate process, in a separate container, or on a separate host.
type Service interface {
	// Name is the name of the service.
	Name() string

	// Configure reads from env vars, runs very quickly, and has no side effects. All services'
	// Configure methods are run before any service's Start method.
	//
	// The returned env.Config will be passed to the service's Start method.
	//
	// The returned debugserver endpoints will be added to the global debugserver.
	Configure() (env.Config, []debugserver.Endpoint)

	// Start starts the service.
	//
	// When start returns or ready is called the service will be marked as
	// ready.
	//
	// TODO(sqs): TODO(single-binary): make it monitorable with goroutine.Whatever interfaces.
	Start(ctx context.Context, observationCtx *observation.Context, ready ReadyFunc, c env.Config) error
}

// ReadyFunc is called in (Service).Start to signal that the service is ready
// to serve clients, even if Start has not returned. It is optional to call
// ready, on Start returning the service will be marked as ready. It is safe
// to call ready multiple times.
type ReadyFunc func()
