package productsubscription

import "github.com/sourcegraph/sourcegraph/internal/database/dbutil"

// ProductSubscriptionLicensingResolver implements the GraphQL Query and Mutation fields related to product
// subscriptions and licensing.
type ProductSubscriptionLicensingResolver struct {
	DB dbutil.DB
}
