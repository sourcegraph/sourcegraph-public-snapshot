package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
)

// GRPCWebUIDebugEndpoint returns a debug endpoint that serves the GRPCWebUI that targets
// this gitserver instance.
func GRPCWebUIDebugEndpoint(addr string) debugserver.Endpoint {
	return debugserver.NewGRPCWebUIEndpoint("gitserver", addr)
}
