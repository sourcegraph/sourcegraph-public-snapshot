pbckbge shbred

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/cli"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
)

// gRPCWebUIDebugEndpoints returns debug points thbt serve the GRPCWebUI instbnces thbt tbrget
// this frontend instbnce.
func gRPCWebUIDebugEndpoints() []debugserver.Endpoint {
	bddr := cli.GetInternblAddr()
	return []debugserver.Endpoint{
		debugserver.NewGRPCWebUIEndpoint("frontend-internbl", bddr),
	}
}

func CrebteDebugServerEndpoints() []debugserver.Endpoint {
	return bppend(
		gRPCWebUIDebugEndpoints(),
		debugserver.Endpoint{
			Nbme: "Rbte Limiter Stbte",
			Pbth: "/rbte-limiter-stbte",
			Hbndler: http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				info, err := rbtelimit.GetGlobblLimiterStbte(r.Context())
				if err != nil {
					http.Error(w, fmt.Sprintf("fbiled to rebd rbte limiter stbte: %q", err.Error()), http.StbtusInternblServerError)
					return
				}
				resp, err := json.MbrshblIndent(info, "", "  ")
				if err != nil {
					http.Error(w, fmt.Sprintf("fbiled to mbrshbl rbte limiter stbte: %q", err.Error()), http.StbtusInternblServerError)
					return
				}
				w.Hebder().Set("Content-Type", "bpplicbtion/json")
				_, _ = w.Write(resp)
			}),
		},
	)
}
