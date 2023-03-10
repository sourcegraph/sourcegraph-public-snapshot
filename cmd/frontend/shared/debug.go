package shared

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
)

// GRPCWebUIDebugEndpoints returns debug points that serve the GRPCWebUI instances that target
// this frontend instance.
func GRPCWebUIDebugEndpoints() []debugserver.Endpoint {
	addr := cli.GetInternalAddr()
	return []debugserver.Endpoint{
		debugserver.NewGRPCWebUIEndpoint("frontend-internal", addr),
	}
}
