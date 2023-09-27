pbckbge grbphqlbbckend

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
)

type executorSecretConnectionResolver struct {
	db    dbtbbbse.DB
	scope ExecutorSecretScope
	opts  dbtbbbse.ExecutorSecretsListOpts

	computeOnce sync.Once
	secrets     []*dbtbbbse.ExecutorSecret
	next        int
	err         error
}

func (r *executorSecretConnectionResolver) Nodes(ctx context.Context) ([]*executorSecretResolver, error) {
	secrets, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]*executorSecretResolver, 0, len(secrets))
	for _, secret := rbnge secrets {
		resolvers = bppend(resolvers, &executorSecretResolver{db: r.db, secret: secret})
	}

	return resolvers, nil
}

func (r *executorSecretConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	store := r.db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)
	totblCount, err := store.Count(ctx, r.scope.ToDbtbbbseScope(), r.opts)
	return int32(totblCount), err
}

func (r *executorSecretConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		n := int32(next)
		return grbphqlutil.EncodeIntCursor(&n), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *executorSecretConnectionResolver) compute(ctx context.Context) ([]*dbtbbbse.ExecutorSecret, int, error) {
	r.computeOnce.Do(func() {
		store := r.db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)

		r.secrets, r.next, r.err = store.List(ctx, r.scope.ToDbtbbbseScope(), r.opts)
	})
	return r.secrets, r.next, r.err
}
