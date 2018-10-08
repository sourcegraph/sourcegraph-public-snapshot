package productsubscription

import (
	"context"
	"time"

	"github.com/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
)

func (r *productSubscription) InvoiceItem(ctx context.Context) (graphqlbackend.ProductSubscriptionInvoiceItem, error) {
	if r.v.BillingSubscriptionID == nil {
		return nil, nil
	}

	params := &stripe.SubscriptionParams{Params: stripe.Params{Context: ctx}}
	params.AddExpand("plan.product")
	billingSub, err := sub.Get(*r.v.BillingSubscriptionID, params)
	if err != nil {
		return nil, err
	}
	return &productSubscriptionInvoiceItem{billingSub: billingSub}, nil
}

type productSubscriptionInvoiceItem struct {
	billingSub *stripe.Subscription
}

var _ graphqlbackend.ProductSubscriptionInvoiceItem = &productSubscriptionInvoiceItem{}

func (r *productSubscriptionInvoiceItem) Plan() (graphqlbackend.ProductPlan, error) {
	return billing.ToProductPlan(r.billingSub.Plan)
}

func (r *productSubscriptionInvoiceItem) UserCount() int32 {
	return int32(r.billingSub.Quantity)
}

func (r *productSubscriptionInvoiceItem) ExpiresAt() string {
	return time.Unix(r.billingSub.CurrentPeriodEnd, 0).Format(time.RFC3339)
}
