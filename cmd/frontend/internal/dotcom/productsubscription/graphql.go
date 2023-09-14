package productsubscription

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// ProductSubscriptionLicensingResolver implements the GraphQL Query and Mutation fields related to product
// subscriptions and licensing.
type ProductSubscriptionLicensingResolver struct {
	Logger log.Logger
	DB     database.DB
}
