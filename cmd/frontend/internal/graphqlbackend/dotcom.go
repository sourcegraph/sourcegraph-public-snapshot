package graphqlbackend

import (
	"github.com/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	"github.com/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func init() {
	// Contribute the GraphQL types DotcomMutation and DotcomQuery.
	graphqlbackend.Dotcom = dotcomResolver{}
}

// dotcomResolver implements the GraphQL types DotcomMutation and DotcomQuery.
type dotcomResolver struct {
	productsubscription.ProductSubscriptionLicensingResolver
	billing.BillingResolver
}
