package billing

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	stripe "github.com/stripe/stripe-go"
)

// ToSubscriptionItemsParams converts a value of GraphQL type ProductSubscriptionInput into a
// subscription item parameter for the billing system.
func ToSubscriptionItemsParams(input graphqlbackend.ProductSubscriptionInput) *stripe.SubscriptionItemsParams {
	return &stripe.SubscriptionItemsParams{
		Plan:     stripe.String(input.BillingPlanID),
		Quantity: stripe.Int64(int64(input.UserCount)),
	}
}
