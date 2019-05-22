package productsubscription

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	db_ "github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/event"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/sub"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	graphqlbackend.ProductSubscriptionByID = func(ctx context.Context, id graphql.ID) (graphqlbackend.ProductSubscription, error) {
		return productSubscriptionByID(ctx, id)
	}
}

// productSubscription implements the GraphQL type ProductSubscription.
type productSubscription struct {
	v *dbSubscription
}

// productSubscriptionByID looks up and returns the ProductSubscription with the given GraphQL
// ID. If no such ProductSubscription exists, it returns a non-nil error.
func productSubscriptionByID(ctx context.Context, id graphql.ID) (*productSubscription, error) {
	idString, err := unmarshalProductSubscriptionID(id)
	if err != nil {
		return nil, err
	}
	return productSubscriptionByDBID(ctx, idString)
}

// productSubscriptionByDBID looks up and returns the ProductSubscription with the given database
// ID. If no such ProductSubscription exists, it returns a non-nil error.
func productSubscriptionByDBID(ctx context.Context, id string) (*productSubscription, error) {
	v, err := dbSubscriptions{}.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only site admins and the subscription account's user may view a product subscription.
	if err := backend.CheckSiteAdminOrSameUser(ctx, v.UserID); err != nil {
		return nil, err
	}
	return &productSubscription{v: v}, nil
}

func (r *productSubscription) ID() graphql.ID {
	return marshalProductSubscriptionID(r.v.ID)
}

func marshalProductSubscriptionID(id string) graphql.ID {
	return relay.MarshalID("ProductSubscription", id)
}

func unmarshalProductSubscriptionID(id graphql.ID) (productSubscriptionID string, err error) {
	err = relay.UnmarshalSpec(id, &productSubscriptionID)
	return
}

func (r *productSubscription) UUID() string {
	return r.v.ID
}

func (r *productSubscription) Name() string {
	return fmt.Sprintf("L-%s", strings.ToUpper(strings.Replace(r.v.ID, "-", "", -1)[:10]))
}

func (r *productSubscription) Account(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.v.UserID)
}

func (r *productSubscription) Events(ctx context.Context) ([]graphqlbackend.ProductSubscriptionEvent, error) {
	if r.v.BillingSubscriptionID == nil {
		return []graphqlbackend.ProductSubscriptionEvent{}, nil
	}

	// List all events related to this subscription. The related_object parameter is an undocumented
	// Stripe API.
	params := &stripe.EventListParams{
		ListParams: stripe.ListParams{Context: ctx},
	}
	params.Filters.AddFilter("related_object", "", *r.v.BillingSubscriptionID)
	events := event.List(params)
	var gqlEvents []graphqlbackend.ProductSubscriptionEvent
	for events.Next() {
		gqlEvent, okToShowUser := billing.ToProductSubscriptionEvent(events.Event())
		if okToShowUser {
			gqlEvents = append(gqlEvents, gqlEvent)
		}
	}
	if err := events.Err(); err != nil {
		return nil, err
	}
	return gqlEvents, nil
}

func (r *productSubscription) ActiveLicense(ctx context.Context) (graphqlbackend.ProductLicense, error) {
	l, err := r.activeDBLicense(ctx)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, nil
	}
	return &productLicense{v: l}, err
}

func (r *productSubscription) activeDBLicense(ctx context.Context) (*dbLicense, error) {
	// Return newest license.
	licenses, err := dbLicenses{}.List(ctx, dbLicensesListOptions{
		ProductSubscriptionID: r.v.ID,
		LimitOffset:           &db_.LimitOffset{Limit: 1},
	})
	if err != nil {
		return nil, err
	}
	if len(licenses) == 0 {
		return nil, nil
	}
	return licenses[0], nil
}

func (r *productSubscription) ProductLicenses(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.ProductLicenseConnection, error) {
	// 🚨 SECURITY: Only site admins may list historical product licenses (to reduce confusion
	// around old license reuse). Other viewers should use ProductSubscription.activeLicense.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	opt := dbLicensesListOptions{ProductSubscriptionID: r.v.ID}
	args.Set(&opt.LimitOffset)
	return &productLicenseConnection{opt: opt}, nil
}

func (r *productSubscription) CreatedAt() string {
	return r.v.CreatedAt.Format(time.RFC3339)
}

func (r *productSubscription) IsArchived() bool { return r.v.ArchivedAt != nil }

func (r *productSubscription) URL(ctx context.Context) (string, error) {
	accountUser, err := r.Account(ctx)
	if err != nil {
		return "", err
	}
	return *accountUser.SettingsURL() + "/subscriptions/" + string(r.v.ID), nil
}

func (r *productSubscription) URLForSiteAdmin(ctx context.Context) *string {
	// 🚨 SECURITY: Only site admins may see this URL. Currently it does not contain any sensitive
	// info, but there is no need to show it to non-site admins.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil
	}
	u := fmt.Sprintf("/site-admin/dotcom/product/subscriptions/%s", r.v.ID)
	return &u
}

func (r *productSubscription) URLForSiteAdminBilling(ctx context.Context) (*string, error) {
	// 🚨 SECURITY: Only site admins may see this URL, which might contain the subscription's billing ID.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if id := r.v.BillingSubscriptionID; id != nil {
		u := billing.SubscriptionURL(*id)
		return &u, nil
	}
	return nil, nil
}

func (ProductSubscriptionLicensingResolver) CreateProductSubscription(ctx context.Context, args *graphqlbackend.CreateProductSubscriptionArgs) (graphqlbackend.ProductSubscription, error) {
	// 🚨 SECURITY: Only site admins may create product subscriptions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	user, err := graphqlbackend.UserByID(ctx, args.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := dbSubscriptions{}.Create(ctx, user.DatabaseID())
	if err != nil {
		return nil, err
	}
	return productSubscriptionByDBID(ctx, id)
}

func (ProductSubscriptionLicensingResolver) SetProductSubscriptionBilling(ctx context.Context, args *graphqlbackend.SetProductSubscriptionBillingArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may update product subscriptions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	// Ensure the args refer to valid subscriptions in the database and in the billing system.
	dbSub, err := productSubscriptionByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	if args.BillingSubscriptionID != nil {
		if _, err := sub.Get(*args.BillingSubscriptionID, &stripe.SubscriptionParams{Params: stripe.Params{Context: ctx}}); err != nil {
			return nil, err
		}
	}

	stringValue := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	if err := (dbSubscriptions{}).Update(ctx, dbSub.v.ID, dbSubscriptionUpdate{
		billingSubscriptionID: &sql.NullString{
			String: stringValue(args.BillingSubscriptionID),
			Valid:  args.BillingSubscriptionID != nil,
		},
	}); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (ProductSubscriptionLicensingResolver) CreatePaidProductSubscription(ctx context.Context, args *graphqlbackend.CreatePaidProductSubscriptionArgs) (*graphqlbackend.CreatePaidProductSubscriptionResult, error) {
	user, err := graphqlbackend.UserByID(ctx, args.AccountID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Users may only create paid product subscriptions for themselves. Site admins may
	// create them for any user.
	if err := backend.CheckSiteAdminOrSameUser(ctx, user.DatabaseID()); err != nil {
		return nil, err
	}

	// Determine which license tags and min quantity to use for the purchased plan. Do this early on
	// because it's the most likely place for a stupid mistake to cause a bug, and doing it early
	// means the user hasn't been charged if there is an error.
	licenseTags, minQuantity, err := billing.InfoForProductPlan(ctx, args.ProductSubscription.BillingPlanID)
	if err != nil {
		return nil, err
	}
	if minQuantity != nil && args.ProductSubscription.UserCount < *minQuantity {
		args.ProductSubscription.UserCount = *minQuantity
	}

	// Create the subscription in our database first, before processing payment. If payment fails,
	// users can retry payment on the already created subscription.
	subID, err := dbSubscriptions{}.Create(ctx, user.DatabaseID())
	if err != nil {
		return nil, err
	}

	// Get the billing customer for the current user, and update it to use the payment source
	// provided to us.
	custID, err := billing.GetOrAssignUserCustomerID(ctx, user.DatabaseID())
	if err != nil {
		return nil, err
	}
	custUpdateParams := &stripe.CustomerParams{
		Params: stripe.Params{Context: ctx},
	}
	custUpdateParams.SetSource(args.PaymentToken)
	if _, err := customer.Update(custID, custUpdateParams); err != nil {
		return nil, err
	}

	// Create the billing subscription.
	billingSub, err := sub.New(&stripe.SubscriptionParams{
		Params:   stripe.Params{Context: ctx},
		Customer: stripe.String(custID),
		Items:    []*stripe.SubscriptionItemsParams{billing.ToSubscriptionItemsParams(args.ProductSubscription)},
	})
	if err != nil {
		return nil, err
	}

	// Link the billing subscription with the subscription in our database.
	if err := (dbSubscriptions{}).Update(ctx, subID, dbSubscriptionUpdate{
		billingSubscriptionID: &sql.NullString{
			String: billingSub.ID,
			Valid:  true,
		},
	}); err != nil {
		return nil, err
	}

	// Generate a new license key for the subscription.
	if _, err := generateProductLicenseForSubscription(ctx, subID, &graphqlbackend.ProductLicenseInput{
		Tags:      licenseTags,
		UserCount: args.ProductSubscription.UserCount,
		ExpiresAt: int32(billingSub.CurrentPeriodEnd),
	}); err != nil {
		return nil, err
	}

	sub, err := productSubscriptionByDBID(ctx, subID)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.CreatePaidProductSubscriptionResult{ProductSubscriptionValue: sub}, nil
}

func (ProductSubscriptionLicensingResolver) UpdatePaidProductSubscription(ctx context.Context, args *graphqlbackend.UpdatePaidProductSubscriptionArgs) (*graphqlbackend.UpdatePaidProductSubscriptionResult, error) {
	subToUpdate, err := productSubscriptionByID(ctx, args.SubscriptionID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins and the subscription's account owner may update product
	// subscriptions.
	if err := backend.CheckSiteAdminOrSameUser(ctx, subToUpdate.v.UserID); err != nil {
		return nil, err
	}

	// Determine which license tags and min quantity to use for the purchased plan. Do this early on
	// because it's the most likely place for a stupid mistake to cause a bug, and doing it early
	// means the user hasn't been charged if there is an error.
	licenseTags, minQuantity, err := billing.InfoForProductPlan(ctx, args.Update.BillingPlanID)
	if err != nil {
		return nil, err
	}
	if minQuantity != nil && args.Update.UserCount < *minQuantity {
		args.Update.UserCount = *minQuantity
	}

	params := &stripe.SubscriptionParams{
		Params:  stripe.Params{Context: ctx},
		Items:   []*stripe.SubscriptionItemsParams{billing.ToSubscriptionItemsParams(args.Update)},
		Prorate: stripe.Bool(true),
	}

	// Get the billing customer for the current user, and update it to use the payment source
	// provided to us.
	custID, err := billing.GetOrAssignUserCustomerID(ctx, subToUpdate.v.UserID)
	if err != nil {
		return nil, err
	}
	custUpdateParams := &stripe.CustomerParams{
		Params: stripe.Params{Context: ctx},
	}
	custUpdateParams.SetSource(args.PaymentToken)
	if _, err := customer.Update(custID, custUpdateParams); err != nil {
		return nil, err
	}

	if subToUpdate.v.BillingSubscriptionID == nil {
		return nil, errors.New("unable to update product subscription that has no associated billing information")
	}
	subParams := &stripe.SubscriptionParams{Params: stripe.Params{Context: ctx}}
	subParams.AddExpand("plan")
	billingSubToUpdate, err := sub.Get(*subToUpdate.v.BillingSubscriptionID, subParams)
	if err != nil {
		return nil, err
	}
	idToReplace, err := billing.GetSubscriptionItemIDToReplace(billingSubToUpdate, custID)
	if err != nil {
		return nil, err
	}
	params.Items[0].ID = stripe.String(idToReplace)

	// Forbid self-service downgrades. (Reason: We can't revoke licenses, so we want to manually
	// intervene to ensure that customers who downgrade are not using the previous license.)
	{
		planParams := &stripe.PlanParams{Params: stripe.Params{Context: ctx}}
		afterPlan, err := plan.Get(args.Update.BillingPlanID, planParams)
		if err != nil {
			return nil, err
		}
		if isDowngradeRequiringManualIntervention(int32(billingSubToUpdate.Quantity), billingSubToUpdate.Plan.Amount, args.Update.UserCount, afterPlan.Amount) {
			return nil, errors.New("self-service downgrades are not yet supported")
		}
	}

	// Update the billing subscription.
	billingSub, err := sub.Update(*subToUpdate.v.BillingSubscriptionID, params)
	if err != nil {
		return nil, err
	}

	// Generate an invoice and charge so that payment is performed immediately. See
	// https://stripe.com/docs/billing/subscriptions/upgrading-downgrading.
	//
	// TODO(sqs): use webhooks to ensure the subscription is rolled back if the invoice payment
	// fails.
	{
		inv, err := invoice.New(&stripe.InvoiceParams{
			Params:       stripe.Params{Context: ctx},
			Customer:     stripe.String(custID),
			Subscription: stripe.String(*subToUpdate.v.BillingSubscriptionID),
		})
		if err != nil {
			return nil, err
		}
		if _, err := invoice.Pay(inv.ID, &stripe.InvoicePayParams{
			Params: stripe.Params{Context: ctx},
		}); err != nil {
			return nil, err
		}
	}

	// Generate a new license key for the subscription with the updated parameters.
	if _, err := generateProductLicenseForSubscription(ctx, subToUpdate.v.ID, &graphqlbackend.ProductLicenseInput{
		Tags:      licenseTags,
		UserCount: args.Update.UserCount,
		ExpiresAt: int32(billingSub.CurrentPeriodEnd),
	}); err != nil {
		return nil, err
	}

	return &graphqlbackend.UpdatePaidProductSubscriptionResult{ProductSubscriptionValue: subToUpdate}, nil
}

func (ProductSubscriptionLicensingResolver) ArchiveProductSubscription(ctx context.Context, args *graphqlbackend.ArchiveProductSubscriptionArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may archive product subscriptions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	sub, err := productSubscriptionByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	if err := (dbSubscriptions{}).Archive(ctx, sub.v.ID); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (ProductSubscriptionLicensingResolver) ProductSubscription(ctx context.Context, args *graphqlbackend.ProductSubscriptionArgs) (graphqlbackend.ProductSubscription, error) {
	// 🚨 SECURITY: Only site admins and the subscription's account owner may get a product
	// subscription. This check is performed in productSubscriptionByDBID.
	return productSubscriptionByDBID(ctx, args.UUID)
}

func (ProductSubscriptionLicensingResolver) ProductSubscriptions(ctx context.Context, args *graphqlbackend.ProductSubscriptionsArgs) (graphqlbackend.ProductSubscriptionConnection, error) {
	var accountUser *graphqlbackend.UserResolver
	if args.Account != nil {
		var err error
		accountUser, err = graphqlbackend.UserByID(ctx, *args.Account)
		if err != nil {
			return nil, err
		}
	}

	// 🚨 SECURITY: Users may only list their own product subscriptions. Site admins may list
	// licenses for all users, or for any other user.
	if accountUser == nil {
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
	} else {
		if err := backend.CheckSiteAdminOrSameUser(ctx, accountUser.DatabaseID()); err != nil {
			return nil, err
		}
	}

	var opt dbSubscriptionsListOptions
	if accountUser != nil {
		opt.UserID = accountUser.DatabaseID()
	}

	orderByExpiresAt := args.OrderBy == "SUBSCRIPTION_ACTIVE_LICENSE_EXPIRES_AT"
	// Only set limit and offset if ordering by default created_at field. If ordering by
	// license expiration, must fetch ALL records from the db, and compute the ordering in Go.
	if !orderByExpiresAt {
		args.ConnectionArgs.Set(&opt.LimitOffset)
	}
	return &productSubscriptionConnection{opt: opt, orderByExpiresAt: orderByExpiresAt}, nil
}

// productSubscriptionConnection implements the GraphQL type ProductSubscriptionConnection.
//
// 🚨 SECURITY: When instantiating a productSubscriptionConnection value, the caller MUST
// check permissions.
type productSubscriptionConnection struct {
	opt dbSubscriptionsListOptions

	// special handling for non-SQL ordering option
	orderByExpiresAt bool

	// cache results because they are used by multiple fields
	once    sync.Once
	results sortSubscriptions
	err     error
}

type subscriptionWithActiveLicense struct {
	subscription  *dbSubscription
	activeLicense *dbLicense
}

type sortSubscriptions []*subscriptionWithActiveLicense

func (s sortSubscriptions) Len() int {
	return len(s)
}

func (s sortSubscriptions) Less(i, j int) bool {
	// Subscriptions with no license should always be last.
	if s[i].activeLicense == nil {
		return false
	}
	if s[j].activeLicense == nil {
		return true
	}

	l1 := &productLicense{v: s[i].activeLicense}
	l1Info, err := l1.Info()
	if err != nil {
		log15.Error("graphqlbackend.productLicense.Info() failed", "error", err)
		return false
	}
	l2 := &productLicense{v: s[j].activeLicense}
	l2Info, err := l2.Info()
	// Subscriptions with no license should be at the end of the list.
	if err != nil {
		log15.Error("graphqlbackend.productLicense.Info() failed", "error", err)
		return false
	}
	return l1Info.ExpiresAtValue.Before(l2Info.ExpiresAtValue)
}
func (s sortSubscriptions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (r *productSubscriptionConnection) compute(ctx context.Context) (sortSubscriptions, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		var dbSubs []*dbSubscription
		dbSubs, r.err = dbSubscriptions{}.List(ctx, opt2)
		if r.err != nil {
			return
		}

		r.results = make([]*subscriptionWithActiveLicense, len(dbSubs))
		for i, s := range dbSubs {
			r.results[i] = &subscriptionWithActiveLicense{subscription: s}
		}

		if r.orderByExpiresAt {
			for i, s := range r.results {
				ps := &productSubscription{v: s.subscription}
				r.results[i].activeLicense, r.err = ps.activeDBLicense(ctx)
				if r.err != nil {
					return
				}
			}

			sort.Sort(r.results)
		}
	})
	return r.results, r.err
}

func (r *productSubscriptionConnection) Nodes(ctx context.Context) ([]graphqlbackend.ProductSubscription, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.ProductSubscription
	for _, result := range results {
		l = append(l, &productSubscription{v: result.subscription})
	}
	return l, nil
}

func (r *productSubscriptionConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbSubscriptions{}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *productSubscriptionConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(results) > r.opt.Limit), nil
}
