package productsubscription

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/sub"
)

type productSubscriptionPreviewInvoice struct {
	amountDue     int32
	prorationDate int32
}

func (r *productSubscriptionPreviewInvoice) AmountDue() int32     { return r.amountDue }
func (r *productSubscriptionPreviewInvoice) ProrationDate() int32 { return r.prorationDate }

func (ProductSubscriptionLicensingResolver) PreviewProductSubscriptionInvoice(ctx context.Context, args *graphqlbackend.PreviewProductSubscriptionInvoiceArgs) (graphqlbackend.ProductSubscriptionPreviewInvoice, error) {
	accountUser, err := graphqlbackend.UserByID(ctx, args.Account)
	if err != nil {
		return nil, err
	}
	custID, err := billing.GetOrAssignUserCustomerID(ctx, accountUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Users may only preview invoices for their own product subscriptions. Site admins
	// may preview invoices for all product subscriptions.
	if err := backend.CheckSiteAdminOrSameUser(ctx, accountUser.SourcegraphID()); err != nil {
		return nil, err
	}

	var (
		subToUpdate        *productSubscription
		billingSubToUpdate *stripe.Subscription
	)
	if args.SubscriptionToUpdate != nil {
		var err error
		subToUpdate, err = productSubscriptionByID(ctx, *args.SubscriptionToUpdate)
		if err != nil {
			return nil, err
		}
		// 🚨 SECURITY: Only site admins and the subscription's account owner may preview invoices
		// for product subscriptions.
		if err := backend.CheckSiteAdminOrSameUser(ctx, subToUpdate.v.UserID); err != nil {
			return nil, err
		}

		// 🚨 SECURITY: Ensure that the subscription is owned by the account (i.e., that the
		// parameters are internally consistent). These checks are redundant for site admins, but
		// it's good to be robust against bugs.
		if subToUpdate.v.UserID != accountUser.SourcegraphID() {
			return nil, errors.New("product subscription's account owner does not match the provided account parameter")
		}
		if subToUpdate.v.BillingSubscriptionID == nil {
			return nil, errors.New("unable to get preview invoice for product subscription that has no associated billing information")
		}
		billingSubToUpdate, err = sub.Get(*subToUpdate.v.BillingSubscriptionID, &stripe.SubscriptionParams{Params: stripe.Params{Context: ctx}})
		if err != nil {
			return nil, err
		}
		if billingSubToUpdate.Customer.ID != custID {
			return nil, errors.New("product subscription's billing customer does not match the provided account parameter")
		}
		if len(billingSubToUpdate.Items.Data) != 1 {
			return nil, fmt.Errorf("product subscription has unexpected number of invoice items (got %d, want 1)", len(billingSubToUpdate.Items.Data))
		}
	}

	// Get the preview invoice.
	prorationDate := time.Now().Unix()
	params := &stripe.InvoiceParams{
		Params:            stripe.Params{Context: ctx},
		Customer:          stripe.String(custID),
		SubscriptionItems: []*stripe.SubscriptionItemsParams{billing.ToSubscriptionItemsParams(args.ProductSubscription)},
	}
	if billingSubToUpdate != nil {
		params.SubscriptionProrationDate = stripe.Int64(prorationDate)
		params.SubscriptionItems[0].ID = stripe.String(billingSubToUpdate.Items.Data[0].ID)
	}
	invoice, err := invoice.GetNext(params)
	if err != nil {
		return nil, err
	}

	// Calculate the cost.
	amountDue := int64(0)
	for _, invoiceItem := range invoice.Lines.Data {
		// When updating an existing subscription, only include invoice items that are affected by
		// the update (== whose proration date is the same as the one we set on the update params).
		if billingSubToUpdate != nil && invoiceItem.Period.Start != prorationDate {
			continue
		}
		amountDue += invoiceItem.Amount
	}

	return &productSubscriptionPreviewInvoice{
		amountDue:     int32(amountDue),
		prorationDate: int32(prorationDate),
	}, nil
}
