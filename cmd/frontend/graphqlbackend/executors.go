package graphqlbackend

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func unmarshalExecutorID(id graphql.ID) (executorID int64, err error) {
	err = relay.UnmarshalSpec(id, &executorID)
	return
}

type ExecutorsListArgs struct {
	Query  *string
	Active *bool
	First  int32
	After  *string
}

func (r *schemaResolver) Executors(ctx context.Context, args ExecutorsListArgs) (*executorConnectionResolver, error) {
	// 🚨 SECURITY: Only site-admins may view executor details
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	opts := database.ExecutorStoreListOptions{
		Offset: offset,
		Limit:  int(args.First),
	}
	if args.Query != nil {
		opts.Query = *args.Query
	}
	if args.Active != nil {
		opts.Active = *args.Active
	}
	execs, err := tx.Executors().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	totalCount, err := tx.Executors().Count(ctx, opts)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*ExecutorResolver, 0, len(execs))
	for _, executor := range execs {
		resolvers = append(resolvers, &ExecutorResolver{executor: executor})
	}

	nextOffset := graphqlutil.NextOffset(offset, len(execs), totalCount)

	executorConnection := &executorConnectionResolver{
		resolvers:  resolvers,
		totalCount: totalCount,
		nextOffset: nextOffset,
	}

	return executorConnection, nil
}

func (r *schemaResolver) AreExecutorsConfigured() bool {
	return conf.Get().ExecutorsAccessToken != ""
}

func executorByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*ExecutorResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalExecutorID(gqlID)
	if err != nil {
		return nil, err
	}

	executor, ok, err := db.Executors().GetByID(ctx, int(id))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return NewExecutorResolver(executor), nil
}

type CreateExecutorSecretArgs struct {
	Key       string
	Value     string
	Scope     string
	Namespace *graphql.ID
}

func (r *schemaResolver) CreateExecutorSecret(ctx context.Context, args CreateExecutorSecretArgs) (*ExecutorSecretResolver, error) {

	return nil, nil
}

type ExecutorSecretsListArgs struct {
	Scope string
	First *int32
	After *string
}

func (o ExecutorSecretsListArgs) LimitOffset() (*database.LimitOffset, error) {
	limit := &database.LimitOffset{}
	if o.First != nil {
		limit.Limit = int(*o.First)
	}
	if o.After != nil {
		offset, err := strconv.Atoi(*o.After)
		if err != nil {
			return nil, errors.Wrap(err, "parsing cursor")
		}
		limit.Offset = offset
	}
	return limit, nil
}

func (r *schemaResolver) ExecutorSecrets(args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// TODO: perms?

	limit, err := args.LimitOffset()
	if err != nil {
		return nil, err
	}
	return &executorSecretConnectionResolver{
		db:    r.db,
		scope: args.Scope,
		opts: database.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NamespaceUserID: 0,
			NamespaceOrgID:  0,
		},
	}, nil
}

func (r *UserResolver) ExecutorSecrets(args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// TODO: perms?

	limit, err := args.LimitOffset()
	if err != nil {
		return nil, err
	}
	return &executorSecretConnectionResolver{
		db:    r.db,
		scope: args.Scope,
		opts: database.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NamespaceUserID: r.user.ID,
			NamespaceOrgID:  0,
		},
	}, nil
}

func (r *OrgResolver) ExecutorSecrets(args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// TODO: perms?

	limit, err := args.LimitOffset()
	if err != nil {
		return nil, err
	}
	return &executorSecretConnectionResolver{
		db:    r.db,
		scope: args.Scope,
		opts: database.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NamespaceUserID: 0,
			NamespaceOrgID:  r.org.ID,
		},
	}, nil
}

type executorSecretConnectionResolver struct {
	db    database.DB
	scope string
	opts  database.ExecutorSecretsListOpts
}

func (r *executorSecretConnectionResolver) Nodes() ([]*ExecutorSecretResolver, error) {
	return nil, nil
}

func (r *executorSecretConnectionResolver) TotalCount() (int32, error) {
	return 0, nil
}

func (r *executorSecretConnectionResolver) PageInfo() (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
