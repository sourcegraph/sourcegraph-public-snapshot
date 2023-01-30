package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var namespacePermissionColumns = []*sqlf.Query{
	sqlf.Sprintf("namespace_permissions.id"),
	sqlf.Sprintf("namespace_permissions.namespace"),
	sqlf.Sprintf("namespace_permissions.resource_id"),
	sqlf.Sprintf("namespace_permissions.action"),
	sqlf.Sprintf("namespace_permissions.user_id"),
	sqlf.Sprintf("namespace_permissions.created_at"),
}

var namespacePermissionInsertColums = []*sqlf.Query{
	sqlf.Sprintf("namespace"),
	sqlf.Sprintf("resource_id"),
	sqlf.Sprintf("action"),
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
		%s,
		%s
	)
	RETURNING %s
`

func (n *namespacePermissionStore) Create(ctx context.Context, opts CreateNamespacePermissionOpts) (*types.NamespacePermission, error) {
	q := sqlf.Sprintf(
		namespacePermissionCreateQueryFmtStr,
		sqlf.Join(namespacePermissionInsertColums, ", "),
		opts.Namespace,
		opts.ResourceID,
		opts.Action,
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

func (n *namespacePermissionStore) Delete(ctx context.Context, opts DeleteNamespacePermissionOpts) error {
	if opts.ID == 0 {
		return errors.New("missing namespace permission id")
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
SELECT %s FROM namespace_permissions WHERE id = %s
`

func (n *namespacePermissionStore) Get(ctx context.Context, opts GetNamespacePermissionOpts) (*types.NamespacePermission, error) {
	if opts.ID == 0 {
		return nil, errors.New("missing namespace permission id")
	}

	q := sqlf.Sprintf(
		namespacePermissionGetQueryFmtStr,
		sqlf.Join(namespacePermissionColumns, ", "),
		opts.ID,
	)

	np, err := scanNamespacePermission(n.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NamespacePermissionNotFoundErr{ID: opts.ID}
		}
		return nil, errors.Wrap(err, "scanning role")
	}

	return np, nil
}

func scanNamespacePermission(sc dbutil.Scanner) (*types.NamespacePermission, error) {
	var np types.NamespacePermission
	if err := sc.Scan(
		&np.ID,
		&np.Namespace,
		&np.ResourceID,
		&np.Action,
		&np.UserID,
		&np.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &np, nil
}

type GetNamespacePermissionOpts struct {
	ID int64
}

type DeleteNamespacePermissionOpts struct {
	ID int64
}

type CreateNamespacePermissionOpts struct {
	Namespace  string
	ResourceID int64
	Action     string
	UserID     int32
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
