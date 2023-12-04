package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
)

type executorSecretConnectionResolver struct {
	db    database.DB
	scope ExecutorSecretScope
	opts  database.ExecutorSecretsListOpts

	computeOnce sync.Once
	secrets     []*database.ExecutorSecret
	next        int
	err         error
}

func (r *executorSecretConnectionResolver) Nodes(ctx context.Context) ([]*executorSecretResolver, error) {
	secrets, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*executorSecretResolver, 0, len(secrets))
	for _, secret := range secrets {
		resolvers = append(resolvers, &executorSecretResolver{db: r.db, secret: secret})
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
