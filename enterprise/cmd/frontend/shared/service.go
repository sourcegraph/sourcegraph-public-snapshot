package shared

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	frontend_shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"

	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
)

type svc struct{}

func (svc) Name() string { return "frontend" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	frontend_shared.CLILoadConfig()
	codeintel.LoadConfig()
	return nil, frontend_shared.GRPCWebUIDebugEndpoints()
}

func setAllowAnonymousUsageContextKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if info, err := licensing.GetConfiguredProductLicenseInfo(); err == nil && info != nil {
			ctx = context.WithValue(r.Context(), auth.AllowAnonymousRequestContextKey, info.HasTag(licensing.AllowAnonymousUsageTag))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var extraContextMiddleware = &auth.Middleware{
	API: setAllowAnonymousUsageContextKey,
	App: setAllowAnonymousUsageContextKey,
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	return frontend_shared.CLIMain(ctx, observationCtx, ready, EnterpriseSetupHook, extraContextMiddleware)
}

var Service service.Service = svc{}
