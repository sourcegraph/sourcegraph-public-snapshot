package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func InitDotcom(db dbutil.DB) {
	// TODO(efritz) - de-globalize assignments in this function
	// Contribute the GraphQL types DotcomMutation and DotcomQuery.
	graphqlbackend.Dotcom = dotcomResolver{
		ProductSubscriptionLicensingResolver: productsubscription.ProductSubscriptionLicensingResolver{
			DB: db,
		},
	}
}

// dotcomResolver implements the GraphQL types DotcomMutation and DotcomQuery.
type dotcomResolver struct {
	productsubscription.ProductSubscriptionLicensingResolver
	billing.BillingResolver
}
