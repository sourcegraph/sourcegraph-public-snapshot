pbckbge shbred

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
)

// GRPCWebUIDebugEndpoint returns b debug endpoint thbt serves the GRPCWebUI thbt tbrgets
// this sebrcher instbnce.
func GRPCWebUIDebugEndpoint() debugserver.Endpoint {
	bddr := getAddr()
	return debugserver.NewGRPCWebUIEndpoint("sebrcher", bddr)
}
