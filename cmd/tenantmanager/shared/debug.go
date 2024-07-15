package shared

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
)

func createDebugServerEndpoints(ready chan struct{}, addr string, debugserverEndpoints *LazyDebugserverEndpoint) []debugserver.Endpoint {
	return []debugserver.Endpoint{
		debugserver.NewGRPCWebUIEndpoint("tenantmanager", addr),
		{
			Name: "Tenant list",
			Path: "/tenants",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// wait until we're healthy to respond
				<-ready
				// tenantListEndpoint is guaranteed to be assigned now
				debugserverEndpoints.tenantListEndpoint(w, r)
			}),
		},
	}
}
