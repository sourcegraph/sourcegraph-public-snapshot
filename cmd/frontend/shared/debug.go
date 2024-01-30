package shared

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
)

func createDebugServerEndpoints(ready chan struct{}, debugserverEndpoints *cli.LazyDebugserverEndpoint) []debugserver.Endpoint {
	return []debugserver.Endpoint{
		debugserver.NewGRPCWebUIEndpoint("frontend-internal", cli.GetInternalAddr()),
		debugserver.Endpoint{
			Name: "Rate Limiter State",
			Path: "/rate-limiter-state",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// wait until we're healthy to respond
				<-ready
				// globalRateLimiterState is guaranteed to be assigned now
				debugserverEndpoints.GlobalRateLimiterState.ServeHTTP(w, r)
			}),
		},
		{
			Name: "Gitserver Repo Status",
			Path: "/gitserver-repo-status",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// wait until we're healthy to respond
				<-ready
				// gitserverReposStatusEndpoint is guaranteed to be assigned now
				debugserverEndpoints.GitserverReposStatusEndpoint.ServeHTTP(w, r)
			}),
		},
		{
			Name: "List Authz Providers",
			Path: "/list-authz-providers",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// wait until we're healthy to respond
				<-ready
				// listAuthzProvidersEndpoint is guaranteed to be assigned now
				debugserverEndpoints.ListAuthzProvidersEndpoint.ServeHTTP(w, r)
			}),
		},
	}
}
