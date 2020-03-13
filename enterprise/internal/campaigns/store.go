package campaigns

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
	"github.com/segmentio/fasthash/fnv1"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// Store exposes methods to read and write campaigns domain models
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

// Clock returns the clock used by the Store.
func (s *Store) Clock() func() time.Time { return s.now }

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

// ProcessPendingChangesetJobs attempts to fetch one pending changeset job.
// A pending job is one that has never been started and its plan is not cancelled.
// If found, 'process' is called. We guarantee that if process is called it will have exclusive global access to
// the job. All operations on the job should be done using the supplied store as they will run in a transaction.
// Returning an error will roll back the transaction.
// NOTE: It should not be called from within an existing transaction
func (s *Store) ProcessPendingChangesetJobs(ctx context.Context, process func(ctx context.Context, s *Store, job campaigns.ChangesetJob) error) (didRun bool, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return false, errors.Wrap(err, "starting transaction")
	}
	defer tx.Done(&err)
	q := sqlf.Sprintf(getPendingChangesetJobQuery)
	var job campaigns.ChangesetJob
	_, count, err := tx.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanChangesetJob(&job, sc)
		if err != nil {
			return 0, 0, errors.Wrap(err, "scanning campaign job row")
		}
		return job.ID, 1, nil
	})
	if err != nil {
		return false, errors.Wrap(err, "querying for pending campaign job")
	}
	if count == 0 {
		return false, nil
	}
	err = process(ctx, tx, job)
	return true, err
}

const getPendingChangesetJobQuery = `
UPDATE changeset_jobs j SET started_at = now() WHERE id = (
	SELECT j.id FROM changeset_jobs j
	JOIN campaigns c ON c.id = j.campaign_id
	JOIN campaign_plans p ON p.id = c.campaign_plan_id
	WHERE j.started_at IS NULL
	AND p.canceled_at IS NULL
	ORDER BY j.id ASC
	FOR UPDATE SKIP LOCKED LIMIT 1
)
RETURNING j.id,
  j.campaign_id,
  j.campaign_job_id,
  j.changeset_id,
  j.branch,
  j.error,
  j.started_at,
  j.finished_at,
  j.created_at,
  j.updated_at
`

// ProcessPendingCampaignJob attempts to fetch one pending campaign job. If found, 'process'
// is called. We guarantee that if process is called it will have exclusive global access to the job.
// All operations on the job should be done using the supplied store as they will run in a transaction.
// Returning an error will roll back the transaction.
// NOTE: It should not be called from within an existing transaction
func (s *Store) ProcessPendingCampaignJob(ctx context.Context, process func(ctx context.Context, s *Store, job campaigns.CampaignJob) error) (didRun bool, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return false, errors.Wrap(err, "starting transaction")
	}
	defer tx.Done(&err)
	q := sqlf.Sprintf(getPendingCampaignJobQuery)
	var job campaigns.CampaignJob
	_, count, err := tx.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignJob(&job, sc)
		if err != nil {
			return 0, 0, errors.Wrap(err, "scanning campaign job row")
		}
		return job.ID, 1, nil
	})
	if err != nil {
		return false, errors.Wrap(err, "querying for pending campaign job")
	}
	if count == 0 {
		return false, nil
	}
	err = process(ctx, tx, job)
	return true, err
}

const getPendingCampaignJobQuery = `
UPDATE campaign_jobs c SET started_at = now() WHERE id = (
	SELECT c.id FROM campaign_jobs c
	JOIN campaign_plans p ON p.id = c.campaign_plan_id
	WHERE c.started_at IS NULL
	AND p.canceled_at IS NULL
	ORDER BY c.id ASC
	FOR UPDATE SKIP LOCKED LIMIT 1
)
RETURNING c.id,
  c.campaign_plan_id,
  c.repo_id,
  c.rev,
  c.base_ref,
  c.diff,
  c.description,
  c.error,
  c.started_at,
  c.finished_at,
  c.created_at,
  c.updated_at
`

// Done terminates the underlying Tx in a Store either by committing or rolling
// back based on the value pointed to by the first given error pointer.
// It's a no-op if the `Store` is not operating within a transaction,
// which can only be done via `Transact`.
//
// When the error value pointed to by the first given `err` is nil, or when no error
// pointer is given, the transaction is committed. Otherwise, it's rolled-back.
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

var NoTransactionError = errors.New("Not in a transaction")

var lockNamespace = int32(fnv1.HashString32("campaigns"))

// TryAcquireAdvisoryLock will attempt to acquire an advisory lock using key
// and is non blocking. If a lock is acquired, "true, nil" will be returned.
// It must be called from within a transaction or "false, NoTransactionError" is returned
func (s *Store) TryAcquireAdvisoryLock(ctx context.Context, key string) (bool, error) {
	_, ok := s.db.(dbutil.Tx)
	if !ok {
		return false, NoTransactionError
	}
	q := lockQuery(key)
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return false, err
	}

	if !rows.Next() {
		return false, rows.Err()
	}

	locked := false
	if err = rows.Scan(&locked); err != nil {
		return false, err
	}

	if err = rows.Close(); err != nil {
		return false, err
	}

	return locked, nil
}

func lockQuery(key string) *sqlf.Query {
	// Postgres advisory lock ids are a global namespace within one database.
	// It's very unlikely that another part of our application uses a lock
	// namespace identically to this one. It's equally unlikely that there are
	// lock id conflicts for different permissions, but if it'd happen, no safety
	// guarantees would be violated, since those two different users would simply
	// have to wait on the other's update to finish, using stale permissions until
	// it would.
	lockID := int32(fnv1.HashString32(key))
	return sqlf.Sprintf(
		lockQueryFmtStr,
		lockNamespace,
		lockID,
	)
}

const lockQueryFmtStr = `
-- source: enterprise/internal/campaigns/store/store.go:TryAcquireAdvisoryLock
SELECT pg_try_advisory_xact_lock(%s, %s)
`

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
func (s *Store) CreateChangesets(ctx context.Context, cs ...*campaigns.Changeset) error {
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

		return cs[i].ID, 1, err
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
      external_service_type text,
      external_branch       text,
      external_deleted_at   timestamptz,
      external_updated_at   timestamptz,
      external_state        text,
      external_review_state text,
      external_check_state  text
    )
  )
  WITH ORDINALITY
)
`

var createChangesetsQueryFmtstr = changesetBatchQueryPrefix + `,
-- source: enterprise/internal/campaigns/store.go:CreateChangesets
changed AS (
  INSERT INTO changesets (
    repo_id,
    created_at,
    updated_at,
    metadata,
    campaign_ids,
    external_id,
    external_service_type,
    external_branch,
    external_deleted_at,
    external_updated_at,
    external_state,
    external_review_state,
    external_check_state
  )
  SELECT
    repo_id,
    created_at,
    updated_at,
    metadata,
    campaign_ids,
    external_id,
    external_service_type,
    external_branch,
    external_deleted_at,
    external_updated_at,
    external_state,
    external_review_state,
    external_check_state
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
  COALESCE(changed.external_service_type, existing.external_service_type) AS external_service_type,
  COALESCE(changed.external_branch, existing.external_branch) AS external_branch,
  COALESCE(changed.external_deleted_at, existing.external_deleted_at) AS external_deleted_at,
  COALESCE(changed.external_updated_at, existing.external_updated_at) AS external_updated_at,
  COALESCE(changed.external_state, existing.external_state) AS external_state,
  COALESCE(changed.external_review_state, existing.external_review_state) AS external_review_state,
  COALESCE(changed.external_check_state, existing.external_check_state) AS external_check_state
FROM changed
RIGHT JOIN batch ON batch.repo_id = changed.repo_id
AND batch.external_id = changed.external_id
LEFT JOIN changesets existing ON existing.repo_id = batch.repo_id
AND existing.external_id = batch.external_id
ORDER BY batch.ordinality
`

func (s *Store) createChangesetsQuery(cs []*campaigns.Changeset) (*sqlf.Query, error) {
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

func batchChangesetsQuery(fmtstr string, cs []*campaigns.Changeset) (*sqlf.Query, error) {
	type record struct {
		ID                  int64                           `json:"id"`
		RepoID              api.RepoID                      `json:"repo_id"`
		CreatedAt           time.Time                       `json:"created_at"`
		UpdatedAt           time.Time                       `json:"updated_at"`
		Metadata            json.RawMessage                 `json:"metadata"`
		CampaignIDs         json.RawMessage                 `json:"campaign_ids"`
		ExternalID          string                          `json:"external_id"`
		ExternalServiceType string                          `json:"external_service_type"`
		ExternalBranch      string                          `json:"external_branch"`
		ExternalDeletedAt   *time.Time                      `json:"external_deleted_at"`
		ExternalUpdatedAt   *time.Time                      `json:"external_updated_at"`
		ExternalState       *campaigns.ChangesetState       `json:"external_state"`
		ExternalReviewState *campaigns.ChangesetReviewState `json:"external_review_state"`
		ExternalCheckState  *campaigns.ChangesetCheckState  `json:"external_check_state"`
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

		r := record{
			ID:                  c.ID,
			RepoID:              c.RepoID,
			CreatedAt:           c.CreatedAt,
			UpdatedAt:           c.UpdatedAt,
			Metadata:            metadata,
			CampaignIDs:         campaignIDs,
			ExternalID:          c.ExternalID,
			ExternalServiceType: c.ExternalServiceType,
			ExternalBranch:      c.ExternalBranch,
			ExternalDeletedAt:   nullTimeColumn(c.ExternalDeletedAt),
			ExternalUpdatedAt:   nullTimeColumn(c.ExternalUpdatedAt),
		}
		if len(c.ExternalState) > 0 {
			r.ExternalState = &c.ExternalState
		}
		if len(c.ExternalReviewState) > 0 {
			r.ExternalReviewState = &c.ExternalReviewState
		}
		if len(c.ExternalCheckState) > 0 {
			r.ExternalCheckState = &c.ExternalCheckState
		}

		records = append(records, r)
	}

	batch, err := json.MarshalIndent(records, "    ", "    ")
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(fmtstr, string(batch)), nil
}

// DeleteChangeset deletes the Changeset with the given ID.
func (s *Store) DeleteChangeset(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deleteChangesetQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteChangesetQueryFmtstr = `
DELETE FROM changesets WHERE id = %s
`

// CountChangesetsOpts captures the query options needed for
// counting changesets.
type CountChangesetsOpts struct {
	CampaignID          int64
	ExternalState       *campaigns.ChangesetState
	ExternalReviewState *campaigns.ChangesetReviewState
	ExternalCheckState  *campaigns.ChangesetCheckState
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
-- source: enterprise/internal/campaigns/store.go:CountChangesets
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

	if opts.ExternalState != nil {
		preds = append(preds, sqlf.Sprintf("external_state = %s", *opts.ExternalState))
	}
	if opts.ExternalReviewState != nil {
		preds = append(preds, sqlf.Sprintf("external_review_state = %s", *opts.ExternalReviewState))
	}
	if opts.ExternalCheckState != nil {
		preds = append(preds, sqlf.Sprintf("external_check_state = %s", *opts.ExternalCheckState))
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
func (s *Store) GetChangeset(ctx context.Context, opts GetChangesetOpts) (*campaigns.Changeset, error) {
	q := getChangesetQuery(&opts)

	var c campaigns.Changeset
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
-- source: enterprise/internal/campaigns/store.go:GetChangeset
SELECT
  id,
  repo_id,
  created_at,
  updated_at,
  metadata,
  campaign_ids,
  external_id,
  external_service_type,
  external_branch,
  external_deleted_at,
  external_updated_at,
  external_state,
  external_review_state,
  external_check_state
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

// ListChangesetSyncData returns sync timing data on all non-externally-deleted changesets.
func (s *Store) ListChangesetSyncData(ctx context.Context) ([]campaigns.ChangesetSyncData, error) {
	q := listChangesetSyncData()
	results := make([]campaigns.ChangesetSyncData, 0)
	_, _, err := s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var h campaigns.ChangesetSyncData
		if err = scanChangesetSyncData(&h, sc); err != nil {
			return 0, 0, err
		}
		results = append(results, h)
		return h.ChangesetID, 1, err
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func scanChangesetSyncData(h *campaigns.ChangesetSyncData, s scanner) error {
	return s.Scan(
		&h.ChangesetID,
		&h.UpdatedAt,
		&dbutil.NullTime{Time: &h.LatestEvent},
		&dbutil.NullTime{Time: &h.ExternalUpdatedAt},
	)
}

func listChangesetSyncData() *sqlf.Query {
	return sqlf.Sprintf(`
SELECT changesets.id,
       changesets.updated_at,
       max(ce.updated_at) as latest_event,
       changesets.external_updated_at
FROM changesets
LEFT JOIN changeset_events ce ON changesets.id = ce.changeset_id
GROUP BY changesets.id
ORDER BY changesets.id ASC
`)
}

// ListChangesetsOpts captures the query options needed for
// listing changesets.
type ListChangesetsOpts struct {
	Cursor              int64
	Limit               int
	CampaignID          int64
	IDs                 []int64
	WithoutDeleted      bool
	ExternalState       *campaigns.ChangesetState
	ExternalReviewState *campaigns.ChangesetReviewState
	ExternalCheckState  *campaigns.ChangesetCheckState
}

// ListChangesets lists Changesets with the given filters.
func (s *Store) ListChangesets(ctx context.Context, opts ListChangesetsOpts) (cs []*campaigns.Changeset, next int64, err error) {
	q := listChangesetsQuery(&opts)

	cs = make([]*campaigns.Changeset, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.Changeset
		if err = scanChangeset(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return c.ID, 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listChangesetsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:ListChangesets
SELECT
  id,
  repo_id,
  created_at,
  updated_at,
  metadata,
  campaign_ids,
  external_id,
  external_service_type,
  external_branch,
  external_deleted_at,
  external_updated_at,
  external_state,
  external_review_state,
  external_check_state
FROM changesets
WHERE %s
ORDER BY id ASC
`

const defaultListLimit = 50

func listChangesetsQuery(opts *ListChangesetsOpts) *sqlf.Query {
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

	if opts.WithoutDeleted {
		preds = append(preds, sqlf.Sprintf("external_deleted_at IS NULL"))
	}

	if opts.ExternalState != nil {
		preds = append(preds, sqlf.Sprintf("external_state = %s", *opts.ExternalState))
	}
	if opts.ExternalReviewState != nil {
		preds = append(preds, sqlf.Sprintf("external_review_state = %s", *opts.ExternalReviewState))
	}
	if opts.ExternalCheckState != nil {
		preds = append(preds, sqlf.Sprintf("external_check_state = %s", *opts.ExternalCheckState))
	}

	return sqlf.Sprintf(
		listChangesetsQueryFmtstr+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

// UpdateChangesets updates the given Changesets.
func (s *Store) UpdateChangesets(ctx context.Context, cs ...*campaigns.Changeset) error {
	q, err := s.updateChangesetsQuery(cs)
	if err != nil {
		return err
	}

	i := -1
	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		i++
		err = scanChangeset(cs[i], sc)
		return cs[i].ID, 1, err
	})
}

const updateChangesetsQueryFmtstr = changesetBatchQueryPrefix + `,
-- source: enterprise/internal/campaigns/store.go:UpdateChangesets
changed AS (
  UPDATE changesets
  SET
    repo_id               = batch.repo_id,
    created_at            = batch.created_at,
    updated_at            = batch.updated_at,
    metadata              = batch.metadata,
    campaign_ids          = batch.campaign_ids,
    external_id           = batch.external_id,
    external_service_type = batch.external_service_type,
    external_branch       = batch.external_branch,
	external_deleted_at   = batch.external_deleted_at,
	external_updated_at   = batch.external_updated_at,
    external_state        = batch.external_state,
    external_review_state = batch.external_review_state,
    external_check_state  = batch.external_check_state
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
  changed.external_service_type,
  changed.external_branch,
  changed.external_deleted_at,
  changed.external_updated_at,
  changed.external_state,
  changed.external_review_state,
  changed.external_check_state
FROM changed
LEFT JOIN batch ON batch.repo_id = changed.repo_id
AND batch.external_id = changed.external_id
ORDER BY batch.ordinality
`

func (s *Store) updateChangesetsQuery(cs []*campaigns.Changeset) (*sqlf.Query, error) {
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
	Kind        campaigns.ChangesetEventKind
	Key         string
}

// GetChangesetEvent gets a changeset matching the given options.
func (s *Store) GetChangesetEvent(ctx context.Context, opts GetChangesetEventOpts) (*campaigns.ChangesetEvent, error) {
	q := getChangesetEventQuery(&opts)

	var c campaigns.ChangesetEvent
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
-- source: enterprise/internal/campaigns/store.go:GetChangesetEvent
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
func (s *Store) ListChangesetEvents(ctx context.Context, opts ListChangesetEventsOpts) (cs []*campaigns.ChangesetEvent, next int64, err error) {
	q := listChangesetEventsQuery(&opts)

	cs = make([]*campaigns.ChangesetEvent, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.ChangesetEvent
		if err = scanChangesetEvent(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return c.ID, 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listChangesetEventsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:ListChangesetEvents
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
-- source: enterprise/internal/campaigns/store.go:CountChangesetEvents
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
func (s *Store) UpsertChangesetEvents(ctx context.Context, cs ...*campaigns.ChangesetEvent) (err error) {
	q, err := s.upsertChangesetEventsQuery(cs)
	if err != nil {
		return err
	}

	i := -1
	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		i++
		err = scanChangesetEvent(cs[i], sc)
		return cs[i].ID, 1, err
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
-- source: enterprise/internal/campaigns/store.go:UpsertChangesetEvents
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

func (s *Store) upsertChangesetEventsQuery(es []*campaigns.ChangesetEvent) (*sqlf.Query, error) {
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

func batchChangesetEventsQuery(fmtstr string, es []*campaigns.ChangesetEvent) (*sqlf.Query, error) {
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
func (s *Store) CreateCampaign(ctx context.Context, c *campaigns.Campaign) error {
	q, err := s.createCampaignQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaign(c, sc)
		return c.ID, 1, err
	})
}

var createCampaignQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CreateCampaign
INSERT INTO campaigns (
  name,
  description,
  branch,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
  id,
  name,
  description,
  branch,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
`

func (s *Store) createCampaignQuery(c *campaigns.Campaign) (*sqlf.Query, error) {
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
		c.Branch,
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
func (s *Store) UpdateCampaign(ctx context.Context, c *campaigns.Campaign) error {
	q, err := s.updateCampaignQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaign(c, sc)
		return c.ID, 1, err
	})
}

var updateCampaignQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:UpdateCampaign
UPDATE campaigns
SET (
  name,
  description,
  branch,
  author_id,
  namespace_user_id,
  namespace_org_id,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  id,
  name,
  description,
  branch,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
`

func (s *Store) updateCampaignQuery(c *campaigns.Campaign) (*sqlf.Query, error) {
	changesetIDs, err := jsonSetColumn(c.ChangesetIDs)
	if err != nil {
		return nil, err
	}

	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateCampaignQueryFmtstr,
		c.Name,
		c.Description,
		c.Branch,
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
-- source: enterprise/internal/campaigns/store.go:DeleteCampaign
DELETE FROM campaigns WHERE id = %s
`

// CountCampaignsOpts captures the query options needed for
// counting campaigns.
type CountCampaignsOpts struct {
	ChangesetID int64
	State       campaigns.CampaignState
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
-- source: enterprise/internal/campaigns/store.go:CountCampaigns
SELECT COUNT(id)
FROM campaigns
WHERE %s
`

func countCampaignsQuery(opts *CountCampaignsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ChangesetID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_ids ? %s", opts.ChangesetID))
	}

	switch opts.State {
	case campaigns.CampaignStateOpen:
		preds = append(preds, sqlf.Sprintf("closed_at IS NULL"))
	case campaigns.CampaignStateClosed:
		preds = append(preds, sqlf.Sprintf("closed_at IS NOT NULL"))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countCampaignsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetCampaignOpts captures the query options needed for getting a Campaign
type GetCampaignOpts struct {
	ID             int64
	CampaignPlanID int64
}

// GetCampaign gets a campaign matching the given options.
func (s *Store) GetCampaign(ctx context.Context, opts GetCampaignOpts) (*campaigns.Campaign, error) {
	q := getCampaignQuery(&opts)

	var c campaigns.Campaign
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
-- source: enterprise/internal/campaigns/store.go:GetCampaign
SELECT
  id,
  name,
  description,
  branch,
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

	if opts.CampaignPlanID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_plan_id = %s", opts.CampaignPlanID))
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
	State       campaigns.CampaignState
}

// ListCampaigns lists Campaigns with the given filters.
func (s *Store) ListCampaigns(ctx context.Context, opts ListCampaignsOpts) (cs []*campaigns.Campaign, next int64, err error) {
	q := listCampaignsQuery(&opts)

	cs = make([]*campaigns.Campaign, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.Campaign
		if err = scanCampaign(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return c.ID, 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listCampaignsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:ListCampaigns
SELECT
  id,
  name,
  description,
  branch,
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

	switch opts.State {
	case campaigns.CampaignStateOpen:
		preds = append(preds, sqlf.Sprintf("closed_at IS NULL"))
	case campaigns.CampaignStateClosed:
		preds = append(preds, sqlf.Sprintf("closed_at IS NOT NULL"))
	}

	return sqlf.Sprintf(
		listCampaignsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

// CreateCampaignPlan creates the given CampaignPlan.
func (s *Store) CreateCampaignPlan(ctx context.Context, c *campaigns.CampaignPlan) error {
	q, err := s.createCampaignPlanQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignPlan(c, sc)
		return c.ID, 1, err
	})
}

var createCampaignPlanQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CreateCampaignPlan
INSERT INTO campaign_plans (
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at,
  user_id
)
VALUES (%s, %s, %s, %s, %s, %s)
RETURNING
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at,
  user_id
`

func (s *Store) createCampaignPlanQuery(c *campaigns.CampaignPlan) (*sqlf.Query, error) {
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
		c.UserID,
	), nil
}

// UpdateCampaignPlan updates the given CampaignPlan.
func (s *Store) UpdateCampaignPlan(ctx context.Context, c *campaigns.CampaignPlan) error {
	q, err := s.updateCampaignPlanQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignPlan(c, sc)
		return c.ID, 1, err
	})
}

var updateCampaignPlanQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:UpdateCampaignPlan
UPDATE campaign_plans
SET (
  campaign_type,
  arguments,
  canceled_at,
  updated_at,
  user_id
) = (%s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at,
  user_id
`

func (s *Store) updateCampaignPlanQuery(c *campaigns.CampaignPlan) (*sqlf.Query, error) {
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
		c.UserID,
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
-- source: enterprise/internal/campaigns/store.go:DeleteCampaignPlan
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
-- source: enterprise/internal/campaigns/store.go:DeleteExpiredCampaignPlans
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
-- todo: what is the case we want to handle here?
NOT EXISTS (
	SELECT 1
	FROM
	action_executions
	WHERE
	action_executions.campaign_plan = campaign_plans.id
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
-- source: enterprise/internal/campaigns/store.go:CountCampaignPlans
SELECT COUNT(id)
FROM campaign_plans
`

// GetCampaignPlanOpts captures the query options needed for getting a CampaignPlan
type GetCampaignPlanOpts struct {
	ID int64
}

// GetCampaignPlan gets a code mod matching the given options.
func (s *Store) GetCampaignPlan(ctx context.Context, opts GetCampaignPlanOpts) (*campaigns.CampaignPlan, error) {
	q := getCampaignPlanQuery(&opts)

	var c campaigns.CampaignPlan
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
-- source: enterprise/internal/campaigns/store.go:GetCampaignPlan
SELECT
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at,
  user_id
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

// GetCampaignPlanStatus gets the campaigns.BackgroundProcessStatus for a CampaignPlan
func (s *Store) GetCampaignPlanStatus(ctx context.Context, id int64) (*campaigns.BackgroundProcessStatus, error) {
	return s.queryBackgroundProcessStatus(ctx, sqlf.Sprintf(
		getCampaignPlanStatusQueryFmtstr,
		id,
		sqlf.Sprintf("campaign_plan_id = %s", id),
	))
}

var getCampaignPlanStatusQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:GetCampaignPlanStatus
SELECT
  (SELECT canceled_at IS NOT NULL FROM campaign_plans WHERE id = %s) AS canceled,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE finished_at IS NULL) AS pending,
  COUNT(*) FILTER (WHERE finished_at IS NOT NULL) AS completed,
  array_agg(error) FILTER (WHERE error != '') AS errors
FROM campaign_jobs
WHERE %s
LIMIT 1;
`

// GetCampaignStatus gets the campaigns.BackgroundProcessStatus for a Campaign
func (s *Store) GetCampaignStatus(ctx context.Context, id int64) (*campaigns.BackgroundProcessStatus, error) {
	return s.queryBackgroundProcessStatus(ctx, sqlf.Sprintf(
		getCampaignStatusQueryFmtstr,
		sqlf.Sprintf("campaign_id = %s", id),
	))
}

func (s *Store) queryBackgroundProcessStatus(ctx context.Context, q *sqlf.Query) (*campaigns.BackgroundProcessStatus, error) {
	var status campaigns.BackgroundProcessStatus
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanBackgroundProcessStatus(&status, sc)
	})
	if err != nil {
		return nil, err
	}

	status.ProcessState = campaigns.BackgroundProcessStateCompleted
	switch {
	case status.Canceled:
		status.ProcessState = campaigns.BackgroundProcessStateCanceled
	case status.Pending > 0:
		status.ProcessState = campaigns.BackgroundProcessStateProcessing
	case status.Completed == status.Total && len(status.ProcessErrors) == 0:
		status.ProcessState = campaigns.BackgroundProcessStateCompleted
	case status.Completed == status.Total && len(status.ProcessErrors) != 0:
		status.ProcessState = campaigns.BackgroundProcessStateErrored
	}
	return &status, nil
}

var getCampaignStatusQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:GetCampaignStatus
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
func (s *Store) ListCampaignPlans(ctx context.Context, opts ListCampaignPlansOpts) (cs []*campaigns.CampaignPlan, next int64, err error) {
	q := listCampaignPlansQuery(&opts)

	cs = make([]*campaigns.CampaignPlan, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.CampaignPlan
		if err = scanCampaignPlan(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return c.ID, 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listCampaignPlansQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:ListCampaignPlans
SELECT
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at,
  user_id
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
// Due to a unique constraint in the DB it is safe to call this more than once
// with the same input. Only one job will be added and the other calls will return an error
func (s *Store) CreateCampaignJob(ctx context.Context, c *campaigns.CampaignJob) error {
	q, err := s.createCampaignJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignJob(c, sc)
		return c.ID, 1, err
	})
}

var createCampaignJobQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CreateCampaignJob
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

func (s *Store) createCampaignJobQuery(c *campaigns.CampaignJob) (*sqlf.Query, error) {
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
func (s *Store) UpdateCampaignJob(ctx context.Context, c *campaigns.CampaignJob) error {
	q, err := s.updateCampaignJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignJob(c, sc)
		return c.ID, 1, err
	})
}

var updateCampaignJobQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:UpdateCampaignJob
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

func (s *Store) updateCampaignJobQuery(c *campaigns.CampaignJob) (*sqlf.Query, error) {
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
-- source: enterprise/internal/campaigns/store.go:DeleteCampaignJob
DELETE FROM campaign_jobs WHERE id = %s
`

// CountCampaignJobsOpts captures the query options needed for
// counting campaign jobs
type CountCampaignJobsOpts struct {
	CampaignPlanID int64
	OnlyFinished   bool
	OnlyWithDiff   bool

	// If this is set to a Campaign ID only the CampaignJobs are returned that
	// are _not_ associated with a successfully completed ChangesetJob (meaning
	// that a Changeset on the codehost was created) for the given Campaign.
	OnlyUnpublishedInCampaign int64
}

// CountCampaignJobs returns the number of CampaignJobs in the database.
func (s *Store) CountCampaignJobs(ctx context.Context, opts CountCampaignJobsOpts) (count int64, _ error) {
	q := countCampaignJobsQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countCampaignJobsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CountCampaignJobs
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

	if opts.OnlyUnpublishedInCampaign != 0 {
		preds = append(preds, onlyUnpublishedInCampaignQuery(opts.OnlyUnpublishedInCampaign))
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
func (s *Store) GetCampaignJob(ctx context.Context, opts GetCampaignJobOpts) (*campaigns.CampaignJob, error) {
	q := getCampaignJobQuery(&opts)

	var c campaigns.CampaignJob
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
-- source: enterprise/internal/campaigns/store.go:GetCampaignJob
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

	// If this is set to a Campaign ID only the CampaignJobs are returned that
	// are _not_ associated with a successfully completed ChangesetJob (meaning
	// that a Changeset on the codehost was created) for the given Campaign.
	OnlyUnpublishedInCampaign int64
}

// ListCampaignJobs lists CampaignJobs with the given filters.
func (s *Store) ListCampaignJobs(ctx context.Context, opts ListCampaignJobsOpts) (cs []*campaigns.CampaignJob, next int64, err error) {
	q := listCampaignJobsQuery(&opts)

	cs = make([]*campaigns.CampaignJob, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.CampaignJob
		if err = scanCampaignJob(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return c.ID, 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listCampaignJobsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:ListCampaignJobs
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

	if opts.OnlyUnpublishedInCampaign != 0 {
		preds = append(preds, onlyUnpublishedInCampaignQuery(opts.OnlyUnpublishedInCampaign))
	}

	return sqlf.Sprintf(
		listCampaignJobsQueryFmtstr+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

var onlyUnpublishedInCampaignQueryFmtstr = `
NOT EXISTS (
  SELECT 1
  FROM changeset_jobs
  WHERE
    campaign_job_id = campaign_jobs.id
  AND
    campaign_id = %s
  AND
    changeset_id IS NOT NULL
)
`

func onlyUnpublishedInCampaignQuery(campaignID int64) *sqlf.Query {
	return sqlf.Sprintf(onlyUnpublishedInCampaignQueryFmtstr, campaignID)
}

// CreateChangesetJob creates the given ChangesetJob.
func (s *Store) CreateChangesetJob(ctx context.Context, c *campaigns.ChangesetJob) error {
	q, err := s.createChangesetJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanChangesetJob(c, sc)
		return c.ID, 1, err
	})
}

var createChangesetJobQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CreateChangesetJob
INSERT INTO changeset_jobs (
  campaign_id,
  campaign_job_id,
  changeset_id,
  branch,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  branch,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
`

func (s *Store) createChangesetJobQuery(c *campaigns.ChangesetJob) (*sqlf.Query, error) {
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
		c.Branch,
		nullStringColumn(c.Error),
		nullTimeColumn(c.StartedAt),
		nullTimeColumn(c.FinishedAt),
		c.CreatedAt,
		c.UpdatedAt,
	), nil
}

// UpdateChangesetJob updates the given ChangesetJob.
func (s *Store) UpdateChangesetJob(ctx context.Context, c *campaigns.ChangesetJob) error {
	q, err := s.updateChangesetJobQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanChangesetJob(c, sc)
		return c.ID, 1, err
	})
}

var updateChangesetJobQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:UpdateChangesetJob
UPDATE changeset_jobs
SET (
  campaign_id,
  campaign_job_id,
  changeset_id,
  branch,
  error,
  started_at,
  finished_at,
  updated_at
) = (%s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  branch,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
`

func (s *Store) updateChangesetJobQuery(c *campaigns.ChangesetJob) (*sqlf.Query, error) {
	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateChangesetJobQueryFmtstr,
		c.CampaignID,
		c.CampaignJobID,
		nullInt64Column(c.ChangesetID),
		c.Branch,
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
-- source: enterprise/internal/campaigns/store.go:DeleteChangesetJob
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
-- source: enterprise/internal/campaigns/store.go:CountChangesetJobs
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

// GetLatestChangesetJobCreatedAt returns the most recent created_at time for all changeset jobs
// for a campaign. But only if they have all been created, one for each CampaignJob belonging to the CampaignPlan attached to the Campaign. If not, it returns a zero time.Time.
func (s *Store) GetLatestChangesetJobCreatedAt(ctx context.Context, campaignID int64) (time.Time, error) {
	q := sqlf.Sprintf(getLatestChangesetJobPublishedAtFmtstr, campaignID)
	var createdAt time.Time
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&dbutil.NullTime{Time: &createdAt})
		if err != nil {
			return 0, 0, err
		}
		return 0, 1, nil
	})
	if err != nil {
		return createdAt, err
	}
	return createdAt, nil
}

var getLatestChangesetJobPublishedAtFmtstr = `
SELECT
  max(changeset_jobs.created_at)
FROM campaign_jobs
INNER JOIN campaigns ON campaign_jobs.campaign_plan_id = campaigns.campaign_plan_id
LEFT JOIN changeset_jobs ON changeset_jobs.campaign_job_id = campaign_jobs.id
WHERE campaigns.id = %s
HAVING count(*) FILTER (WHERE changeset_jobs.created_at IS NULL) = 0;
`

// GetChangesetJobOpts captures the query options needed for getting a ChangesetJob
type GetChangesetJobOpts struct {
	ID            int64
	CampaignJobID int64
	CampaignID    int64
	ChangesetID   int64
}

// GetChangesetJob gets a ChangesetJob matching the given options.
func (s *Store) GetChangesetJob(ctx context.Context, opts GetChangesetJobOpts) (*campaigns.ChangesetJob, error) {
	q := getChangesetJobQuery(&opts)

	var c campaigns.ChangesetJob
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
-- source: enterprise/internal/campaigns/store.go:GetChangesetJob
SELECT
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  branch,
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

	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_id = %s", opts.CampaignID))
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
// listing changeset jobs.
type ListChangesetJobsOpts struct {
	CampaignID     int64
	CampaignPlanID int64
	Cursor         int64
	Limit          int
}

// ListChangesetJobs lists ChangesetJobs with the given filters.
func (s *Store) ListChangesetJobs(ctx context.Context, opts ListChangesetJobsOpts) (cs []*campaigns.ChangesetJob, next int64, err error) {
	q := listChangesetJobsQuery(&opts)

	cs = make([]*campaigns.ChangesetJob, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.ChangesetJob
		if err = scanChangesetJob(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return c.ID, 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listChangesetJobsQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ListChangesetJobs
SELECT
  changeset_jobs.id,
  changeset_jobs.campaign_id,
  changeset_jobs.campaign_job_id,
  changeset_jobs.changeset_id,
  changeset_jobs.branch,
  changeset_jobs.error,
  changeset_jobs.started_at,
  changeset_jobs.finished_at,
  changeset_jobs.created_at,
  changeset_jobs.updated_at
FROM changeset_jobs
`

var listChangesetJobsQueryFmtstrConditions = `
WHERE %s
ORDER BY changeset_jobs.id ASC
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
		sqlf.Sprintf("changeset_jobs.id >= %s", opts.Cursor),
	}

	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_jobs.campaign_id = %s", opts.CampaignID))
	}

	var joinClause string
	if opts.CampaignPlanID != 0 {
		joinClause = "JOIN campaigns ON changeset_jobs.campaign_id = campaigns.id"

		preds = append(preds, sqlf.Sprintf("campaigns.campaign_plan_id = %s", opts.CampaignPlanID))
	}

	queryTemplate := listChangesetJobsQueryFmtstrSelect + joinClause +
		listChangesetJobsQueryFmtstrConditions + limitClause

	return sqlf.Sprintf(queryTemplate, sqlf.Join(preds, "\n AND "))
}

// ResetFailedChangesetJobs resets the Error, StartedAt and FinishedAt fields
// of the ChangesetJobs belonging to the Campaign with the given ID that
// resulted in an error.
func (s *Store) ResetFailedChangesetJobs(ctx context.Context, campaignID int64) (err error) {
	q := resetChangesetJobsQuery(campaignID, true)

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		return 0, 1, nil
	})
}

// ResetChangesetJobs resets the Error, StartedAt and FinishedAt fields
// of all ChangesetJobs belonging to the Campaign with the given ID.
func (s *Store) ResetChangesetJobs(ctx context.Context, campaignID int64) (err error) {
	q := resetChangesetJobsQuery(campaignID, false)

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		return 0, 1, nil
	})
}

func resetChangesetJobsQuery(campaignID int64, onlyErrored bool) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("campaign_id = %s", campaignID),
	}

	if onlyErrored {
		preds = append(preds, sqlf.Sprintf("error != ''"))
	}

	return sqlf.Sprintf(
		resetChangesetJobsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

var resetChangesetJobsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:resetChangesetJobsQuery
UPDATE changeset_jobs
SET
  error = '',
  started_at = NULL,
  finished_at = NULL
WHERE %s
`

// GetGithubExternalIDForRefs allows us to find the external id for GitHub pull requests based on
// a slice of head refs. We need this in order to match incoming status webhooks to pull requests as
// the only information they provide is the remote branch
func (s *Store) GetGithubExternalIDForRefs(ctx context.Context, refs []string) ([]string, error) {
	queryFmtString := `
SELECT external_id FROM changesets
WHERE external_service_type = 'github'
AND external_branch IN (%s)
ORDER BY id ASC
`
	inClause := make([]*sqlf.Query, 0, len(refs))
	for _, ref := range refs {
		if ref == "" {
			continue
		}
		inClause = append(inClause, sqlf.Sprintf("%s", ref))
	}
	q := sqlf.Sprintf(queryFmtString, sqlf.Join(inClause, ","))
	ids := make([]string, 0, len(refs))
	_, _, err := s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var s string
		err = sc.Scan(&s)
		if err != nil {
			return 0, 0, err
		}
		ids = append(ids, s)
		return 0, 1, nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
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

func scanChangeset(t *campaigns.Changeset, s scanner) error {
	var metadata json.RawMessage

	var (
		externalState       string
		externamReviewState string
		externalCheckState  string
	)
	err := s.Scan(
		&t.ID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&metadata,
		&dbutil.JSONInt64Set{Set: &t.CampaignIDs},
		&t.ExternalID,
		&t.ExternalServiceType,
		&t.ExternalBranch,
		&dbutil.NullTime{Time: &t.ExternalDeletedAt},
		&dbutil.NullTime{Time: &t.ExternalUpdatedAt},
		&dbutil.NullString{S: &externalState},
		&dbutil.NullString{S: &externamReviewState},
		&dbutil.NullString{S: &externalCheckState},
	)
	if err != nil {
		return errors.Wrap(err, "scanning changeset")
	}

	t.ExternalState = campaigns.ChangesetState(externalState)
	t.ExternalReviewState = campaigns.ChangesetReviewState(externamReviewState)
	t.ExternalCheckState = campaigns.ChangesetCheckState(externalCheckState)

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

func scanChangesetEvent(e *campaigns.ChangesetEvent, s scanner) error {
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

	e.Metadata, err = campaigns.NewChangesetEventMetadata(e.Kind)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(metadata, e.Metadata); err != nil {
		return errors.Wrapf(err, "scanChangesetEvent: failed to unmarshal %q metadata", e.Kind)
	}

	return nil
}

func scanCampaign(c *campaigns.Campaign, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.Name,
		&c.Description,
		&dbutil.NullString{S: &c.Branch},
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

func scanCampaignPlan(c *campaigns.CampaignPlan, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.CampaignType,
		&c.Arguments,
		&dbutil.NullTime{Time: &c.CanceledAt},
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.UserID,
	)
}

func scanCampaignJob(c *campaigns.CampaignJob, s scanner) error {
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

func scanChangesetJob(c *campaigns.ChangesetJob, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.CampaignID,
		&c.CampaignJobID,
		&dbutil.NullInt64{N: &c.ChangesetID},
		&c.Branch,
		&dbutil.NullString{S: &c.Error},
		&dbutil.NullTime{Time: &c.StartedAt},
		&dbutil.NullTime{Time: &c.FinishedAt},
		&c.CreatedAt,
		&c.UpdatedAt,
	)
}

func scanBackgroundProcessStatus(b *campaigns.BackgroundProcessStatus, s scanner) error {
	return s.Scan(
		&b.Canceled,
		&b.Total,
		&b.Pending,
		&b.Completed,
		pq.Array(&b.ProcessErrors),
	)
}

func scanAction(a *campaigns.Action, s scanner) error {
	return s.Scan(
		&a.ID,
		&a.CampaignID,
		&a.Schedule,
		&a.CancelPrevious,
		&a.SavedSearchID,
		&a.Steps,
		&a.EnvStr,
	)
}

func scanActionExecution(a *campaigns.ActionExecution, s scanner) error {
	return s.Scan(
		&a.ID,
		&a.Steps,
		&a.EnvStr,
		&a.InvokationReason,
		&a.CampaignPlanID,
		&a.ActionID,
		&dbutil.NullTime{Time: &a.ExecutionStart},
		&dbutil.NullTime{Time: &a.ExecutionEnd},
	)
}

func scanCount(count *int64, s scanner) error {
	return s.Scan(
		count,
	)
}

func scanActionJob(a *campaigns.ActionJob, s scanner) error {
	return s.Scan(
		&a.ID,
		&a.Log,
		&dbutil.NullTime{Time: &a.ExecutionStart},
		&dbutil.NullTime{Time: &a.ExecutionEnd},
		&dbutil.NullTime{Time: &a.RunnerSeenAt},
		&a.Patch,
		&a.State,
		&a.RepoID,
		&a.ExecutionID,
		&a.BaseRevision,
		&a.BaseReference,
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

// ListActionsOpts captures the query options needed for
// listing actions.
type ListActionsOpts struct {
	Cursor int64
	Limit  int
}

// ListActions lists Actions with the given filters.
func (s *Store) ListActions(ctx context.Context, opts ListActionsOpts) (actions []*campaigns.Action, totalCount int64, err error) {
	q := listActionsQuery(&opts)

	actions = make([]*campaigns.Action, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var a campaigns.Action
		if err = scanAction(&a, sc); err != nil {
			return 0, 0, err
		}
		actions = append(actions, &a)
		return a.ID, 1, err
	})

	q = sqlf.Sprintf("SELECT COUNT(*) FROM actions")
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		if err = scanCount(&totalCount, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, err
	})
	if err != nil {
		return nil, 0, err
	}

	return actions, totalCount, err
}

var listActionsQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ListActions
SELECT
	actions.id,
	actions.campaign,
	actions.schedule,
	actions.cancel_previous,
	actions.saved_search,
	actions.steps,
	actions.env
FROM actions
`

var listActionsQueryFmtstrConditions = `
WHERE %s
ORDER BY actions.id ASC
`

func listActionsQuery(opts *ListActionsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var limitClause string
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}

	preds := []*sqlf.Query{
		sqlf.Sprintf("actions.id >= %s", opts.Cursor),
	}

	queryTemplate := listActionsQueryFmtstrSelect + listActionsQueryFmtstrConditions + limitClause

	return sqlf.Sprintf(queryTemplate, sqlf.Join(preds, "\n AND "))
}

// UpdateActionJobOpts captures the query options needed for
// listing actions.
type UpdateActionJobOpts struct {
	ID             int64
	State          *campaigns.ActionJobState
	Log            *string
	Patch          *string
	ExecutionStart *time.Time
	ExecutionEnd   *time.Time
	RunnerSeenAt   *time.Time
}

// UpdateActionJob lists Actions with the given filters.
func (s *Store) UpdateActionJob(ctx context.Context, opts UpdateActionJobOpts) (*campaigns.ActionJob, error) {
	q := updateActionJobQuery(&opts)

	var job campaigns.ActionJob
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanActionJob(&job, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &job, err
}

var updateActionJobQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:UpdateActionJob
UPDATE
	action_jobs
SET %s
WHERE action_jobs.id = %d
RETURNING
	id,
	log,
	execution_start,
	execution_end,
	runner_seen_at,
	patch,
	state,
	repository,
	execution,
	base_revision,
	base_reference
`

func updateActionJobQuery(opts *UpdateActionJobOpts) *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.State != nil {
		preds = append(preds, sqlf.Sprintf("state = %s", string(*opts.State)))
	}
	if opts.Log != nil {
		preds = append(preds, sqlf.Sprintf("log = (CASE WHEN log IS NOT NULL THEN log ELSE '' END) || %s", *opts.Log))
	}
	if opts.Patch != nil {
		preds = append(preds, sqlf.Sprintf("patch = %s", *opts.Patch))
	}
	if opts.ExecutionStart != nil {
		if (*opts.ExecutionStart).IsZero() {
			preds = append(preds, sqlf.Sprintf("execution_start = NULL"))
		} else {
			// todo may throw
			time, _ := (*opts.ExecutionStart).MarshalText()
			preds = append(preds, sqlf.Sprintf("execution_start = %s", time))
		}
	}
	if opts.ExecutionEnd != nil {
		if (*opts.ExecutionEnd).IsZero() {
			preds = append(preds, sqlf.Sprintf("execution_end = NULL"))
		} else {
			// todo may throw
			time, _ := (*opts.ExecutionEnd).MarshalText()
			preds = append(preds, sqlf.Sprintf("execution_end = %s", time))
		}
	}
	if opts.RunnerSeenAt != nil {
		if (*opts.RunnerSeenAt).IsZero() {
			preds = append(preds, sqlf.Sprintf("runner_seen_at = NULL"))
		} else {
			// todo may throw
			time, _ := (*opts.RunnerSeenAt).MarshalText()
			preds = append(preds, sqlf.Sprintf("runner_seen_at = %s", time))
		}
	}

	queryTemplate := updateActionJobQueryFmtstrSelect

	return sqlf.Sprintf(queryTemplate, sqlf.Join(preds, ",\n "), opts.ID)
}

// ClearActionJobOpts captures the query options needed for clearing an action job
type ClearActionJobOpts struct {
	ID int64
}

// ClearActionJob resets an action job so it is retried.
func (s *Store) ClearActionJob(ctx context.Context, opts ClearActionJobOpts) error {
	q := clearActionJobQuery(&opts)
	_, _, err := s.query(ctx, q, nil)
	return err
}

var clearActionJobQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ClearActionJob
UPDATE
	action_jobs
SET
	log = NULL,
	execution_start = NULL,
	execution_end = NULL,
	runner_seen_at = NULL,
	patch = NULL,
	state = 'PENDING'
WHERE action_jobs.id = %d
`

func clearActionJobQuery(opts *ClearActionJobOpts) *sqlf.Query {
	queryTemplate := clearActionJobQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.ID)
}

// PullActionJob resets an action job so it is eventually retried by a runner.
func (s *Store) PullActionJob(ctx context.Context) (*campaigns.ActionJob, error) {
	q := pullActionJobQuery()

	var job campaigns.ActionJob
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanActionJob(&job, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	return &job, err
}

var pullActionJobQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:PullActionJob
UPDATE
	action_jobs
SET
	execution_start = NOW(),
	state = 'RUNNING',
	runner_seen_at = NOW()
WHERE
	-- todo: just the IN query is not atomic and jobs were pulled twice. the second AND will yield a null job though, wasting one pullActionJob request, it's just a hot fix
	action_jobs.id IN (SELECT id FROM action_jobs WHERE action_jobs.state = 'PENDING' ORDER BY action_jobs.id ASC LIMIT 1)
	AND action_jobs.state = 'PENDING'
RETURNING
	id,
	log,
	execution_start,
	execution_end,
	runner_seen_at,
	patch,
	state,
	repository,
	execution,
	base_revision,
	base_reference
`

func pullActionJobQuery() *sqlf.Query {
	queryTemplate := pullActionJobQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate)
}

type ActionJobByIDOpts struct {
	ID int64
}

// ActionJobByID resets an action job so it is eventually retried by a runner.
func (s *Store) ActionJobByID(ctx context.Context, opts ActionJobByIDOpts) (*campaigns.ActionJob, error) {
	q := actionJobByIDQuery(&opts)

	var job campaigns.ActionJob
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanActionJob(&job, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &job, err
}

var actionJobByIDQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ActionJobByID
SELECT
	action_jobs.id,
	action_jobs.log,
	action_jobs.execution_start,
	action_jobs.execution_end,
	action_jobs.runner_seen_at,
	action_jobs.patch,
	action_jobs.state,
	action_jobs.repository,
	action_jobs.execution,
	action_jobs.base_revision,
	action_jobs.base_reference
FROM
	action_jobs
WHERE
	action_jobs.id = %d
`

func actionJobByIDQuery(opts *ActionJobByIDOpts) *sqlf.Query {
	queryTemplate := actionJobByIDQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.ID)
}

type ActionByIDOpts struct {
	ID int64
}

// ActionByID resets an action job so it is eventually retried by a runner.
func (s *Store) ActionByID(ctx context.Context, opts ActionByIDOpts) (*campaigns.Action, error) {
	q := actionByIDQuery(&opts)

	var a campaigns.Action
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanAction(&a, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &a, err
}

var actionByIDQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ActionByID
SELECT
	actions.id,
	actions.campaign,
	actions.schedule,
	actions.cancel_previous,
	actions.saved_search,
	actions.steps,
	actions.env
FROM
	actions
WHERE
	actions.id = %d
`

func actionByIDQuery(opts *ActionByIDOpts) *sqlf.Query {
	queryTemplate := actionByIDQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.ID)
}

type ActionExecutionByIDOpts struct {
	ID int64
}

// ActionExecutionByID resets an action job so it is eventually retried by a runner.
func (s *Store) ActionExecutionByID(ctx context.Context, opts ActionExecutionByIDOpts) (*campaigns.ActionExecution, error) {
	q := actionExecutionByIDQuery(&opts)

	var a campaigns.ActionExecution
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanActionExecution(&a, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &a, err
}

var actionExecutionByIDQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ActionExecutionByID
SELECT
	action_executions.id,
	action_executions.steps,
	action_executions.env,
	action_executions.invokation_reason,
	action_executions.campaign_plan,
	action_executions.action,
	(SELECT MIN(action_jobs.execution_start) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_start,
	(SELECT MAX(action_jobs.execution_end) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_end
FROM
	action_executions
WHERE
	action_executions.id = %d
`

func actionExecutionByIDQuery(opts *ActionExecutionByIDOpts) *sqlf.Query {
	queryTemplate := actionExecutionByIDQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.ID)
}

// ListActionExecutionsOpts captures the query options needed for
// listing action executions.
type ListActionExecutionsOpts struct {
	Cursor   int64
	Limit    int
	ActionID *int64
}

// ListActionExecutions lists ActionExecutions with the given filters.
func (s *Store) ListActionExecutions(ctx context.Context, opts ListActionExecutionsOpts) (actionExecutions []*campaigns.ActionExecution, totalCount int64, err error) {
	q := listActionExecutionsQuery(&opts)

	actionExecutions = make([]*campaigns.ActionExecution, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var a campaigns.ActionExecution
		if err = scanActionExecution(&a, sc); err != nil {
			return 0, 0, err
		}
		actionExecutions = append(actionExecutions, &a)
		return a.ID, 1, err
	})

	q = sqlf.Sprintf("SELECT COUNT(*) FROM action_executions")
	countTemplate := "SELECT COUNT(*) FROM action_executions"
	if opts.ActionID != nil {
		countTemplate = countTemplate + " WHERE action = %d"
		q = sqlf.Sprintf(countTemplate, opts.ActionID)
	} else {
		q = sqlf.Sprintf(countTemplate)
	}
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		if err = scanCount(&totalCount, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, err
	})
	if err != nil {
		return nil, 0, err
	}

	return actionExecutions, totalCount, err
}

var listActionExecutionsQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ListActionExecutions
SELECT
	action_executions.id,
	action_executions.steps,
	action_executions.env,
	action_executions.invokation_reason,
	action_executions.campaign_plan,
	action_executions.action,
	(SELECT MIN(action_jobs.execution_start) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_start,
	(SELECT MAX(action_jobs.execution_end) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_end
FROM action_executions
`

var listActionExecutionsQueryFmtstrConditions = `
WHERE %s
ORDER BY action_executions.id ASC
`

func listActionExecutionsQuery(opts *ListActionExecutionsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var limitClause string
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}

	preds := []*sqlf.Query{
		sqlf.Sprintf("action_executions.id >= %s", opts.Cursor),
	}

	if opts.ActionID != nil {
		preds = append(preds, sqlf.Sprintf("action_executions.action = %s", opts.ActionID))
	}

	queryTemplate := listActionExecutionsQueryFmtstrSelect + listActionExecutionsQueryFmtstrConditions + limitClause

	return sqlf.Sprintf(queryTemplate, sqlf.Join(preds, "\n AND "))
}

type CreateActionExecutionOpts struct {
	InvokationReason campaigns.ActionExecutionInvokationReason
	Steps            string
	EnvStr           string
	ActionID         int64
}

// CreateActionExecution resets an action job so it is eventually retried by a runner.
func (s *Store) CreateActionExecution(ctx context.Context, opts CreateActionExecutionOpts) (*campaigns.ActionExecution, error) {
	q := createActionExecutionQuery(&opts)

	var a campaigns.ActionExecution
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanActionExecution(&a, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &a, err
}

var createActionExecutionQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:CreateActionExecution
INSERT INTO
	action_executions
	(steps, env, invokation_reason, action)
VALUES
	(%s, %s::json, %s, %d)
RETURNING
	action_executions.id,
	action_executions.steps,
	action_executions.env,
	action_executions.invokation_reason,
	action_executions.campaign_plan,
	action_executions.action,
	(SELECT MIN(action_jobs.execution_start) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_start,
	(SELECT MAX(action_jobs.execution_end) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_end
`

func createActionExecutionQuery(opts *CreateActionExecutionOpts) *sqlf.Query {
	queryTemplate := createActionExecutionQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.Steps, opts.EnvStr, string(opts.InvokationReason), opts.ActionID)
}

type CreateActionJobOpts struct {
	RepositoryID  int64
	ExecutionID   int64
	BaseRevision  string
	BaseReference string
}

// CreateActionJob resets an action job so it is eventually retried by a runner.
func (s *Store) CreateActionJob(ctx context.Context, opts CreateActionJobOpts) (*campaigns.ActionJob, error) {
	q := createActionJobQuery(&opts)

	var a campaigns.ActionJob
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanActionJob(&a, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &a, err
}

var createActionJobQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:CreateActionJob
INSERT INTO
	action_jobs
	(state, repository, execution, base_revision, base_reference)
VALUES
	('PENDING', %d, %d, %s, %s)
RETURNING
	action_jobs.id,
	action_jobs.log,
	action_jobs.execution_start,
	action_jobs.execution_end,
	action_jobs.runner_seen_at,
	action_jobs.patch,
	action_jobs.state,
	action_jobs.repository,
	action_jobs.execution,
	action_jobs.base_revision,
	action_jobs.base_reference
`

func createActionJobQuery(opts *CreateActionJobOpts) *sqlf.Query {
	queryTemplate := createActionJobQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.RepositoryID, opts.ExecutionID, opts.BaseRevision, opts.BaseReference)
}

// ListActionJobsOpts captures the query options needed for
// listing action executions.
type ListActionJobsOpts struct {
	Cursor      int64
	Limit       int
	ExecutionID *int64
}

// ListActionJobs lists ActionJobs with the given filters.
func (s *Store) ListActionJobs(ctx context.Context, opts ListActionJobsOpts) (actionJobs []*campaigns.ActionJob, totalCount int64, err error) {
	q := listActionJobsQuery(&opts)

	actionJobs = make([]*campaigns.ActionJob, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var a campaigns.ActionJob
		if err = scanActionJob(&a, sc); err != nil {
			return 0, 0, err
		}
		actionJobs = append(actionJobs, &a)
		return a.ID, 1, err
	})

	q = sqlf.Sprintf("SELECT COUNT(*) FROM action_jobs")
	countTemplate := "SELECT COUNT(*) FROM action_jobs"
	if opts.ExecutionID != nil {
		countTemplate = countTemplate + " WHERE execution = %d"
		q = sqlf.Sprintf(countTemplate, opts.ExecutionID)
	} else {
		q = sqlf.Sprintf(countTemplate)
	}
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		if err = scanCount(&totalCount, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, err
	})
	if err != nil {
		return nil, 0, err
	}

	return actionJobs, totalCount, err
}

var listActionJobsQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ListActionJobs
SELECT
	action_jobs.id,
	action_jobs.log,
	action_jobs.execution_start,
	action_jobs.execution_end,
	action_jobs.runner_seen_at,
	action_jobs.patch,
	action_jobs.state,
	action_jobs.repository,
	action_jobs.execution,
	action_jobs.base_revision,
	action_jobs.base_reference
FROM action_jobs
`

var listActionJobsQueryFmtstrConditions = `
WHERE %s
ORDER BY action_jobs.id ASC
`

func listActionJobsQuery(opts *ListActionJobsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var limitClause string
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}

	preds := []*sqlf.Query{
		sqlf.Sprintf("action_jobs.id >= %s", opts.Cursor),
	}

	if opts.ExecutionID != nil {
		preds = append(preds, sqlf.Sprintf("action_jobs.execution = %s", opts.ExecutionID))
	}

	queryTemplate := listActionJobsQueryFmtstrSelect + listActionJobsQueryFmtstrConditions + limitClause

	return sqlf.Sprintf(queryTemplate, sqlf.Join(preds, "\n AND "))
}

type CreateActionOpts struct {
	Steps string
}

// CreateAction creates a new action in the database.
func (s *Store) CreateAction(ctx context.Context, opts CreateActionOpts) (*campaigns.Action, error) {
	q := createActionQuery(&opts)

	var a campaigns.Action
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanAction(&a, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &a, err
}

// todo: pass env
var createActionQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:CreateAction
INSERT INTO
	actions
	(steps, env)
VALUES
	(%s, '[]'::json)
RETURNING
	actions.id,
	actions.campaign,
	actions.schedule,
	actions.cancel_previous,
	actions.saved_search,
	actions.steps,
	actions.env
`

func createActionQuery(opts *CreateActionOpts) *sqlf.Query {
	queryTemplate := createActionQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.Steps)
}

type UpdateActionOpts struct {
	ActionID int64
	Steps    string
}

// UpdateAction creates a new action in the database.
func (s *Store) UpdateAction(ctx context.Context, opts UpdateActionOpts) (*campaigns.Action, error) {
	q := updateActionQuery(&opts)

	var a campaigns.Action
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanAction(&a, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &a, err
}

// todo: pass env
var updateActionQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:UpdateAction
UPDATE
	actions
SET
	steps = %s
WHERE
	actions.id = %d
RETURNING
	actions.id,
	actions.campaign,
	actions.schedule,
	actions.cancel_previous,
	actions.saved_search,
	actions.steps,
	actions.env
`

func updateActionQuery(opts *UpdateActionOpts) *sqlf.Query {
	queryTemplate := updateActionQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.Steps, opts.ActionID)
}

type UpdateActionExecutionOpts struct {
	ExecutionID    int64
	CampaignPlanID int64
}

// UpdateActionExecution creates a new action in the database.
func (s *Store) UpdateActionExecution(ctx context.Context, opts UpdateActionExecutionOpts) (*campaigns.ActionExecution, error) {
	q := updateActionExecutionQuery(&opts)

	var a campaigns.ActionExecution
	_, _, err := s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err := scanActionExecution(&a, sc); err != nil {
			return 0, 0, err
		}
		return 0, 0, nil
	})

	// todo handle empty ie not-found error

	return &a, err
}

// todo: pass env
var updateActionExecutionQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:UpdateActionExecution
UPDATE
	action_executions
SET
	campaign_plan = %s
WHERE
	action_executions.id = %d
RETURNING
	action_executions.id,
	action_executions.steps,
	action_executions.env,
	action_executions.invokation_reason,
	action_executions.campaign_plan,
	action_executions.action,
	(SELECT MIN(action_jobs.execution_start) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_start,
	(SELECT MAX(action_jobs.execution_end) FROM action_jobs WHERE action_jobs.execution = action_executions.id) AS execution_end
`

func updateActionExecutionQuery(opts *UpdateActionExecutionOpts) *sqlf.Query {
	queryTemplate := updateActionExecutionQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.CampaignPlanID, opts.ExecutionID)
}

// todo: this can be combined with the generic ListActions store method. This was just quicker for now as the count query needs the new handling as well
// ListActionsBySavedSearchQueryOpts captures the query options needed for
// listing action executions.
type ListActionsBySavedSearchQueryOpts struct {
	SavedSearchQuery string
}

// ListActionsBySavedSearchQuery lists ActionExecutions with the given filters.
func (s *Store) ListActionsBySavedSearchQuery(ctx context.Context, opts ListActionsBySavedSearchQueryOpts) (actions []*campaigns.Action, err error) {
	q := listActionsBySavedSearchQueryQuery(&opts)

	actions = make([]*campaigns.Action, 0)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var a campaigns.Action
		if err = scanAction(&a, sc); err != nil {
			return 0, 0, err
		}
		actions = append(actions, &a)
		return a.ID, 1, err
	})

	if err != nil {
		return nil, err
	}

	return actions, err
}

var listActionsBySavedSearchQueryQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ListActionsBySavedSearchQuery
SELECT
	actions.id,
	actions.campaign,
	actions.schedule,
	actions.cancel_previous,
	actions.saved_search,
	actions.steps,
	actions.env
FROM
	actions
WHERE
	actions.saved_search IS NOT NULL
AND
	actions.saved_search IN (SELECT id FROM saved_searches WHERE query = %s)
ORDER BY actions.id ASC
`

func listActionsBySavedSearchQueryQuery(opts *ListActionsBySavedSearchQueryOpts) *sqlf.Query {
	queryTemplate := listActionsBySavedSearchQueryQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.SavedSearchQuery)
}

// ActionExecutionStatusOpts captures the query options needed for
// getting an action execution's status.
type ActionExecutionStatusOpts struct {
	ExecutionID int64
}

// Result of the query for the status of an execution
type ActionExecutionStatusResult struct {
	Total    int64
	Canceled bool
	Pending  int64
	Errored  bool
}

// ActionExecutionStatus lists ActionExecutions with the given filters.
func (s *Store) ActionExecutionStatus(ctx context.Context, opts ActionExecutionStatusOpts) (result *ActionExecutionStatusResult, err error) {
	q := actionExecutionStatusQuery(&opts)
	result = &ActionExecutionStatusResult{}
	_, _, err = s.query(ctx, q, func(sc scanner) (_, _ int64, err error) {
		if err = sc.Scan(
			&result.Total,
			&result.Canceled,
			&result.Pending,
			&result.Errored,
		); err != nil {
			return 0, 0, err
		}
		return 0, 0, err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

var actionExecutionStatusQueryFmtstrSelect = `
-- source: enterprise/internal/campaigns/store.go:ActionExecutionStatus
SELECT
	COUNT(*) AS total,
	COUNT(CASE action_jobs.state WHEN 'CANCELED' THEN 1 ELSE NULL END) > 0 AS canceled,
	COUNT(CASE action_jobs.state WHEN 'PENDING' THEN 1 WHEN 'RUNNING' THEN 1 ELSE NULL END) AS pending,
	COUNT(CASE action_jobs.state WHEN 'ERRORED' THEN 1 ELSE NULL END) > 0 AS errored
FROM
	action_jobs
WHERE
	action_jobs.execution = %d
`

func actionExecutionStatusQuery(opts *ActionExecutionStatusOpts) *sqlf.Query {
	queryTemplate := actionExecutionStatusQueryFmtstrSelect
	return sqlf.Sprintf(queryTemplate, opts.ExecutionID)
}
