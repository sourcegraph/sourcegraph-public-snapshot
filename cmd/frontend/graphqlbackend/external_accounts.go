package graphqlbackend

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func (r *siteResolver) ExternalAccounts(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
	User        *graphql.ID
	ServiceType *string
	ServiceID   *string
	ClientID    *string
}) (*externalAccountConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can list all external accounts.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var opt database.ExternalAccountsListOptions
	if args.ServiceType != nil {
		opt.ServiceType = *args.ServiceType
	}
	if args.ServiceID != nil {
		opt.ServiceID = *args.ServiceID
	}
	if args.ClientID != nil {
		opt.ClientID = *args.ClientID
	}
	if args.User != nil {
		var err error
		opt.UserID, err = UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &externalAccountConnectionResolver{db: r.db, opt: opt}, nil
}

func (r *UserResolver) ExternalAccounts(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*externalAccountConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins and the user can list a user's external accounts.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}

	opt := database.ExternalAccountsListOptions{
		UserID: r.user.ID,
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &externalAccountConnectionResolver{db: r.db, opt: opt}, nil
}

// externalAccountConnectionResolver resolves a list of external accounts.
//
// 🚨 SECURITY: When instantiating an externalAccountConnectionResolver value, the caller MUST check
// permissions.
type externalAccountConnectionResolver struct {
	db  dbutil.DB
	opt database.ExternalAccountsListOptions

	// cache results because they are used by multiple fields
	once             sync.Once
	externalAccounts []*extsvc.Account
	err              error
}

func (r *externalAccountConnectionResolver) compute(ctx context.Context) ([]*extsvc.Account, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.externalAccounts, r.err = database.GlobalExternalAccounts.List(ctx, opt2)
	})
	return r.externalAccounts, r.err
}

func (r *externalAccountConnectionResolver) Nodes(ctx context.Context) ([]*externalAccountResolver, error) {
	externalAccounts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []*externalAccountResolver
	for _, externalAccount := range externalAccounts {
		l = append(l, &externalAccountResolver{db: r.db, account: *externalAccount})
	}
	return l, nil
}

func (r *externalAccountConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := database.GlobalExternalAccounts.Count(ctx, r.opt)
	return int32(count), err
}

func (r *externalAccountConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	externalAccounts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(externalAccounts) > r.opt.Limit), nil
}

func (r *schemaResolver) DeleteExternalAccount(ctx context.Context, args *struct {
	ExternalAccount graphql.ID
}) (*EmptyResponse, error) {
	id, err := unmarshalExternalAccountID(args.ExternalAccount)
	if err != nil {
		return nil, err
	}
	account, err := database.GlobalExternalAccounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user and site admins should be able to see a user's external accounts.
	if err := backend.CheckSiteAdminOrSameUser(ctx, account.UserID); err != nil {
		return nil, err
	}

	if err := database.GlobalExternalAccounts.Delete(ctx, account.ID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
