package dotcom

import (
	"context"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	subscriptionlicensechecksv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1/v1connect"
)

var (
	enableOnlineLicenseChecks = env.MustGetBool("DOTCOM_ENABLE_ONLINE_LICENSE_CHECKS", true,
		"If false, online license checks from instances always return successfully.")

	onlineLicenseChecksEnterprisePortal = env.Get("DOTCOM_ONLINE_LICENSE_CHECKS_ENTERPRISE_PORTAL_ADDR",
		"https://enterprise-portal.sourcegraph.com",
		"Enterprise Portal instance to target for the legacy dotcom online license check forwarding.")
)

// dotcomRootResolver implements the GraphQL types DotcomMutation and DotcomQuery.
type dotcomRootResolver struct {
	productsubscription.CodyGatewayDotcomUserResolver
}

func (d dotcomRootResolver) Dotcom() graphqlbackend.DotcomResolver {
	return d
}

func (d dotcomRootResolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{}
}

var _ graphqlbackend.DotcomRootResolver = dotcomRootResolver{}

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	// Only enabled on Sourcegraph.com.
	if dotcom.SourcegraphDotComMode() {
		enterpriseServices.DotcomRootResolver = dotcomRootResolver{
			CodyGatewayDotcomUserResolver: productsubscription.CodyGatewayDotcomUserResolver{
				Logger: observationCtx.Logger.Scoped("codygatewayuser"),
				DB:     db,
			},
		}
		enterpriseServices.NewDotcomLicenseCheckHandler = func() http.Handler {
			ep, err := url.Parse(onlineLicenseChecksEnterprisePortal)
			if err == nil && ep.Host == "127.0.0.1" {
				return productsubscription.NewLicenseCheckHandler(db, enableOnlineLicenseChecks,
					subscriptionlicensechecksv1connect.NewSubscriptionLicenseChecksServiceClient(
						httpcli.InternalDoer,
						onlineLicenseChecksEnterprisePortal,
					))
			}
			return productsubscription.NewLicenseCheckHandler(db, enableOnlineLicenseChecks,
				subscriptionlicensechecksv1connect.NewSubscriptionLicenseChecksServiceClient(
					httpcli.UncachedExternalDoer,
					onlineLicenseChecksEnterprisePortal,
				))
		}
	}
	return nil
}
