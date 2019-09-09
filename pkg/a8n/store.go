package a8n

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
)

// Store exposes methods to read and write a8n domain models
// from persistent storage.
type Store struct {
	db dbutil.DB
}

// NewStore returns a new Store backed by the given db.
func NewStore(db dbutil.DB) *Store {
	return &Store{db: db}
}

// CreateThread creates the given Thread.
func (s *Store) CreateThread(ctx context.Context, t *Thread) error {
	q, err := createThreadQuery(t)
	if err != nil {
		return err
	}

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	_, _, err = scanAll(rows, func(sc scanner) (last, count int64, err error) {
		err = scanThread(t, sc)
		return int64(t.ID), 1, err
	})

	return err
}

var createThreadQueryFmtstr = `
-- source: pkg/a8n/store.go:CreateThread
INSERT INTO threads (
	campaign_id,
	repo_id,
	created_at,
	updated_at,
	metadata
)
VALUES (%s, %s, %s, %s, %s)
RETURNING
	id,
	campaign_id,
	repo_id,
	created_at,
	updated_at,
	metadata
`

func createThreadQuery(t *Thread) (*sqlf.Query, error) {
	metadata, err := metadataColumn(t.Metadata)
	if err != nil {
		return nil, err
	}

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	}

	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}

	return sqlf.Sprintf(
		createThreadQueryFmtstr,
		t.CampaignID,
		t.RepoID,
		t.CreatedAt,
		t.UpdatedAt,
		metadata,
	), nil
}

// CountThreadsOpts captures the query options needed for
// counting threads.
type CountThreadsOpts struct {
	CampaignID int64
}

// CountThreads returns the number of threads in the database.
func (s *Store) CountThreads(ctx context.Context, opts CountThreadsOpts) (int64, error) {
	q := countThreadsQuery(&opts)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return 0, err
	}

	_, count, err := scanAll(rows, func(sc scanner) (_, count int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})

	return count, err
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
		preds = append(preds, sqlf.Sprintf("campaign_id = %s", opts.CampaignID))
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

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, 0, err
	}

	cs = make([]*Thread, 0, opts.Limit)
	_, _, err = scanAll(rows, func(sc scanner) (last, count int64, err error) {
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
	campaign_id,
	repo_id,
	created_at,
	updated_at,
	metadata
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
		preds = append(preds, sqlf.Sprintf("campaign_id = %s", opts.CampaignID))
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
	q := createCampaignQuery(c)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	_, _, err = scanAll(rows, func(sc scanner) (last, count int64, err error) {
		err = scanCampaign(c, sc)
		return int64(c.ID), 1, err
	})

	return err
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
	updated_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING
	id,
	name,
	description,
	author_id,
	namespace_user_id,
	namespace_org_id,
	created_at,
	updated_at
`

func createCampaignQuery(c *Campaign) *sqlf.Query {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
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
	)
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

// UpdateCampaign updates the given Campaign.
func (s *Store) UpdateCampaign(ctx context.Context, c *Campaign) error {
	panic("not implemented")
}

// CountCampaigns returns the number of campaigns in the database.
func (s *Store) CountCampaigns(ctx context.Context) (int64, error) {
	q := countCampaignsQuery

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return 0, err
	}

	_, count, err := scanAll(rows, func(sc scanner) (_, count int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})

	return count, err
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

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, 0, err
	}

	cs = make([]*Campaign, 0, opts.Limit)
	_, _, err = scanAll(rows, func(sc scanner) (last, count int64, err error) {
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
	updated_at
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
		&t.CampaignID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Metadata,
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
