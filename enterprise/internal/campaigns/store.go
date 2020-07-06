package campaigns

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/segmentio/fasthash/fnv1"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// seededRand is used to populate the RandID fields on CampaignSpec and
// ChangesetSpec when creating them.
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

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
// A pending job is one that has never been started.
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
			return 0, 0, errors.Wrap(err, "scanning changeset job row")
		}
		return job.ID, 1, nil
	})
	if err != nil {
		return false, errors.Wrap(err, "querying for pending changeset job")
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
	WHERE j.started_at IS NULL AND c.patch_set_id IS NOT NULL
	ORDER BY j.updated_at ASC
	FOR UPDATE SKIP LOCKED LIMIT 1
)
RETURNING j.id,
  j.campaign_id,
  j.patch_id,
  j.changeset_id,
  j.branch,
  j.error,
  j.started_at,
  j.finished_at,
  j.created_at,
  j.updated_at
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
      external_check_state  text,
      created_by_campaign   boolean,
      added_to_campaign     boolean,
      diff_stat_added       integer,
      diff_stat_changed     integer,
      diff_stat_deleted     integer,
      sync_state            jsonb
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
    external_check_state,
    created_by_campaign,
    added_to_campaign,
    diff_stat_added,
    diff_stat_changed,
    diff_stat_deleted,
    sync_state
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
    external_check_state,
    created_by_campaign,
    added_to_campaign,
    diff_stat_added,
    diff_stat_changed,
    diff_stat_deleted,
    sync_state
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
  COALESCE(changed.external_check_state, existing.external_check_state) AS external_check_state,
  COALESCE(changed.created_by_campaign, existing.created_by_campaign) AS created_by_campaign,
  COALESCE(changed.added_to_campaign, existing.added_to_campaign) AS added_to_campaign,
  COALESCE(changed.diff_stat_added, existing.diff_stat_added) AS diff_stat_added,
  COALESCE(changed.diff_stat_changed, existing.diff_stat_changed) AS diff_stat_changed,
  COALESCE(changed.diff_stat_deleted, existing.diff_stat_deleted) AS diff_stat_deleted,
  COALESCE(changed.sync_state, existing.sync_state) AS sync_state
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
		CreatedByCampaign   bool                            `json:"created_by_campaign"`
		AddedToCampaign     bool                            `json:"added_to_campaign"`
		DiffStatAdded       *int32                          `json:"diff_stat_added"`
		DiffStatChanged     *int32                          `json:"diff_stat_changed"`
		DiffStatDeleted     *int32                          `json:"diff_stat_deleted"`
		SyncState           json.RawMessage                 `json:"sync_state"`
	}

	records := make([]record, 0, len(cs))

	for _, c := range cs {
		metadata, err := jsonbColumn(c.Metadata)
		if err != nil {
			return nil, err
		}

		campaignIDs, err := jsonSetColumn(c.CampaignIDs)
		if err != nil {
			return nil, err
		}

		syncState, err := json.Marshal(c.SyncState)
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
			CreatedByCampaign:   c.CreatedByCampaign,
			AddedToCampaign:     c.AddedToCampaign,
			DiffStatAdded:       c.DiffStatAdded,
			DiffStatChanged:     c.DiffStatChanged,
			DiffStatDeleted:     c.DiffStatDeleted,
			SyncState:           syncState,
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
SELECT COUNT(changesets.id)
FROM changesets
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE %s
`

func countChangesetsQuery(opts *CountChangesetsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("changesets.campaign_ids ? %s", opts.CampaignID))
	}

	if opts.ExternalState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_state = %s", *opts.ExternalState))
	}
	if opts.ExternalReviewState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_review_state = %s", *opts.ExternalReviewState))
	}
	if opts.ExternalCheckState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_check_state = %s", *opts.ExternalCheckState))
	}

	return sqlf.Sprintf(countChangesetsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetChangesetOpts captures the query options needed for getting a Changeset
type GetChangesetOpts struct {
	ID                  int64
	RepoID              api.RepoID
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
  changesets.id,
  changesets.repo_id,
  changesets.created_at,
  changesets.updated_at,
  changesets.metadata,
  changesets.campaign_ids,
  changesets.external_id,
  changesets.external_service_type,
  changesets.external_branch,
  changesets.external_deleted_at,
  changesets.external_updated_at,
  changesets.external_state,
  changesets.external_review_state,
  changesets.external_check_state,
  changesets.created_by_campaign,
  changesets.added_to_campaign,
  changesets.diff_stat_added,
  changesets.diff_stat_changed,
  changesets.diff_stat_deleted,
  changesets.sync_state
FROM changesets
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE %s
LIMIT 1
`

func getChangesetQuery(opts *GetChangesetOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("changesets.id = %s", opts.ID))
	}

	if opts.RepoID != 0 {
		preds = append(preds, sqlf.Sprintf("changesets.repo_id = %s", opts.RepoID))
	}

	if opts.ExternalID != "" && opts.ExternalServiceType != "" {
		preds = append(preds,
			sqlf.Sprintf("changesets.external_id = %s", opts.ExternalID),
			sqlf.Sprintf("changesets.external_service_type = %s", opts.ExternalServiceType),
		)
	}

	return sqlf.Sprintf(getChangesetsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

type ListChangesetSyncDataOpts struct {
	// Return only the supplied changesets. If empty, all changesets are returned
	ChangesetIDs []int64
}

// ListChangesetSyncData returns sync data on all non-externally-deleted changesets
// that are part of at least one open campaign.
func (s *Store) ListChangesetSyncData(ctx context.Context, opts ListChangesetSyncDataOpts) ([]campaigns.ChangesetSyncData, error) {
	q := listChangesetSyncData(opts)
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
		&h.RepoExternalServiceID,
	)
}

func listChangesetSyncData(opts ListChangesetSyncDataOpts) *sqlf.Query {
	fmtString := `
 SELECT changesets.id,
        changesets.updated_at,
        max(ce.updated_at) AS latest_event,
        changesets.external_updated_at,
        r.external_service_id
 FROM changesets
 LEFT JOIN changeset_events ce ON changesets.id = ce.changeset_id
 JOIN campaigns ON campaigns.changeset_ids ? changesets.id::TEXT
 JOIN repo r ON changesets.repo_id = r.id
 WHERE %s
 GROUP BY changesets.id, r.id
 ORDER BY changesets.id ASC
`

	preds := []*sqlf.Query{
		sqlf.Sprintf("campaigns.closed_at IS NULL"),
		sqlf.Sprintf("r.deleted_at IS NULL"),
	}
	if len(opts.ChangesetIDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(opts.ChangesetIDs))
		for _, id := range opts.ChangesetIDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds, sqlf.Sprintf("changesets.id IN (%s)", sqlf.Join(ids, ",")))
	}

	return sqlf.Sprintf(fmtString, sqlf.Join(preds, "\n AND"))
}

// ListChangesetsOpts captures the query options needed for
// listing changesets.
type ListChangesetsOpts struct {
	Cursor               int64
	Limit                int
	CampaignID           int64
	IDs                  []int64
	WithoutDeleted       bool
	ExternalState        *campaigns.ChangesetState
	ExternalReviewState  *campaigns.ChangesetReviewState
	ExternalCheckState   *campaigns.ChangesetCheckState
	OnlyWithoutDiffStats bool
}

// ListChangesets lists Changesets with the given filters.
func (s *Store) ListChangesets(ctx context.Context, opts ListChangesetsOpts) (cs campaigns.Changesets, next int64, err error) {
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
  changesets.id,
  changesets.repo_id,
  changesets.created_at,
  changesets.updated_at,
  changesets.metadata,
  changesets.campaign_ids,
  changesets.external_id,
  changesets.external_service_type,
  changesets.external_branch,
  changesets.external_deleted_at,
  changesets.external_updated_at,
  changesets.external_state,
  changesets.external_review_state,
  changesets.external_check_state,
  changesets.created_by_campaign,
  changesets.added_to_campaign,
  changesets.diff_stat_added,
  changesets.diff_stat_changed,
  changesets.diff_stat_deleted,
  changesets.sync_state
FROM changesets
INNER JOIN repo ON repo.id = changesets.repo_id
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
		sqlf.Sprintf("changesets.id >= %s", opts.Cursor),
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.CampaignID != 0 {
		preds = append(preds, sqlf.Sprintf("changesets.campaign_ids ? %s", opts.CampaignID))
	}

	if len(opts.IDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(opts.IDs))
		for _, id := range opts.IDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds, sqlf.Sprintf("changesets.id IN (%s)", sqlf.Join(ids, ",")))
	}

	if opts.WithoutDeleted {
		preds = append(preds, sqlf.Sprintf("changesets.external_deleted_at IS NULL"))
	}

	if opts.ExternalState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_state = %s", *opts.ExternalState))
	}
	if opts.ExternalReviewState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_review_state = %s", *opts.ExternalReviewState))
	}
	if opts.ExternalCheckState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_check_state = %s", *opts.ExternalCheckState))
	}

	if opts.OnlyWithoutDiffStats {
		preds = append(preds, sqlf.Sprintf("(changesets.diff_stat_added IS NULL OR changesets.diff_stat_changed IS NULL OR changesets.diff_stat_deleted IS NULL)"))
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
    external_check_state  = batch.external_check_state,
    created_by_campaign   = batch.created_by_campaign,
    added_to_campaign     = batch.added_to_campaign,
    diff_stat_added       = batch.diff_stat_added,
    diff_stat_changed     = batch.diff_stat_changed,
    diff_stat_deleted     = batch.diff_stat_deleted,
    sync_state            = batch.sync_state
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
  changed.external_check_state,
  changed.created_by_campaign,
  changed.added_to_campaign,
  changed.diff_stat_added,
  changed.diff_stat_changed,
  changed.diff_stat_deleted,
  changed.sync_state
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
		metadata, err := jsonbColumn(e.Metadata)
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
  patch_set_id,
  closed_at,
  campaign_spec_id
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
  patch_set_id,
  closed_at,
  campaign_spec_id
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
		nullInt64Column(c.PatchSetID),
		nullTimeColumn(c.ClosedAt),
		nullInt64Column(c.CampaignSpecID),
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
  patch_set_id,
  closed_at,
  campaign_spec_id
) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
  patch_set_id,
  closed_at,
  campaign_spec_id
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
		nullInt64Column(c.PatchSetID),
		nullTimeColumn(c.ClosedAt),
		nullInt64Column(c.CampaignSpecID),
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
	HasPatchSet *bool
	// Only return campaigns where author_id is the given.
	OnlyForAuthor int32
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

	if opts.HasPatchSet != nil {
		if *opts.HasPatchSet {
			preds = append(preds, sqlf.Sprintf("patch_set_id IS NOT NULL"))
		} else {
			preds = append(preds, sqlf.Sprintf("patch_set_id IS NULL"))
		}
	}

	if opts.OnlyForAuthor != 0 {
		preds = append(preds, sqlf.Sprintf("author_id = %d", opts.OnlyForAuthor))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countCampaignsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetCampaignOpts captures the query options needed for getting a Campaign
type GetCampaignOpts struct {
	ID         int64
	PatchSetID int64

	NamespaceUserID int32
	NamespaceOrgID  int32

	CampaignSpecID   int64
	CampaignSpecName string
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

var getCampaignsQueryFmtstrPre = `
-- source: enterprise/internal/campaigns/store.go:GetCampaign
SELECT
  campaigns.id,
  campaigns.name,
  campaigns.description,
  campaigns.branch,
  campaigns.author_id,
  campaigns.namespace_user_id,
  campaigns.namespace_org_id,
  campaigns.created_at,
  campaigns.updated_at,
  campaigns.changeset_ids,
  campaigns.patch_set_id,
  campaigns.closed_at,
  campaigns.campaign_spec_id
FROM campaigns
`

var getCampaignsQueryFmtstrPost = `
WHERE %s
LIMIT 1
`

func getCampaignQuery(opts *GetCampaignOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("campaigns.id = %s", opts.ID))
	}

	if opts.PatchSetID != 0 {
		preds = append(preds, sqlf.Sprintf("campaigns.patch_set_id = %s", opts.PatchSetID))
	}

	if opts.CampaignSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("campaigns.campaign_spec_id = %s", opts.CampaignSpecID))
	}

	if opts.NamespaceUserID != 0 {
		preds = append(preds, sqlf.Sprintf("campaigns.namespace_user_id = %s", opts.NamespaceUserID))
	}

	if opts.NamespaceOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("campaigns.namespace_org_id = %s", opts.NamespaceOrgID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	var joinClause string
	if opts.CampaignSpecName != "" {
		joinClause = "JOIN campaign_specs ON campaigns.campaign_spec_id = campaign_specs.id"
		cond := fmt.Sprintf(`campaign_specs.spec @> '{"name": %q}'`, opts.CampaignSpecName)
		preds = append(preds, sqlf.Sprintf(cond))

	}
	return sqlf.Sprintf(
		getCampaignsQueryFmtstrPre+joinClause+getCampaignsQueryFmtstrPost,
		sqlf.Join(preds, "\n AND "),
	)
}

// ListCampaignsOpts captures the query options needed for
// listing campaigns.
type ListCampaignsOpts struct {
	ChangesetID int64
	Cursor      int64
	Limit       int
	State       campaigns.CampaignState
	HasPatchSet *bool
	// Only return campaigns where author_id is the given.
	OnlyForAuthor int32
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
  patch_set_id,
  closed_at,
  campaign_spec_id
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

	if opts.HasPatchSet != nil {
		if *opts.HasPatchSet {
			preds = append(preds, sqlf.Sprintf("patch_set_id IS NOT NULL"))
		} else {
			preds = append(preds, sqlf.Sprintf("patch_set_id IS NULL"))
		}
	}

	if opts.OnlyForAuthor != 0 {
		preds = append(preds, sqlf.Sprintf("author_id = %d", opts.OnlyForAuthor))
	}

	return sqlf.Sprintf(
		listCampaignsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

// CreatePatchSet creates the given PatchSet.
func (s *Store) CreatePatchSet(ctx context.Context, c *campaigns.PatchSet) error {
	q, err := s.createPatchSetQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanPatchSet(c, sc)
		return c.ID, 1, err
	})
}

var createPatchSetQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CreatePatchSet
INSERT INTO patch_sets (
  created_at,
  updated_at,
  user_id
)
VALUES (%s, %s, %s)
RETURNING
  id,
  created_at,
  updated_at,
  user_id
`

func (s *Store) createPatchSetQuery(c *campaigns.PatchSet) (*sqlf.Query, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return sqlf.Sprintf(
		createPatchSetQueryFmtstr,
		c.CreatedAt,
		c.UpdatedAt,
		c.UserID,
	), nil
}

// UpdatePatchSet updates the given PatchSet.
func (s *Store) UpdatePatchSet(ctx context.Context, c *campaigns.PatchSet) error {
	q, err := s.updatePatchSetQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanPatchSet(c, sc)
		return c.ID, 1, err
	})
}

var updatePatchSetQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:UpdatePatchSet
UPDATE patch_sets
SET (
  updated_at,
  user_id
) = (%s, %s)
WHERE id = %s
RETURNING
  id,
  created_at,
  updated_at,
  user_id
`

func (s *Store) updatePatchSetQuery(c *campaigns.PatchSet) (*sqlf.Query, error) {
	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updatePatchSetQueryFmtstr,
		c.UpdatedAt,
		c.UserID,
		c.ID,
	), nil
}

// DeletePatchSet deletes the PatchSet with the given ID.
func (s *Store) DeletePatchSet(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deletePatchSetQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deletePatchSetQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:DeletePatchSet
DELETE FROM patch_sets WHERE id = %s
`

const PatchSetTTL = 7 * 24 * time.Hour

// DeleteExpiredPatchSets deletes PatchSets that have not been attached to a Campaign within PatchSetTTL.
func (s *Store) DeleteExpiredPatchSets(ctx context.Context) error {
	expirationTime := s.now().Add(-PatchSetTTL)
	q := sqlf.Sprintf(deleteExpiredPatchSetsQueryFmtstr, expirationTime)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteExpiredPatchSetsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:DeleteExpiredPatchSets
DELETE FROM
  patch_sets
WHERE
  created_at < %s
AND
NOT EXISTS (
  SELECT 1
  FROM
    campaigns
  WHERE
    campaigns.patch_set_id = patch_sets.id
  )
AND
NOT EXISTS (
  SELECT 1
  FROM
    changeset_jobs
  JOIN patches ON patches.id = changeset_jobs.patch_id
  JOIN changesets ON changesets.id = changeset_jobs.changeset_id
  WHERE
    (SELECT COUNT(*) FROM jsonb_object_keys(changesets.campaign_ids)) > 0
);
`

// CountPatchSets returns the number of code mods in the database.
func (s *Store) CountPatchSets(ctx context.Context) (count int64, _ error) {
	q := sqlf.Sprintf(countPatchSetsQueryFmtstr)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countPatchSetsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CountPatchSets
SELECT COUNT(id)
FROM patch_sets
`

// GetPatchSetOpts captures the query options needed for getting a PatchSet
type GetPatchSetOpts struct {
	ID int64
}

// GetPatchSet gets a code mod matching the given options.
func (s *Store) GetPatchSet(ctx context.Context, opts GetPatchSetOpts) (*campaigns.PatchSet, error) {
	q := getPatchSetQuery(&opts)

	var c campaigns.PatchSet
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanPatchSet(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getPatchSetsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:GetPatchSet
SELECT
  id,
  created_at,
  updated_at,
  user_id
FROM patch_sets
WHERE %s
LIMIT 1
`

func getPatchSetQuery(opts *GetPatchSetOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getPatchSetsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetCampaignStatusOpts captures the query options needed for getting the
// BackgroundProcessStatus for a Campaign.
type GetCampaignStatusOpts struct {
	ID int64
	// When ExcludeErrors is set the ProcessErrors slice of the
	// BackgroundProcessStatus returned by GetCampaignStatus won't be
	// populated.
	ExcludeErrors bool

	// ExcludeErrorsInRepos filters out error messages from ChangesetJobs that
	// are associated with Patches that have the given repository IDs set in
	// `patches.repo_id`.
	// This is used to filter out error messages from repositories the user
	// doesn't have access to.
	ExcludeErrorsInRepos []api.RepoID
}

// GetCampaignStatus gets the campaigns.BackgroundProcessStatus for a Campaign
func (s *Store) GetCampaignStatus(ctx context.Context, opts GetCampaignStatusOpts) (*campaigns.BackgroundProcessStatus, error) {
	q := getCampaignStatusQuery(&opts)
	return s.queryBackgroundProcessStatus(ctx, q)
}

func getCampaignStatusQuery(opts *GetCampaignStatusOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("campaign_id = %s", opts.ID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	var errorsPreds []*sqlf.Query
	if opts.ExcludeErrors {
		errorsPreds = append(errorsPreds, sqlf.Sprintf("FALSE"))
	}

	if len(opts.ExcludeErrorsInRepos) > 0 {
		ids := make([]*sqlf.Query, 0, len(opts.ExcludeErrorsInRepos))

		for _, repoID := range opts.ExcludeErrorsInRepos {
			ids = append(ids, sqlf.Sprintf("%s", repoID))
		}

		joined := sqlf.Join(ids, ",")

		errorsPreds = append(errorsPreds, sqlf.Sprintf("patches.repo_id NOT IN (%s)", joined))
		errorsPreds = append(errorsPreds, sqlf.Sprintf("error != ''"))
	}

	if len(errorsPreds) == 0 {
		errorsPreds = append(errorsPreds, sqlf.Sprintf("error != ''"))
	}

	return sqlf.Sprintf(
		getCampaignStatusQueryFmtstr,
		sqlf.Join(errorsPreds, " AND "),
		sqlf.Join(preds, "\n AND "),
	)
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
	case status.Completed == status.Total && status.Failed == 0:
		status.ProcessState = campaigns.BackgroundProcessStateCompleted
	case status.Completed == status.Total && status.Failed > 0:
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
  COUNT(*) FILTER (WHERE error != '') AS failed,
  array_agg(error) FILTER (WHERE %s) AS errors
FROM changeset_jobs
JOIN patches ON patches.id = changeset_jobs.patch_id
WHERE %s
LIMIT 1
`

// ListPatchSetsOpts captures the query options needed for
// listing code mods.
type ListPatchSetsOpts struct {
	Cursor int64
	Limit  int
}

// ListPatchSets lists PatchSets with the given filters.
func (s *Store) ListPatchSets(ctx context.Context, opts ListPatchSetsOpts) (cs []*campaigns.PatchSet, next int64, err error) {
	q := listPatchSetsQuery(&opts)

	cs = make([]*campaigns.PatchSet, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.PatchSet
		if err = scanPatchSet(&c, sc); err != nil {
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

var listPatchSetsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:ListPatchSets
SELECT
  id,
  created_at,
  updated_at,
  user_id
FROM patch_sets
WHERE %s
ORDER BY id ASC
LIMIT %s
`

func listPatchSetsQuery(opts *ListPatchSetsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	preds := []*sqlf.Query{
		sqlf.Sprintf("id >= %s", opts.Cursor),
	}

	return sqlf.Sprintf(
		listPatchSetsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

// CreatePatch creates the given Patch.
// Due to a unique constraint in the DB it is safe to call this more than once
// with the same input. Only one job will be added and the other calls will return an error
func (s *Store) CreatePatch(ctx context.Context, c *campaigns.Patch) error {
	q, err := s.createPatchQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanPatch(c, sc)
		return c.ID, 1, err
	})
}

var createPatchQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CreatePatch
INSERT INTO patches (
  patch_set_id,
  repo_id,
  rev,
  base_ref,
  diff,
  diff_stat_added,
  diff_stat_deleted,
  diff_stat_changed,
  created_at,
  updated_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
  id,
  patch_set_id,
  repo_id,
  rev,
  base_ref,
  diff,
  diff_stat_added,
  diff_stat_deleted,
  diff_stat_changed,
  created_at,
  updated_at
`

func (s *Store) createPatchQuery(c *campaigns.Patch) (*sqlf.Query, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return sqlf.Sprintf(
		createPatchQueryFmtstr,
		c.PatchSetID,
		c.RepoID,
		c.Rev,
		c.BaseRef,
		c.Diff,
		c.DiffStatAdded,
		c.DiffStatDeleted,
		c.DiffStatChanged,
		c.CreatedAt,
		c.UpdatedAt,
	), nil
}

// UpdatePatch updates the given Patch.
func (s *Store) UpdatePatch(ctx context.Context, c *campaigns.Patch) error {
	q, err := s.updatePatchQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanPatch(c, sc)
		return c.ID, 1, err
	})
}

var updatePatchQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:UpdatePatch
UPDATE patches
SET (
  patch_set_id,
  repo_id,
  rev,
  base_ref,
  diff,
  diff_stat_added,
  diff_stat_deleted,
  diff_stat_changed,
  updated_at
) = (%s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  id,
  patch_set_id,
  repo_id,
  rev,
  base_ref,
  diff,
  diff_stat_added,
  diff_stat_deleted,
  diff_stat_changed,
  created_at,
  updated_at
`

func (s *Store) updatePatchQuery(c *campaigns.Patch) (*sqlf.Query, error) {
	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updatePatchQueryFmtstr,
		c.PatchSetID,
		c.RepoID,
		c.Rev,
		c.BaseRef,
		c.Diff,
		c.DiffStatAdded,
		c.DiffStatDeleted,
		c.DiffStatChanged,
		c.UpdatedAt,
		c.ID,
	), nil
}

// DeletePatch deletes the Patch with the given ID.
func (s *Store) DeletePatch(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deletePatchQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deletePatchQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:DeletePatch
DELETE FROM patches WHERE id = %s
`

// CountPatchesOpts captures the query options needed for counting patches.
type CountPatchesOpts struct {
	PatchSetID   int64
	OnlyWithDiff bool

	// When set to a Campaign ID, only patches that do not have ChangesetJobs
	// associated with that Campaign are returned. The state of the
	// ChangesetJobs is not checked. This is mutually exclusive with
	// OnlyUnpublishedInCampaign.
	OnlyWithoutChangesetJob int64

	// If this is set to a Campaign ID only the Patches are returned that are
	// _not_ associated with a successfully completed ChangesetJob (meaning
	// that a Changeset on the codehost was created) for the given Campaign.
	OnlyUnpublishedInCampaign int64
}

// CountPatches returns the number of Patches in the database.
func (s *Store) CountPatches(ctx context.Context, opts CountPatchesOpts) (count int64, _ error) {
	q := countPatchesQuery(&opts)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countPatchesQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CountPatches
SELECT
	COUNT(patches.id)
FROM patches
INNER JOIN repo on repo.id = patches.repo_id
WHERE %s
`

func countPatchesQuery(opts *CountPatchesOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	if opts.PatchSetID != 0 {
		preds = append(preds, sqlf.Sprintf("patches.patch_set_id = %s", opts.PatchSetID))
	}

	if opts.OnlyWithDiff {
		preds = append(preds, sqlf.Sprintf("patches.diff != ''"))
	}

	if opts.OnlyWithoutChangesetJob != 0 {
		preds = append(preds, notInCampaignQuery(opts.OnlyWithoutChangesetJob))
	}

	if opts.OnlyUnpublishedInCampaign != 0 {
		preds = append(preds, onlyUnpublishedInCampaignQuery(opts.OnlyUnpublishedInCampaign))
	}

	return sqlf.Sprintf(countPatchesQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetPatchOpts captures the query options needed for getting a Patch
type GetPatchOpts struct {
	ID int64
}

// GetPatch gets a code mod matching the given options.
func (s *Store) GetPatch(ctx context.Context, opts GetPatchOpts) (*campaigns.Patch, error) {
	q := getPatchQuery(&opts)

	var c campaigns.Patch
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanPatch(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getPatchesQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:GetPatch
SELECT
  patches.id,
  patches.patch_set_id,
  patches.repo_id,
  patches.rev,
  patches.base_ref,
  patches.diff,
  patches.diff_stat_added,
  patches.diff_stat_deleted,
  patches.diff_stat_changed,
  patches.created_at,
  patches.updated_at
FROM patches
INNER JOIN repo ON repo.id = patches.repo_id
WHERE %s
LIMIT 1
`

func getPatchQuery(opts *GetPatchOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("patches.id = %s", opts.ID))
	}

	return sqlf.Sprintf(getPatchesQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListPatchesOpts captures the query options needed for
// listing code mods.
type ListPatchesOpts struct {
	PatchSetID   int64
	Cursor       int64
	Limit        int
	OnlyWithDiff bool

	// When set to a Campaign ID, only patches that do not have ChangesetJobs
	// associated with that Campaign are returned. The state of the
	// ChangesetJobs is not checked. This is mutually exclusive with
	// OnlyUnpublishedInCampaign.
	OnlyWithoutChangesetJob int64

	// If this is set to a Campaign ID only the Patches are returned that are
	// _not_ associated with a successfully completed ChangesetJob (meaning that
	// a Changeset on the codehost was created) for the given Campaign. This is
	// mutually exclusive with OnlyWithoutChangesetJob.
	OnlyUnpublishedInCampaign int64

	// If this is set only the Patches where diff_stat_added OR
	// diff_stat_changed OR diff_stat_deleted are NULL.
	OnlyWithoutDiffStats bool

	// If this is set, the patches.diff column is not loaded. The idea is to
	// speed up the query, since diffs can become quite large and require
	// memory allocations that can be unnecessary if only the other columns are
	// used.
	NoDiff bool
}

// ListPatches lists Patches with the given filters.
func (s *Store) ListPatches(ctx context.Context, opts ListPatchesOpts) (cs []*campaigns.Patch, next int64, err error) {
	q := listPatchesQuery(&opts)

	cs = make([]*campaigns.Patch, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.Patch
		if err = scanPatch(&c, sc); err != nil {
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

var listPatchesQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:ListPatches
SELECT
  patches.id,
  patches.patch_set_id,
  patches.repo_id,
  patches.rev,
  patches.base_ref,
  %s,
  patches.diff_stat_added,
  patches.diff_stat_deleted,
  patches.diff_stat_changed,
  patches.created_at,
  patches.updated_at
FROM patches
INNER JOIN repo ON repo.id = patches.repo_id
WHERE %s
ORDER BY id ASC
`

func listPatchesQuery(opts *ListPatchesOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var limitClause string
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}

	preds := []*sqlf.Query{
		sqlf.Sprintf("patches.id >= %s", opts.Cursor),
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.PatchSetID != 0 {
		preds = append(preds, sqlf.Sprintf("patches.patch_set_id = %s", opts.PatchSetID))
	}

	if opts.OnlyWithoutChangesetJob != 0 {
		preds = append(preds, notInCampaignQuery(opts.OnlyWithoutChangesetJob))
	}

	if opts.OnlyWithDiff {
		preds = append(preds, sqlf.Sprintf("patches.diff != ''"))
	}

	if opts.OnlyUnpublishedInCampaign != 0 {
		preds = append(preds, onlyUnpublishedInCampaignQuery(opts.OnlyUnpublishedInCampaign))
	}

	if opts.OnlyWithoutDiffStats {
		preds = append(preds, sqlf.Sprintf("(patches.diff_stat_added IS NULL OR patches.diff_stat_deleted IS NULL OR patches.diff_stat_changed IS NULL)"))
	}

	// To replace a field within a SELECT, we need to avoid extra escaping,
	// which we can do by ensuring it's a sqlf.Query already by using
	// sqlf.Sprintf, even though there's no actual formatting to be done.
	diffSrc := sqlf.Sprintf("patches.diff")
	if opts.NoDiff {
		diffSrc = sqlf.Sprintf("''")
	}

	return sqlf.Sprintf(
		listPatchesQueryFmtstr+limitClause,
		diffSrc,
		sqlf.Join(preds, "\n AND "),
	)
}

const onlyNotInCampaignQueryFmtstr = `
NOT EXISTS (
	SELECT 1
	FROM changeset_jobs
	WHERE
	  changeset_jobs.patch_id = patches.id
	AND
	  changeset_jobs.campaign_id = %s
)
`

func notInCampaignQuery(campaignID int64) *sqlf.Query {
	return sqlf.Sprintf(onlyNotInCampaignQueryFmtstr, campaignID)
}

var onlyUnpublishedInCampaignQueryFmtstr = `
NOT EXISTS (
  SELECT 1
  FROM changeset_jobs
  WHERE
    changeset_jobs.patch_id = patches.id
  AND
    changeset_jobs.campaign_id = %s
  AND
    changeset_jobs.changeset_id IS NOT NULL
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
  patch_id,
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
  patch_id,
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
		c.PatchID,
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
  patch_id,
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
  patch_id,
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
		c.PatchID,
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

// GetChangesetJobOpts captures the query options needed for getting a ChangesetJob
type GetChangesetJobOpts struct {
	ID          int64
	PatchID     int64
	CampaignID  int64
	ChangesetID int64
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
  patch_id,
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

	if opts.PatchID != 0 {
		preds = append(preds, sqlf.Sprintf("patch_id = %s", opts.PatchID))
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
	CampaignID int64
	PatchSetID int64
	Cursor     int64
	Limit      int
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
  changeset_jobs.patch_id,
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
	if opts.PatchSetID != 0 {
		joinClause = "JOIN campaigns ON changeset_jobs.campaign_id = campaigns.id"

		preds = append(preds, sqlf.Sprintf("campaigns.patch_set_id = %s", opts.PatchSetID))
	}

	queryTemplate := listChangesetJobsQueryFmtstrSelect + joinClause +
		listChangesetJobsQueryFmtstrConditions + limitClause

	return sqlf.Sprintf(queryTemplate, sqlf.Join(preds, "\n AND "))
}

type ResetChangesetJobsOpts struct {
	// The CampaignID of the ChangesetJobs to be reset.
	CampaignID int64

	// When PatchIDs is set, only the ChangesetJobs with the given
	// PatchIDs are reset.
	PatchIDs []int64

	// If OnlyFailed is set, only ChangesetJobs were Error != '' are reset.
	OnlyFailed bool
}

// ResetFailedChangesetJobs resets the Error, StartedAt and FinishedAt fields
// of the ChangesetJobs matching the conditions in ResetChangesetJobsOpts.
func (s *Store) ResetChangesetJobs(ctx context.Context, opts ResetChangesetJobsOpts) (err error) {
	if opts.CampaignID == 0 {
		return errors.New("CampaignID cannot be zero")
	}

	q := resetChangesetJobsQuery(opts)

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		return 0, 1, nil
	})
}

func resetChangesetJobsQuery(opts ResetChangesetJobsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("campaign_id = %s", opts.CampaignID),
	}

	if opts.OnlyFailed {
		preds = append(preds, sqlf.Sprintf("error != ''"))
	}

	if len(opts.PatchIDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(opts.PatchIDs))
		for _, id := range opts.PatchIDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds, sqlf.Sprintf("patch_id IN (%s)", sqlf.Join(ids, ",")))
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

// GetChangesetExternalIDs allows us to find the external ids for pull requests based on
// a slice of head refs. We need this in order to match incoming webhooks to pull requests as
// the only information they provide is the remote branch
func (s *Store) GetChangesetExternalIDs(ctx context.Context, spec api.ExternalRepoSpec, refs []string) ([]string, error) {
	queryFmtString := `
	SELECT cs.external_id FROM changesets cs
	JOIN repo r ON cs.repo_id = r.id
	WHERE cs.external_service_type = %s
	AND cs.external_branch IN (%s)
	AND r.external_id = %s
	AND r.external_service_type = %s
	AND r.external_service_id = %s
	AND r.deleted_at IS NULL
	ORDER BY cs.id ASC;
	`

	inClause := make([]*sqlf.Query, 0, len(refs))
	for _, ref := range refs {
		if ref == "" {
			continue
		}
		inClause = append(inClause, sqlf.Sprintf("%s", ref))
	}
	q := sqlf.Sprintf(queryFmtString, spec.ServiceType, sqlf.Join(inClause, ","), spec.ID, spec.ServiceType, spec.ServiceID)
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

// GetRepoIDsForFailedChangesetJobs returns the repository IDs of patches that
// are associated with failed changeset jobs belonging to the specified
// campaign.
// The repository IDs are used to get filtered by our authzFilter so we can
// then only load the error messages of the changeset jobs belonging to
// repositories that are NOT filtered out.
func (s *Store) GetRepoIDsForFailedChangesetJobs(ctx context.Context, campaign int64) ([]api.RepoID, error) {
	const queryFmtString = `
	SELECT patches.repo_id
	FROM changeset_jobs
	JOIN patches ON patches.id = changeset_jobs.patch_id
	WHERE
	  changeset_jobs.campaign_id = %s
	AND
	  changeset_jobs.error != ''
	AND
	  changeset_jobs.finished_at IS NOT NULL;
	`

	q := sqlf.Sprintf(queryFmtString, campaign)
	var ids []api.RepoID
	_, _, err := s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var id api.RepoID
		err = sc.Scan(&id)
		if err != nil {
			return 0, 0, err
		}
		ids = append(ids, id)
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
	var metadata, syncState json.RawMessage

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
		&t.CreatedByCampaign,
		&t.AddedToCampaign,
		&t.DiffStatAdded,
		&t.DiffStatChanged,
		&t.DiffStatDeleted,
		&syncState,
	)
	if err != nil {
		return errors.Wrap(err, "scanning changeset")
	}

	t.ExternalState = campaigns.ChangesetState(externalState)
	t.ExternalReviewState = campaigns.ChangesetReviewState(externamReviewState)
	t.ExternalCheckState = campaigns.ChangesetCheckState(externalCheckState)

	switch t.ExternalServiceType {
	case extsvc.TypeGitHub:
		t.Metadata = new(github.PullRequest)
	case extsvc.TypeBitbucketServer:
		t.Metadata = new(bitbucketserver.PullRequest)
	case extsvc.TypeGitLab:
		t.Metadata = new(gitlab.MergeRequest)
	default:
		return errors.New("unknown external service type")
	}

	if err = json.Unmarshal(metadata, t.Metadata); err != nil {
		return errors.Wrapf(err, "scanChangeset: failed to unmarshal %q metadata", t.ExternalServiceType)
	}
	if err = json.Unmarshal(syncState, &t.SyncState); err != nil {
		return errors.Wrapf(err, "scanChangeset: failed to unmarshal sync state: %s", syncState)
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
		&dbutil.NullString{S: &c.Description},
		&dbutil.NullString{S: &c.Branch},
		&c.AuthorID,
		&dbutil.NullInt32{N: &c.NamespaceUserID},
		&dbutil.NullInt32{N: &c.NamespaceOrgID},
		&c.CreatedAt,
		&c.UpdatedAt,
		&dbutil.JSONInt64Set{Set: &c.ChangesetIDs},
		&dbutil.NullInt64{N: &c.PatchSetID},
		&dbutil.NullTime{Time: &c.ClosedAt},
		&dbutil.NullInt64{N: &c.CampaignSpecID},
	)
}

func scanPatchSet(c *campaigns.PatchSet, s scanner) error {
	return s.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt, &c.UserID)
}

func scanPatch(c *campaigns.Patch, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.PatchSetID,
		&c.RepoID,
		&c.Rev,
		&c.BaseRef,
		&c.Diff,
		&c.DiffStatAdded,
		&c.DiffStatDeleted,
		&c.DiffStatChanged,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
}

func scanChangesetJob(c *campaigns.ChangesetJob, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.CampaignID,
		&c.PatchID,
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
		&b.Failed,
		pq.Array(&b.ProcessErrors),
	)
}

func jsonbColumn(metadata interface{}) (msg json.RawMessage, err error) {
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
