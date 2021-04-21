package dotcom

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/stripeutil"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// dotcomRootResolver implements the GraphQL types DotcomMutation and DotcomQuery.
type dotcomRootResolver struct {
	productsubscription.ProductSubscriptionLicensingResolver
	billing.BillingResolver
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

func Init(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services) error {
	stripeEnabled := stripeutil.ValidateAndPublishConfig()
	// Only enabled on Sourcegraph.com or when Stripe is configured correctly.
	if envvar.SourcegraphDotComMode() || stripeEnabled {
		enterpriseServices.DotcomResolver = dotcomRootResolver{
			ProductSubscriptionLicensingResolver: productsubscription.ProductSubscriptionLicensingResolver{
				DB: db,
			},
			BillingResolver: billing.BillingResolver{DB: db},
		}
	}
	return nil
}
