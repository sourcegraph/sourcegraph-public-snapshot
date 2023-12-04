package shared

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

// gRPCWebUIDebugEndpoints returns debug points that serve the GRPCWebUI instances that target
// this frontend instance.
func gRPCWebUIDebugEndpoints() []debugserver.Endpoint {
	addr := cli.GetInternalAddr()
	return []debugserver.Endpoint{
		debugserver.NewGRPCWebUIEndpoint("frontend-internal", addr),
	}
}

func CreateDebugServerEndpoints() []debugserver.Endpoint {
	return append(
		gRPCWebUIDebugEndpoints(),
		debugserver.Endpoint{
			Name: "Rate Limiter State",
			Path: "/rate-limiter-state",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				info, err := ratelimit.GetGlobalLimiterState(r.Context())
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to read rate limiter state: %q", err.Error()), http.StatusInternalServerError)
					return
				}
				resp, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to marshal rate limiter state: %q", err.Error()), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(resp)
			}),
		},
	)
}
