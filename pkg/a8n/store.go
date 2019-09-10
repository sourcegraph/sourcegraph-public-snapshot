package a8n

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
)

// Store exposes methods to read and write a8n domain models
// from persistent storage.
type Store struct {
	db  dbutil.DB
	now func() time.Time
}

// NewStore returns a new Store backed by the given db.
func NewStore(db dbutil.DB) *Store {
	return &Store{db: db, now: func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}}
}

// Transact returns a Store whose methods operate within the context of a transaction.
// This method will return an error if the underlying DB cannot be interface upgraded
// to a TxBeginner.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	if _, ok := s.db.(dbutil.Tx); ok { // Already in a Tx.
		return nil, errors.New("store: already in a transaction")
	}

	tb, ok := s.db.(dbutil.TxBeginner)
	if !ok { // Not a Tx nor a TxBeginner, error.
		return nil, errors.New("store: not transactable")
	}

	tx, err := tb.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "store: BeginTx")
	}

	return &Store{db: tx, now: s.now}, nil
}

// Done terminates the underlying Tx in a Store either by committing or rolling
// back based on the value pointed to by the first given error pointer.
// It's a no-op if the `Store` is not operating within a transaction,
// which can only be done via `Transact`.
//
// When the error value pointed to by the first given `err` is nil, or when no error
// pointer is given, the transaction is commited. Otherwise, it's rolled-back.
func (s *Store) Done(errs ...*error) {
	switch tx, ok := s.db.(dbutil.Tx); {
	case !ok:
		return
	case len(errs) == 0:
		_ = tx.Commit()
	case errs[0] != nil && *errs[0] != nil:
		_ = tx.Rollback()
	default:
		_ = tx.Commit()
	}
}

// CreateThread creates the given Thread.
func (s *Store) CreateThread(ctx context.Context, t *Thread) error {
	q, err := s.createThreadQuery(t)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanThread(t, sc)
		return int64(t.ID), 1, err
	})
}

var createThreadQueryFmtstr = `
-- source: pkg/a8n/store.go:CreateThread
INSERT INTO threads (
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids
)
VALUES (%s, %s, %s, %s, %s)
RETURNING
	id,
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids
`

func (s *Store) createThreadQuery(t *Thread) (*sqlf.Query, error) {
	metadata, err := metadataColumn(t.Metadata)
	if err != nil {
		return nil, err
	}

	campaignIDs, err := jsonSetColumn(t.CampaignIDs)
	if err != nil {
		return nil, err
	}

	if t.CreatedAt.IsZero() {
		t.CreatedAt = s.now()
	}

	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}

	return sqlf.Sprintf(
		createThreadQueryFmtstr,
		t.RepoID,
		t.CreatedAt,
		t.UpdatedAt,
		metadata,
		campaignIDs,
	), nil
}

// CountThreadsOpts captures the query options needed for
// counting threads.
type CountThreadsOpts struct {
	CampaignID int64
}

// CountThreads returns the number of threads in the database.
func (s *Store) CountThreads(ctx context.Context, opts CountThreadsOpts) (count int64, _ error) {
	q := countThreadsQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countThreadsQueryFmtstr = `
-- source: pkg/a8n/store.go:ListThreads
SELECT COUNT(id)
FROM threads
WHERE %s
`

func countThreadsQuery(opts *CountThreadsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_ids ? %s", opts.CampaignID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countThreadsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListThreadsOpts captures the query options needed for
// listing threads.
type ListThreadsOpts struct {
	Limit      int
	CampaignID int64
}

// ListThreads lists Threads with the given filters.
func (s *Store) ListThreads(ctx context.Context, opts ListThreadsOpts) (cs []*Thread, next int64, err error) {
	q := listThreadsQuery(&opts)

	cs = make([]*Thread, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c Thread
		if err = scanThread(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return int64(c.ID), 1, err
	})

	if len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listThreadsQueryFmtstr = `
-- source: pkg/a8n/store.go:ListThreads
SELECT
	id,
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids
FROM threads
WHERE %s
ORDER BY id ASC
LIMIT %s
`

const defaultListLimit = 50

func listThreadsQuery(opts *ListThreadsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var preds []*sqlf.Query
	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_ids ? %s", opts.CampaignID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listThreadsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

// CreateCampaign creates the given Campaign.
func (s *Store) CreateCampaign(ctx context.Context, c *Campaign) error {
	q, err := s.createCampaignQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaign(c, sc)
		return int64(c.ID), 1, err
	})
}

var createCampaignQueryFmtstr = `
-- source: pkg/a8n/store.go:CreateCampaign
INSERT INTO campaigns (
	name,
	description,
	author_id,
	namespace_user_id,
	namespace_org_id,
	created_at,
	updated_at,
	thread_ids
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
	id,
	name,
	description,
	author_id,
	namespace_user_id,
	namespace_org_id,
	created_at,
	updated_at,
	thread_ids
`

func (s *Store) createCampaignQuery(c *Campaign) (*sqlf.Query, error) {
	threadIDs, err := jsonSetColumn(c.ThreadIDs)
	if err != nil {
		return nil, err
	}

	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return sqlf.Sprintf(
		createCampaignQueryFmtstr,
		c.Name,
		c.Description,
		c.AuthorID,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		c.CreatedAt,
		c.UpdatedAt,
		threadIDs,
	), nil
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

// UpdateCampaign updates the given Campaign.
func (s *Store) UpdateCampaign(ctx context.Context, c *Campaign) error {
	q, err := s.updateCampaignQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaign(c, sc)
		return int64(c.ID), 1, err
	})
}

var updateCampaignQueryFmtstr = `
-- source: pkg/a8n/store.go:UpdateCampaign
UPDATE campaigns
SET (
	name,
	description,
	author_id,
	namespace_user_id,
	namespace_org_id,
	updated_at,
	thread_ids
) = (%s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
	id,
	name,
	description,
	author_id,
	namespace_user_id,
	namespace_org_id,
	created_at,
	updated_at,
	thread_ids
`

func (s *Store) updateCampaignQuery(c *Campaign) (*sqlf.Query, error) {
	threadIDs, err := jsonSetColumn(c.ThreadIDs)
	if err != nil {
		return nil, err
	}

	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateCampaignQueryFmtstr,
		c.Name,
		c.Description,
		c.AuthorID,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		c.UpdatedAt,
		threadIDs,
		c.ID,
	), nil
}

// CountCampaigns returns the number of campaigns in the database.
func (s *Store) CountCampaigns(ctx context.Context) (count int64, _ error) {
	q := countCampaignsQuery
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countCampaignsQuery = sqlf.Sprintf(`
-- source: pkg/a8n/store.go:CountCampaigns
SELECT COUNT(id) FROM campaigns
`)

// ListCampaignsOpts captures the query options needed for
// listing campaigns.
type ListCampaignsOpts struct {
	Limit int
}

// ListCampaigns lists Campaigns with the given filters.
func (s *Store) ListCampaigns(ctx context.Context, opts ListCampaignsOpts) (cs []*Campaign, next int64, err error) {
	q := listCampaignsQuery(&opts)

	cs = make([]*Campaign, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c Campaign
		if err = scanCampaign(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return int64(c.ID), 1, err
	})

	if len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listCampaignsQueryFmtstr = `
-- source: pkg/a8n/store.go:ListCampaigns
SELECT
	id,
	name,
	description,
	author_id,
	namespace_user_id,
	namespace_org_id,
	created_at,
	updated_at,
	thread_ids
FROM campaigns
ORDER BY id ASC
LIMIT %s
`

func listCampaignsQuery(opts *ListCampaignsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++
	return sqlf.Sprintf(listCampaignsQueryFmtstr, opts.Limit)
}

func (s *Store) exec(ctx context.Context, q *sqlf.Query, sc scanFunc) error {
	_, _, err := s.query(ctx, q, sc)
	return err
}

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scanFunc) (last, count int64, err error) {
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return 0, 0, err
	}
	return scanAll(rows, sc)
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...interface{}) error
}

// a scanFunc scans one or more rows from a scanner, returning
// the last id column scanned and the count of scanned rows.
type scanFunc func(scanner) (last, count int64, err error)

func scanAll(rows *sql.Rows, scan scanFunc) (last, count int64, err error) {
	defer closeErr(rows, &err)

	last = -1
	for rows.Next() {
		var n int64
		if last, n, err = scan(rows); err != nil {
			return last, count, err
		}
		count += n
	}

	return last, count, rows.Err()
}

func closeErr(c io.Closer, err *error) {
	if e := c.Close(); err != nil && *err == nil {
		*err = e
	}
}

func scanThread(t *Thread, s scanner) error {
	t.Metadata = json.RawMessage{}
	return s.Scan(
		&t.ID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Metadata,
		&dbutil.JSONInt64Set{Set: &t.CampaignIDs},
	)
}

func scanCampaign(c *Campaign, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.Name,
		&c.Description,
		&c.AuthorID,
		&dbutil.NullInt32{N: &c.NamespaceUserID},
		&dbutil.NullInt32{N: &c.NamespaceOrgID},
		&c.CreatedAt,
		&c.UpdatedAt,
		&dbutil.JSONInt64Set{Set: &c.ThreadIDs},
	)
}

func metadataColumn(metadata interface{}) (msg json.RawMessage, err error) {
	switch m := metadata.(type) {
	case nil:
		msg = json.RawMessage("{}")
	case string:
		msg = json.RawMessage(m)
	case []byte:
		msg = m
	case json.RawMessage:
		msg = m
	default:
		msg, err = json.MarshalIndent(m, "        ", "    ")
	}
	return
}

func jsonSetColumn(ids []int64) ([]byte, error) {
	set := make(map[int64]*struct{}, len(ids))
	for _, id := range ids {
		set[id] = nil
	}
	return json.Marshal(set)
}
