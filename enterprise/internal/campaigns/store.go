package campaigns

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/segmentio/fasthash/fnv1"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
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
	*basestore.Store
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
	handle := basestore.NewHandleWithDB(db)
	return &Store{Store: basestore.NewWithHandle(handle), now: clock}
}

// Clock returns the clock used by the Store.
func (s *Store) Clock() func() time.Time { return s.now }

// DB returns the underlying dbutil.DB that this Store was
// instantiated with.
// It's here for legacy reason to pass the dbutil.DB to a repos.Store while
// repos.Store doesn't accept a basestore.TransactableHandle yet.
func (s *Store) DB() dbutil.DB { return s.Handle().DB() }

var _ basestore.ShareableStore = &Store{}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *Store) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new Store with the given basestore.Shareable store as the
// underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{Store: s.Store.With(other), now: s.now}
}

// Transact creates a new transaction.
// It's required to implement this method and wrap the Transact method of the
// underlying basestore.Store.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &Store{Store: txBase, now: s.now}, nil
}

var NoTransactionError = errors.New("Not in a transaction")

var lockNamespace = int32(fnv1.HashString32("campaigns"))

// TryAcquireAdvisoryLock will attempt to acquire an advisory lock using key
// and is non blocking. If a lock is acquired, "true, nil" will be returned.
// It must be called from within a transaction or "false, NoTransactionError" is returned
func (s *Store) TryAcquireAdvisoryLock(ctx context.Context, key string) (bool, error) {
	if ok := s.Store.InTransaction(); !ok {
		return false, NoTransactionError
	}

	q := lockQuery(key)
	ok, _, err := basestore.ScanFirstBool(s.Store.Query(ctx, q))
	if err != nil || !ok {
		return false, err
	}
	return true, nil
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

// changesetColumns are used by by workerutil.Worker that loads changesets from
// the database for processing by the reconciler.
var changesetColumns = []*sqlf.Query{
	sqlf.Sprintf("changesets.id"),
	sqlf.Sprintf("changesets.repo_id"),
	sqlf.Sprintf("changesets.created_at"),
	sqlf.Sprintf("changesets.updated_at"),
	sqlf.Sprintf("changesets.metadata"),
	sqlf.Sprintf("changesets.campaign_ids"),
	sqlf.Sprintf("changesets.external_id"),
	sqlf.Sprintf("changesets.external_service_type"),
	sqlf.Sprintf("changesets.external_branch"),
	sqlf.Sprintf("changesets.external_deleted_at"),
	sqlf.Sprintf("changesets.external_updated_at"),
	sqlf.Sprintf("changesets.external_state"),
	sqlf.Sprintf("changesets.external_review_state"),
	sqlf.Sprintf("changesets.external_check_state"),
	sqlf.Sprintf("changesets.created_by_campaign"),
	sqlf.Sprintf("changesets.added_to_campaign"),
	sqlf.Sprintf("changesets.diff_stat_added"),
	sqlf.Sprintf("changesets.diff_stat_changed"),
	sqlf.Sprintf("changesets.diff_stat_deleted"),
	sqlf.Sprintf("changesets.sync_state"),
	sqlf.Sprintf("changesets.owned_by_campaign_id"),
	sqlf.Sprintf("changesets.current_spec_id"),
	sqlf.Sprintf("changesets.previous_spec_id"),
	sqlf.Sprintf("changesets.publication_state"),
	sqlf.Sprintf("changesets.reconciler_state"),
	sqlf.Sprintf("changesets.failure_message"),
	sqlf.Sprintf("changesets.started_at"),
	sqlf.Sprintf("changesets.finished_at"),
	sqlf.Sprintf("changesets.process_after"),
	sqlf.Sprintf("changesets.num_resets"),
}

// CreateChangesets creates the given Changesets in a loop.
// If inserting one changeset fails, the error is returned. If you want to
// ensure that all-or-none changesets are inserted, use a transaction.
func (s *Store) CreateChangesets(ctx context.Context, cs ...*campaigns.Changeset) error {
	for _, c := range cs {
		if err := s.CreateChangeset(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// CreateChangeset creates the given Changeset.
func (s *Store) CreateChangeset(ctx context.Context, c *campaigns.Changeset) error {
	q, err := s.createChangesetQuery(c)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc scanner) error { return scanChangeset(c, sc) })
}

var createChangesetQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:CreateChangeset
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
  sync_state,

  owned_by_campaign_id,
  current_spec_id,
  previous_spec_id,
  publication_state,
  reconciler_state,
  failure_message,
  started_at,
  finished_at,
  process_after,
  num_resets
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
changesets_repo_external_id_unique
DO NOTHING
RETURNING
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
  external_check_state,
  created_by_campaign,
  added_to_campaign,
  diff_stat_added,
  diff_stat_changed,
  diff_stat_deleted,
  sync_state,

  owned_by_campaign_id,
  current_spec_id,
  previous_spec_id,
  publication_state,
  reconciler_state,
  failure_message,
  started_at,
  finished_at,
  process_after,
  num_resets
`

func (s *Store) createChangesetQuery(c *campaigns.Changeset) (*sqlf.Query, error) {
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

	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return sqlf.Sprintf(
		createChangesetQueryFmtstr,
		c.RepoID,
		c.CreatedAt,
		c.UpdatedAt,
		metadata,
		campaignIDs,
		nullStringColumn(c.ExternalID),
		c.ExternalServiceType,
		nullStringColumn(c.ExternalBranch),
		nullTimeColumn(c.ExternalDeletedAt),
		nullTimeColumn(c.ExternalUpdatedAt),
		nullStringColumn(string(c.ExternalState)),
		nullStringColumn(string(c.ExternalReviewState)),
		nullStringColumn(string(c.ExternalCheckState)),
		c.CreatedByCampaign,
		c.AddedToCampaign,
		c.DiffStatAdded,
		c.DiffStatChanged,
		c.DiffStatDeleted,
		syncState,
		nullInt64Column(c.OwnedByCampaignID),
		nullInt64Column(c.CurrentSpecID),
		nullInt64Column(c.PreviousSpecID),
		c.PublicationState,
		c.ReconcilerState,
		c.FailureMessage,
		c.StartedAt,
		c.FinishedAt,
		c.ProcessAfter,
		c.NumResets,
	), nil
}

// DeleteChangeset deletes the Changeset with the given ID.
func (s *Store) DeleteChangeset(ctx context.Context, id int64) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteChangesetQueryFmtstr, id))
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
func (s *Store) CountChangesets(ctx context.Context, opts CountChangesetsOpts) (int, error) {
	return s.queryCount(ctx, countChangesetsQuery(&opts))
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
	err := s.query(ctx, q, func(sc scanner) error { return scanChangeset(&c, sc) })
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
  changesets.sync_state,

  changesets.owned_by_campaign_id,
  changesets.current_spec_id,
  changesets.previous_spec_id,
  changesets.publication_state,
  changesets.reconciler_state,
  changesets.failure_message,
  changesets.started_at,
  changesets.finished_at,
  changesets.process_after,
  changesets.num_resets
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
	err := s.query(ctx, q, func(sc scanner) (err error) {
		var h campaigns.ChangesetSyncData
		if err = scanChangesetSyncData(&h, sc); err != nil {
			return err
		}
		results = append(results, h)
		return err
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
		sqlf.Sprintf("changesets.publication_state = %s", campaigns.ChangesetPublicationStatePublished),
		sqlf.Sprintf("changesets.reconciler_state = %s", campaigns.ReconcilerStateCompleted),
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
	PublicationState     *campaigns.ChangesetPublicationState
	ReconcilerState      *campaigns.ReconcilerState
	ExternalState        *campaigns.ChangesetExternalState
	ExternalReviewState  *campaigns.ChangesetReviewState
	ExternalCheckState   *campaigns.ChangesetCheckState
	OnlyWithoutDiffStats bool
}

// ListChangesets lists Changesets with the given filters.
func (s *Store) ListChangesets(ctx context.Context, opts ListChangesetsOpts) (cs campaigns.Changesets, next int64, err error) {
	q := listChangesetsQuery(&opts)

	cs = make([]*campaigns.Changeset, 0, opts.Limit)
	err = s.query(ctx, q, func(sc scanner) (err error) {
		var c campaigns.Changeset
		if err = scanChangeset(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
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
  changesets.sync_state,

  changesets.owned_by_campaign_id,
  changesets.current_spec_id,
  changesets.previous_spec_id,
  changesets.publication_state,
  changesets.reconciler_state,
  changesets.failure_message,
  changesets.started_at,
  changesets.finished_at,
  changesets.process_after,
  changesets.num_resets
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

	if opts.PublicationState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.publication_state = %s", *opts.PublicationState))
	}
	if opts.ReconcilerState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.reconciler_state = %s", *opts.ReconcilerState))
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
	for _, c := range cs {
		if err := s.UpdateChangeset(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// UpdateChangeset updates the given Changeset.
func (s *Store) UpdateChangeset(ctx context.Context, cs *campaigns.Changeset) error {
	q, err := s.updateChangesetQuery(cs)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc scanner) (err error) {
		return scanChangeset(cs, sc)
	})
}

var updateChangesetQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_changeset_specs.go:UpdateChangeset
UPDATE changesets
SET (
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
  sync_state,

  owned_by_campaign_id,
  current_spec_id,
  previous_spec_id,
  publication_state,
  reconciler_state,
  failure_message,
  started_at,
  finished_at,
  process_after,
  num_resets
) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
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
  external_check_state,
  created_by_campaign,
  added_to_campaign,
  diff_stat_added,
  diff_stat_changed,
  diff_stat_deleted,
  sync_state,

  owned_by_campaign_id,
  current_spec_id,
  previous_spec_id,
  publication_state,
  reconciler_state,
  failure_message,
  started_at,
  finished_at,
  process_after,
  num_resets
`

func (s *Store) updateChangesetQuery(c *campaigns.Changeset) (*sqlf.Query, error) {
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

	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateChangesetQueryFmtstr,
		c.RepoID,
		c.CreatedAt,
		c.UpdatedAt,
		metadata,
		campaignIDs,
		nullStringColumn(c.ExternalID),
		c.ExternalServiceType,
		nullStringColumn(c.ExternalBranch),
		nullTimeColumn(c.ExternalDeletedAt),
		nullTimeColumn(c.ExternalUpdatedAt),
		nullStringColumn(string(c.ExternalState)),
		nullStringColumn(string(c.ExternalReviewState)),
		nullStringColumn(string(c.ExternalCheckState)),
		c.CreatedByCampaign,
		c.AddedToCampaign,
		c.DiffStatAdded,
		c.DiffStatChanged,
		c.DiffStatDeleted,
		syncState,
		nullInt64Column(c.OwnedByCampaignID),
		nullInt64Column(c.CurrentSpecID),
		nullInt64Column(c.PreviousSpecID),
		c.PublicationState,
		c.ReconcilerState,
		c.FailureMessage,
		c.StartedAt,
		c.FinishedAt,
		c.ProcessAfter,
		c.NumResets,
		// ID
		c.ID,
	), nil
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
	err := s.query(ctx, q, func(sc scanner) error {
		return scanChangesetEvent(&c, sc)
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
	err = s.query(ctx, q, func(sc scanner) (err error) {
		var c campaigns.ChangesetEvent
		if err = scanChangesetEvent(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
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
func (s *Store) CountChangesetEvents(ctx context.Context, opts CountChangesetEventsOpts) (int, error) {
	return s.queryCount(ctx, countChangesetEventsQuery(&opts))
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
	return s.query(ctx, q, func(sc scanner) (err error) {
		i++
		return scanChangesetEvent(cs[i], sc)
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

	return s.query(ctx, q, func(sc scanner) (err error) {
		return scanCampaign(c, sc)
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

	return s.query(ctx, q, func(sc scanner) (err error) { return scanCampaign(c, sc) })
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
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteCampaignQueryFmtstr, id))
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
func (s *Store) CountCampaigns(ctx context.Context, opts CountCampaignsOpts) (int, error) {
	return s.queryCount(ctx, countCampaignsQuery(&opts))
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

	CampaignSpecID int64
	Name           string
}

// GetCampaign gets a campaign matching the given options.
func (s *Store) GetCampaign(ctx context.Context, opts GetCampaignOpts) (*campaigns.Campaign, error) {
	q := getCampaignQuery(&opts)

	var c campaigns.Campaign
	err := s.query(ctx, q, func(sc scanner) error {
		return scanCampaign(&c, sc)
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

	if opts.Name != "" {
		preds = append(preds, sqlf.Sprintf("campaigns.name = %s", opts.Name))

	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getCampaignsQueryFmtstr,
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
	err = s.query(ctx, q, func(sc scanner) error {
		var c campaigns.Campaign
		if err := scanCampaign(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
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
	return basestore.ScanStrings(s.Store.Query(ctx, q))
}

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scanFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scanAll(rows, sc)
}

func (s *Store) queryCount(ctx context.Context, q *sqlf.Query) (int, error) {
	count, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	if err != nil || !ok {
		return count, err
	}
	return count, nil
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...interface{}) error
}

// a scanFunc scans one or more rows from a scanner, returning
// the last id column scanned and the count of scanned rows.
type scanFunc func(scanner) (err error)

func scanAll(rows *sql.Rows, scan scanFunc) (err error) {
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err = scan(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func scanFirstChangeset(rows *sql.Rows, err error) (*campaigns.Changeset, bool, error) {
	changesets, err := scanChangesets(rows, err)
	if err != nil || len(changesets) == 0 {
		return &campaigns.Changeset{}, false, err
	}
	return changesets[0], true, nil
}

func scanChangesets(rows *sql.Rows, queryErr error) ([]*campaigns.Changeset, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	var cs []*campaigns.Changeset

	return cs, scanAll(rows, func(sc scanner) (err error) {
		var c campaigns.Changeset
		if err = scanChangeset(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})
}

func scanChangeset(t *campaigns.Changeset, s scanner) error {
	var metadata, syncState json.RawMessage

	var (
		externalState       string
		externalReviewState string
		externalCheckState  string
		failureMessage      string
	)
	err := s.Scan(
		&t.ID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&metadata,
		&dbutil.JSONInt64Set{Set: &t.CampaignIDs},
		&dbutil.NullString{S: &t.ExternalID},
		&t.ExternalServiceType,
		&dbutil.NullString{S: &t.ExternalBranch},
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
		&dbutil.NullInt64{N: &t.OwnedByCampaignID},
		&dbutil.NullInt64{N: &t.CurrentSpecID},
		&dbutil.NullInt64{N: &t.PreviousSpecID},
		&t.PublicationState,
		&t.ReconcilerState,
		&dbutil.NullString{S: &failureMessage},
		&t.StartedAt,
		&t.FinishedAt,
		&t.ProcessAfter,
		&t.NumResets,
	)
	if err != nil {
		return errors.Wrap(err, "scanning changeset")
	}

	t.ExternalState = campaigns.ChangesetExternalState(externalState)
	t.ExternalReviewState = campaigns.ChangesetReviewState(externalReviewState)
	t.ExternalCheckState = campaigns.ChangesetCheckState(externalCheckState)
	if failureMessage != "" {
		t.FailureMessage = &failureMessage
	}

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
