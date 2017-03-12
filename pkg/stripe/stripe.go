package stripe

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
	"gopkg.in/inconshreveable/log15.v2"
)

func init() {
	stripeSecretTestKey := "sk_test_QHDBfU09USr4SVaJPZJEGruf"
	stripe.Key = env.Get("STRIPE_SECRET_KEY", stripeSecretTestKey, "The secret/private stripe API key.")

	stripePublicTestKey := "pk_test_Vo5BwrEkrXCM2ULouAd5ZBZz"
	StripePublicKey = env.Get("Stripe_PUBLIC_KEY", stripePublicTestKey, "The public Stripe API key.")
}

var StripePublicKey string

var ErrNotCustomer = errors.New("this user is not yet a customer")

const stripeMetaDataKey = "stripe_account_id"
const stripeOrgNameKey = "github_organization_name"

// getCustomerID retrieves the user's Stripe customerID from Auth0.
func getCustomerID(ctx context.Context) (*string, error) {
	appMeta, err := auth.GetAppMetadata(ctx)
	if err != nil {
		return nil, err
	}
	stripeID, ok := appMeta[stripeMetaDataKey].(string)
	if !ok {
		return nil, ErrNotCustomer
	}
	return &stripeID, nil
}

// setCustomerID saves the Stripe customerID in Auth0 metadata.
func setCustomerID(ctx context.Context, customerID string) error {
	actor := auth.ActorFromContext(ctx)
	return auth.SetAppMetadata(ctx, actor.UID, stripeMetaDataKey, customerID)
}

// getStripeSubscription retrieves the user's subscription. For now, a user can
// have at most one subscription.
func getStripeSubscription(ctx context.Context) *stripe.Sub {
	customerID, err := getCustomerID(ctx)
	if err != nil {
		return nil
	}
	customer, err := customer.Get(*customerID, nil)
	if err != nil {
		return nil
	}
	subs := customer.Subs.Values
	if len(subs) == 0 {
		return nil
	}
	if len(subs) > 1 {
		log15.Error("Expected customer to have at most one subscription, but got more. Manual intervention required for customer", "customerID", customerID)
		return nil
	}
	return subs[0]
}

type Plan struct {
	Cost        uint64
	RenewalDate int64
	OrgName     string
	Seats       uint64
}

// Get the user payment plan from Stripe.
func GetPlan(ctx context.Context) *Plan {
	sub := getStripeSubscription(ctx)
	if sub == nil {
		return nil
	}
	seats := sub.Quantity
	return &Plan{
		RenewalDate: sub.PeriodEnd,
		OrgName:     sub.Meta[stripeOrgNameKey],
		Cost:        sub.Plan.Amount * seats,
		Seats:       seats,
	}
}

func getOrCreateCustomer(ctx context.Context) (*string, error) {
	customerID, err := getCustomerID(ctx)
	if err == nil {
		return customerID, nil
	}
	if err != nil && err != ErrNotCustomer {
		return nil, err
	}
	actor := auth.ActorFromContext(ctx)

	// Create the customer
	customer, err := customer.New(&stripe.CustomerParams{
		Params: stripe.Params{
			Meta: map[string]string{"UID": actor.UID},
		},
	})
	if err != nil {
		return nil, err
	}
	err = setCustomerID(ctx, customer.ID)
	if err != nil {
		return nil, err
	}
	return &customer.ID, nil
}

// CancelSubscription cancels the user's subscription. It will not have an
// effect on the current billing cycle.
func CancelSubscription(ctx context.Context) error {
	subscription := getStripeSubscription(ctx)
	if subscription == nil {
		return errors.New("subscription does not exist")
	}
	_, err := sub.Cancel(subscription.ID, nil)
	return err
}

// SetTokenSourceForCustomer updates the user's payment method.
func SetTokenSourceForCustomer(ctx context.Context, token string) error {
	customerID, err := getOrCreateCustomer(ctx)
	if err != nil {
		return err
	}
	_, err = customer.Update(*customerID, &stripe.CustomerParams{
		Source: &stripe.SourceParams{Token: token},
	})
	return err
}

// orgPlan is the name of the Stripe plan that new subscribers will use.
const orgPlan = "orgplan-3/8/2016"

// SubscribeOrganization subscribes the user to an organization plan, with cost
// proportional to the number of seats.
func SubscribeOrganization(ctx context.Context, token string, org string, seats uint64) error {
	customerID, err := getOrCreateCustomer(ctx)
	if err != nil {
		return err
	}
	if err := SetTokenSourceForCustomer(ctx, token); err != nil {
		return err
	}
	params := &stripe.SubParams{
		Customer: *customerID,
		Plan:     orgPlan,
		Quantity: seats,
		Params: stripe.Params{
			Meta: map[string]string{stripeOrgNameKey: org},
		},
	}
	subscription := getStripeSubscription(ctx)
	if subscription != nil {
		log15.Error("Customer tried to subscribe for multiple plans", "customerID", customerID)
		return errors.New("only one organization allowed per customer")
	}
	_, err = sub.New(params)
	return err
}
