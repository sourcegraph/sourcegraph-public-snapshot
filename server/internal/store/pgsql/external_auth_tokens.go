package pgsql

import (
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
		`CREATE UNIQUE INDEX ext_auth_token_token_host ON ext_auth_token(token, host, client_id);`,
		`ALTER TABLE ext_auth_token ALTER COLUMN first_auth_failure_at TYPE timestamp with time zone USING first_auth_failure_at::timestamp with time zone;`,
	)
}

// ExternalAuthTokens is a DB-backed implementation of the ExternalAuthTokens store.
type ExternalAuthTokens struct{}

var _ store.ExternalAuthTokens = (*ExternalAuthTokens)(nil)

func (s *ExternalAuthTokens) GetUserToken(ctx context.Context, user int, host, clientID string) (*auth.ExternalAuthToken, error) {
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

func (s *ExternalAuthTokens) SetUserToken(ctx context.Context, tok *auth.ExternalAuthToken) error {
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
