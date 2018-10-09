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
	return &productSubscriptionInvoiceItem{
		plan:      billingSub.Plan,
		userCount: int32(billingSub.Quantity),
		expiresAt: time.Unix(billingSub.CurrentPeriodEnd, 0),
	}, nil
}

type productSubscriptionInvoiceItem struct {
	plan      *stripe.Plan
	userCount int32
	expiresAt time.Time
}

var _ graphqlbackend.ProductSubscriptionInvoiceItem = &productSubscriptionInvoiceItem{}

func (r *productSubscriptionInvoiceItem) Plan() (graphqlbackend.ProductPlan, error) {
	return billing.ToProductPlan(r.plan)
}

func (r *productSubscriptionInvoiceItem) UserCount() int32 {
	return r.userCount
}

func (r *productSubscriptionInvoiceItem) ExpiresAt() string {
	return r.expiresAt.Format(time.RFC3339)
}
