package graphqlbackend

import (
	"context"
	"strings"

	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var executorSecretKeyPattern = regexp.MustCompile("^[A-Z][A-Z0-9_]*$")

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

func (r *schemaResolver) CreateExecutorSecret(ctx context.Context, args CreateExecutorSecretArgs) (*executorSecretResolver, error) {
	var userID, orgID int32
	if args.Namespace != nil {
		if err := UnmarshalNamespaceID(*args.Namespace, &userID, &orgID); err != nil {
			return nil, err
		}
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	// 🚨 SECURITY: Check namespace access.
	if err := checkNamespaceAccess(ctx, r.db, userID, orgID); err != nil {
		return nil, err
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	if len(args.Key) == 0 {
		return nil, errors.New("key cannot be empty string")
	}

	if !executorSecretKeyPattern.Match([]byte(args.Key)) {
		return nil, errors.New("invalid key format, should be a valid env var name")
	}

	if len(args.Value) == 0 {
		return nil, errors.New("value cannot be empty string")
	}

	secret := &database.ExecutorSecret{
		Key:             args.Key,
		CreatorID:       a.UID,
		NamespaceUserID: userID,
		NamespaceOrgID:  orgID,
	}
	if err := store.Create(ctx, args.Scope.ToDatabaseScope(), secret, args.Value); err != nil {
		if err == database.ErrDuplicateExecutorSecret {
			return nil, &ErrDuplicateExecutorSecret{}
		}
		return nil, err
	}

	return &executorSecretResolver{db: r.db, secret: secret}, nil
}

type ErrDuplicateExecutorSecret struct{}

func (e ErrDuplicateExecutorSecret) Error() string {
	return "multiple secrets with the same key in the same namespace not allowed"
}

func (e ErrDuplicateExecutorSecret) Extensions() map[string]any {
	return map[string]any{"code": "ErrDuplicateExecutorSecret"}
}

type UpdateExecutorSecretArgs struct {
	ID    graphql.ID
	Scope ExecutorSecretScope
	Value string
}

func (r *schemaResolver) UpdateExecutorSecret(ctx context.Context, args UpdateExecutorSecretArgs) (_ *executorSecretResolver, err error) {
	scope, id, err := unmarshalExecutorSecretID(args.ID)
	if err != nil {
		return nil, err
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	if scope != args.Scope {
		return nil, errors.New("scope mismatch")
	}

	if len(args.Value) == 0 {
		return nil, errors.New("value cannot be empty string")
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	secret, err := tx.GetByID(ctx, args.Scope.ToDatabaseScope(), id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check namespace access.
	if err := checkNamespaceAccess(ctx, r.db, secret.NamespaceUserID, secret.NamespaceOrgID); err != nil {
		return nil, err
	}

	if err := tx.Update(ctx, args.Scope.ToDatabaseScope(), secret, args.Value); err != nil {
		return nil, err
	}

	return &executorSecretResolver{db: r.db, secret: secret}, nil
}

type DeleteExecutorSecretArgs struct {
	ID    graphql.ID
	Scope ExecutorSecretScope
}

func (r *schemaResolver) DeleteExecutorSecret(ctx context.Context, args DeleteExecutorSecretArgs) (_ *EmptyResponse, err error) {
	scope, id, err := unmarshalExecutorSecretID(args.ID)
	if err != nil {
		return nil, err
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	if scope != args.Scope {
		return nil, errors.New("scope mismatch")
	}

	store := r.db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)

	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	secret, err := tx.GetByID(ctx, args.Scope.ToDatabaseScope(), id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check namespace access.
	if err := checkNamespaceAccess(ctx, r.db, secret.NamespaceUserID, secret.NamespaceOrgID); err != nil {
		return nil, err
	}

	if err := tx.Delete(ctx, args.Scope.ToDatabaseScope(), id); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type ExecutorSecretsListArgs struct {
	Scope ExecutorSecretScope
	First int32
	After *string
}

func (o ExecutorSecretsListArgs) LimitOffset() (*database.LimitOffset, error) {
	limit := &database.LimitOffset{Limit: int(o.First)}
	if o.After != nil {
		offset, err := graphqlutil.DecodeIntCursor(o.After)
		if err != nil {
			return nil, err
		}
		limit.Offset = offset
	}
	return limit, nil
}

// ExecutorSecrets returns the global executor secrets.
func (r *schemaResolver) ExecutorSecrets(ctx context.Context, args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only allow access to list global secrets if the user is admin.
	// This is not terribly bad, since the secrets are also part of the user's namespace
	// secrets, but this endpoint is useless to non-admins.
	if err := checkNamespaceAccess(ctx, r.db, 0, 0); err != nil {
		return nil, err
	}

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

func (r *UserResolver) ExecutorSecrets(ctx context.Context, args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only allow access to list secrets if the user has access to the namespace.
	if err := checkNamespaceAccess(ctx, r.db, r.user.ID, 0); err != nil {
		return nil, err
	}

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

func (r *OrgResolver) ExecutorSecrets(ctx context.Context, args ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only allow access to list secrets if the user has access to the namespace.
	if err := checkNamespaceAccess(ctx, r.db, 0, r.org.ID); err != nil {
		return nil, err
	}

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

func checkNamespaceAccess(ctx context.Context, db database.DB, namespaceUserID, namespaceOrgID int32) error {
	if namespaceUserID != 0 {
		return auth.CheckSiteAdminOrSameUser(ctx, db, namespaceUserID)
	}
	if namespaceOrgID != 0 {
		return auth.CheckOrgAccessOrSiteAdmin(ctx, db, namespaceOrgID)
	}

	return auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}
