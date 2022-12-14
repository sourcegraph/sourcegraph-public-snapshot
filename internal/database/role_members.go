package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var roleMemberInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("role_id"),
}

type RoleMemberStore interface {
	basestore.ShareableStore

	Create(ctx context.Context, roleID, userID int32) (*types.RoleMembership, error)
	Transact(context.Context) (RoleMemberStore, error)
	With(basestore.ShareableStore) RoleMemberStore
	// Remove(ctx context.Context, roleID, userID int32) error
	// GetByRoleID(ctx context.Context, roleID int32) ([]*types.RoleMembership, error)
	// GetByUserID(ctx context.Context, userID int32) ([]*types.RoleMembership, error)
	// GetByRoleIDAndUserID(ctx context.Context, roleID, userID int32) ([]*types.RoleMembership, error)
}

type roleMemberStore struct {
	*basestore.Store
}

func RoleMembersWith(other basestore.ShareableStore) RoleMemberStore {
	return &roleMemberStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (r *roleMemberStore) With(other basestore.ShareableStore) RoleMemberStore {
	return &roleMemberStore{Store: r.Store.With(other)}
}

func (r *roleMemberStore) Transact(ctx context.Context) (RoleMemberStore, error) {
	tx, err := r.Store.Transact(ctx)
	return &roleMemberStore{Store: tx}, err
}

const roleMemberCreateQueryFmtStr = `
INSERT INTO
	user_roles (%s)
	VALUES (
		%s,
		%s
	)
	RETURNING %s;
`

func (r *roleMemberStore) Create(ctx context.Context, roleID, userID int32) (*types.RoleMembership, error) {
	q := sqlf.Sprintf(
		roleMemberCreateQueryFmtStr,
		userID,
		roleID,
		sqlf.Join(roleMemberInsertColumns, ", "),
	)

	rm, err := scanRoleMember(r.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning user role")
	}
	return rm, nil
}

func scanRoleMember(sc dbutil.Scanner) (*types.RoleMembership, error) {
	var rm types.RoleMembership
	if err := sc.Scan(
		&rm.RoleID,
		&rm.UserID,
		&rm.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &rm, nil
}
