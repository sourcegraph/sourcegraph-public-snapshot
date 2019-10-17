package a8n

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// Store exposes methods to read and write a8n domain models
// from persistent storage.
type Store struct {
	db  dbutil.DB
	now func() time.Time
}

// NewStore returns a new Store backed by the given db.
func NewStore(db dbutil.DB) *Store {
	return NewStoreWithClock(db, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
}

// NewStoreWithClock returns a new Store backed by the given db and
// clock for timestamps.
func NewStoreWithClock(db dbutil.DB, clock func() time.Time) *Store {
	return &Store{db: db, now: clock}
}

// Transact returns a Store whose methods operate within the context of a transaction.
// This method will return an error if the underlying DB cannot be interface upgraded
// to a TxBeginner.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	if _, ok := s.db.(dbutil.Tx); ok { // Already in a Tx.
		return s, nil
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

// AlreadyExistError is returned by CreateChangesets in case a subset of the
// given changesets already existed in the database and were not inserted but
// returned
type AlreadyExistError struct {
	ChangesetIDs []int64
}

func (e AlreadyExistError) Error() string {
	return fmt.Sprintf("Changesets already exist: %v", e.ChangesetIDs)
}

// CreateChangesets creates the given Changesets. If a subset of the given
// Changesets with the same RepoID and ExternalID already exists in the
// database, it overwrites the fields of the affected changeset pointers with
// the values contained in the database and returns an AlreadyExistError.
func (s *Store) CreateChangesets(ctx context.Context, cs ...*a8n.Changeset) error {
	q, err := s.createChangesetsQuery(cs)
	if err != nil {
		return err
	}

	exist := []int64{}
	i := -1
	err = s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		i++

		createdAt := cs[i].CreatedAt

		err = scanChangeset(cs[i], sc)
		if err != nil {
			return 0, 0, err
		}

		// Check whether the Changeset already existed in the database or not.
		// We use CreatedAt for this, which `createChangesetsQuery` sets to
		// now() if it wasn't set. If that value is not returned from the
		// database we know that an already existing one is returned.
		if cs[i].CreatedAt != createdAt {
			exist = append(exist, cs[i].ID)
		}

		return int64(cs[i].ID), 1, err
	})
	if err != nil {
		return err
	}

	if len(exist) != 0 {
		return AlreadyExistError{ChangesetIDs: exist}
	}

	return nil
}

const changesetBatchQueryPrefix = `
WITH batch AS (
  SELECT * FROM ROWS FROM (
  json_to_recordset(%s)
  AS (
      id                    bigint,
      repo_id               integer,
      created_at            timestamptz,
      updated_at            timestamptz,
      metadata              jsonb,
      campaign_ids          jsonb,
      external_id           text,
      external_service_type text
    )
  )
  WITH ORDINALITY
)
`

var createChangesetsQueryFmtstr = changesetBatchQueryPrefix + `,
-- source: pkg/a8n/store.go:CreateChangesets
changed AS (
  INSERT INTO changesets (
    repo_id,
    created_at,
    updated_at,
    metadata,
    campaign_ids,
    external_id,
    external_service_type
  )
  SELECT
    repo_id,
    created_at,
    updated_at,
    metadata,
    campaign_ids,
    external_id,
    external_service_type
  FROM batch
  ON CONFLICT ON CONSTRAINT
    changesets_repo_external_id_unique
  DO NOTHING
  RETURNING changesets.*
)
` + batchCreateChangesetsQuerySuffix

const batchCreateChangesetsQuerySuffix = `
SELECT
  COALESCE(changed.id, existing.id) AS id,
  COALESCE(changed.repo_id, existing.repo_id) AS repo_id,
  COALESCE(changed.created_at, existing.created_at) AS created_at,
  COALESCE(changed.updated_at, existing.updated_at) AS updated_at,
  COALESCE(changed.metadata, existing.metadata) AS metadata,
  COALESCE(changed.campaign_ids, existing.campaign_ids) AS campaign_ids,
  COALESCE(changed.external_id, existing.external_id) AS external_id,
  COALESCE(changed.external_service_type, existing.external_service_type) AS external_service_type
FROM changed
RIGHT JOIN batch ON batch.repo_id = changed.repo_id
AND batch.external_id = changed.external_id
LEFT JOIN changesets existing ON existing.repo_id = batch.repo_id
AND existing.external_id = batch.external_id
ORDER BY batch.ordinality
`

func (s *Store) createChangesetsQuery(cs []*a8n.Changeset) (*sqlf.Query, error) {
	now := s.now()
	for _, c := range cs {
		if c.CreatedAt.IsZero() {
			c.CreatedAt = now
		}

		if c.UpdatedAt.IsZero() {
			c.UpdatedAt = c.CreatedAt
		}
	}
	return batchChangesetsQuery(createChangesetsQueryFmtstr, cs)
}

func batchChangesetsQuery(fmtstr string, cs []*a8n.Changeset) (*sqlf.Query, error) {
	type record struct {
		ID                  int64           `json:"id"`
		RepoID              int32           `json:"repo_id"`
		CreatedAt           time.Time       `json:"created_at"`
		UpdatedAt           time.Time       `json:"updated_at"`
		Metadata            json.RawMessage `json:"metadata"`
		CampaignIDs         json.RawMessage `json:"campaign_ids"`
		ExternalID          string          `json:"external_id"`
		ExternalServiceType string          `json:"external_service_type"`
	}

	records := make([]record, 0, len(cs))

	for _, c := range cs {
		metadata, err := metadataColumn(c.Metadata)
		if err != nil {
			return nil, err
		}

		campaignIDs, err := jsonSetColumn(c.CampaignIDs)
		if err != nil {
			return nil, err
		}

		records = append(records, record{
			ID:                  c.ID,
			RepoID:              c.RepoID,
			CreatedAt:           c.CreatedAt,
			UpdatedAt:           c.UpdatedAt,
			Metadata:            metadata,
			CampaignIDs:         campaignIDs,
			ExternalID:          c.ExternalID,
			ExternalServiceType: c.ExternalServiceType,
		})
	}

	batch, err := json.MarshalIndent(records, "    ", "    ")
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(fmtstr, string(batch)), nil
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
	ID                  int64
	ExternalID          string
	ExternalServiceType string
}

// ErrNoResults is returned by Store method calls that found no results.
var ErrNoResults = errors.New("no results")

// GetChangeset gets a changeset matching the given options.
func (s *Store) GetChangeset(ctx context.Context, opts GetChangesetOpts) (*a8n.Changeset, error) {
	q := getChangesetQuery(&opts)

	var c a8n.Changeset
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
  external_id,
  external_service_type
FROM changesets
WHERE %s
LIMIT 1
`

func getChangesetQuery(opts *GetChangesetOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.ExternalID != "" && opts.ExternalServiceType != "" {
		preds = append(preds,
			sqlf.Sprintf("external_id = %s", opts.ExternalID),
			sqlf.Sprintf("external_service_type = %s", opts.ExternalServiceType),
		)
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getChangesetsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListChangesetsOpts captures the query options needed for
// listing changesets.
type ListChangesetsOpts struct {
	Cursor     int64
	Limit      int
	CampaignID int64
	IDs        []int64
}

// ListChangesets lists Changesets with the given filters.
func (s *Store) ListChangesets(ctx context.Context, opts ListChangesetsOpts) (cs []*a8n.Changeset, next int64, err error) {
	q := listChangesetsQuery(&opts)

	cs = make([]*a8n.Changeset, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c a8n.Changeset
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
  external_id,
  external_service_type
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

	preds := []*sqlf.Query{
		sqlf.Sprintf("id >= %s", opts.Cursor),
	}

	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_ids ? %s", opts.CampaignID))
	}

	if len(opts.IDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(opts.IDs))
		for _, id := range opts.IDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds, sqlf.Sprintf("id IN (%s)", sqlf.Join(ids, ",")))
	}

	return sqlf.Sprintf(
		listChangesetsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

// UpdateChangesets updates the given Changesets.
func (s *Store) UpdateChangesets(ctx context.Context, cs ...*a8n.Changeset) error {
	q, err := s.updateChangesetsQuery(cs)
	if err != nil {
		return err
	}

	i := -1
	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		i++
		err = scanChangeset(cs[i], sc)
		return int64(cs[i].ID), 1, err
	})
}

const updateChangesetsQueryFmtstr = changesetBatchQueryPrefix + `,
-- source: pkg/a8n/store.go:UpdateChangesets
changed AS (
  UPDATE changesets
  SET
    repo_id               = batch.repo_id,
    created_at            = batch.created_at,
    updated_at            = batch.updated_at,
    metadata              = batch.metadata,
    campaign_ids          = batch.campaign_ids,
    external_id           = batch.external_id,
    external_service_type = batch.external_service_type
  FROM batch
  WHERE changesets.id = batch.id
  RETURNING changesets.*
)
` + batchChangesetsQuerySuffix

const batchChangesetsQuerySuffix = `
SELECT
  changed.id,
  changed.repo_id,
  changed.created_at,
  changed.updated_at,
  changed.metadata,
  changed.campaign_ids,
  changed.external_id,
  changed.external_service_type
FROM changed
LEFT JOIN batch ON batch.repo_id = changed.repo_id
AND batch.external_id = changed.external_id
ORDER BY batch.ordinality
`

func (s *Store) updateChangesetsQuery(cs []*a8n.Changeset) (*sqlf.Query, error) {
	now := s.now()
	for _, c := range cs {
		c.UpdatedAt = now
	}
	return batchChangesetsQuery(updateChangesetsQueryFmtstr, cs)
}

// GetChangesetEventOpts captures the query options needed for getting a ChangesetEvent
type GetChangesetEventOpts struct {
	ID          int64
	ChangesetID int64
	Kind        a8n.ChangesetEventKind
	Key         string
}

// GetChangesetEvent gets a changeset matching the given options.
func (s *Store) GetChangesetEvent(ctx context.Context, opts GetChangesetEventOpts) (*a8n.ChangesetEvent, error) {
	q := getChangesetEventQuery(&opts)

	var c a8n.ChangesetEvent
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanChangesetEvent(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getChangesetEventsQueryFmtstr = `
-- source: pkg/a8n/store.go:GetChangesetEvent
SELECT
    id,
    changeset_id,
    kind,
    key,
    created_at,
    updated_at,
    metadata
FROM changeset_events
WHERE %s
LIMIT 1
`

func getChangesetEventQuery(opts *GetChangesetEventOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.ChangesetID != 0 && opts.Kind != "" && opts.Key != "" {
		preds = append(preds,
			sqlf.Sprintf("changeset_id = %s", opts.ChangesetID),
			sqlf.Sprintf("kind = %s", opts.Kind),
			sqlf.Sprintf("key = %s", opts.Key),
		)
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getChangesetEventsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListChangesetEventsOpts captures the query options needed for
// listing changeset events.
type ListChangesetEventsOpts struct {
	ChangesetIDs []int64
	Cursor       int64
	Limit        int
}

// ListChangesetEvents lists ChangesetEvents with the given filters.
func (s *Store) ListChangesetEvents(ctx context.Context, opts ListChangesetEventsOpts) (cs []*a8n.ChangesetEvent, next int64, err error) {
	q := listChangesetEventsQuery(&opts)

	cs = make([]*a8n.ChangesetEvent, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c a8n.ChangesetEvent
		if err = scanChangesetEvent(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return int64(c.ID), 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listChangesetEventsQueryFmtstr = `
-- source: pkg/a8n/store.go:ListChangesetEvents
SELECT
    id,
    changeset_id,
    kind,
    key,
    created_at,
    updated_at,
    metadata
FROM changeset_events
WHERE %s
ORDER BY id ASC
`

func listChangesetEventsQuery(opts *ListChangesetEventsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var limitClause string
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}

	preds := []*sqlf.Query{
		sqlf.Sprintf("id >= %s", opts.Cursor),
	}

	if len(opts.ChangesetIDs) != 0 {
		ids := make([]*sqlf.Query, 0, len(opts.ChangesetIDs))
		for _, id := range opts.ChangesetIDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds,
			sqlf.Sprintf("changeset_id IN (%s)", sqlf.Join(ids, ",")))
	}

	return sqlf.Sprintf(
		listChangesetEventsQueryFmtstr+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

// CountChangesetEventsOpts captures the query options needed for
// counting changeset events.
type CountChangesetEventsOpts struct {
	ChangesetID int64
}

// CountChangesetEvents returns the number of changeset events in the database.
func (s *Store) CountChangesetEvents(ctx context.Context, opts CountChangesetEventsOpts) (count int64, _ error) {
	q := countChangesetEventsQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countChangesetEventsQueryFmtstr = `
-- source: pkg/a8n/store.go:CountChangesetEvents
SELECT COUNT(id)
FROM changeset_events
WHERE %s
`

func countChangesetEventsQuery(opts *CountChangesetEventsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ChangesetID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_id = %s", opts.ChangesetID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countChangesetEventsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// UpsertChangesetEvents creates or updates the given ChangesetEvents.
func (s *Store) UpsertChangesetEvents(ctx context.Context, cs ...*a8n.ChangesetEvent) (err error) {
	q, err := s.upsertChangesetEventsQuery(cs)
	if err != nil {
		return err
	}

	i := -1
	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		i++
		err = scanChangesetEvent(cs[i], sc)
		return int64(cs[i].ID), 1, err
	})
}

const changesetEventsBatchQueryPrefix = `
WITH batch AS (
  SELECT * FROM ROWS FROM (
  json_to_recordset(%s)
  AS (
      id           bigint,
      changeset_id integer,
      kind         text,
      key          text,
      created_at   timestamptz,
      updated_at   timestamptz,
      metadata     jsonb
    )
  )
  WITH ORDINALITY
)
`

const batchChangesetEventsQuerySuffix = `
SELECT
  changed.id,
  changed.changeset_id,
  changed.kind,
  changed.key,
  changed.created_at,
  changed.updated_at,
  changed.metadata
FROM changed
LEFT JOIN batch
ON batch.changeset_id = changed.changeset_id
AND batch.kind = changed.kind
AND batch.key = changed.key
ORDER BY batch.ordinality
`

var upsertChangesetEventsQueryFmtstr = changesetEventsBatchQueryPrefix + `,
-- source: pkg/a8n/store.go:UpsertChangesetEvents
changed AS (
  INSERT INTO changeset_events (
    changeset_id,
    kind,
    key,
    created_at,
    updated_at,
    metadata
  )
  SELECT
    changeset_id,
    kind,
    key,
    created_at,
    updated_at,
    metadata
  FROM batch
  ON CONFLICT ON CONSTRAINT
    changeset_events_changeset_id_kind_key_unique
  DO UPDATE
  SET
    metadata   = excluded.metadata,
    updated_at = excluded.updated_at
  RETURNING changeset_events.*
)
` + batchChangesetEventsQuerySuffix

func (s *Store) upsertChangesetEventsQuery(es []*a8n.ChangesetEvent) (*sqlf.Query, error) {
	now := s.now()
	for _, e := range es {
		if e.CreatedAt.IsZero() {
			e.CreatedAt = now
		}

		if !e.UpdatedAt.After(e.CreatedAt) {
			e.UpdatedAt = now
		}
	}
	return batchChangesetEventsQuery(upsertChangesetEventsQueryFmtstr, es)
}

func batchChangesetEventsQuery(fmtstr string, es []*a8n.ChangesetEvent) (*sqlf.Query, error) {
	type record struct {
		ID          int64           `json:"id"`
		ChangesetID int64           `json:"changeset_id"`
		Kind        string          `json:"kind"`
		Key         string          `json:"key"`
		CreatedAt   time.Time       `json:"created_at"`
		UpdatedAt   time.Time       `json:"updated_at"`
		Metadata    json.RawMessage `json:"metadata"`
	}

	records := make([]record, 0, len(es))

	for _, e := range es {
		metadata, err := metadataColumn(e.Metadata)
		if err != nil {
			return nil, err
		}

		records = append(records, record{
			ID:          e.ID,
			ChangesetID: e.ChangesetID,
			Kind:        string(e.Kind),
			Key:         e.Key,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
			Metadata:    metadata,
		})
	}

	batch, err := json.MarshalIndent(records, "    ", "    ")
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(fmtstr, string(batch)), nil
}

// CreateCampaign creates the given Campaign.
func (s *Store) CreateCampaign(ctx context.Context, c *a8n.Campaign) error {
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

func (s *Store) createCampaignQuery(c *a8n.Campaign) (*sqlf.Query, error) {
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
func (s *Store) UpdateCampaign(ctx context.Context, c *a8n.Campaign) error {
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

func (s *Store) updateCampaignQuery(c *a8n.Campaign) (*sqlf.Query, error) {
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

// DeleteCampaign deletes the Campaign with the given ID.
func (s *Store) DeleteCampaign(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deleteCampaignQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteCampaignQueryFmtstr = `
-- source: pkg/a8n/store.go:DeleteCampaign
DELETE FROM campaigns WHERE id = %s
`

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
func (s *Store) GetCampaign(ctx context.Context, opts GetCampaignOpts) (*a8n.Campaign, error) {
	q := getCampaignQuery(&opts)

	var c a8n.Campaign
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
	Cursor      int64
	Limit       int
}

// ListCampaigns lists Campaigns with the given filters.
func (s *Store) ListCampaigns(ctx context.Context, opts ListCampaignsOpts) (cs []*a8n.Campaign, next int64, err error) {
	q := listCampaignsQuery(&opts)

	cs = make([]*a8n.Campaign, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c a8n.Campaign
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

	preds := []*sqlf.Query{
		sqlf.Sprintf("id >= %s", opts.Cursor),
	}

	if opts.ChangesetID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_ids ? %s", opts.ChangesetID))
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

func scanChangeset(t *a8n.Changeset, s scanner) error {
	var metadata json.RawMessage

	err := s.Scan(
		&t.ID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&metadata,
		&dbutil.JSONInt64Set{Set: &t.CampaignIDs},
		&t.ExternalID,
		&t.ExternalServiceType,
	)
	if err != nil {
		return err
	}

	switch t.ExternalServiceType {
	case github.ServiceType:
		t.Metadata = new(github.PullRequest)
	case bitbucketserver.ServiceType:
		t.Metadata = new(bitbucketserver.PullRequest)
	default:
		return errors.New("unknown external service type")
	}

	if err = json.Unmarshal(metadata, t.Metadata); err != nil {
		return errors.Wrapf(err, "scanChangeset: failed to unmarshal %q metadata", t.ExternalServiceType)
	}

	return nil
}

func scanChangesetEvent(e *a8n.ChangesetEvent, s scanner) error {
	var metadata json.RawMessage

	err := s.Scan(
		&e.ID,
		&e.ChangesetID,
		&e.Kind,
		&e.Key,
		&e.CreatedAt,
		&e.UpdatedAt,
		&metadata,
	)
	if err != nil {
		return err
	}

	switch e.Kind {
	case a8n.ChangesetEventKindGitHubAssigned:
		e.Metadata = new(github.AssignedEvent)
	case a8n.ChangesetEventKindGitHubClosed:
		e.Metadata = new(github.ClosedEvent)
	case a8n.ChangesetEventKindGitHubCommented:
		e.Metadata = new(github.IssueComment)
	case a8n.ChangesetEventKindGitHubRenamedTitle:
		e.Metadata = new(github.RenamedTitleEvent)
	case a8n.ChangesetEventKindGitHubMerged:
		e.Metadata = new(github.MergedEvent)
	case a8n.ChangesetEventKindGitHubReviewed:
		e.Metadata = new(github.PullRequestReview)
	case a8n.ChangesetEventKindGitHubReviewCommented:
		e.Metadata = new(github.PullRequestReviewComment)
	case a8n.ChangesetEventKindGitHubReopened:
		e.Metadata = new(github.ReopenedEvent)
	case a8n.ChangesetEventKindGitHubReviewDismissed:
		e.Metadata = new(github.ReviewDismissedEvent)
	case a8n.ChangesetEventKindGitHubReviewRequestRemoved:
		e.Metadata = new(github.ReviewRequestRemovedEvent)
	case a8n.ChangesetEventKindGitHubReviewRequested:
		e.Metadata = new(github.ReviewRequestedEvent)
	case a8n.ChangesetEventKindGitHubUnassigned:
		e.Metadata = new(github.UnassignedEvent)
	default:
		panic(errors.Errorf("unknown changeset event kind for %T", e))
	}

	if err = json.Unmarshal(metadata, e.Metadata); err != nil {
		return errors.Wrapf(err, "scanChangesetEvent: failed to unmarshal %q metadata", e.Kind)
	}

	return nil
}

func scanCampaign(c *a8n.Campaign, s scanner) error {
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
