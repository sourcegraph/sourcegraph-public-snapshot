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

// DB returns the underlying dbutil.DB that this Store was
// instantiated with.
func (s *Store) DB() dbutil.DB { return s.db }

// CreateChangeset creates the given Changeset.
func (s *Store) CreateChangeset(ctx context.Context, t *Changeset) error {
	q, err := s.createChangesetQuery(t)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanChangeset(t, sc)
		return int64(t.ID), 1, err
	})
}

var createChangesetQueryFmtstr = `
-- source: pkg/a8n/store.go:CreateChangeset
INSERT INTO changesets (
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids,
	external_id
)
VALUES (%s, %s, %s, %s, %s, %s)
RETURNING
	id,
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids,
	external_id
`

func (s *Store) createChangesetQuery(t *Changeset) (*sqlf.Query, error) {
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
		createChangesetQueryFmtstr,
		t.RepoID,
		t.CreatedAt,
		t.UpdatedAt,
		metadata,
		campaignIDs,
		t.ExternalID,
	), nil
}

// CountChangesetsOpts captures the query options needed for
// counting changesets.
type CountChangesetsOpts struct {
	CampaignID int64
}

// CountChangesets returns the number of changesets in the database.
func (s *Store) CountChangesets(ctx context.Context, opts CountChangesetsOpts) (count int64, _ error) {
	q := countChangesetsQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countChangesetsQueryFmtstr = `
-- source: pkg/a8n/store.go:CountChangesets
SELECT COUNT(id)
FROM changesets
WHERE %s
`

func countChangesetsQuery(opts *CountChangesetsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_ids ? %s", opts.CampaignID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countChangesetsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetChangesetOpts captures the query options needed for getting a Changeset
type GetChangesetOpts struct {
	ID int64
}

// ErrNoResults is returned by Store method calls that found no results.
var ErrNoResults = errors.New("no results")

// GetChangeset gets a changeset matching the given options.
func (s *Store) GetChangeset(ctx context.Context, opts GetChangesetOpts) (*Changeset, error) {
	q := getChangesetQuery(&opts)

	var c Changeset
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanChangeset(&c, sc)
	})

	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getChangesetsQueryFmtstr = `
-- source: pkg/a8n/store.go:GetChangeset
SELECT
	id,
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids,
	external_id
FROM changesets
WHERE %s
LIMIT 1
`

func getChangesetQuery(opts *GetChangesetOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getChangesetsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListChangesetsOpts captures the query options needed for
// listing changesets.
type ListChangesetsOpts struct {
	Limit      int
	CampaignID int64
}

// ListChangesets lists Changesets with the given filters.
func (s *Store) ListChangesets(ctx context.Context, opts ListChangesetsOpts) (cs []*Changeset, next int64, err error) {
	q := listChangesetsQuery(&opts)

	cs = make([]*Changeset, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c Changeset
		if err = scanChangeset(&c, sc); err != nil {
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

var listChangesetsQueryFmtstr = `
-- source: pkg/a8n/store.go:ListChangesets
SELECT
	id,
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids,
	external_id
FROM changesets
WHERE %s
ORDER BY id ASC
LIMIT %s
`

const defaultListLimit = 50

func listChangesetsQuery(opts *ListChangesetsOpts) *sqlf.Query {
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
		listChangesetsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

// UpdateChangeset updates the given Changeset.
func (s *Store) UpdateChangeset(ctx context.Context, c *Changeset) error {
	q, err := s.updateChangesetQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanChangeset(c, sc)
		return int64(c.ID), 1, err
	})
}

var updateChangesetQueryFmtstr = `
-- source: pkg/a8n/store.go:UpdateChangeset
UPDATE changesets
SET (
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids,
	external_id
) = (%s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
	id,
	repo_id,
	created_at,
	updated_at,
	metadata,
	campaign_ids,
	external_id
`

func (s *Store) updateChangesetQuery(c *Changeset) (*sqlf.Query, error) {
	metadata, err := metadataColumn(c.Metadata)
	if err != nil {
		return nil, err
	}

	campaignIDs, err := jsonSetColumn(c.CampaignIDs)
	if err != nil {
		return nil, err
	}

	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateChangesetQueryFmtstr,
		c.RepoID,
		c.CreatedAt,
		c.UpdatedAt,
		metadata,
		campaignIDs,
		c.ExternalID,
		c.ID,
	), nil
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
	changeset_ids
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
	changeset_ids
`

func (s *Store) createCampaignQuery(c *Campaign) (*sqlf.Query, error) {
	changesetIDs, err := jsonSetColumn(c.ChangesetIDs)
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
		changesetIDs,
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
	changeset_ids
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
	changeset_ids
`

func (s *Store) updateCampaignQuery(c *Campaign) (*sqlf.Query, error) {
	changesetIDs, err := jsonSetColumn(c.ChangesetIDs)
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
		changesetIDs,
		c.ID,
	), nil
}

// CountCampaignsOpts captures the query options needed for
// counting campaigns.
type CountCampaignsOpts struct {
	ChangesetID int64
}

// CountCampaigns returns the number of campaigns in the database.
func (s *Store) CountCampaigns(ctx context.Context, opts CountCampaignsOpts) (count int64, _ error) {
	q := countCampaignsQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countCampaignsQueryFmtstr = `
-- source: pkg/a8n/store.go:CountCampaigns
SELECT COUNT(id)
FROM campaigns
WHERE %s
`

func countCampaignsQuery(opts *CountCampaignsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ChangesetID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_ids ? %s", opts.ChangesetID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countCampaignsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetCampaignOpts captures the query options needed for getting a Campaign
type GetCampaignOpts struct {
	ID int64
}

// GetCampaign gets a campaign matching the given options.
func (s *Store) GetCampaign(ctx context.Context, opts GetCampaignOpts) (*Campaign, error) {
	q := getCampaignQuery(&opts)

	var c Campaign
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanCampaign(&c, sc)
	})

	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getCampaignsQueryFmtstr = `
-- source: pkg/a8n/store.go:GetCampaign
SELECT
	id,
	name,
	description,
	author_id,
	namespace_user_id,
	namespace_org_id,
	created_at,
	updated_at,
	changeset_ids
FROM campaigns
WHERE %s
LIMIT 1
`

func getCampaignQuery(opts *GetCampaignOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getCampaignsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListCampaignsOpts captures the query options needed for
// listing campaigns.
type ListCampaignsOpts struct {
	ChangesetID int64
	Limit       int
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
	changeset_ids
FROM campaigns
WHERE %s
ORDER BY id ASC
LIMIT %s
`

func listCampaignsQuery(opts *ListCampaignsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var preds []*sqlf.Query
	if opts.ChangesetID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_ids ? %s", opts.ChangesetID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listCampaignsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
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

func scanChangeset(t *Changeset, s scanner) error {
	t.Metadata = json.RawMessage{}
	return s.Scan(
		&t.ID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Metadata,
		&dbutil.JSONInt64Set{Set: &t.CampaignIDs},
		&t.ExternalID,
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
		&dbutil.JSONInt64Set{Set: &c.ChangesetIDs},
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
