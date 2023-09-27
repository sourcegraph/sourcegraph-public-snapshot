pbckbge shbred

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
)

// GRPCWebUIDebugEndpoint returns b debug endpoint thbt serves the GRPCWebUI thbt tbrgets
// this symbols instbnce.
func GRPCWebUIDebugEndpoint() debugserver.Endpoint {
	return debugserver.NewGRPCWebUIEndpoint("symbols", bddr)
}
