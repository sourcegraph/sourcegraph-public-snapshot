package graphqlbackend

import (
	"context"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
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

type ExecutorSecretScope string

const (
	ExecutorSecretScopeBatches ExecutorSecretScope = "BATCHES"
)

func (s ExecutorSecretScope) ToDatabaseScope() database.ExecutorSecretScope {
	return database.ExecutorSecretScope(strings.ToLower(string(s)))
}

type CreateExecutorSecretArgs struct {
	Key       string
	Value     string
	Scope     ExecutorSecretScope
	Namespace *graphql.ID
}

func (r *schemaResolver) CreateExecutorSecret(ctx context.Context, args CreateExecutorSecretArgs) (*ExecutorSecretResolver, error) {
	var userID, orgID int32
	if args.Namespace != nil {
		if err := UnmarshalNamespaceID(*args.Namespace, &userID, &orgID); err != nil {
			return nil, err
		}
	}
	// TODO: namespace access check.

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	secret := &database.ExecutorSecret{
		Key:             args.Key,
		CreatorID:       a.UID,
		NamespaceUserID: userID,
		NamespaceOrgID:  orgID,
	}
	if err := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey).Create(ctx, args.Scope.ToDatabaseScope(), secret, args.Value); err != nil {
		return nil, err
	}
	return &ExecutorSecretResolver{db: r.db, secret: secret}, nil
}

type UpdateExecutorSecretArgs struct {
	ID    graphql.ID
	Scope ExecutorSecretScope
	Value string
}

func (r *schemaResolver) UpdateExecutorSecret(ctx context.Context, args UpdateExecutorSecretArgs) (*ExecutorSecretResolver, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	id, err := unmarshalExecutorSecretID(args.ID)
	if err != nil {
		return nil, err
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	secret, err := tx.GetByID(ctx, args.Scope.ToDatabaseScope(), id)
	if err != nil {
		return nil, err
	}
	if err := store.Update(ctx, args.Scope.ToDatabaseScope(), secret, args.Value); err != nil {
		return nil, err
	}

	return &ExecutorSecretResolver{db: r.db, secret: secret}, nil
}

type DeleteExecutorSecretArgs struct {
	ID    graphql.ID
	Scope ExecutorSecretScope
}

func (r *schemaResolver) DeleteExecutorSecret(ctx context.Context, args DeleteExecutorSecretArgs) (*EmptyResponse, error) {
	id, err := unmarshalExecutorSecretID(args.ID)
	if err != nil {
		return nil, err
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	// Delete handles access of the actor to the secret properly.
	if err := store.Delete(ctx, args.Scope.ToDatabaseScope(), id); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type ExecutorSecretsListArgs struct {
	Scope ExecutorSecretScope
	First *int32
	After *string
}

func (o ExecutorSecretsListArgs) LimitOffset() (*database.LimitOffset, error) {
	limit := &database.LimitOffset{}
	if o.First != nil {
		limit.Limit = int(*o.First)
	}
	if o.After != nil {
		offset, err := graphqlutil.DecodeIntCursor(o.After)
		if err != nil {
			return nil, err
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

func executorSecretByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*ExecutorSecretResolver, error) {
	// TODO: perms?

	id, err := unmarshalExecutorSecretID(gqlID)
	if err != nil {
		return nil, err
	}

	// TODO: this should not hard code the scope.
	secret, err := db.ExecutorSecrets(keyring.Default().ExecutorSecretKey).GetByID(ctx, database.ExecutorSecretScopeBatches, id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &ExecutorSecretResolver{db: db, secret: secret}, nil
}

type executorSecretConnectionResolver struct {
	db    database.DB
	scope ExecutorSecretScope
	opts  database.ExecutorSecretsListOpts

	computeOnce sync.Once
	secrets     []*database.ExecutorSecret
	next        int
	err         error
}

func (r *executorSecretConnectionResolver) Nodes(ctx context.Context) ([]*ExecutorSecretResolver, error) {
	secrets, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*ExecutorSecretResolver, 0, len(secrets))
	for _, secret := range secrets {
		resolvers = append(resolvers, &ExecutorSecretResolver{db: r.db, secret: secret})
	}

	return resolvers, nil
}

func (r *executorSecretConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)
	totalCount, err := store.Count(ctx, r.scope.ToDatabaseScope(), r.opts)
	return int32(totalCount), err
}

func (r *executorSecretConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		n := int32(next)
		return graphqlutil.EncodeIntCursor(&n), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *executorSecretConnectionResolver) compute(ctx context.Context) ([]*database.ExecutorSecret, int, error) {
	r.computeOnce.Do(func() {
		store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

		r.secrets, r.next, r.err = store.List(ctx, r.scope.ToDatabaseScope(), r.opts)
	})
	return r.secrets, r.next, r.err
}
