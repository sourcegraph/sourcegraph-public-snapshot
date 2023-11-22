package graphqlbackend

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	gext "github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit/externalaccount"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *siteResolver) ExternalAccounts(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
	User        *graphql.ID
	ServiceType *string
	ServiceID   *string
	ClientID    *string
},
) (*externalAccountConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can list all external accounts.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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
},
) (*externalAccountConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins and the user can list a user's external accounts.
	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
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
	db  database.DB
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

		r.externalAccounts, r.err = r.db.UserExternalAccounts().List(ctx, opt2)
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
	count, err := r.db.UserExternalAccounts().Count(ctx, r.opt)
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
},
) (*EmptyResponse, error) {
	disallow := featureflag.FromContext(ctx).GetBoolOr("disallow-user-external-account-deletion", false)
	if disallow {
		// NOTE: The error message will be directly shown to the user, so we need to make
		// it read like a sentence.
		return nil, errors.New("Self-serve unlinking external account is not allowed, please contact Sourcegraph support.")
	}

	id, err := unmarshalExternalAccountID(args.ExternalAccount)
	if err != nil {
		return nil, err
	}
	account, err := r.db.UserExternalAccounts().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user and site admins should be able to see a user's external accounts.
	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, account.UserID); err != nil {
		return nil, err
	}

	deleteOpts := database.ExternalAccountsDeleteOptions{IDs: []int32{account.ID}}
	if err := r.db.UserExternalAccounts().Delete(ctx, deleteOpts); err != nil {
		return nil, err
	}

	permssync.SchedulePermsSync(ctx, r.logger, r.db, permssync.ScheduleSyncOpts{
		UserIDs: []int32{account.UserID},
		Reason:  database.ReasonExternalAccountDeleted,
	})

	return &EmptyResponse{}, nil
}

func (r *schemaResolver) AddExternalAccount(ctx context.Context, args *struct {
	ServiceType    string
	ServiceID      string
	AccountDetails string
},
) (*EmptyResponse, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() || a.IsInternal() {
		return nil, auth.ErrNotAuthenticated
	}

	switch args.ServiceType {
	case extsvc.TypeGerrit:
		err := gext.AddGerritExternalAccount(ctx, r.db, a.UID, args.ServiceID, args.AccountDetails)
		if err != nil {
			return nil, err
		}

	case auth.SourcegraphOperatorProviderType:
		err := sourcegraphoperator.AddSourcegraphOperatorExternalAccount(ctx, r.db, a.UID, args.ServiceID, args.AccountDetails)
		if err != nil {
			return nil, errors.Wrap(err, "failed to add Sourcegraph Operator external account")
		}

	default:
		return nil, errors.Newf("unsupported service type %q", args.ServiceType)
	}

	permssync.SchedulePermsSync(ctx, r.logger, r.db, permssync.ScheduleSyncOpts{
		UserIDs: []int32{a.UID},
		Reason:  database.ReasonExternalAccountAdded,
	})

	return &EmptyResponse{}, nil
}
