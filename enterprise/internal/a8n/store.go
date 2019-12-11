package a8n

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
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
-- source: internal/a8n/store.go:CreateChangesets
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
-- source: internal/a8n/store.go:CountChangesets
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
-- source: internal/a8n/store.go:GetChangeset
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
-- source: internal/a8n/store.go:ListChangesets
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
-- source: internal/a8n/store.go:UpdateChangesets
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
-- source: internal/a8n/store.go:GetChangesetEvent
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
-- source: internal/a8n/store.go:ListChangesetEvents
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
-- source: internal/a8n/store.go:CountChangesetEvents
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
-- source: internal/a8n/store.go:UpsertChangesetEvents
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
-- source: internal/a8n/store.go:CreateCampaign
INSERT INTO campaigns (
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
  id,
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
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
		nullInt64Column(c.CampaignPlanID),
		nullTimeColumn(c.ClosedAt),
	), nil
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

func nullInt64Column(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func nullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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
-- source: internal/a8n/store.go:UpdateCampaign
UPDATE campaigns
SET (
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
) = (%s, %s, %s, %s, %s, %s, %s, %s, %s)
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
  changeset_ids,
  campaign_plan_id,
  closed_at
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
		nullInt64Column(c.CampaignPlanID),
		nullTimeColumn(c.ClosedAt),
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
-- source: internal/a8n/store.go:DeleteCampaign
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
-- source: internal/a8n/store.go:CountCampaigns
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
-- source: internal/a8n/store.go:GetCampaign
SELECT
  id,
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
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
-- source: internal/a8n/store.go:ListCampaigns
SELECT
  id,
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
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

// CreateCampaignPlan creates the given CampaignPlan.
func (s *Store) CreateCampaignPlan(ctx context.Context, c *a8n.CampaignPlan) error {
	q, err := s.createCampaignPlanQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignPlan(c, sc)
		return int64(c.ID), 1, err
	})
}

var createCampaignPlanQueryFmtstr = `
-- source: internal/a8n/store.go:CreateCampaignPlan
INSERT INTO campaign_plans (
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
)
VALUES (%s, %s, %s, %s, %s)
RETURNING
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
`

func (s *Store) createCampaignPlanQuery(c *a8n.CampaignPlan) (*sqlf.Query, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	arguments, err := metadataColumn(c.Arguments)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		createCampaignPlanQueryFmtstr,
		c.CampaignType,
		arguments,
		nullTimeColumn(c.CanceledAt),
		c.CreatedAt,
		c.UpdatedAt,
	), nil
}

// UpdateCampaignPlan updates the given CampaignPlan.
func (s *Store) UpdateCampaignPlan(ctx context.Context, c *a8n.CampaignPlan) error {
	q, err := s.updateCampaignPlanQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignPlan(c, sc)
		return int64(c.ID), 1, err
	})
}

var updateCampaignPlanQueryFmtstr = `
-- source: internal/a8n/store.go:UpdateCampaignPlan
UPDATE campaign_plans
SET (
  campaign_type,
  arguments,
  canceled_at,
  updated_at
) = (%s, %s, %s, %s)
WHERE id = %s
RETURNING
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
`

func (s *Store) updateCampaignPlanQuery(c *a8n.CampaignPlan) (*sqlf.Query, error) {
	c.UpdatedAt = s.now()

	arguments, err := metadataColumn(c.Arguments)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		updateCampaignPlanQueryFmtstr,
		c.CampaignType,
		arguments,
		nullTimeColumn(c.CanceledAt),
		c.UpdatedAt,
		c.ID,
	), nil
}

// DeleteCampaignPlan deletes the CampaignPlan with the given ID.
func (s *Store) DeleteCampaignPlan(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deleteCampaignPlanQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteCampaignPlanQueryFmtstr = `
-- source: internal/a8n/store.go:DeleteCampaignPlan
DELETE FROM campaign_plans WHERE id = %s
`

const CampaignPlanTTL = 1 * time.Hour

// DeleteExpiredCampaignPlans deletes CampaignPlans that have finished execution
// but have not been attached to a Campaign within CampaignPlanTTL.
func (s *Store) DeleteExpiredCampaignPlans(ctx context.Context) error {
	expirationTime := s.now().Add(-CampaignPlanTTL)
	q := sqlf.Sprintf(deleteExpiredCampaignPlansQueryFmtstr, expirationTime)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteExpiredCampaignPlansQueryFmtstr = `
-- source: internal/a8n/store.go:DeleteExpiredCampaignPlans
DELETE FROM
  campaign_plans
WHERE
NOT EXISTS (
  SELECT 1
  FROM
  campaigns
  WHERE
  campaigns.campaign_plan_id = campaign_plans.id
)
AND
NOT EXISTS (
  SELECT id
  FROM
  campaign_jobs
  WHERE
  campaign_jobs.campaign_plan_id = campaign_plans.id
  AND
  (
    campaign_jobs.finished_at IS NULL
    OR
    campaign_jobs.finished_at > %s
  )
);
`

// CountCampaignPlans returns the number of code mods in the database.
func (s *Store) CountCampaignPlans(ctx context.Context) (count int64, _ error) {
	q := sqlf.Sprintf(countCampaignPlansQueryFmtstr)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countCampaignPlansQueryFmtstr = `
-- source: internal/a8n/store.go:CountCampaignPlans
SELECT COUNT(id)
FROM campaign_plans
`

// GetCampaignPlanOpts captures the query options needed for getting a CampaignPlan
type GetCampaignPlanOpts struct {
	ID int64
}

// GetCampaignPlan gets a code mod matching the given options.
func (s *Store) GetCampaignPlan(ctx context.Context, opts GetCampaignPlanOpts) (*a8n.CampaignPlan, error) {
	q := getCampaignPlanQuery(&opts)

	var c a8n.CampaignPlan
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanCampaignPlan(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getCampaignPlansQueryFmtstr = `
-- source: internal/a8n/store.go:GetCampaignPlan
SELECT
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
FROM campaign_plans
WHERE %s
LIMIT 1
`

func getCampaignPlanQuery(opts *GetCampaignPlanOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getCampaignPlansQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetCampaignPlanStatus gets the a8n.BackgroundProcessStatus for a CampaignPlan
func (s *Store) GetCampaignPlanStatus(ctx context.Context, id int64) (*a8n.BackgroundProcessStatus, error) {
	return s.queryBackgroundProcessStatus(ctx, sqlf.Sprintf(
		getCampaignPlanStatusQueryFmtstr,
		id,
		sqlf.Sprintf("campaign_plan_id = %s", id),
	))
}

var getCampaignPlanStatusQueryFmtstr = `
-- source: internal/a8n/store.go:GetCampaignPlanStatus
SELECT
  (SELECT canceled_at IS NOT NULL FROM campaign_plans WHERE id = %s) AS canceled,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE finished_at IS NULL) AS pending,
  COUNT(*) FILTER (WHERE finished_at IS NOT NULL AND (diff != '' OR error != '')) AS completed,
  array_agg(error) FILTER (WHERE error != '') AS errors
FROM campaign_jobs
WHERE %s
LIMIT 1;
`

// GetCampaignStatus gets the a8n.BackgroundProcessStatus for a Campaign
func (s *Store) GetCampaignStatus(ctx context.Context, id int64) (*a8n.BackgroundProcessStatus, error) {
	return s.queryBackgroundProcessStatus(ctx, sqlf.Sprintf(
		getCampaignStatusQueryFmtstr,
		sqlf.Sprintf("campaign_id = %s", id),
	))
}

func (s *Store) queryBackgroundProcessStatus(ctx context.Context, q *sqlf.Query) (*a8n.BackgroundProcessStatus, error) {
	var status a8n.BackgroundProcessStatus
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanBackgroundProcessStatus(&status, sc)
	})
	if err != nil {
		return nil, err
	}

	status.ProcessState = a8n.BackgroundProcessStateCompleted
	switch {
	case status.Canceled:
		status.ProcessState = a8n.BackgroundProcessStateCanceled
	case status.Pending > 0:
		status.ProcessState = a8n.BackgroundProcessStateProcessing
	case status.Completed == status.Total && len(status.ProcessErrors) == 0:
		status.ProcessState = a8n.BackgroundProcessStateCompleted
	case status.Completed == status.Total && len(status.ProcessErrors) != 0:
		status.ProcessState = a8n.BackgroundProcessStateErrored
	}
	return &status, nil
}

var getCampaignStatusQueryFmtstr = `
-- source: internal/a8n/store.go:GetCampaignStatus
SELECT
  -- canceled is here so that this can be used with scanBackgroundProcessStatus
  false AS canceled,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE finished_at IS NULL) AS pending,
  COUNT(*) FILTER (WHERE finished_at IS NOT NULL) AS completed,
  array_agg(error) FILTER (WHERE error != '') AS errors
FROM changeset_jobs
WHERE %s
LIMIT 1
`

// ListCampaignPlansOpts captures the query options needed for
// listing code mods.
type ListCampaignPlansOpts struct {
	Cursor int64
	Limit  int
}

// ListCampaignPlans lists CampaignPlans with the given filters.
func (s *Store) ListCampaignPlans(ctx context.Context, opts ListCampaignPlansOpts) (cs []*a8n.CampaignPlan, next int64, err error) {
	q := listCampaignPlansQuery(&opts)

	cs = make([]*a8n.CampaignPlan, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c a8n.CampaignPlan
		if err = scanCampaignPlan(&c, sc); err != nil {
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

var listCampaignPlansQueryFmtstr = `
-- source: internal/a8n/store.go:ListCampaignPlans
SELECT
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
FROM campaign_plans
WHERE %s
ORDER BY id ASC
LIMIT %s
`

func listCampaignPlansQuery(opts *ListCampaignPlansOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	preds := []*sqlf.Query{
		sqlf.Sprintf("id >= %s", opts.Cursor),
	}

	return sqlf.Sprintf(
		listCampaignPlansQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

// CreateCampaignJob creates the given CampaignJob.
func (s *Store) CreateCampaignJob(ctx context.Context, c *a8n.CampaignJob) error {
	q, err := s.createCampaignJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignJob(c, sc)
		return int64(c.ID), 1, err
	})
}

var createCampaignJobQueryFmtstr = `
-- source: internal/a8n/store.go:CreateCampaignJob
INSERT INTO campaign_jobs (
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
`

func (s *Store) createCampaignJobQuery(c *a8n.CampaignJob) (*sqlf.Query, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return sqlf.Sprintf(
		createCampaignJobQueryFmtstr,
		c.CampaignPlanID,
		c.RepoID,
		c.Rev,
		c.BaseRef,
		c.Diff,
		c.Description,
		c.Error,
		nullTimeColumn(c.StartedAt),
		nullTimeColumn(c.FinishedAt),
		c.CreatedAt,
		c.UpdatedAt,
	), nil
}

// UpdateCampaignJob updates the given CampaignJob.
func (s *Store) UpdateCampaignJob(ctx context.Context, c *a8n.CampaignJob) error {
	q, err := s.updateCampaignJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignJob(c, sc)
		return int64(c.ID), 1, err
	})
}

var updateCampaignJobQueryFmtstr = `
-- source: internal/a8n/store.go:UpdateCampaignJob
UPDATE campaign_jobs
SET (
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  updated_at
) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
`

func (s *Store) updateCampaignJobQuery(c *a8n.CampaignJob) (*sqlf.Query, error) {
	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateCampaignJobQueryFmtstr,
		c.CampaignPlanID,
		c.RepoID,
		c.Rev,
		c.BaseRef,
		c.Diff,
		c.Description,
		c.Error,
		c.StartedAt,
		c.FinishedAt,
		c.UpdatedAt,
		c.ID,
	), nil
}

// DeleteCampaignJob deletes the CampaignJob with the given ID.
func (s *Store) DeleteCampaignJob(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deleteCampaignJobQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteCampaignJobQueryFmtstr = `
-- source: internal/a8n/store.go:DeleteCampaignJob
DELETE FROM campaign_jobs WHERE id = %s
`

// CountCampaignJobsOpts captures the query options needed for
// counting code mods.
type CountCampaignJobsOpts struct {
	CampaignPlanID int64
	OnlyFinished   bool
	OnlyWithDiff   bool
}

// CountCampaignJobs returns the number of code mods in the database.
func (s *Store) CountCampaignJobs(ctx context.Context, opts CountCampaignJobsOpts) (count int64, _ error) {
	q := countCampaignJobsQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countCampaignJobsQueryFmtstr = `
-- source: internal/a8n/store.go:CountCampaignJobs
SELECT COUNT(id)
FROM campaign_jobs
WHERE %s
`

func countCampaignJobsQuery(opts *CountCampaignJobsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.CampaignPlanID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_plan_id = %s", opts.CampaignPlanID))
	}

	if opts.OnlyFinished {
		preds = append(preds, sqlf.Sprintf("finished_at IS NOT NULL"))
	}

	if opts.OnlyWithDiff {
		preds = append(preds, sqlf.Sprintf("diff != ''"))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countCampaignJobsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetCampaignJobOpts captures the query options needed for getting a CampaignJob
type GetCampaignJobOpts struct {
	ID int64
}

// GetCampaignJob gets a code mod matching the given options.
func (s *Store) GetCampaignJob(ctx context.Context, opts GetCampaignJobOpts) (*a8n.CampaignJob, error) {
	q := getCampaignJobQuery(&opts)

	var c a8n.CampaignJob
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanCampaignJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getCampaignJobsQueryFmtstr = `
-- source: internal/a8n/store.go:GetCampaignJob
SELECT
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM campaign_jobs
WHERE %s
LIMIT 1
`

func getCampaignJobQuery(opts *GetCampaignJobOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getCampaignJobsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListCampaignJobsOpts captures the query options needed for
// listing code mods.
type ListCampaignJobsOpts struct {
	CampaignPlanID int64
	Cursor         int64
	Limit          int
	OnlyFinished   bool
	OnlyWithDiff   bool
}

// ListCampaignJobs lists CampaignJobs with the given filters.
func (s *Store) ListCampaignJobs(ctx context.Context, opts ListCampaignJobsOpts) (cs []*a8n.CampaignJob, next int64, err error) {
	q := listCampaignJobsQuery(&opts)

	cs = make([]*a8n.CampaignJob, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c a8n.CampaignJob
		if err = scanCampaignJob(&c, sc); err != nil {
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

var listCampaignJobsQueryFmtstr = `
-- source: internal/a8n/store.go:ListCampaignJobs
SELECT
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM campaign_jobs
WHERE %s
ORDER BY id ASC
`

func listCampaignJobsQuery(opts *ListCampaignJobsOpts) *sqlf.Query {
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

	if opts.CampaignPlanID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_plan_id = %s", opts.CampaignPlanID))
	}

	if opts.OnlyFinished {
		preds = append(preds, sqlf.Sprintf("finished_at IS NOT NULL"))
	}

	if opts.OnlyWithDiff {
		preds = append(preds, sqlf.Sprintf("diff != ''"))
	}

	return sqlf.Sprintf(
		listCampaignJobsQueryFmtstr+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

// CreateChangesetJob creates the given ChangesetJob.
func (s *Store) CreateChangesetJob(ctx context.Context, c *a8n.ChangesetJob) error {
	q, err := s.createChangesetJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanChangesetJob(c, sc)
		return int64(c.ID), 1, err
	})
}

var createChangesetJobQueryFmtstr = `
-- source: internal/a8n/store.go:CreateChangesetJob
INSERT INTO changeset_jobs (
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
`

func (s *Store) createChangesetJobQuery(c *a8n.ChangesetJob) (*sqlf.Query, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return sqlf.Sprintf(
		createChangesetJobQueryFmtstr,
		c.CampaignID,
		c.CampaignJobID,
		nullInt64Column(c.ChangesetID),
		nullStringColumn(c.Error),
		nullTimeColumn(c.StartedAt),
		nullTimeColumn(c.FinishedAt),
		c.CreatedAt,
		c.UpdatedAt,
	), nil
}

// UpdateChangesetJob updates the given ChangesetJob.
func (s *Store) UpdateChangesetJob(ctx context.Context, c *a8n.ChangesetJob) error {
	q, err := s.updateChangesetJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanChangesetJob(c, sc)
		return int64(c.ID), 1, err
	})
}

var updateChangesetJobQueryFmtstr = `
-- source: internal/a8n/store.go:UpdateChangesetJob
UPDATE changeset_jobs
SET (
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  updated_at
) = (%s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
`

func (s *Store) updateChangesetJobQuery(c *a8n.ChangesetJob) (*sqlf.Query, error) {
	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateChangesetJobQueryFmtstr,
		c.CampaignID,
		c.CampaignJobID,
		nullInt64Column(c.ChangesetID),
		nullStringColumn(c.Error),
		nullTimeColumn(c.StartedAt),
		nullTimeColumn(c.FinishedAt),
		c.UpdatedAt,
		c.ID,
	), nil
}

// DeleteChangesetJob deletes the ChangesetJob with the given ID.
func (s *Store) DeleteChangesetJob(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deleteChangesetJobQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteChangesetJobQueryFmtstr = `
-- source: internal/a8n/store.go:DeleteChangesetJob
DELETE FROM changeset_jobs WHERE id = %s
`

// CountChangesetJobsOpts captures the query options needed for
// counting code mods.
type CountChangesetJobsOpts struct {
	CampaignID int64
}

// CountChangesetJobs returns the number of code mods in the database.
func (s *Store) CountChangesetJobs(ctx context.Context, opts CountChangesetJobsOpts) (count int64, _ error) {
	q := countChangesetJobsQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countChangesetJobsQueryFmtstr = `
-- source: internal/a8n/store.go:CountChangesetJobs
SELECT COUNT(id)
FROM changeset_jobs
WHERE %s
`

func countChangesetJobsQuery(opts *CountChangesetJobsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_id = %s", opts.CampaignID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countChangesetJobsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetChangesetJobOpts captures the query options needed for getting a ChangesetJob
type GetChangesetJobOpts struct {
	ID            int64
	CampaignJobID int64
	ChangesetID   int64
}

// GetChangesetJob gets a code mod matching the given options.
func (s *Store) GetChangesetJob(ctx context.Context, opts GetChangesetJobOpts) (*a8n.ChangesetJob, error) {
	q := getChangesetJobQuery(&opts)

	var c a8n.ChangesetJob
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanChangesetJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getChangesetJobsQueryFmtstr = `
-- source: internal/a8n/store.go:GetChangesetJob
SELECT
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM changeset_jobs
WHERE %s
LIMIT 1
`

func getChangesetJobQuery(opts *GetChangesetJobOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.CampaignJobID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_job_id = %s", opts.CampaignJobID))
	}

	if opts.ChangesetID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_id = %s", opts.ChangesetID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getChangesetJobsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListChangesetJobsOpts captures the query options needed for
// listing code mods.
type ListChangesetJobsOpts struct {
	CampaignID int64
	Cursor     int64
	Limit      int
}

// ListChangesetJobs lists ChangesetJobs with the given filters.
func (s *Store) ListChangesetJobs(ctx context.Context, opts ListChangesetJobsOpts) (cs []*a8n.ChangesetJob, next int64, err error) {
	q := listChangesetJobsQuery(&opts)

	cs = make([]*a8n.ChangesetJob, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c a8n.ChangesetJob
		if err = scanChangesetJob(&c, sc); err != nil {
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

var listChangesetJobsQueryFmtstr = `
-- source: internal/a8n/store.go:ListChangesetJobs
SELECT
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM changeset_jobs
WHERE %s
ORDER BY id ASC
`

func listChangesetJobsQuery(opts *ListChangesetJobsOpts) *sqlf.Query {
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

	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_id = %s", opts.CampaignID))
	}

	return sqlf.Sprintf(
		listChangesetJobsQueryFmtstr+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

// ResetFailedChangesetJobs resets the Error, StartedAt and FinishedAt fields
// of the ChangesetJobs belonging to the Campaign with the given ID.
func (s *Store) ResetFailedChangesetJobs(ctx context.Context, campaignID int64) (err error) {
	q := sqlf.Sprintf(resetFailedChangesetJobsQueryFmtstr, campaignID)

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		return 0, 1, nil
	})
}

var resetFailedChangesetJobsQueryFmtstr = `
-- source: internal/a8n/store.go:ResetFailedChangesetJobs
UPDATE changeset_jobs
SET
  error = '',
  started_at = NULL,
  finished_at = NULL
WHERE campaign_id = %s
`

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

	e.Metadata, err = a8n.NewChangesetEventMetadata(e.Kind)
	if err != nil {
		return err
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
		&dbutil.NullInt64{N: &c.CampaignPlanID},
		&dbutil.NullTime{Time: &c.ClosedAt},
	)
}

func scanCampaignPlan(c *a8n.CampaignPlan, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.CampaignType,
		&c.Arguments,
		&dbutil.NullTime{Time: &c.CanceledAt},
		&c.CreatedAt,
		&c.UpdatedAt,
	)
}

func scanCampaignJob(c *a8n.CampaignJob, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.CampaignPlanID,
		&c.RepoID,
		&c.Rev,
		&c.BaseRef,
		&c.Diff,
		&c.Description,
		&c.Error,
		&dbutil.NullTime{Time: &c.StartedAt},
		&dbutil.NullTime{Time: &c.FinishedAt},
		&c.CreatedAt,
		&c.UpdatedAt,
	)
}

func scanChangesetJob(c *a8n.ChangesetJob, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.CampaignID,
		&c.CampaignJobID,
		&dbutil.NullInt64{N: &c.ChangesetID},
		&dbutil.NullString{S: &c.Error},
		&dbutil.NullTime{Time: &c.StartedAt},
		&dbutil.NullTime{Time: &c.FinishedAt},
		&c.CreatedAt,
		&c.UpdatedAt,
	)
}

func scanBackgroundProcessStatus(b *a8n.BackgroundProcessStatus, s scanner) error {
	return s.Scan(
		&b.Canceled,
		&b.Total,
		&b.Pending,
		&b.Completed,
		pq.Array(&b.ProcessErrors),
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
