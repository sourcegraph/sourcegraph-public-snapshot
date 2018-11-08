// Package graphqlbackend injects enterprise GraphQL resolvers into our main
// graphqlbackend package. It does this as a side-effect of being imported.
package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
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
