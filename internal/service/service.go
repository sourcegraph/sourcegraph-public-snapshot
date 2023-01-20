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
	// TODO(sqs): TODO(single-binary): make it monitorable with goroutine.Whatever interfaces.
	Start(context.Context, *observation.Context, ReadyFunc, env.Config) error
}

// ReadyFunc is called in (Service).Start when the service is ready to start serving clients.
type ReadyFunc func()
