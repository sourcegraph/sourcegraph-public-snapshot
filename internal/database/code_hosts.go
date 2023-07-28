package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ListCodeHostsOpts struct {
	LimitOffset

	// Only list code hosts with the given ID. This makes if effectively a getByID.
	ID int32
	// Cursor is used for pagination, it's the ID of the next code host to look at.
	Cursor int32
	// IncludeDeleted makes to so that we return deleted code hosts as well. Note:
	// For now deletion is a virtual concept where we check if all referencing
	// external services are soft or hard deleted.
	IncludeDeleted bool
	// Search is an optional string to search through kind and URL.
	Search string
}

// CodeHostStore provides access to the code_hosts table.
type CodeHostStore interface {
	basestore.ShareableStore

	With(other basestore.ShareableStore) CodeHostStore
	WithTransact(context.Context, func(CodeHostStore) error) error

	GetCodeHostByID(ctx context.Context, id int32) (*types.CodeHost, error)
	ListCodeHosts(ctx context.Context, opts ListCodeHostsOpts) (chs []*types.CodeHost, next int32, err error)
	CreateCodeHost(ctx context.Context, ch *types.CodeHost) error
}

// CodeHostsWith instantiates and returns a new CodeHostStore using the other stores
// handle.
func CodeHostsWith(other basestore.ShareableStore) CodeHostStore {
	return &codeHostStore{Store: basestore.NewWithHandle(other.Handle())}
}

type codeHostStore struct {
	*basestore.Store
}

func (s *codeHostStore) With(other basestore.ShareableStore) CodeHostStore {
	return &codeHostStore{Store: s.Store.With(other)}
}

func (s *codeHostStore) copy() *codeHostStore {
	return &codeHostStore{
		Store: s.Store,
	}
}

func (s *codeHostStore) WithTransact(ctx context.Context, f func(CodeHostStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		c := s.copy()
		c.Store = tx
		return f(c)
	})
}

var codeHostColumnExpressions = []*sqlf.Query{
	sqlf.Sprintf("code_hosts.id"),
	sqlf.Sprintf("code_hosts.kind"),
	sqlf.Sprintf("code_hosts.url"),
	sqlf.Sprintf("code_hosts.api_rate_limit_quota"),
	sqlf.Sprintf("code_hosts.api_rate_limit_interval_seconds"),
	sqlf.Sprintf("code_hosts.git_rate_limit_quota"),
	sqlf.Sprintf("code_hosts.git_rate_limit_interval_seconds"),
	sqlf.Sprintf("code_hosts.created_at"),
	sqlf.Sprintf("code_hosts.updated_at"),
}

type errCodeHostNotFound struct {
	id int32
}

func (e errCodeHostNotFound) Error() string {
	if e.id != 0 {
		return fmt.Sprintf("code host with id %d not found", e.id)
	}
	return "code host not found"
}

func (errCodeHostNotFound) NotFound() bool {
	return true
}

func (s *codeHostStore) GetCodeHostByID(ctx context.Context, id int32) (*types.CodeHost, error) {
	chs, _, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{LimitOffset: LimitOffset{Limit: 1}, ID: id})
	if err != nil {
		return nil, err
	}
	if len(chs) != 1 {
		return nil, errCodeHostNotFound{}
	}
	return chs[0], nil
}

func (s *codeHostStore) CreateCodeHost(ctx context.Context, ch *types.CodeHost) error {
	query := createCodeHostQuery(ch)

	row := s.QueryRow(ctx, query)
	if err := scanCodeHost(row, ch); err != nil {
		if err == sql.ErrNoRows {
			// TODO: ch should still contain the latest database values when we return here.
			return nil
		}
		return err
	}

	return nil
}

func createCodeHostQuery(ch *types.CodeHost) *sqlf.Query {
	return sqlf.Sprintf(
		createCodeHostQueryFmtstr,
		ch.Kind,
		ch.URL,
		ch.APIRateLimitQuota,
		ch.APIRateLimitIntervalSeconds,
		ch.GitRateLimitQuota,
		ch.GitRateLimitIntervalSeconds,
		sqlf.Join(codeHostColumnExpressions, ","),
		sqlf.Join(codeHostColumnExpressions, ","),
		ch.URL,
	)
}

const createCodeHostQueryFmtstr = `
WITH inserted AS (
	INSERT INTO
		code_hosts (kind, url, api_rate_limit_quota, api_rate_limit_interval_seconds, git_rate_limit_quota, git_rate_limit_interval_seconds)
	VALUES (%s, %s, %s, %s, %s, %s)
	ON CONFLICT(url) DO NOTHING
	RETURNING
		%s
)
SELECT * FROM inserted

UNION
SELECT
	%s
FROM code_hosts
WHERE url = %s
`

func (s *codeHostStore) ListCodeHosts(ctx context.Context, opts ListCodeHostsOpts) (chs []*types.CodeHost, next int32, err error) {
	query := listCodeHostsQuery(opts)

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var ch types.CodeHost
		err = scanCodeHost(rows, &ch)
		if err != nil {
			return nil, 0, err
		}

		chs = append(chs, &ch)
	}

	return
}

func listCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	conds := []*sqlf.Query{}

	if !opts.IncludeDeleted {
		// Don't show code hosts for which all external services are soft-deleted or hard-deleted.
		conds = append(conds, sqlf.Sprintf("(EXISTS (SELECT 1 FROM external_services WHERE external_services.code_host_id = code_hosts.id AND deleted_at IS NULL))"))
	}

	if opts.ID > 0 {
		conds = append(conds, sqlf.Sprintf("code_hosts.id = %s", opts.ID))
	}

	if opts.Cursor > 0 {
		conds = append(conds, sqlf.Sprintf("code_hosts.id >= %s", opts.Cursor))
	}

	if opts.Search != "" {
		conds = append(conds, sqlf.Sprintf("(code_hosts.kind ILIKE %s OR code_hosts.url ILIKE %s)", "%"+opts.Search+"%", "%"+opts.Search+"%"))
	}

	return sqlf.Sprintf(listCodeHostsQueryFmtstr, sqlf.Join(codeHostColumnExpressions, ","), sqlf.Join(conds, "AND"), opts.LimitOffset.SQL())
}

const listCodeHostsQueryFmtstr = `
SELECT
	%s
FROM
	code_hosts
WHERE
	%s
ORDER BY id ASC
%s
`

func scanCodeHost(rows dbutil.Scanner, ch *types.CodeHost) error {
	return rows.Scan(
		&ch.ID,
		&ch.Kind,
		&ch.URL,
		&ch.APIRateLimitQuota,
		&ch.APIRateLimitIntervalSeconds,
		&ch.GitRateLimitQuota,
		&ch.GitRateLimitIntervalSeconds,
		&dbutil.NullTime{Time: &ch.CreatedAt},
		&dbutil.NullTime{Time: &ch.UpdatedAt},
	)
}
