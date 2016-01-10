package pgsql

import (
	"database/sql"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/dbutil"
)

// directory is a DB-backed implementation of the Directory store.
type directory struct{}

var _ store.Directory = (*directory)(nil)

func (s *directory) GetUserByEmail(ctx context.Context, email string) (*sourcegraph.UserSpec, error) {
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
