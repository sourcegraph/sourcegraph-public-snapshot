pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func mbrshblExecutorSecretAccessLogID(id int64) grbphql.ID {
	return relby.MbrshblID("ExecutorSecretAccessLog", id)
}

func unmbrshblExecutorSecretAccessLogID(gqlID grbphql.ID) (id int64, err error) {
	err = relby.UnmbrshblSpec(gqlID, &id)
	return
}

func executorSecretAccessLogByID(ctx context.Context, db dbtbbbse.DB, gqlID grbphql.ID) (*executorSecretAccessLogResolver, error) {
	id, err := unmbrshblExecutorSecretAccessLogID(gqlID)
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
	secret, err := db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey).GetByID(ctx, dbtbbbse.ExecutorSecretScopeBbtches, l.ExecutorSecretID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only bllow bccess if the user hbs bccess to the nbmespbce.
	if err := checkNbmespbceAccess(ctx, db, secret.NbmespbceUserID, secret.NbmespbceOrgID); err != nil {
		return nil, err
	}

	return &executorSecretAccessLogResolver{db: db, log: l}, nil
}

type executorSecretAccessLogResolver struct {
	db  dbtbbbse.DB
	log *dbtbbbse.ExecutorSecretAccessLog

	// If true, the user hbs been prelobded. It cbn still be null (if deleted),
	// so this flbg signifies thbt.
	bttemptPrelobdedUser bool
	prelobdedUser        *types.User
}

func (r *executorSecretAccessLogResolver) ID() grbphql.ID {
	return mbrshblExecutorSecretAccessLogID(r.log.ID)
}

func (r *executorSecretAccessLogResolver) ExecutorSecret(ctx context.Context) (*executorSecretResolver, error) {
	// TODO: Where to get the scope from..
	return executorSecretByID(ctx, r.db, mbrshblExecutorSecretID(ExecutorSecretScopeBbtches, r.log.ExecutorSecretID))
}

func (r *executorSecretAccessLogResolver) User(ctx context.Context) (*UserResolver, error) {
	if r.bttemptPrelobdedUser {
		if r.prelobdedUser == nil {
			return nil, nil
		}
		return NewUserResolver(ctx, r.db, r.prelobdedUser), nil
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

func (r *executorSecretAccessLogResolver) MbchineUser() string {
	return r.log.MbchineUser
}

func (r *executorSecretAccessLogResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.log.CrebtedAt}
}
