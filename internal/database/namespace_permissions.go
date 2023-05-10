package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var namespacePermissionColumns = []*sqlf.Query{
	sqlf.Sprintf("namespace_permissions.id"),
	sqlf.Sprintf("namespace_permissions.namespace"),
	sqlf.Sprintf("namespace_permissions.resource_id"),
	sqlf.Sprintf("namespace_permissions.user_id"),
	sqlf.Sprintf("namespace_permissions.created_at"),
}

var namespacePermissionInsertColums = []*sqlf.Query{
	sqlf.Sprintf("namespace"),
	sqlf.Sprintf("resource_id"),
	sqlf.Sprintf("user_id"),
}

type NamespacePermissionStore interface {
	basestore.ShareableStore

	// Create inserts the given namespace permission into the database.
	Create(context.Context, CreateNamespacePermissionOpts) (*types.NamespacePermission, error)
	// Deletes removes an existing namespace permission from the database.
	Delete(context.Context, DeleteNamespacePermissionOpts) error
	// Get returns the NamespacePermission matching the ID provided in the options.
	Get(context.Context, GetNamespacePermissionOpts) (*types.NamespacePermission, error)
}

type namespacePermissionStore struct {
	*basestore.Store
}

func NamespacePermissionsWith(other basestore.ShareableStore) NamespacePermissionStore {
	return &namespacePermissionStore{Store: basestore.NewWithHandle(other.Handle())}
}

var _ NamespacePermissionStore = &namespacePermissionStore{}

const namespacePermissionCreateQueryFmtStr = `
INSERT INTO
	namespace_permissions (%s)
	VALUES (
		%s,
		%s,
		%s
	)
	RETURNING %s
`

type CreateNamespacePermissionOpts struct {
	Namespace  rtypes.PermissionNamespace
	ResourceID int64
	UserID     int32
}

func (n *namespacePermissionStore) Create(ctx context.Context, opts CreateNamespacePermissionOpts) (*types.NamespacePermission, error) {
	if opts.ResourceID == 0 {
		return nil, errors.New("resource id is required")
	}

	if opts.UserID == 0 {
		return nil, errors.New("user id is required")
	}

	if !opts.Namespace.Valid() {
		return nil, errors.New("valid namespace is required")
	}

	q := sqlf.Sprintf(
		namespacePermissionCreateQueryFmtStr,
		sqlf.Join(namespacePermissionInsertColums, ", "),
		opts.Namespace,
		opts.ResourceID,
		opts.UserID,
		sqlf.Join(namespacePermissionColumns, ", "),
	)

	np, err := scanNamespacePermission(n.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning namespace permission")
	}

	return np, nil
}

const namespacePermissionDeleteQueryFmtStr = `
DELETE FROM namespace_permissions
WHERE id = %s
`

type DeleteNamespacePermissionOpts struct {
	ID int64
}

func (n *namespacePermissionStore) Delete(ctx context.Context, opts DeleteNamespacePermissionOpts) error {
	if opts.ID == 0 {
		return errors.New("namespace permission id is required")
	}

	q := sqlf.Sprintf(
		namespacePermissionDeleteQueryFmtStr,
		opts.ID,
	)

	result, err := n.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrap(&NamespacePermissionNotFoundErr{ID: opts.ID}, "failed to delete namespace permission")
	}

	return nil
}

const namespacePermissionGetQueryFmtStr = `
SELECT %s FROM namespace_permissions WHERE %s
`

// When querying namespace permissions, you need to provide one of the following
// 1. The ID belonging to the namespace to be retrieved.
// 2. The Namespace, ResourceID and UserID associated with the namespace permission.
type GetNamespacePermissionOpts struct {
	ID         int64
	Namespace  rtypes.PermissionNamespace
	ResourceID int64
	UserID     int32
}

func (n *namespacePermissionStore) Get(ctx context.Context, opts GetNamespacePermissionOpts) (*types.NamespacePermission, error) {
	if !isGetNamsepaceOptsValid(opts) {
		return nil, errors.New("missing namespace permission query")
	}

	var conds []*sqlf.Query
	if opts.ID != 0 {
		conds = append(conds, sqlf.Sprintf("id = %s", opts.ID))
	} else {
		conds = append(conds, sqlf.Sprintf("namespace = %s", opts.Namespace))
		conds = append(conds, sqlf.Sprintf("user_id = %s", opts.UserID))
		conds = append(conds, sqlf.Sprintf("resource_id = %s", opts.ResourceID))
	}

	q := sqlf.Sprintf(
		namespacePermissionGetQueryFmtStr,
		sqlf.Join(namespacePermissionColumns, ", "),
		sqlf.Join(conds, " AND "),
	)

	np, err := scanNamespacePermission(n.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NamespacePermissionNotFoundErr{ID: opts.ID}
		}
		return nil, errors.Wrap(err, "scanning namespace permission")
	}

	return np, nil
}

// isGetNamsepaceOptsValid is used to validate the options passed into the `namespacePermissionStore.Get` method are valid.
// One of the conditions below need to be valid to execute a Get query.
// 1. ID is provided
// 2. Namespace, UserID and ResourceID is provided.
func isGetNamsepaceOptsValid(opts GetNamespacePermissionOpts) bool {
	areNonIDOptsValid := opts.Namespace.Valid() && opts.UserID != 0 && opts.ResourceID != 0
	if areNonIDOptsValid || opts.ID != 0 {
		return true
	}
	return false
}

func scanNamespacePermission(sc dbutil.Scanner) (*types.NamespacePermission, error) {
	var np types.NamespacePermission
	if err := sc.Scan(
		&np.ID,
		&np.Namespace,
		&np.ResourceID,
		&np.UserID,
		&np.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &np, nil
}

type NamespacePermissionNotFoundErr struct {
	ID int64
}

func (e *NamespacePermissionNotFoundErr) Error() string {
	return fmt.Sprintf("namespace permission with id %d not found", e.ID)
}

func (e *NamespacePermissionNotFoundErr) NotFound() bool {
	return true
}
