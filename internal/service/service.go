// Package service defines a service that runs as part of the Sourcegraph application. Examples
// include frontend, gitserver, and repo-updater.
package service

import (
	"context"

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
	Configure() env.Config

	// Start starts the service.
	// TODO(sqs): make it monitorable with goroutine.Whatever interfaces.
	Start(context.Context, *observation.Context, env.Config) error
}
