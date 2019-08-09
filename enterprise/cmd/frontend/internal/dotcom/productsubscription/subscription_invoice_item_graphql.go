package productsubscription

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_658(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
