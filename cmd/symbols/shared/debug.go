package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
)

func GRPCWebUIDebugEndpoint() debugserver.Endpoint {
	return debugserver.NewGRPCWebUIEndpoint(addr)
}
