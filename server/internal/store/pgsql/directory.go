package pgsql

import (
	"database/sql"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil"
)

// Directory is a DB-backed implementation of the Directory store.
type Directory struct{}

var _ store.Directory = (*Directory)(nil)

func (s *Directory) GetUserByEmail(ctx context.Context, email string) (*sourcegraph.UserSpec, error) {
	q := `SELECT uid FROM user_email WHERE (NOT blacklisted) AND email=$1 ORDER BY uid ASC LIMIT 1`
	uid, err := dbutil.SelectInt(dbh(ctx), q, email)
	switch {
	case err == sql.ErrNoRows:
		return nil, &store.UserNotFoundError{Login: "email=" + email}
	case err != nil:
		return nil, err
	default:
		return &sourcegraph.UserSpec{UID: int32(uid)}, nil
	}
}
