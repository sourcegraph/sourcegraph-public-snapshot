package localstore

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

// directory is a DB-backed implementation of the Directory store.
type directory struct{}

var _ store.Directory = (*directory)(nil)

func (s *directory) GetUserByEmail(ctx context.Context, email string) (*sourcegraph.UserSpec, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Directory.GetUserByEmail"); err != nil {
		return nil, err
	}
	q := `SELECT uid FROM user_email WHERE (NOT blacklisted) AND email=$1 ORDER BY uid ASC LIMIT 1`
	uid, err := appDBH(ctx).SelectInt(q, email)
	switch {
	case err != nil:
		return nil, err
	case uid == 0:
		return nil, &store.UserNotFoundError{Login: "email=" + email}
	default:
		return &sourcegraph.UserSpec{UID: int32(uid)}, nil
	}
}
