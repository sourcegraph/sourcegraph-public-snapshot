package billing

import "github.com/sourcegraph/sourcegraph/internal/database/dbutil"

// BillingResolver implements the GraphQL Query and Mutation fields related to billing.
type BillingResolver struct {
	DB dbutil.DB
}
