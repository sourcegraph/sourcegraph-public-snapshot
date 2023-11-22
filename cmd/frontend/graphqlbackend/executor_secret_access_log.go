package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func marshalExecutorSecretAccessLogID(id int64) graphql.ID {
	return relay.MarshalID("ExecutorSecretAccessLog", id)
}

func unmarshalExecutorSecretAccessLogID(gqlID graphql.ID) (id int64, err error) {
	err = relay.UnmarshalSpec(gqlID, &id)
	return
}

func executorSecretAccessLogByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*executorSecretAccessLogResolver, error) {
	id, err := unmarshalExecutorSecretAccessLogID(gqlID)
	if err != nil {
		return nil, err
	}

	l, err := db.ExecutorSecretAccessLogs().GetByID(ctx, id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// TODO: How to get scope.
	secret, err := db.ExecutorSecrets(keyring.Default().ExecutorSecretKey).GetByID(ctx, database.ExecutorSecretScopeBatches, l.ExecutorSecretID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only allow access if the user has access to the namespace.
	if err := checkNamespaceAccess(ctx, db, secret.NamespaceUserID, secret.NamespaceOrgID); err != nil {
		return nil, err
	}

	return &executorSecretAccessLogResolver{db: db, log: l}, nil
}

type executorSecretAccessLogResolver struct {
	db  database.DB
	log *database.ExecutorSecretAccessLog

	// If true, the user has been preloaded. It can still be null (if deleted),
	// so this flag signifies that.
	attemptPreloadedUser bool
	preloadedUser        *types.User
}

func (r *executorSecretAccessLogResolver) ID() graphql.ID {
	return marshalExecutorSecretAccessLogID(r.log.ID)
}

func (r *executorSecretAccessLogResolver) ExecutorSecret(ctx context.Context) (*executorSecretResolver, error) {
	// TODO: Where to get the scope from..
	return executorSecretByID(ctx, r.db, marshalExecutorSecretID(ExecutorSecretScopeBatches, r.log.ExecutorSecretID))
}

func (r *executorSecretAccessLogResolver) User(ctx context.Context) (*UserResolver, error) {
	if r.attemptPreloadedUser {
		if r.preloadedUser == nil {
			return nil, nil
		}
		return NewUserResolver(ctx, r.db, r.preloadedUser), nil
	}

	if r.log.UserID == nil {
		return nil, nil
	}

	u, err := UserByIDInt32(ctx, r.db, *r.log.UserID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *executorSecretAccessLogResolver) MachineUser() string {
	return r.log.MachineUser
}

func (r *executorSecretAccessLogResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.log.CreatedAt}
}
