package productsubscription

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// productSubscription implements the GraphQL type ProductSubscription.
type productSubscription struct {
	db database.DB
	v  *dbSubscription
}

// ProductSubscriptionByID looks up and returns the ProductSubscription with the given GraphQL
// ID. If no such ProductSubscription exists, it returns a non-nil error.
func (p ProductSubscriptionLicensingResolver) ProductSubscriptionByID(ctx context.Context, id graphql.ID) (graphqlbackend.ProductSubscription, error) {
	return productSubscriptionByID(ctx, p.DB, id)
}

// productSubscriptionByID looks up and returns the ProductSubscription with the given GraphQL
// ID. If no such ProductSubscription exists, it returns a non-nil error.
//
// 🚨 SECURITY: This checks that the actor has appropriate permissions on a product subscription.
func productSubscriptionByID(ctx context.Context, db database.DB, id graphql.ID) (*productSubscription, error) {
	idString, err := unmarshalProductSubscriptionID(id)
	if err != nil {
		return nil, err
	}
	return productSubscriptionByDBID(ctx, db, idString)
}

// productSubscriptionByDBID looks up and returns the ProductSubscription with the given database
// ID. If no such ProductSubscription exists, it returns a non-nil error.
//
// 🚨 SECURITY: This checks that the actor has appropriate permissions on a product subscription.
func productSubscriptionByDBID(ctx context.Context, db database.DB, id string) (*productSubscription, error) {
	v, err := dbSubscriptions{db: db}.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only site admins and the subscription account's user may view a product subscription.
	if err := serviceAccountOrOwnerOrSiteAdmin(ctx, db, &v.UserID); err != nil {
		return nil, err
	}
	return &productSubscription{v: v, db: db}, nil
}

func (r *productSubscription) ID() graphql.ID {
	return marshalProductSubscriptionID(r.v.ID)
}

const ProductSubscriptionIDKind = "ProductSubscription"

func marshalProductSubscriptionID(id string) graphql.ID {
	return relay.MarshalID(ProductSubscriptionIDKind, id)
}

func unmarshalProductSubscriptionID(id graphql.ID) (productSubscriptionID string, err error) {
	err = relay.UnmarshalSpec(id, &productSubscriptionID)
	return
}

func (r *productSubscription) UUID() string {
	return r.v.ID
}

func (r *productSubscription) Name() string {
	return fmt.Sprintf("L-%s", strings.ToUpper(strings.ReplaceAll(r.v.ID, "-", "")[:10]))
}

func (r *productSubscription) Account(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.db, r.v.UserID)
	if errcode.IsNotFound(err) {
		// NOTE: It is possible that the user has been deleted, but we do not want to
		// lose information of the product subscription because of that.
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *productSubscription) ActiveLicense(ctx context.Context) (graphqlbackend.ProductLicense, error) {
	// Return newest license.
	active, err := dbLicenses{db: r.db}.Active(ctx, r.v.ID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, nil
	}
	return &productLicense{db: r.db, v: active}, nil
}

func (r *productSubscription) ProductLicenses(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.ProductLicenseConnection, error) {
	// 🚨 SECURITY: Only site admins may list historical product licenses (to reduce confusion
	// around old license reuse). Other viewers should use ProductSubscription.activeLicense.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	opt := dbLicensesListOptions{ProductSubscriptionID: r.v.ID}
	args.Set(&opt.LimitOffset)
	return &productLicenseConnection{db: r.db, opt: opt}, nil
}

func (r *productSubscription) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.v.CreatedAt}
}

func (r *productSubscription) IsArchived() bool { return r.v.ArchivedAt != nil }

func (r *productSubscription) URL(ctx context.Context) (string, error) {
	accountUser, err := r.Account(ctx)
	if err != nil {
		return "", err
	}
	return *accountUser.SettingsURL() + "/subscriptions/" + r.v.ID, nil
}

func (r *productSubscription) URLForSiteAdmin(ctx context.Context) *string {
	// 🚨 SECURITY: Only site admins may see this URL. Currently it does not contain any sensitive
	// info, but there is no need to show it to non-site admins.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil
	}
	u := fmt.Sprintf("/site-admin/dotcom/product/subscriptions/%s", r.v.ID)
	return &u
}

func (r ProductSubscriptionLicensingResolver) CreateProductSubscription(ctx context.Context, args *graphqlbackend.CreateProductSubscriptionArgs) (graphqlbackend.ProductSubscription, error) {
	// 🚨 SECURITY: Only site admins may create product subscriptions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.DB); err != nil {
		return nil, err
	}

	user, err := graphqlbackend.UserByID(ctx, r.DB, args.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := dbSubscriptions{db: r.DB}.Create(ctx, user.DatabaseID(), user.Username())
	if err != nil {
		return nil, err
	}
	return productSubscriptionByDBID(ctx, r.DB, id)
}

func (r ProductSubscriptionLicensingResolver) ArchiveProductSubscription(ctx context.Context, args *graphqlbackend.ArchiveProductSubscriptionArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may archive product subscriptions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.DB); err != nil {
		return nil, err
	}

	sub, err := productSubscriptionByID(ctx, r.DB, args.ID)
	if err != nil {
		return nil, err
	}
	if err := (dbSubscriptions{db: r.DB}).Archive(ctx, sub.v.ID); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r ProductSubscriptionLicensingResolver) ProductSubscription(ctx context.Context, args *graphqlbackend.ProductSubscriptionArgs) (graphqlbackend.ProductSubscription, error) {
	// 🚨 SECURITY: Only site admins and the subscription's account owner may get a product
	// subscription. This check is performed in productSubscriptionByDBID.
	return productSubscriptionByDBID(ctx, r.DB, args.UUID)
}

func (r ProductSubscriptionLicensingResolver) ProductSubscriptions(ctx context.Context, args *graphqlbackend.ProductSubscriptionsArgs) (graphqlbackend.ProductSubscriptionConnection, error) {
	var accountUser *graphqlbackend.UserResolver
	var accountUserID *int32
	if args.Account != nil {
		var err error
		accountUser, err = graphqlbackend.UserByID(ctx, r.DB, *args.Account)
		if err != nil {
			return nil, err
		}
		id := accountUser.DatabaseID()
		accountUserID = &id
	}

	// 🚨 SECURITY: Users may only list their own product subscriptions. Site admins may list
	// licenses for all users, or for any other user.
	if err := serviceAccountOrOwnerOrSiteAdmin(ctx, r.DB, accountUserID); err != nil {
		return nil, err
	}
	var opt dbSubscriptionsListOptions
	if accountUser != nil {
		opt.UserID = accountUser.DatabaseID()
	}

	if args.Query != nil {
		// 🚨 SECURITY: Only site admins may query or view license for all users, or for any other user.
		// Note this check is currently repetitive with the check above. However, it is duplicated here to
		// ensure it remains in effect if the code path above chagnes.
		if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.DB); err != nil {
			return nil, err
		}
		opt.Query = *args.Query
	}

	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &productSubscriptionConnection{logger: r.logger, db: r.DB, opt: opt}, nil
}

// productSubscriptionConnection implements the GraphQL type ProductSubscriptionConnection.
//
// 🚨 SECURITY: When instantiating a productSubscriptionConnection value, the caller MUST
// check permissions.
type productSubscriptionConnection struct {
	logger log.Logger
	opt    dbSubscriptionsListOptions
	db     database.DB

	// cache results because they are used by multiple fields
	once    sync.Once
	results []*dbSubscription
	err     error
}

func (r *productSubscriptionConnection) compute(ctx context.Context) ([]*dbSubscription, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.results, r.err = dbSubscriptions{db: r.db}.List(ctx, opt2)
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
		l = append(l, &productSubscription{db: r.db, v: result})
	}
	return l, nil
}

func (r *productSubscriptionConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbSubscriptions{db: r.db}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *productSubscriptionConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(results) > r.opt.Limit), nil
}
