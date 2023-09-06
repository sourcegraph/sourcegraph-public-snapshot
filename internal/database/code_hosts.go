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
	*LimitOffset

	// Only list code hosts with the given ID. This makes if effectively a getByID.
	ID int32
	// Only list code hosts with the given URL. This makes if effectively a getByURL.
	URL string
	// Cursor is used for pagination, it's the ID of the next code host to look at.
	Cursor int32
	// IncludeDeleted causes deleted code hosts to be returned as well. Note:
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
	Count(ctx context.Context, opts ListCodeHostsOpts) (int32, error)

	// GetByID gets the code host matching the specified ID.
	GetByID(ctx context.Context, id int32) (*types.CodeHost, error)
	// GetByURL gets the code host matching the specified url.
	GetByURL(ctx context.Context, url string) (*types.CodeHost, error)
	// List lists all code hosts matching the specified options.
	List(ctx context.Context, opts ListCodeHostsOpts) (chs []*types.CodeHost, next int32, err error)
	// Create creates a new code host in the db.
	//
	// If a code host with the given url already exists, it returns the existing code host.
	Create(ctx context.Context, ch *types.CodeHost) error
	// Update updates a code host, it uses the id field to match.
	Update(ctx context.Context, ch *types.CodeHost) error
	// Delete deletes the code host specified by the id.
	Delete(ctx context.Context, id int32) error
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

func (s *codeHostStore) GetByID(ctx context.Context, id int32) (*types.CodeHost, error) {
	chs, _, err := s.List(ctx, ListCodeHostsOpts{LimitOffset: &LimitOffset{Limit: 1}, ID: id})
	if err != nil {
		return nil, err
	}
	if len(chs) != 1 {
		return nil, errCodeHostNotFound{}
	}
	return chs[0], nil
}

func (s *codeHostStore) GetByURL(ctx context.Context, url string) (*types.CodeHost, error) {
	// We would normally parse the URL here to verify its valid, but some code hosts have connections that
	// have multiple URLs and in the code host table, they are represented by a code_hosts.url of:
	// python/gomodules/etc...
	chs, _, err := s.List(ctx, ListCodeHostsOpts{LimitOffset: &LimitOffset{Limit: 1}, URL: url})
	if err != nil {
		return nil, err
	}
	if len(chs) != 1 {
		return nil, errCodeHostNotFound{}
	}
	return chs[0], nil
}

func (s *codeHostStore) Create(ctx context.Context, ch *types.CodeHost) error {
	query := createCodeHostQuery(ch)
	row := s.QueryRow(ctx, query)
	return scan(row, ch)
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

func (s *codeHostStore) List(ctx context.Context, opts ListCodeHostsOpts) (chs []*types.CodeHost, next int32, err error) {
	query := listCodeHostsQuery(opts)
	chs, err = basestore.NewSliceScanner(scanCodeHosts)(s.Query(ctx, query))
	if err != nil || chs == nil {
		// Return an empty list in case of no results
		return []*types.CodeHost{}, 0, err
	}

	// Set the next value if we were able to fetch Limit + 1
	if opts.LimitOffset != nil && opts.Limit > 0 && len(chs) == opts.Limit+1 {
		next = chs[len(chs)-1].ID
		chs = chs[:len(chs)-1]
	}
	return
}

func listCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	// We fetch an extra one so that we have the `next` value
	newLimitOffset := opts.LimitOffset
	if newLimitOffset != nil && newLimitOffset.Limit > 0 {
		newLimitOffset.Limit += 1
	}

	return sqlf.Sprintf(listCodeHostsQueryFmtstr, sqlf.Join(codeHostColumnExpressions, ","), listCodeHostsWhereQuery(opts), newLimitOffset.SQL())
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

func (s *codeHostStore) Update(ctx context.Context, ch *types.CodeHost) error {
	query := updateCodeHostQuery(ch)

	row := s.QueryRow(ctx, query)
	err := scan(row, ch)
	if err != nil {
		if err == sql.ErrNoRows {
			return errCodeHostNotFound{ch.ID}
		}
		return err
	}
	return nil
}

func updateCodeHostQuery(ch *types.CodeHost) *sqlf.Query {
	return sqlf.Sprintf(
		updateCodeHostQueryFmtstr,
		ch.Kind,
		ch.URL,
		ch.APIRateLimitQuota,
		ch.APIRateLimitIntervalSeconds,
		ch.GitRateLimitQuota,
		ch.GitRateLimitIntervalSeconds,
		ch.ID,
		sqlf.Join(codeHostColumnExpressions, ","),
	)
}

const updateCodeHostQueryFmtstr = `
UPDATE code_hosts
	SET
		kind = %s,
		url = %s,
		api_rate_limit_quota = %s,
		api_rate_limit_interval_seconds = %s,
		git_rate_limit_quota = %s,
		git_rate_limit_interval_seconds = %s
	WHERE
		id = %s
	RETURNING
		%s`

func (s *codeHostStore) Delete(ctx context.Context, id int32) error {
	query := deleteCodeHostQuery(id)

	row := s.QueryRow(ctx, query)

	if err := row.Err(); err != nil {
		return err
	}
	return nil
}

func deleteCodeHostQuery(id int32) *sqlf.Query {
	return sqlf.Sprintf(
		deleteCodeHostQueryFmtstr,
		id,
	)
}

const deleteCodeHostQueryFmtstr = `
DELETE FROM code_hosts
WHERE
	id = %s
`

func (e *codeHostStore) Count(ctx context.Context, opts ListCodeHostsOpts) (int32, error) {
	q := countCodeHostsQuery(opts)
	count, _, err := basestore.ScanFirstInt(e.Query(ctx, q))
	return int32(count), err
}

func countCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	return sqlf.Sprintf(countCodeHostsQueryFmtstr, listCodeHostsWhereQuery(opts))
}

const countCodeHostsQueryFmtstr = `
SELECT
	count(*)
FROM
	code_hosts
WHERE
	%s
`

func listCodeHostsWhereQuery(opts ListCodeHostsOpts) *sqlf.Query {
	conds := []*sqlf.Query{}

	if !opts.IncludeDeleted {
		// Don't show code hosts for which all external services are soft-deleted or hard-deleted.
		conds = append(conds, sqlf.Sprintf("(EXISTS (SELECT 1 FROM external_services WHERE external_services.code_host_id = code_hosts.id AND deleted_at IS NULL))"))
	}

	if opts.ID > 0 {
		conds = append(conds, sqlf.Sprintf("code_hosts.id = %s", opts.ID))
	}

	if opts.URL != "" {
		conds = append(conds, sqlf.Sprintf("code_hosts.url = %s", opts.URL))
	}

	if opts.Cursor > 0 {
		conds = append(conds, sqlf.Sprintf("code_hosts.id >= %s", opts.Cursor))
	}

	if opts.Search != "" {
		conds = append(conds, sqlf.Sprintf("(code_hosts.kind ILIKE %s OR code_hosts.url ILIKE %s)", "%"+opts.Search+"%", "%"+opts.Search+"%"))
	}
	if len(conds) == 0 {
		return sqlf.Sprintf("TRUE")
	}

	return sqlf.Join(conds, "AND")
}

func scanCodeHosts(s dbutil.Scanner) (*types.CodeHost, error) {
	var ch types.CodeHost
	err := scan(s, &ch)
	if err != nil {
		return &types.CodeHost{}, err
	}
	return &ch, nil
}

func scan(s dbutil.Scanner, ch *types.CodeHost) error {
	return s.Scan(
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
