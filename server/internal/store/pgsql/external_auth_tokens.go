package pgsql

import (
	"fmt"
	"strings"

	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/dbutil"
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
	var toks []*auth.ExternalAuthToken
	err := dbh(ctx).Select(&toks, `SELECT * FROM ext_auth_token WHERE "user"=$1 AND "host"=$2 AND client_id=$3`, user, host, clientID)
	if err != nil {
		return nil, err
	}
	if len(toks) == 0 {
		return nil, auth.ErrNoExternalAuthToken
	}
	return toks[0], nil
}

func (s *externalAuthTokens) SetUserToken(ctx context.Context, tok *auth.ExternalAuthToken) error {
	return dbutil.Transact(dbh(ctx), func(tx modl.SqlExecutor) error {
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
	var args []interface{}
	arg := func(a interface{}) string {
		v := modl.PostgresDialect{}.BindVar(len(args))
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
	err := dbh(ctx).Select(&toks, sql, args...)
	if err != nil {
		return nil, err
	}
	return toks, nil
}
