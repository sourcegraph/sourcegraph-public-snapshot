package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
)

// GRPCWebUIDebugEndpoint returns a debug endpoint that serves the GRPCWebUI that targets
// this searcher instance.
func GRPCWebUIDebugEndpoint(addr string) debugserver.Endpoint {
	return debugserver.NewGRPCWebUIEndpoint("searcher", addr)
}
