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
		ID                  int64                             `json:"id"`
		RepoID              api.RepoID                        `json:"repo_id"`
		CreatedAt           time.Time                         `json:"created_at"`
		UpdatedAt           time.Time                         `json:"updated_at"`
		Metadata            json.RawMessage                   `json:"metadata"`
		CampaignIDs         json.RawMessage                   `json:"campaign_ids"`
		ExternalID          string                            `json:"external_id"`
		ExternalServiceType string                            `json:"external_service_type"`
		ExternalBranch      string                            `json:"external_branch"`
		ExternalDeletedAt   *time.Time                        `json:"external_deleted_at"`
		ExternalUpdatedAt   *time.Time                        `json:"external_updated_at"`
		ExternalState       *campaigns.ChangesetExternalState `json:"external_state"`
		ExternalReviewState *campaigns.ChangesetReviewState   `json:"external_review_state"`
		ExternalCheckState  *campaigns.ChangesetCheckState    `json:"external_check_state"`
		CreatedByCampaign   bool                              `json:"created_by_campaign"`
		AddedToCampaign     bool                              `json:"added_to_campaign"`
		DiffStatAdded       *int32                            `json:"diff_stat_added"`
		DiffStatChanged     *int32                            `json:"diff_stat_changed"`
		DiffStatDeleted     *int32                            `json:"diff_stat_deleted"`
		SyncState           json.RawMessage                   `json:"sync_state"`
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
	ExternalState       *campaigns.ChangesetExternalState
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
	ExternalState        *campaigns.ChangesetExternalState
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
  closed_at,
  campaign_spec_id
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
  closed_at,
  campaign_spec_id
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
	ID int64

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

	if opts.OnlyForAuthor != 0 {
		preds = append(preds, sqlf.Sprintf("author_id = %d", opts.OnlyForAuthor))
	}

	return sqlf.Sprintf(
		listCampaignsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		opts.Limit,
	)
}

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
		externalReviewState string
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
		&dbutil.NullString{S: &externalReviewState},
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

	t.ExternalState = campaigns.ChangesetExternalState(externalState)
	t.ExternalReviewState = campaigns.ChangesetReviewState(externalReviewState)
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
		&dbutil.NullTime{Time: &c.ClosedAt},
		&dbutil.NullInt64{N: &c.CampaignSpecID},
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
