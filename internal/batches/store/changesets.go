package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	adobatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/azuredevops"
	gerritbatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/perforce"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches/search"
	bbcs "github.com/sourcegraph/sourcegraph/internal/batches/sources/bitbucketcloud"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var changesetStringColumns = SQLColumns{
	"id",
	"repo_id",
	"created_at",
	"updated_at",
	"metadata",
	"batch_change_ids",
	"external_id",
	"external_service_type",
	"external_branch",
	"external_fork_name",
	"external_fork_namespace",
	"external_deleted_at",
	"external_updated_at",
	"external_state",
	"external_review_state",
	"external_check_state",
	"commit_verification",
	"diff_stat_added",
	"diff_stat_deleted",
	"sync_state",
	"owned_by_batch_change_id",
	"current_spec_id",
	"previous_spec_id",
	"publication_state",
	"ui_publication_state",
	"reconciler_state",
	// computed_state is calculated by a Postgres function called changesets_computed_state_ensure. The value is
	// determined by the combination of reconciler_state, publication_state, and external_state.
	"computed_state",
	"failure_message",
	"started_at",
	"finished_at",
	"process_after",
	"num_resets",
	"num_failures",
	"closing",
	"syncer_error",
	"detached_at",
	"previous_failure_message",
}

// ChangesetColumns are used by the changeset related Store methods and by
// workerutil.Worker to load changesets from the database for processing by
// the reconciler.
var ChangesetColumns = []*sqlf.Query{
	sqlf.Sprintf("changesets.id"),
	sqlf.Sprintf("changesets.repo_id"),
	sqlf.Sprintf("changesets.created_at"),
	sqlf.Sprintf("changesets.updated_at"),
	sqlf.Sprintf("changesets.metadata"),
	sqlf.Sprintf("changesets.batch_change_ids"),
	sqlf.Sprintf("changesets.external_id"),
	sqlf.Sprintf("changesets.external_service_type"),
	sqlf.Sprintf("changesets.external_branch"),
	sqlf.Sprintf("changesets.external_fork_name"),
	sqlf.Sprintf("changesets.external_fork_namespace"),
	sqlf.Sprintf("changesets.external_deleted_at"),
	sqlf.Sprintf("changesets.external_updated_at"),
	sqlf.Sprintf("changesets.external_state"),
	sqlf.Sprintf("changesets.external_review_state"),
	sqlf.Sprintf("changesets.external_check_state"),
	sqlf.Sprintf("changesets.commit_verification"),
	sqlf.Sprintf("changesets.diff_stat_added"),
	sqlf.Sprintf("changesets.diff_stat_deleted"),
	sqlf.Sprintf("changesets.sync_state"),
	sqlf.Sprintf("changesets.owned_by_batch_change_id"),
	sqlf.Sprintf("changesets.current_spec_id"),
	sqlf.Sprintf("changesets.previous_spec_id"),
	sqlf.Sprintf("changesets.publication_state"),
	sqlf.Sprintf("changesets.ui_publication_state"),
	sqlf.Sprintf("changesets.reconciler_state"),
	// computed_state is calculated by a Postgres function called changesets_computed_state_ensure. The value is
	// determined by the combination of reconciler_state, publication_state, and external_state.
	sqlf.Sprintf("changesets.computed_state"),
	sqlf.Sprintf("changesets.failure_message"),
	sqlf.Sprintf("changesets.started_at"),
	sqlf.Sprintf("changesets.finished_at"),
	sqlf.Sprintf("changesets.process_after"),
	sqlf.Sprintf("changesets.num_resets"),
	sqlf.Sprintf("changesets.num_failures"),
	sqlf.Sprintf("changesets.closing"),
	sqlf.Sprintf("changesets.syncer_error"),
	sqlf.Sprintf("changesets.detached_at"),
	sqlf.Sprintf("changesets.previous_failure_message"),
}

// changesetInsertColumns is the list of changeset columns that are modified in
// Store.UpdateChangeset.
var changesetInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("metadata"),
	sqlf.Sprintf("batch_change_ids"),
	sqlf.Sprintf("detached_at"),
	sqlf.Sprintf("external_id"),
	sqlf.Sprintf("external_service_type"),
	sqlf.Sprintf("external_branch"),
	sqlf.Sprintf("external_fork_name"),
	sqlf.Sprintf("external_fork_namespace"),
	sqlf.Sprintf("external_deleted_at"),
	sqlf.Sprintf("external_updated_at"),
	sqlf.Sprintf("external_state"),
	sqlf.Sprintf("external_review_state"),
	sqlf.Sprintf("external_check_state"),
	sqlf.Sprintf("commit_verification"),
	sqlf.Sprintf("diff_stat_added"),
	sqlf.Sprintf("diff_stat_deleted"),
	sqlf.Sprintf("sync_state"),
	sqlf.Sprintf("owned_by_batch_change_id"),
	sqlf.Sprintf("current_spec_id"),
	sqlf.Sprintf("previous_spec_id"),
	sqlf.Sprintf("publication_state"),
	sqlf.Sprintf("ui_publication_state"),
	sqlf.Sprintf("reconciler_state"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("closing"),
	sqlf.Sprintf("syncer_error"),
	// We additionally store the result of changeset.Title() in a column, so
	// the business logic for determining it is in one place and the field is
	// indexable for searching.
	sqlf.Sprintf("external_title"),
	sqlf.Sprintf("previous_failure_message"),
}

// changesetCodeHostStateInsertColumns are the columns that Store.UpdateChangesetCodeHostState uses to update a changeset
// with state change on a code host.
var changesetCodeHostStateInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("metadata"),
	sqlf.Sprintf("external_branch"),
	sqlf.Sprintf("external_fork_name"),
	sqlf.Sprintf("external_fork_namespace"),
	sqlf.Sprintf("external_deleted_at"),
	sqlf.Sprintf("external_updated_at"),
	sqlf.Sprintf("external_state"),
	sqlf.Sprintf("external_review_state"),
	sqlf.Sprintf("external_check_state"),
	sqlf.Sprintf("diff_stat_added"),
	sqlf.Sprintf("diff_stat_deleted"),
	sqlf.Sprintf("sync_state"),
	sqlf.Sprintf("syncer_error"),
	// We additionally store the result of changeset.Title() in a column, so
	// the business logic for determining it is in one place and the field is
	// indexable for searching.
	sqlf.Sprintf("external_title"),
}

// changesetInsertStringColumns is the list of column names that are used by Store.CreateChangesets for insertion.
var changesetInsertStringColumns = []string{
	"repo_id",
	"created_at",
	"updated_at",
	"metadata",
	"batch_change_ids",
	"detached_at",
	"external_id",
	"external_service_type",
	"external_branch",
	"external_fork_name",
	"external_fork_namespace",
	"external_deleted_at",
	"external_updated_at",
	"external_state",
	"external_review_state",
	"external_check_state",
	"commit_verification",
	"diff_stat_added",
	"diff_stat_deleted",
	"sync_state",
	"owned_by_batch_change_id",
	"current_spec_id",
	"previous_spec_id",
	"publication_state",
	"ui_publication_state",
	"reconciler_state",
	"failure_message",
	"started_at",
	"finished_at",
	"process_after",
	"num_resets",
	"num_failures",
	"closing",
	"syncer_error",
	"external_title",
	"previous_failure_message",
}

// temporaryChangesetInsertColumns is the list of column names used by Store.UpdateChangesetsForApply to insert into
// a temporary table.
var temporaryChangesetInsertColumns = []string{
	"id",
	"batch_change_ids",
	"detached_at",
	"diff_stat_added",
	"diff_stat_deleted",
	"current_spec_id",
	"previous_spec_id",
	"ui_publication_state",
	"reconciler_state",
	"failure_message",
	"num_resets",
	"num_failures",
	"closing",
	"syncer_error",
}

// CreateChangeset creates the given Changesets.
func (s *Store) CreateChangeset(ctx context.Context, cs ...*btypes.Changeset) (err error) {
	ctx, _, endObservation := s.operations.createChangeset.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("count", len(cs)),
	}})
	defer endObservation(1, observation.Args{})

	inserter := func(inserter *batch.Inserter) error {
		for _, c := range cs {
			if c.CreatedAt.IsZero() {
				c.CreatedAt = s.now()
			}

			if c.UpdatedAt.IsZero() {
				c.UpdatedAt = c.CreatedAt
			}

			metadata, err := jsonbColumn(c.Metadata)
			if err != nil {
				return err
			}

			batchChanges, err := batchChangesColumn(c)
			if err != nil {
				return err
			}

			syncState, err := json.Marshal(c.SyncState)
			if err != nil {
				return err
			}

			var cv json.RawMessage
			// Don't bother to record the result of verification if it's not even verified.
			if c.CommitVerification != nil && c.CommitVerification.Verified {
				cv, err = jsonbColumn(c.CommitVerification)
			} else {
				cv, err = jsonbColumn(nil)
			}
			if err != nil {
				return err
			}

			// Not being able to find a title is fine, we just have a NULL in the database then.
			title, _ := c.Title()

			uiPublicationState := uiPublicationStateColumn(c)

			if err := inserter.Insert(
				ctx,
				c.RepoID,
				c.CreatedAt,
				c.UpdatedAt,
				metadata,
				batchChanges,
				dbutil.NullTimeColumn(c.DetachedAt),
				dbutil.NullStringColumn(c.ExternalID),
				c.ExternalServiceType,
				dbutil.NullStringColumn(c.ExternalBranch),
				dbutil.NullStringColumn(c.ExternalForkName),
				dbutil.NullStringColumn(c.ExternalForkNamespace),
				dbutil.NullTimeColumn(c.ExternalDeletedAt),
				dbutil.NullTimeColumn(c.ExternalUpdatedAt),
				dbutil.NullStringColumn(string(c.ExternalState)),
				dbutil.NullStringColumn(string(c.ExternalReviewState)),
				dbutil.NullStringColumn(string(c.ExternalCheckState)),
				cv,
				c.DiffStatAdded,
				c.DiffStatDeleted,
				syncState,
				dbutil.NullInt64Column(c.OwnedByBatchChangeID),
				dbutil.NullInt64Column(c.CurrentSpecID),
				dbutil.NullInt64Column(c.PreviousSpecID),
				c.PublicationState,
				uiPublicationState,
				c.ReconcilerState.ToDB(),
				c.FailureMessage,
				dbutil.NullTimeColumn(c.StartedAt),
				dbutil.NullTimeColumn(c.FinishedAt),
				dbutil.NullTimeColumn(c.ProcessAfter),
				c.NumResets,
				c.NumFailures,
				c.Closing,
				c.SyncErrorMessage,
				dbutil.NullStringColumn(title),
				c.PreviousFailureMessage,
			); err != nil {
				return err
			}
		}
		return nil
	}

	i := -1
	return batch.WithInserterWithReturn(
		ctx,
		s.Handle(),
		"changesets",
		batch.MaxNumPostgresParameters,
		changesetInsertStringColumns,
		"",
		changesetStringColumns,
		func(rows dbutil.Scanner) error {
			i++
			return ScanChangeset(cs[i], rows)
		},
		inserter,
	)
}

// DeleteChangeset deletes the Changeset with the given ID.
func (s *Store) DeleteChangeset(ctx context.Context, id int64) (err error) {
	ctx, _, endObservation := s.operations.deleteChangeset.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteChangesetQueryFmtstr, id))
}

var deleteChangesetQueryFmtstr = `
DELETE FROM changesets WHERE id = %s
`

// CountChangesetsOpts captures the query options needed for
// counting changesets.
type CountChangesetsOpts struct {
	BatchChangeID        int64
	OnlyArchived         bool
	IncludeArchived      bool
	ExternalStates       []btypes.ChangesetExternalState
	ExternalReviewState  *btypes.ChangesetReviewState
	ExternalCheckState   *btypes.ChangesetCheckState
	ReconcilerStates     []btypes.ReconcilerState
	OwnedByBatchChangeID int64
	PublicationState     *btypes.ChangesetPublicationState
	TextSearch           []search.TextSearchTerm
	EnforceAuthz         bool
	RepoIDs              []api.RepoID
	States               []btypes.ChangesetState
}

// CountChangesets returns the number of changesets in the database.
func (s *Store) CountChangesets(ctx context.Context, opts CountChangesetsOpts) (count int, err error) {
	ctx, _, endObservation := s.operations.countChangesets.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s))
	if err != nil {
		return 0, errors.Wrap(err, "CountChangesets generating authz query conds")
	}
	return s.queryCount(ctx, countChangesetsQuery(&opts, authzConds))
}

var countChangesetsQueryFmtstr = `
SELECT COUNT(changesets.id)
FROM changesets
INNER JOIN repo ON repo.id = changesets.repo_id
%s -- optional LEFT JOIN to changeset_specs if required
WHERE %s
`

func countChangesetsQuery(opts *CountChangesetsOpts, authzConds *sqlf.Query) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	if opts.BatchChangeID != 0 {
		batchChangeID := strconv.Itoa(int(opts.BatchChangeID))
		preds = append(preds, sqlf.Sprintf("changesets.batch_change_ids ? %s", batchChangeID))
		if opts.OnlyArchived {
			preds = append(preds, archivedInBatchChange(batchChangeID))
		} else if !opts.IncludeArchived {
			preds = append(preds, sqlf.Sprintf("NOT (%s)", archivedInBatchChange(batchChangeID)))
		}
	}
	if opts.PublicationState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.publication_state = %s", *opts.PublicationState))
	}
	if len(opts.ExternalStates) > 0 {
		preds = append(preds, sqlf.Sprintf("changesets.external_state = ANY (%s)", pq.Array(opts.ExternalStates)))
	}
	if len(opts.States) > 0 {
		preds = append(preds, sqlf.Sprintf("changesets.computed_state = ANY (%s)", pq.Array(opts.States)))
	}
	if opts.ExternalReviewState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_review_state = %s", *opts.ExternalReviewState))
	}
	if opts.ExternalCheckState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_check_state = %s", *opts.ExternalCheckState))
	}
	if len(opts.ReconcilerStates) != 0 {
		// TODO: Would be nice if we could use this with pq.Array.
		states := make([]*sqlf.Query, len(opts.ReconcilerStates))
		for i, reconcilerState := range opts.ReconcilerStates {
			states[i] = sqlf.Sprintf("%s", reconcilerState.ToDB())
		}
		preds = append(preds, sqlf.Sprintf("changesets.reconciler_state IN (%s)", sqlf.Join(states, ",")))
	}
	if opts.OwnedByBatchChangeID != 0 {
		preds = append(preds, sqlf.Sprintf("changesets.owned_by_batch_change_id = %s", opts.OwnedByBatchChangeID))
	}
	if opts.EnforceAuthz {
		preds = append(preds, authzConds)
	}
	if len(opts.RepoIDs) > 0 {
		preds = append(preds, sqlf.Sprintf("repo.id = ANY (%s)", pq.Array(opts.RepoIDs)))
	}

	join := sqlf.Sprintf("")
	if len(opts.TextSearch) != 0 {
		// TextSearch predicates require changeset_specs to be joined into the
		// query as well.
		join = sqlf.Sprintf("LEFT JOIN changeset_specs ON changesets.current_spec_id = changeset_specs.id")

		for _, term := range opts.TextSearch {
			preds = append(preds, textSearchTermToClause(
				term,
				// The COALESCE() is required to handle the actual title on the
				// changeset, if it has been published or if it's tracked.
				sqlf.Sprintf("COALESCE(changesets.external_title, changeset_specs.title)"),
				sqlf.Sprintf("repo.name"),
			))
		}
	}

	return sqlf.Sprintf(countChangesetsQueryFmtstr, join, sqlf.Join(preds, "\n AND "))
}

// GetChangesetByID is a convenience method if only the ID needs to be passed in. It's also used for abstraction in
// the testing package.
func (s *Store) GetChangesetByID(ctx context.Context, id int64) (*btypes.Changeset, error) {
	return s.GetChangeset(ctx, GetChangesetOpts{ID: id})
}

// GetChangesetOpts captures the query options needed for getting a Changeset
type GetChangesetOpts struct {
	ID                  int64
	RepoID              api.RepoID
	ExternalID          string
	ExternalServiceType string
	ExternalBranch      string
	ReconcilerState     btypes.ReconcilerState
	PublicationState    btypes.ChangesetPublicationState
}

// GetChangeset gets a changeset matching the given options.
func (s *Store) GetChangeset(ctx context.Context, opts GetChangesetOpts) (ch *btypes.Changeset, err error) {
	ctx, _, endObservation := s.operations.getChangeset.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getChangesetQuery(&opts)

	var c btypes.Changeset
	err = s.query(ctx, q, func(sc dbutil.Scanner) error { return ScanChangeset(&c, sc) })
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}
	return &c, nil
}

var getChangesetsQueryFmtstr = `
SELECT %s FROM changesets
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
	if opts.ExternalBranch != "" {
		preds = append(preds, sqlf.Sprintf("changesets.external_branch = %s", opts.ExternalBranch))
	}
	if opts.ReconcilerState != "" {
		preds = append(preds, sqlf.Sprintf("changesets.reconciler_state = %s", opts.ReconcilerState.ToDB()))
	}
	if opts.PublicationState != "" {
		preds = append(preds, sqlf.Sprintf("changesets.publication_state = %s", opts.PublicationState))
	}

	return sqlf.Sprintf(
		getChangesetsQueryFmtstr,
		sqlf.Join(ChangesetColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

type ListChangesetSyncDataOpts struct {
	// Return only the supplied changesets. If empty, all changesets are returned
	ChangesetIDs []int64

	ExternalServiceID string
}

// ListChangesetSyncData returns sync data on all non-externally-deleted changesets
// that are part of at least one open batch change.
func (s *Store) ListChangesetSyncData(ctx context.Context, opts ListChangesetSyncDataOpts) (sd []*btypes.ChangesetSyncData, err error) {
	ctx, _, endObservation := s.operations.listChangesetSyncData.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listChangesetSyncDataQuery(opts)
	results := make([]*btypes.ChangesetSyncData, 0)
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		var h btypes.ChangesetSyncData
		if err := ScanChangesetSyncData(&h, sc); err != nil {
			return err
		}
		results = append(results, &h)
		return err
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func ScanChangesetSyncData(h *btypes.ChangesetSyncData, s dbutil.Scanner) error {
	return s.Scan(
		&h.ChangesetID,
		&h.UpdatedAt,
		&dbutil.NullTime{Time: &h.LatestEvent},
		&dbutil.NullTime{Time: &h.ExternalUpdatedAt},
		&h.RepoExternalServiceID,
	)
}

const listChangesetSyncDataQueryFmtstr = `
SELECT changesets.id,
	changesets.updated_at,
	max(ce.updated_at) AS latest_event,
	changesets.external_updated_at,
	r.external_service_id
FROM changesets
LEFT JOIN changeset_events ce ON changesets.id = ce.changeset_id
JOIN batch_changes ON changesets.batch_change_ids ? batch_changes.id::TEXT
JOIN repo r ON changesets.repo_id = r.id
WHERE %s
GROUP BY changesets.id, r.id
ORDER BY changesets.id ASC
`

func listChangesetSyncDataQuery(opts ListChangesetSyncDataOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("batch_changes.closed_at IS NULL"),
		sqlf.Sprintf("r.deleted_at IS NULL"),
		sqlf.Sprintf("changesets.publication_state = %s", btypes.ChangesetPublicationStatePublished),
		sqlf.Sprintf("changesets.reconciler_state = %s", btypes.ReconcilerStateCompleted.ToDB()),
	}
	if len(opts.ChangesetIDs) > 0 {
		preds = append(preds, sqlf.Sprintf("changesets.id = ANY (%s)", pq.Array(opts.ChangesetIDs)))
	}

	if opts.ExternalServiceID != "" {
		preds = append(preds, sqlf.Sprintf("r.external_service_id = %s", opts.ExternalServiceID))
	}

	return sqlf.Sprintf(listChangesetSyncDataQueryFmtstr, sqlf.Join(preds, "\n AND"))
}

// ListChangesetsOpts captures the query options needed for listing changesets.
//
// Note that TextSearch is potentially expensive, and should only be specified
// in conjunction with at least one other option (most likely, BatchChangeID).
type ListChangesetsOpts struct {
	LimitOpts
	Cursor               int64
	BatchChangeID        int64
	OnlyArchived         bool
	IncludeArchived      bool
	IDs                  []int64
	States               []btypes.ChangesetState
	PublicationState     *btypes.ChangesetPublicationState
	ReconcilerStates     []btypes.ReconcilerState
	ExternalStates       []btypes.ChangesetExternalState
	ExternalReviewState  *btypes.ChangesetReviewState
	ExternalCheckState   *btypes.ChangesetCheckState
	OwnedByBatchChangeID int64
	TextSearch           []search.TextSearchTerm
	EnforceAuthz         bool
	RepoIDs              []api.RepoID
	BitbucketCloudCommit string
}

// ListChangesets lists Changesets with the given filters.
func (s *Store) ListChangesets(ctx context.Context, opts ListChangesetsOpts) (cs btypes.Changesets, next int64, err error) {
	ctx, _, endObservation := s.operations.listChangesets.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s))
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListChangesets generating authz query conds")
	}
	q := listChangesetsQuery(&opts, authzConds)

	cs = make([]*btypes.Changeset, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		var c btypes.Changeset
		if err = ScanChangeset(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	if opts.Limit != 0 && len(cs) == opts.DBLimit() {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listChangesetsQueryFmtstr = `
SELECT %s FROM changesets
INNER JOIN repo ON repo.id = changesets.repo_id
%s -- optional LEFT JOIN to changeset_specs if required
WHERE %s
ORDER BY id ASC
`

func listChangesetsQuery(opts *ListChangesetsOpts, authzConds *sqlf.Query) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("changesets.id >= %s", opts.Cursor),
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.BatchChangeID != 0 {
		batchChangeID := strconv.Itoa(int(opts.BatchChangeID))
		preds = append(preds, sqlf.Sprintf("changesets.batch_change_ids ? %s", batchChangeID))

		if opts.OnlyArchived {
			preds = append(preds, archivedInBatchChange(batchChangeID))
		} else if !opts.IncludeArchived {
			preds = append(preds, sqlf.Sprintf("NOT (%s)", archivedInBatchChange(batchChangeID)))
		}
	}

	if len(opts.IDs) > 0 {
		preds = append(preds, sqlf.Sprintf("changesets.id = ANY (%s)", pq.Array(opts.IDs)))
	}

	if opts.PublicationState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.publication_state = %s", *opts.PublicationState))
	}
	if len(opts.ReconcilerStates) != 0 {
		states := make([]*sqlf.Query, len(opts.ReconcilerStates))
		for i, reconcilerState := range opts.ReconcilerStates {
			states[i] = sqlf.Sprintf("%s", reconcilerState.ToDB())
		}
		preds = append(preds, sqlf.Sprintf("changesets.reconciler_state IN (%s)", sqlf.Join(states, ",")))
	}
	if len(opts.States) != 0 {
		preds = append(preds, sqlf.Sprintf("changesets.computed_state = ANY(%s)", pq.Array(opts.States)))
	}
	if len(opts.ExternalStates) > 0 {
		preds = append(preds, sqlf.Sprintf("changesets.external_state = ANY (%s)", pq.Array(opts.ExternalStates)))
	}
	if opts.ExternalReviewState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_review_state = %s", *opts.ExternalReviewState))
	}
	if opts.ExternalCheckState != nil {
		preds = append(preds, sqlf.Sprintf("changesets.external_check_state = %s", *opts.ExternalCheckState))
	}
	if opts.OwnedByBatchChangeID != 0 {
		preds = append(preds, sqlf.Sprintf("changesets.owned_by_batch_change_id = %s", opts.OwnedByBatchChangeID))
	}
	if opts.EnforceAuthz {
		preds = append(preds, authzConds)
	}
	if len(opts.RepoIDs) > 0 {
		preds = append(preds, sqlf.Sprintf("repo.id = ANY (%s)", pq.Array(opts.RepoIDs)))
	}
	if len(opts.BitbucketCloudCommit) >= 12 {
		// Bitbucket Cloud commit hashes in PR objects are generally truncated
		// to 12 characters, but this isn't actually documented in the API
		// documentation: they may be anything from 7 up. In practice, we've
		// only observed 12. Given that, we'll look for 7, 12, and the full hash
		// — since this hits an index, this should be relatively cheap.
		preds = append(preds, sqlf.Sprintf(
			"changesets.metadata->'source'->'commit'->>'hash' IN (%s, %s, %s)",
			opts.BitbucketCloudCommit[0:7],
			opts.BitbucketCloudCommit[0:12],
			opts.BitbucketCloudCommit,
		))
	}

	join := sqlf.Sprintf("")
	if len(opts.TextSearch) != 0 {
		// TextSearch predicates require changeset_specs to be joined into the
		// query as well.
		join = sqlf.Sprintf("LEFT JOIN changeset_specs ON changesets.current_spec_id = changeset_specs.id")

		for _, term := range opts.TextSearch {
			preds = append(preds, textSearchTermToClause(
				term,
				// The COALESCE() is required to handle the actual title on the
				// changeset, if it has been published or if it's tracked.
				sqlf.Sprintf("COALESCE(changesets.external_title, changeset_specs.title)"),
				sqlf.Sprintf("repo.name"),
			))
		}
	}

	return sqlf.Sprintf(
		listChangesetsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(ChangesetColumns, ", "),
		join,
		sqlf.Join(preds, "\n AND "),
	)
}

// EnqueueChangeset enqueues the given changeset by resetting all
// worker-related columns and setting its reconciler_state column to the
// `resetState` argument but *only if* the `currentState` matches its current
// `reconciler_state`.
func (s *Store) EnqueueChangeset(ctx context.Context, cs *btypes.Changeset, resetState, currentState btypes.ReconcilerState) (err error) {
	ctx, _, endObservation := s.operations.enqueueChangeset.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(cs.ID)),
	}})
	defer endObservation(1, observation.Args{})

	_, ok, err := basestore.ScanFirstInt(s.Store.Query(
		ctx,
		s.enqueueChangesetQuery(cs, resetState, currentState),
	))
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("cannot re-enqueue changeset not in failed state")
	}

	return nil
}

var enqueueChangesetQueryFmtstr = `
UPDATE changesets
SET
	reconciler_state = %s,
	num_resets = 0,
	num_failures = 0,
	-- Copy over and reset the previous failure message
	previous_failure_message = changesets.failure_message,
	failure_message = NULL,
	syncer_error = NULL,
	updated_at = %s
WHERE
	%s
RETURNING
	changesets.id
`

func (s *Store) enqueueChangesetQuery(cs *btypes.Changeset, resetState, currentState btypes.ReconcilerState) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("id = %s", cs.ID),
	}

	if currentState != "" {
		preds = append(preds, sqlf.Sprintf("reconciler_state = %s", currentState.ToDB()))
	}

	return sqlf.Sprintf(
		enqueueChangesetQueryFmtstr,
		resetState.ToDB(),
		s.now(),
		sqlf.Join(preds, "AND"),
	)
}

// UpdateChangeset updates the given Changeset.
func (s *Store) UpdateChangeset(ctx context.Context, cs *btypes.Changeset) (err error) {
	ctx, _, endObservation := s.operations.updateChangeset.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(cs.ID)),
	}})
	defer endObservation(1, observation.Args{})

	cs.UpdatedAt = s.now()

	q, err := s.changesetWriteQuery(updateChangesetQueryFmtstr, true, cs)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return ScanChangeset(cs, sc)
	})
}

func (s *Store) changesetWriteQuery(q string, includeID bool, c *btypes.Changeset) (*sqlf.Query, error) {
	metadata, err := jsonbColumn(c.Metadata)
	if err != nil {
		return nil, err
	}

	batchChanges, err := batchChangesColumn(c)
	if err != nil {
		return nil, err
	}

	syncState, err := json.Marshal(c.SyncState)
	if err != nil {
		return nil, err
	}

	var cv json.RawMessage
	// Don't bother to record the result of verification if it's not even verified.
	if c.CommitVerification != nil && c.CommitVerification.Verified {
		cv, err = jsonbColumn(c.CommitVerification)
	} else {
		cv, err = jsonbColumn(nil)
	}
	if err != nil {
		return nil, err
	}

	// Not being able to find a title is fine, we just have a NULL in the database then.
	title, _ := c.Title()

	uiPublicationState := uiPublicationStateColumn(c)

	vars := []any{
		sqlf.Join(changesetInsertColumns, ", "),
		c.RepoID,
		c.CreatedAt,
		c.UpdatedAt,
		metadata,
		batchChanges,
		dbutil.NullTimeColumn(c.DetachedAt),
		dbutil.NullStringColumn(c.ExternalID),
		c.ExternalServiceType,
		dbutil.NullStringColumn(c.ExternalBranch),
		dbutil.NullStringColumn(c.ExternalForkName),
		dbutil.NullStringColumn(c.ExternalForkNamespace),
		dbutil.NullTimeColumn(c.ExternalDeletedAt),
		dbutil.NullTimeColumn(c.ExternalUpdatedAt),
		dbutil.NullStringColumn(string(c.ExternalState)),
		dbutil.NullStringColumn(string(c.ExternalReviewState)),
		dbutil.NullStringColumn(string(c.ExternalCheckState)),
		cv,
		c.DiffStatAdded,
		c.DiffStatDeleted,
		syncState,
		dbutil.NullInt64Column(c.OwnedByBatchChangeID),
		dbutil.NullInt64Column(c.CurrentSpecID),
		dbutil.NullInt64Column(c.PreviousSpecID),
		c.PublicationState,
		uiPublicationState,
		c.ReconcilerState.ToDB(),
		c.FailureMessage,
		dbutil.NullTimeColumn(c.StartedAt),
		dbutil.NullTimeColumn(c.FinishedAt),
		dbutil.NullTimeColumn(c.ProcessAfter),
		c.NumResets,
		c.NumFailures,
		c.Closing,
		c.SyncErrorMessage,
		dbutil.NullStringColumn(title),
		c.PreviousFailureMessage,
	}

	if includeID {
		vars = append(vars, c.ID)
	}

	vars = append(vars, sqlf.Join(ChangesetColumns, ", "))

	return sqlf.Sprintf(q, vars...), nil
}

var updateChangesetQueryFmtstr = `
UPDATE changesets
SET (%s) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  %s
`

// UpdateChangesetsForApply updates the provided Changesets.
//
// To efficiently insert a batch of updates to the changesets table, we fist insert the provided changesets to a temorary
// table. The temporary table's columns are only the fields that are updated when applying changesets for a batch change
// (for efficiency reasons).
//
// Once the changesets are in the temporary table, the values are then used to update their "previous" value in the actual
// changesets table.
func (s *Store) UpdateChangesetsForApply(ctx context.Context, cs []*btypes.Changeset) (err error) {
	ctx, _, endObservation := s.operations.updateChangeset.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("count", len(cs)),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create the temporary table
	if err = tx.Exec(ctx, sqlf.Sprintf(updateChangesetsTemporaryTableQuery)); err != nil {
		return err
	}

	inserter := func(inserter *batch.Inserter) error {
		for _, c := range cs {
			batchChanges, _ := batchChangesColumn(c)
			if err != nil {
				return err
			}

			uiPublicationState := uiPublicationStateColumn(c)

			if err := inserter.Insert(
				ctx,
				c.ID,
				batchChanges,
				dbutil.NullTimeColumn(c.DetachedAt),
				c.DiffStatAdded,
				c.DiffStatDeleted,
				dbutil.NullInt64Column(c.CurrentSpecID),
				dbutil.NullInt64Column(c.PreviousSpecID),
				uiPublicationState,
				c.ReconcilerState.ToDB(),
				c.FailureMessage,
				c.NumResets,
				c.NumFailures,
				c.Closing,
				c.SyncErrorMessage,
			); err != nil {
				return err
			}
		}
		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"temp_changesets",
		batch.MaxNumPostgresParameters,
		temporaryChangesetInsertColumns,
		inserter,
	); err != nil {
		return err
	}

	// Insert the values from the temporary table into the target table.
	return tx.Exec(ctx, sqlf.Sprintf(updateChangesetsInsertQuery))
}

const updateChangesetsTemporaryTableQuery = `
CREATE TEMPORARY TABLE temp_changesets (
    id bigint primary key,
    batch_change_ids jsonb DEFAULT '{}'::jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT NOW() NOT NULL,
    detached_at timestamp with time zone,
    diff_stat_added integer,
    diff_stat_deleted integer,
    current_spec_id bigint,
    previous_spec_id bigint,
    ui_publication_state batch_changes_changeset_ui_publication_state,
    reconciler_state text DEFAULT 'queued'::text,
    failure_message text,
	previous_failure_message text,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    closing boolean DEFAULT false NOT NULL,
    syncer_error text
) ON COMMIT DROP
`

const updateChangesetsInsertQuery = `
UPDATE changesets c SET batch_change_ids = source.batch_change_ids, updated_at = source.updated_at,
                        detached_at = source.detached_at, diff_stat_added = source.diff_stat_added,
                        diff_stat_deleted = source.diff_stat_deleted, current_spec_id = source.current_spec_id,
                        previous_spec_id = source.previous_spec_id, ui_publication_state = source.ui_publication_state,
                        reconciler_state = source.reconciler_state, failure_message = source.failure_message,
						previous_failure_message = source.previous_failure_message,
                        num_resets = source.num_resets, num_failures = source.num_failures, closing = source.closing,
                        syncer_error = source.syncer_error
FROM temp_changesets source
WHERE c.id = source.id
`

// UpdateChangesetBatchChanges updates only the `batch_changes` & `updated_at`
// columns of the given Changeset.
func (s *Store) UpdateChangesetBatchChanges(ctx context.Context, cs *btypes.Changeset) (err error) {
	ctx, _, endObservation := s.operations.updateChangesetBatchChanges.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(cs.ID)),
	}})
	defer endObservation(1, observation.Args{})

	batchChanges, err := batchChangesColumn(cs)
	if err != nil {
		return err
	}

	return s.updateChangesetColumn(ctx, cs, "batch_change_ids", batchChanges)
}

// UpdateChangesetUiPublicationState updates only the `ui_publication_state` &
// `updated_at` columns of the given Changeset.
func (s *Store) UpdateChangesetUiPublicationState(ctx context.Context, cs *btypes.Changeset) (err error) {
	ctx, _, endObservation := s.operations.updateChangesetUIPublicationState.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(cs.ID)),
	}})
	defer endObservation(1, observation.Args{})

	uiPublicationState := uiPublicationStateColumn(cs)
	return s.updateChangesetColumn(ctx, cs, "ui_publication_state", uiPublicationState)
}

// UpdateChangesetSCommitVerification records the commit verification object for a commit
// to the Changeset if it was signed and verified.
func (s *Store) UpdateChangesetCommitVerification(ctx context.Context, cs *btypes.Changeset, commit *github.RestCommit) (err error) {
	ctx, _, endObservation := s.operations.updateChangesetCommitVerification.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(cs.ID)),
	}})
	defer endObservation(1, observation.Args{})

	var cv json.RawMessage
	// Don't bother to record the result of verification if it's not even verified.
	if commit.Verification.Verified {
		cv, err = jsonbColumn(commit.Verification)
	} else {
		cv, err = jsonbColumn(nil)
	}
	if err != nil {
		return err
	}

	return s.updateChangesetColumn(ctx, cs, "commit_verification", cv)
}

// updateChangesetColumn updates the column with the given name, setting it to
// the given value, and updating the updated_at column.
func (s *Store) updateChangesetColumn(ctx context.Context, cs *btypes.Changeset, name string, val any) error {
	cs.UpdatedAt = s.now()

	vars := []any{
		sqlf.Sprintf(name),
		cs.UpdatedAt,
		val,
		cs.ID,
		sqlf.Join(ChangesetColumns, ", "),
	}

	q := sqlf.Sprintf(updateChangesetColumnQueryFmtstr, vars...)

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return ScanChangeset(cs, sc)
	})
}

var updateChangesetColumnQueryFmtstr = `
UPDATE changesets
SET (updated_at, %s) = (%s, %s)
WHERE id = %s
RETURNING
  %s
`

// UpdateChangesetCodeHostState updates only the columns of the given Changeset
// that relate to the state of the changeset on the code host, e.g.
// external_branch, external_state, etc.
func (s *Store) UpdateChangesetCodeHostState(ctx context.Context, cs *btypes.Changeset) (err error) {
	ctx, _, endObservation := s.operations.updateChangesetCodeHostState.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(cs.ID)),
	}})
	defer endObservation(1, observation.Args{})

	cs.UpdatedAt = s.now()

	q, err := updateChangesetCodeHostStateQuery(cs)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return ScanChangeset(cs, sc)
	})
}

func updateChangesetCodeHostStateQuery(c *btypes.Changeset) (*sqlf.Query, error) {
	metadata, err := jsonbColumn(c.Metadata)
	if err != nil {
		return nil, err
	}

	syncState, err := json.Marshal(c.SyncState)
	if err != nil {
		return nil, err
	}

	// Not being able to find a title is fine, we just have a NULL in the database then.
	title, _ := c.Title()

	vars := []any{
		sqlf.Join(changesetCodeHostStateInsertColumns, ", "),
		c.UpdatedAt,
		metadata,
		dbutil.NullStringColumn(c.ExternalBranch),
		dbutil.NullStringColumn(c.ExternalForkName),
		dbutil.NullStringColumn(c.ExternalForkNamespace),
		dbutil.NullTimeColumn(c.ExternalDeletedAt),
		dbutil.NullTimeColumn(c.ExternalUpdatedAt),
		dbutil.NullStringColumn(string(c.ExternalState)),
		dbutil.NullStringColumn(string(c.ExternalReviewState)),
		dbutil.NullStringColumn(string(c.ExternalCheckState)),
		c.DiffStatAdded,
		c.DiffStatDeleted,
		syncState,
		c.SyncErrorMessage,
		dbutil.NullStringColumn(title),
		c.ID,
		sqlf.Join(ChangesetColumns, ", "),
	}

	return sqlf.Sprintf(updateChangesetCodeHostStateQueryFmtstr, vars...), nil
}

var updateChangesetCodeHostStateQueryFmtstr = `
UPDATE changesets
SET (%s) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  %s
`

// GetChangesetExternalIDs allows us to find the external ids for pull requests based on
// a slice of head refs. We need this in order to match incoming webhooks to pull requests as
// the only information they provide is the remote branch
func (s *Store) GetChangesetExternalIDs(ctx context.Context, spec api.ExternalRepoSpec, refs []string) (externalIDs []string, err error) {
	ctx, _, endObservation := s.operations.getChangesetExternalIDs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

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

// CanceledChangesetFailureMessage is set on changesets as the FailureMessage
// by CancelQueuedBatchChangeChangesets which is called at the beginning of
// ApplyBatchChange to stop enqueued changesets being processed while we're
// applying the new batch spec.
var CanceledChangesetFailureMessage = "Canceled"

// CancelQueuedBatchChangeChangesets cancels all scheduled, queued, or errored
// changesets that are owned by the given batch change. It blocks until all
// currently processing changesets have finished executing.
func (s *Store) CancelQueuedBatchChangeChangesets(ctx context.Context, batchChangeID int64) (err error) {
	var iterations int
	ctx, _, endObservation := s.operations.cancelQueuedBatchChangeChangesets.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("batchChangeID", int(batchChangeID)),
	}})
	defer endObservation(1, observation.Args{Attrs: []attribute.KeyValue{attribute.Int("iterations", iterations)}})

	// Just for safety, so we don't end up with stray cancel requests bombarding
	// the DB with 10 requests a second forever:
	ctx, cancel := context.WithDeadline(ctx, s.now().Add(2*time.Minute))
	defer cancel()

	for {
		// Note that we don't cancel queued "syncing" changesets, since their
		// owned_by_batch_change_id is not set. That's on purpose. It's okay if they're
		// being processed after this, since they only pull data and not create
		// changesets on the code hosts.
		q := sqlf.Sprintf(
			cancelQueuedBatchChangeChangesetsFmtstr,
			batchChangeID,
			btypes.ReconcilerStateScheduled.ToDB(),
			btypes.ReconcilerStateQueued.ToDB(),
			btypes.ReconcilerStateErrored.ToDB(),
			btypes.ReconcilerStateFailed.ToDB(),
			CanceledChangesetFailureMessage,
			batchChangeID,
			btypes.ReconcilerStateProcessing.ToDB(),
		)

		processing, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
		if err != nil {
			return errors.Wrap(err, "canceling queued batch change changesets failed")
		}
		if !ok || processing == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
		iterations++
	}
	return nil
}

const cancelQueuedBatchChangeChangesetsFmtstr = `
WITH changeset_ids AS (
  SELECT id FROM changesets
  WHERE
    owned_by_batch_change_id = %s
  AND
    reconciler_state IN (%s, %s, %s)
),
updated_records AS (
	UPDATE
	  changesets
	SET
	  reconciler_state = %s,
	  failure_message = %s
	WHERE id IN (SELECT id FROM changeset_ids)
)
SELECT
	COUNT(id) AS remaining_processing
FROM changesets
WHERE
	owned_by_batch_change_id = %d
	AND
	reconciler_state = %s
`

// EnqueueChangesetsToClose updates all changesets that are owned by the given
// batch change to set their reconciler status to 'queued' and the Closing boolean
// to true.
//
// It does not update the changesets that are fully processed and already
// closed/merged.
//
// This will loop until there are no processing rows anymore, or until 2 minutes
// passed.
func (s *Store) EnqueueChangesetsToClose(ctx context.Context, batchChangeID int64) (err error) {
	var iterations int
	ctx, _, endObservation := s.operations.enqueueChangesetsToClose.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("batchChangeID", int(batchChangeID)),
	}})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{attribute.Int("iterations", iterations)}})
	}()

	// Just for safety, so we don't end up with stray cancel requests bombarding
	// the DB with 10 requests a second forever:
	ctx, cancel := context.WithDeadline(ctx, s.now().Add(2*time.Minute))
	defer cancel()

	for {
		q := sqlf.Sprintf(
			enqueueChangesetsToCloseFmtstr,
			batchChangeID,
			btypes.ChangesetPublicationStatePublished,
			btypes.ReconcilerStateCompleted.ToDB(),
			btypes.ChangesetExternalStateClosed,
			btypes.ChangesetExternalStateMerged,
			btypes.ReconcilerStateQueued.ToDB(),
			btypes.ReconcilerStateProcessing.ToDB(),
			btypes.ReconcilerStateProcessing.ToDB(),
		)
		processing, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
		if err != nil {
			return err
		}
		if !ok || processing == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
		iterations++
	}
	return nil
}

const enqueueChangesetsToCloseFmtstr = `
WITH all_matching AS (
	SELECT
		id, reconciler_state
	FROM
		changesets
	WHERE
		owned_by_batch_change_id = %d
		AND
		publication_state = %s
		AND
		NOT (
			reconciler_state = %s
			AND
			(external_state = %s OR external_state = %s)
		)
),
updated_records AS (
	UPDATE
		changesets
	SET
		reconciler_state = %s,
		failure_message = NULL,
		num_resets = 0,
		num_failures = 0,
		closing = TRUE
	WHERE
		changesets.id IN (SELECT id FROM all_matching WHERE NOT all_matching.reconciler_state = %s)
)
SELECT COUNT(id) FROM all_matching WHERE all_matching.reconciler_state = %s
`

// jsonBatchChangeChangesetSet represents a "join table" set as a JSONB object
// where the keys are the ids and the values are json objects holding the properties.
// It implements the sql.Scanner interface so it can be used as a scan destination,
// similar to sql.NullString.
type jsonBatchChangeChangesetSet struct {
	Assocs *[]btypes.BatchChangeAssoc
}

// Scan implements the Scanner interface.
func (n *jsonBatchChangeChangesetSet) Scan(value any) error {
	m := make(map[int64]btypes.BatchChangeAssoc)

	switch value := value.(type) {
	case nil:
	case []byte:
		if err := json.Unmarshal(value, &m); err != nil {
			return err
		}
	default:
		return errors.Errorf("value is not []byte: %T", value)
	}

	if *n.Assocs == nil {
		*n.Assocs = make([]btypes.BatchChangeAssoc, 0, len(m))
	} else {
		*n.Assocs = (*n.Assocs)[:0]
	}

	for id, assoc := range m {
		assoc.BatchChangeID = id
		*n.Assocs = append(*n.Assocs, assoc)
	}

	sort.Slice(*n.Assocs, func(i, j int) bool {
		return (*n.Assocs)[i].BatchChangeID < (*n.Assocs)[j].BatchChangeID
	})

	return nil
}

// Value implements the driver Valuer interface.
func (n jsonBatchChangeChangesetSet) Value() (driver.Value, error) {
	if n.Assocs == nil {
		return nil, nil
	}
	return *n.Assocs, nil
}

func ScanChangeset(t *btypes.Changeset, s dbutil.Scanner) error {
	var metadata, syncState, commitVerification json.RawMessage

	var (
		externalState          string
		externalReviewState    string
		externalCheckState     string
		failureMessage         string
		syncErrorMessage       string
		reconcilerState        string
		previousFailureMessage string
	)
	err := s.Scan(
		&t.ID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&metadata,
		&jsonBatchChangeChangesetSet{Assocs: &t.BatchChanges},
		&dbutil.NullString{S: &t.ExternalID},
		&t.ExternalServiceType,
		&dbutil.NullString{S: &t.ExternalBranch},
		&dbutil.NullString{S: &t.ExternalForkName},
		&dbutil.NullString{S: &t.ExternalForkNamespace},
		&dbutil.NullTime{Time: &t.ExternalDeletedAt},
		&dbutil.NullTime{Time: &t.ExternalUpdatedAt},
		&dbutil.NullString{S: &externalState},
		&dbutil.NullString{S: &externalReviewState},
		&dbutil.NullString{S: &externalCheckState},
		&commitVerification,
		&t.DiffStatAdded,
		&t.DiffStatDeleted,
		&syncState,
		&dbutil.NullInt64{N: &t.OwnedByBatchChangeID},
		&dbutil.NullInt64{N: &t.CurrentSpecID},
		&dbutil.NullInt64{N: &t.PreviousSpecID},
		&t.PublicationState,
		&t.UiPublicationState,
		&reconcilerState,
		&t.State,
		&dbutil.NullString{S: &failureMessage},
		&dbutil.NullTime{Time: &t.StartedAt},
		&dbutil.NullTime{Time: &t.FinishedAt},
		&dbutil.NullTime{Time: &t.ProcessAfter},
		&t.NumResets,
		&t.NumFailures,
		&t.Closing,
		&dbutil.NullString{S: &syncErrorMessage},
		&dbutil.NullTime{Time: &t.DetachedAt},
		&dbutil.NullString{S: &previousFailureMessage},
	)
	if err != nil {
		return errors.Wrap(err, "scanning changeset")
	}

	t.ExternalState = btypes.ChangesetExternalState(externalState)
	t.ExternalReviewState = btypes.ChangesetReviewState(externalReviewState)
	t.ExternalCheckState = btypes.ChangesetCheckState(externalCheckState)
	if failureMessage != "" {
		t.FailureMessage = &failureMessage
	}
	if previousFailureMessage != "" {
		t.PreviousFailureMessage = &previousFailureMessage
	}
	if syncErrorMessage != "" {
		t.SyncErrorMessage = &syncErrorMessage
	}
	t.ReconcilerState = btypes.ReconcilerState(strings.ToUpper(reconcilerState))

	switch t.ExternalServiceType {
	case extsvc.TypeGitHub:
		t.Metadata = new(github.PullRequest)
	case extsvc.TypeBitbucketServer:
		t.Metadata = new(bitbucketserver.PullRequest)
	case extsvc.TypeGitLab:
		t.Metadata = new(gitlab.MergeRequest)
	case extsvc.TypeBitbucketCloud:
		m := new(bbcs.AnnotatedPullRequest)
		// Ensure the inner PR is initialized, it should never be nil.
		m.PullRequest = &bitbucketcloud.PullRequest{}
		t.Metadata = m
	case extsvc.TypeAzureDevOps:
		m := new(adobatches.AnnotatedPullRequest)
		// Ensure the inner PR is initialized, it should never be nil.
		m.PullRequest = &azuredevops.PullRequest{}
		t.Metadata = m
	case extsvc.TypeGerrit:
		m := new(gerritbatches.AnnotatedChange)
		m.Change = &gerrit.Change{}
		t.Metadata = m
	case extsvc.TypePerforce:
		t.Metadata = new(perforce.Changelist)
	case extsvc.TypeGerrit:
		t.Metadata = new(gerrit.Change)
	default:
		return errors.New("unknown external service type")
	}

	if err = json.Unmarshal(metadata, t.Metadata); err != nil {
		return errors.Wrapf(err, "scanChangeset: failed to unmarshal %q metadata", t.ExternalServiceType)
	}
	if err = json.Unmarshal(syncState, &t.SyncState); err != nil {
		return errors.Wrapf(err, "scanChangeset: failed to unmarshal sync state: %s", syncState)
	}
	var cv *github.Verification
	if err = json.Unmarshal(commitVerification, &cv); err != nil {
		return errors.Wrapf(err, "scanChangesetSpecs: failed to unmarshal commitVerification: %s", commitVerification)
	}
	// Only set the commit verification if it's actually verified.
	if cv.Verified {
		t.CommitVerification = cv
	}

	return nil
}

// GetChangesetsStats returns statistics on all the changesets associated to the given batch change,
// or all changesets across the instance.
func (s *Store) GetChangesetsStats(ctx context.Context, batchChangeID int64) (stats btypes.ChangesetsStats, err error) {
	ctx, _, endObservation := s.operations.getChangesetsStats.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("batchChangeID", int(batchChangeID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getChangesetsStatsQuery(batchChangeID)
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		if err := sc.Scan(
			&stats.Total,
			&stats.Retrying,
			&stats.Failed,
			&stats.Scheduled,
			&stats.Processing,
			&stats.Unpublished,
			&stats.Closed,
			&stats.Draft,
			&stats.Merged,
			&stats.Open,
			&stats.Deleted,
			&stats.Archived,
		); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return stats, err
	}
	return stats, nil
}

const getChangesetStatsFmtstr = `
SELECT
	COUNT(*) AS total,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'RETRYING') AS retrying,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'FAILED') AS failed,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'SCHEDULED') AS scheduled,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'PROCESSING') AS processing,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'UNPUBLISHED') AS unpublished,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'CLOSED') AS closed,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'DRAFT') AS draft,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'MERGED') AS merged,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'OPEN') AS open,
	COUNT(*) FILTER (WHERE NOT %s AND changesets.computed_state = 'DELETED') AS deleted,
	COUNT(*) FILTER (WHERE %s) AS archived
FROM changesets
INNER JOIN repo on repo.id = changesets.repo_id
WHERE
	%s
`

// GetRepoChangesetsStats returns statistics on all the changesets associated to the given repo.
func (s *Store) GetRepoChangesetsStats(ctx context.Context, repoID api.RepoID) (stats *btypes.RepoChangesetsStats, err error) {
	ctx, _, endObservation := s.operations.getRepoChangesetsStats.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repoID", int(repoID)),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s))
	if err != nil {
		return nil, errors.Wrap(err, "GetRepoChangesetsStats generating authz query conds")
	}
	q := getRepoChangesetsStatsQuery(int64(repoID), authzConds)
	stats = &btypes.RepoChangesetsStats{}
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		if err := sc.Scan(
			&stats.Total,
			&stats.Unpublished,
			&stats.Draft,
			&stats.Closed,
			&stats.Merged,
			&stats.Open,
		); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return stats, err
	}
	return stats, nil
}

func (s *Store) GetGlobalChangesetsStats(ctx context.Context) (stats *btypes.GlobalChangesetsStats, err error) {
	ctx, _, endObservation := s.operations.getGlobalChangesetsStats.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(getGlobalChangesetsStatsFmtstr)
	stats = &btypes.GlobalChangesetsStats{}
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		if err := sc.Scan(
			&stats.Total,
			&stats.Unpublished,
			&stats.Draft,
			&stats.Closed,
			&stats.Merged,
			&stats.Open,
		); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return stats, err
	}
	return stats, nil
}

func (s *Store) EnqueueNextScheduledChangeset(ctx context.Context) (ch *btypes.Changeset, err error) {
	ctx, _, endObservation := s.operations.enqueueNextScheduledChangeset.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		enqueueNextScheduledChangesetFmtstr,
		btypes.ReconcilerStateScheduled.ToDB(),
		btypes.ReconcilerStateQueued.ToDB(),
		sqlf.Join(ChangesetColumns, ","),
	)

	var c btypes.Changeset
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		return ScanChangeset(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

const enqueueNextScheduledChangesetFmtstr = `
WITH c AS (
	SELECT *
	FROM changesets
	WHERE reconciler_state = %s
	ORDER BY updated_at ASC
	LIMIT 1
)
UPDATE changesets
SET reconciler_state = %s
FROM c
WHERE c.id = changesets.id
RETURNING %s
`

func (s *Store) GetChangesetPlaceInSchedulerQueue(ctx context.Context, id int64) (place int, err error) {
	ctx, _, endObservation := s.operations.getChangesetPlaceInSchedulerQueue.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		getChangesetPlaceInSchedulerQueueFmtstr,
		btypes.ReconcilerStateScheduled.ToDB(),
		id,
	)

	row := s.QueryRow(ctx, q)
	if err := row.Scan(&place); err == sql.ErrNoRows {
		return 0, ErrNoResults
	} else if err != nil {
		return 0, err
	}

	// PostgreSQL returns 1-indexed row numbers, but we want 0-indexed places
	// when calculating schedules.
	return place - 1, nil
}

const getChangesetPlaceInSchedulerQueueFmtstr = `
SELECT
	row_number
FROM (
	SELECT
		id,
		ROW_NUMBER() OVER (ORDER BY updated_at ASC) AS row_number
	FROM
		changesets
	WHERE
		reconciler_state = %s
	) t
WHERE
	id = %d
`

func archivedInBatchChange(batchChangeID string) *sqlf.Query {
	return sqlf.Sprintf(
		"(COALESCE((batch_change_ids->%s->>'isArchived')::bool, false) OR COALESCE((batch_change_ids->%s->>'archive')::bool, false))",
		batchChangeID,
		batchChangeID,
	)
}

func getChangesetsStatsQuery(batchChangeID int64) *sqlf.Query {
	batchChangeIDStr := strconv.Itoa(int(batchChangeID))

	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changesets.batch_change_ids ? %s", batchChangeIDStr),
	}

	archived := archivedInBatchChange(batchChangeIDStr)

	return sqlf.Sprintf(
		getChangesetStatsFmtstr,
		archived, archived,
		archived, archived,
		archived, archived,
		archived, archived,
		archived, archived,
		archived,
		sqlf.Join(preds, " AND "),
	)
}

func getRepoChangesetsStatsQuery(repoID int64, authzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		getRepoChangesetsStatsFmtstr,
		strconv.Itoa(int(repoID)),
		authzConds,
	)
}

const getRepoChangesetsStatsFmtstr = `
SELECT
	COUNT(*) AS total,
	COUNT(*) FILTER (WHERE computed_state = 'UNPUBLISHED') AS unpublished,
	COUNT(*) FILTER (WHERE computed_state = 'DRAFT') AS draft,
	COUNT(*) FILTER (WHERE computed_state = 'CLOSED') AS closed,
	COUNT(*) FILTER (WHERE computed_state = 'MERGED') AS merged,
	COUNT(*) FILTER (WHERE computed_state = 'OPEN') AS open
FROM (
	SELECT
		changesets.id,
		changesets.computed_state
	FROM
		changesets
		INNER JOIN repo ON changesets.repo_id = repo.id
	WHERE
		repo.id = %s
		-- where the changeset is not archived on at least one batch change
		AND jsonb_path_exists (batch_change_ids, '$.* ? ((!exists(@.isArchived) || @.isArchived == false) && (!exists(@.archive) || @.archive == false))')
		-- authz conditions:
		AND %s
) AS fcs;
`

const getGlobalChangesetsStatsFmtstr = `
SELECT
	COUNT(*) AS total,
	COUNT(*) FILTER (WHERE computed_state = 'UNPUBLISHED') AS unpublished,
	COUNT(*) FILTER (WHERE computed_state = 'DRAFT') AS draft,
	COUNT(*) FILTER (WHERE computed_state = 'CLOSED') AS closed,
	COUNT(*) FILTER (WHERE computed_state = 'MERGED') AS merged,
	COUNT(*) FILTER (WHERE computed_state = 'OPEN') AS open
FROM (
	SELECT
		changesets.id,
		changesets.computed_state
	FROM
		changesets
	INNER JOIN repo ON repo.id = changesets.repo_id
	WHERE
		-- where the changeset is not archived on at least one batch change
		jsonb_path_exists (batch_change_ids, '$.* ? ((!exists(@.isArchived) || @.isArchived == false) && (!exists(@.archive) || @.archive == false))')
	AND
		-- where the repo is neither deleted nor blocked
		repo.deleted_at is null and repo.blocked is null
		) AS fcs;
`

func batchChangesColumn(c *btypes.Changeset) ([]byte, error) {
	assocsAsMap := make(map[int64]btypes.BatchChangeAssoc, len(c.BatchChanges))
	for _, assoc := range c.BatchChanges {
		assocsAsMap[assoc.BatchChangeID] = assoc
	}

	return json.Marshal(assocsAsMap)
}

func uiPublicationStateColumn(c *btypes.Changeset) *string {
	var uiPublicationState *string
	if state := c.UiPublicationState; state != nil {
		uiPublicationState = dbutil.NullStringColumn(string(*state))
	}
	return uiPublicationState
}

// CleanDetachedChangesets deletes changesets that have been detached after duration specified.
func (s *Store) CleanDetachedChangesets(ctx context.Context, retention time.Duration) (err error) {
	ctx, _, endObservation := s.operations.cleanDetachedChangesets.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Stringer("Retention", retention),
	}})
	defer endObservation(1, observation.Args{})

	return s.Exec(ctx, sqlf.Sprintf(cleanDetachedChangesetsFmtstr, retention/time.Second))
}

const cleanDetachedChangesetsFmtstr = `
DELETE FROM changesets WHERE detached_at < (NOW() - (%s * interval '1 second'));
`
