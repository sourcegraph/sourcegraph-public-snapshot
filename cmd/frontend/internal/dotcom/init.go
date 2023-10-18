package dotcom

import (
	"context"
	"net/http"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// dotcomRootResolver implements the GraphQL types DotcomMutation and DotcomQuery.
type dotcomRootResolver struct {
	productsubscription.ProductSubscriptionLicensingResolver
	productsubscription.CodyGatewayDotcomUserResolver
}

func (d dotcomRootResolver) Dotcom() graphqlbackend.DotcomResolver {
	return d
}

func (d dotcomRootResolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		productsubscription.ProductLicenseIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return d.ProductLicenseByID(ctx, id)
		},
		productsubscription.ProductSubscriptionIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return d.ProductSubscriptionByID(ctx, id)
		},
	}
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
	if envvar.SourcegraphDotComMode() {
		enterpriseServices.DotcomRootResolver = dotcomRootResolver{
			ProductSubscriptionLicensingResolver: productsubscription.ProductSubscriptionLicensingResolver{
				Logger: observationCtx.Logger.Scoped("productsubscriptions"),
				DB:     db,
			},
			CodyGatewayDotcomUserResolver: productsubscription.CodyGatewayDotcomUserResolver{
				Logger: observationCtx.Logger.Scoped("codygatewayuser"),
				DB:     db,
			},
		}
		enterpriseServices.NewDotcomLicenseCheckHandler = func() http.Handler {
			return productsubscription.NewLicenseCheckHandler(db)
		}
	}
	return nil
}
