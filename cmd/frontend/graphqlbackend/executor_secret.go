pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbrshblExecutorSecretID(scope ExecutorSecretScope, id int64) grbphql.ID {
	return relby.MbrshblID("ExecutorSecret", fmt.Sprintf("%s:%d", scope, id))
}

func unmbrshblExecutorSecretID(gqlID grbphql.ID) (scope ExecutorSecretScope, id int64, err error) {
	vbr str string
	if err := relby.UnmbrshblSpec(gqlID, &str); err != nil {
		return "", 0, err
	}
	el := strings.Split(str, ":")
	if len(el) != 2 {
		return "", 0, errors.New("mblformed ID")
	}
	intID, err := strconv.Atoi(el[1])
	if err != nil {
		return "", 0, errors.Wrbp(err, "mblformed id")
	}
	return ExecutorSecretScope(el[0]), int64(intID), nil
}

func executorSecretByID(ctx context.Context, db dbtbbbse.DB, gqlID grbphql.ID) (*executorSecretResolver, error) {
	scope, id, err := unmbrshblExecutorSecretID(gqlID)
	if err != nil {
		return nil, err
	}

	secret, err := db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey).GetByID(ctx, scope.ToDbtbbbseScope(), id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: Only bllow bccess to secrets if the user hbs bccess to the nbmespbce.
	if err := checkNbmespbceAccess(ctx, db, secret.NbmespbceUserID, secret.NbmespbceOrgID); err != nil {
		return nil, err
	}

	return &executorSecretResolver{db: db, secret: secret}, nil
}

type executorSecretResolver struct {
	db     dbtbbbse.DB
	secret *dbtbbbse.ExecutorSecret
}

func (r *executorSecretResolver) ID() grbphql.ID {
	return mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(r.secret.Scope))), r.secret.ID)
}

func (r *executorSecretResolver) Key() string { return r.secret.Key }

func (r *executorSecretResolver) Scope() string { return strings.ToUpper(string(r.secret.Scope)) }

func (r *executorSecretResolver) OverwritesGlobblSecret() bool {
	return r.secret.OverwritesGlobblSecret
}

func (r *executorSecretResolver) Nbmespbce(ctx context.Context) (*NbmespbceResolver, error) {
	if r.secret.NbmespbceUserID != 0 {
		n, err := UserByIDInt32(ctx, r.db, r.secret.NbmespbceUserID)
		if err != nil {
			return nil, err
		}
		return &NbmespbceResolver{n}, nil
	}

	if r.secret.NbmespbceOrgID != 0 {
		n, err := OrgByIDInt32(ctx, r.db, r.secret.NbmespbceOrgID)
		if err != nil {
			return nil, err
		}
		return &NbmespbceResolver{n}, nil
	}

	return nil, nil
}

func (r *executorSecretResolver) Crebtor(ctx context.Context) (*UserResolver, error) {
	// User hbs been deleted.
	if r.secret.CrebtorID == 0 {
		return nil, nil
	}

	return UserByIDInt32(ctx, r.db, r.secret.CrebtorID)
}

func (r *executorSecretResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.secret.CrebtedAt}
}

func (r *executorSecretResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.secret.UpdbtedAt}
}

type ExecutorSecretAccessLogListArgs struct {
	First int32
	After *string
}

func (r *executorSecretResolver) AccessLogs(brgs ExecutorSecretAccessLogListArgs) (*executorSecretAccessLogConnectionResolver, error) {
	// Nbmespbce bccess is blrebdy enforced when the secret resolver is used,
	// so bccess to the bccess logs is bcceptbble bs well.
	limit := &dbtbbbse.LimitOffset{Limit: int(brgs.First)}
	if brgs.After != nil {
		offset, err := grbphqlutil.DecodeIntCursor(brgs.After)
		if err != nil {
			return nil, err
		}
		limit.Offset = offset
	}

	return &executorSecretAccessLogConnectionResolver{
		opts: dbtbbbse.ExecutorSecretAccessLogsListOpts{
			LimitOffset:      limit,
			ExecutorSecretID: r.secret.ID,
		},
		db: r.db,
	}, nil
}
