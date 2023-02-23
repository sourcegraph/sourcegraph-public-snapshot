package command

import (
	"context"
)

// Runner is the interface between an executor and the host on which commands
// are invoked. Having this interface at this level allows us to use the same
// code paths for local development (via shell + docker) as well as production
// usage (via Firecracker).
type Runner interface {
	// Setup prepares the runner to invoke a series of commands.
	Setup(ctx context.Context) error

	// Teardown disposes of any resources created in Setup.
	Teardown(ctx context.Context) error

	// Run invokes a command in the runner context.
	Run(ctx context.Context) error
}
