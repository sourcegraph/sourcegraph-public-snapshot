package localstore

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

func init() {
	tbl := AppSchema.Map.AddTableWithName(store.ExternalAuthToken{}, "ext_auth_token").SetKeys(false, "User", "Host", "client_id")
	tbl.ColMap("FirstAuthFailureMessage").SetMaxSize(1000)
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		`ALTER TABLE ext_auth_token ALTER COLUMN first_auth_failure_at TYPE timestamp with time zone USING first_auth_failure_at::timestamp with time zone;`,
	)
}

// externalAuthTokens is a DB-backed implementation of the ExternalAuthTokens store.
type externalAuthTokens struct{}

var _ store.ExternalAuthTokens = (*externalAuthTokens)(nil)

func (s *externalAuthTokens) GetUserToken(ctx context.Context, user int, host, clientID string) (*store.ExternalAuthToken, error) {
	if user == 0 {
		return nil, errors.New("no uid specified")
	}
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "ExternalAuthTokens.GetExternalToken", int32(user)); err != nil {
		return nil, err
	}
	var tok store.ExternalAuthToken
	err := appDBH(ctx).SelectOne(&tok, `SELECT * FROM ext_auth_token WHERE "user"=$1 AND "host"=$2 AND client_id=$3`, user, host, clientID)
	if err == sql.ErrNoRows {
		return nil, store.ErrNoExternalAuthToken
	} else if err != nil {
		return nil, err
	}
	return &tok, nil
}

func (s *externalAuthTokens) SetUserToken(ctx context.Context, tok *store.ExternalAuthToken) error {
	if tok.User == 0 {
		return errors.New("no uid specified")
	}
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "ExternalAuthTokens.SetExternalToken", int32(tok.User)); err != nil {
		return err
	}
	return dbutil.Transact(appDBH(ctx), func(tx gorp.SqlExecutor) error {
		ctx = WithAppDBH(ctx, tx)

		if _, err := s.GetUserToken(ctx, tok.User, tok.Host, tok.ClientID); err == store.ErrNoExternalAuthToken {
			return tx.Insert(tok)
		} else if err != nil && err != store.ErrExternalAuthTokenDisabled {
			return err
		}
		_, err := tx.Update(tok)
		return err
	})
}

func (s *externalAuthTokens) ListExternalUsers(ctx context.Context, extUIDs []int, host, clientID string) ([]*store.ExternalAuthToken, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "ExternalAuthTokens.ListExternalUsers"); err != nil {
		return nil, err
	}
	if extUIDs == nil || len(extUIDs) == 0 {
		return []*store.ExternalAuthToken{}, nil
	}
	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}
	var conds []string
	conds = append(conds, "host="+arg(host))
	conds = append(conds, "client_id="+arg(clientID))
	uidBindVars := make([]string, len(extUIDs))
	for i, uid := range extUIDs {
		uidBindVars[i] = arg(uid)
	}
	conds = append(conds, "ext_uid IN ("+strings.Join(uidBindVars, ",")+")")
	whereSQL := "(" + strings.Join(conds, ") AND (") + ")"
	sql := fmt.Sprintf(`SELECT * FROM ext_auth_token WHERE %s`, whereSQL)

	var toks []*store.ExternalAuthToken
	if _, err := appDBH(ctx).Select(&toks, sql, args...); err != nil {
		return nil, err
	}
	return toks, nil
}

func (s *externalAuthTokens) DeleteToken(ctx context.Context, tok *sourcegraph.ExternalTokenSpec) error {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "ExternalAuthTokens.DeleteToken", tok.UID); err != nil {
		return err
	}

	_, err := appDBH(ctx).Exec(`DELETE FROM ext_auth_token WHERE "user"=$1 AND "host"=$2 AND client_id=$3;`, tok.UID, tok.Host, tok.ClientID)
	return err
}
