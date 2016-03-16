package pgsql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil"
)

func init() {
	tbl := Schema.Map.AddTableWithName(auth.ExternalAuthToken{}, "ext_auth_token").SetKeys(false, "User", "Host", "client_id")
	tbl.ColMap("FirstAuthFailureMessage").SetMaxSize(1000)
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE ext_auth_token ALTER COLUMN first_auth_failure_at TYPE timestamp with time zone USING first_auth_failure_at::timestamp with time zone;`,
	)
}

// externalAuthTokens is a DB-backed implementation of the ExternalAuthTokens store.
type externalAuthTokens struct{}

var _ store.ExternalAuthTokens = (*externalAuthTokens)(nil)

func (s *externalAuthTokens) GetUserToken(ctx context.Context, user int, host, clientID string) (*auth.ExternalAuthToken, error) {
	if user == 0 {
		return nil, errors.New("no uid specified")
	}
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "ExternalAuthTokens.GetExternalToken", int32(user)); err != nil {
		return nil, err
	}
	var tok auth.ExternalAuthToken
	err := dbh(ctx).SelectOne(&tok, `SELECT * FROM ext_auth_token WHERE "user"=$1 AND "host"=$2 AND client_id=$3`, user, host, clientID)
	if err == sql.ErrNoRows {
		return nil, auth.ErrNoExternalAuthToken
	} else if err != nil {
		return nil, err
	}
	return &tok, nil
}

func (s *externalAuthTokens) SetUserToken(ctx context.Context, tok *auth.ExternalAuthToken) error {
	if tok.User == 0 {
		return errors.New("no uid specified")
	}
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "ExternalAuthTokens.SetExternalToken", int32(tok.User)); err != nil {
		return err
	}
	return dbutil.Transact(dbh(ctx), func(tx gorp.SqlExecutor) error {
		ctx = NewContext(ctx, tx)

		if _, err := s.GetUserToken(ctx, tok.User, tok.Host, tok.ClientID); err == auth.ErrNoExternalAuthToken {
			return tx.Insert(tok)
		} else if err != nil && err != auth.ErrExternalAuthTokenDisabled {
			return err
		}
		_, err := tx.Update(tok)
		return err
	})
}

func (s *externalAuthTokens) ListExternalUsers(ctx context.Context, extUIDs []int, host, clientID string) ([]*auth.ExternalAuthToken, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "ExternalAuthTokens.ListExternalUsers"); err != nil {
		return nil, err
	}
	if extUIDs == nil || len(extUIDs) == 0 {
		return []*auth.ExternalAuthToken{}, nil
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

	var toks []*auth.ExternalAuthToken
	if _, err := dbh(ctx).Select(&toks, sql, args...); err != nil {
		return nil, err
	}
	return toks, nil
}
