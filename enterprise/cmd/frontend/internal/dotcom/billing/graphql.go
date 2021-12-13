package billing

import "github.com/sourcegraph/sourcegraph/internal/database"

// BillingResolver implements the GraphQL Query and Mutation fields related to billing.
type BillingResolver struct {
	DB database.DB
}
