pbckbge shbred

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
)

// GRPCWebUIDebugEndpoint returns b debug endpoint thbt serves the GRPCWebUI thbt tbrgets
// this gitserver instbnce.
func GRPCWebUIDebugEndpoint(bddr string) debugserver.Endpoint {
	return debugserver.NewGRPCWebUIEndpoint("gitserver", bddr)
}
