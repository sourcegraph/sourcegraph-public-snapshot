package productsubscription

import "github.com/sourcegraph/sourcegraph/internal/database"

// ProductSubscriptionLicensingResolver implements the GraphQL Query and Mutation fields related to product
// subscriptions and licensing.
type ProductSubscriptionLicensingResolver struct {
	DB database.DB
}
