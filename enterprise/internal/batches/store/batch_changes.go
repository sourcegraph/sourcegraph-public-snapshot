package store

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/go-diff/diff"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// batchChangeColumns are used by the batch change related Store methods to insert,
// update and query batches.
var batchChangeColumns = []*sqlf.Query{
	sqlf.Sprintf("batch_changes.id"),
	sqlf.Sprintf("batch_changes.name"),
	sqlf.Sprintf("batch_changes.description"),
	sqlf.Sprintf("batch_changes.initial_applier_id"),
	sqlf.Sprintf("batch_changes.last_applier_id"),
	sqlf.Sprintf("batch_changes.last_applied_at"),
	sqlf.Sprintf("batch_changes.namespace_user_id"),
	sqlf.Sprintf("batch_changes.namespace_org_id"),
	sqlf.Sprintf("batch_changes.created_at"),
	sqlf.Sprintf("batch_changes.updated_at"),
	sqlf.Sprintf("batch_changes.closed_at"),
	sqlf.Sprintf("batch_changes.batch_spec_id"),
}

// batchChangeInsertColumns is the list of batch changes columns that are
// modified in CreateBatchChange and UpdateBatchChange.
// update and query batches.
var batchChangeInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("name"),
	sqlf.Sprintf("description"),
	sqlf.Sprintf("initial_applier_id"),
	sqlf.Sprintf("last_applier_id"),
	sqlf.Sprintf("last_applied_at"),
	sqlf.Sprintf("namespace_user_id"),
	sqlf.Sprintf("namespace_org_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("closed_at"),
	sqlf.Sprintf("batch_spec_id"),
}

// CreateBatchChange creates the given batch change.
func (s *Store) CreateBatchChange(ctx context.Context, c *btypes.BatchChange) (err error) {
	ctx, endObservation := s.operations.createBatchChange.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := s.createBatchChangeQuery(c)

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchChange(c, sc)
	})
}

var createBatchChangeQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:CreateBatchChange
INSERT INTO batch_changes (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

func (s *Store) createBatchChangeQuery(c *btypes.BatchChange) *sqlf.Query {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	return sqlf.Sprintf(
		createBatchChangeQueryFmtstr,
		sqlf.Join(batchChangeInsertColumns, ", "),
		c.Name,
		c.Description,
		nullInt32Column(c.InitialApplierID),
		nullInt32Column(c.LastApplierID),
		c.LastAppliedAt,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		c.CreatedAt,
		c.UpdatedAt,
		nullTimeColumn(c.ClosedAt),
		c.BatchSpecID,
		sqlf.Join(batchChangeColumns, ", "),
	)
}

// UpdateBatchChange updates the given bach change.
func (s *Store) UpdateBatchChange(ctx context.Context, c *btypes.BatchChange) (err error) {
	ctx, endObservation := s.operations.updateBatchChange.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(c.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := s.updateBatchChangeQuery(c)

	return s.query(ctx, q, func(sc dbutil.Scanner) (err error) { return scanBatchChange(c, sc) })
}

var updateBatchChangeQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:UpdateBatchChange
UPDATE batch_changes
SET (%s) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING %s
`

func (s *Store) updateBatchChangeQuery(c *btypes.BatchChange) *sqlf.Query {
	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateBatchChangeQueryFmtstr,
		sqlf.Join(batchChangeInsertColumns, ", "),
		c.Name,
		c.Description,
		nullInt32Column(c.InitialApplierID),
		nullInt32Column(c.LastApplierID),
		c.LastAppliedAt,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		c.CreatedAt,
		c.UpdatedAt,
		nullTimeColumn(c.ClosedAt),
		c.BatchSpecID,
		c.ID,
		sqlf.Join(batchChangeColumns, ", "),
	)
}

// DeleteBatchChange deletes the batch change with the given ID.
func (s *Store) DeleteBatchChange(ctx context.Context, id int64) (err error) {
	ctx, endObservation := s.operations.deleteBatchChange.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteBatchChangeQueryFmtstr, id))
}

var deleteBatchChangeQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:DeleteBatchChange
DELETE FROM batch_changes WHERE id = %s
`

// CountBatchChangesOpts captures the query options needed for
// counting batches.
type CountBatchChangesOpts struct {
	ChangesetID int64
	State       btypes.BatchChangeState
	RepoID      api.RepoID

	InitialApplierID int32

	NamespaceUserID int32
	NamespaceOrgID  int32
}

// CountBatchChanges returns the number of batch changes in the database.
func (s *Store) CountBatchChanges(ctx context.Context, opts CountBatchChangesOpts) (count int, err error) {
	ctx, endObservation := s.operations.countBatchChanges.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repoAuthzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return 0, errors.Wrap(err, "CountBatchChanges generating authz query conds")
	}

	return s.queryCount(ctx, countBatchChangesQuery(&opts, repoAuthzConds))
}

var countBatchChangesQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:CountBatchChanges
SELECT COUNT(batch_changes.id)
FROM batch_changes
%s
WHERE %s
`

func countBatchChangesQuery(opts *CountBatchChangesOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	joins := []*sqlf.Query{
		sqlf.Sprintf("LEFT JOIN users namespace_user ON batch_changes.namespace_user_id = namespace_user.id"),
		sqlf.Sprintf("LEFT JOIN orgs namespace_org ON batch_changes.namespace_org_id = namespace_org.id"),
	}
	preds := []*sqlf.Query{
		sqlf.Sprintf("namespace_user.deleted_at IS NULL"),
		sqlf.Sprintf("namespace_org.deleted_at IS NULL"),
	}

	if opts.ChangesetID != 0 {
		joins = append(joins, sqlf.Sprintf("INNER JOIN changesets ON changesets.batch_change_ids ? batch_changes.id::TEXT"))
		preds = append(preds, sqlf.Sprintf("changesets.id = %s", opts.ChangesetID))
	}

	switch opts.State {
	case btypes.BatchChangeStateOpen:
		preds = append(preds, sqlf.Sprintf("batch_changes.closed_at IS NULL"))
	case btypes.BatchChangeStateClosed:
		preds = append(preds, sqlf.Sprintf("batch_changes.closed_at IS NOT NULL"))
	}

	if opts.InitialApplierID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.initial_applier_id = %d", opts.InitialApplierID))
	}

	if opts.NamespaceUserID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.namespace_user_id = %s", opts.NamespaceUserID))
	}

	if opts.NamespaceOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.namespace_org_id = %s", opts.NamespaceOrgID))
	}

	if opts.RepoID != 0 {
		preds = append(preds, sqlf.Sprintf(`EXISTS(
			SELECT * FROM changesets
			INNER JOIN repo ON changesets.repo_id = repo.id
			WHERE
				changesets.batch_change_ids ? batch_changes.id::TEXT AND
				changesets.repo_id = %s AND
				repo.deleted_at IS NULL AND
				-- authz conditions:
				%s
		)`, opts.RepoID, repoAuthzConds))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countBatchChangesQueryFmtstr, sqlf.Join(joins, "\n"), sqlf.Join(preds, "\n AND "))
}

// GetBatchChangeOpts captures the query options needed for getting a batch change
type GetBatchChangeOpts struct {
	ID int64

	NamespaceUserID int32
	NamespaceOrgID  int32

	BatchSpecID int64
	Name        string
}

// GetBatchChange gets a batch change matching the given options.
func (s *Store) GetBatchChange(ctx context.Context, opts GetBatchChangeOpts) (bc *btypes.BatchChange, err error) {
	ctx, endObservation := s.operations.getBatchChange.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchChangeQuery(&opts)

	var c btypes.BatchChange
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanBatchChange(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchChangesQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:GetBatchChange
SELECT %s FROM batch_changes
LEFT JOIN users namespace_user ON batch_changes.namespace_user_id = namespace_user.id
LEFT JOIN orgs  namespace_org  ON batch_changes.namespace_org_id = namespace_org.id
WHERE %s
LIMIT 1
`

func getBatchChangeQuery(opts *GetBatchChangeOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("namespace_user.deleted_at IS NULL"),
		sqlf.Sprintf("namespace_org.deleted_at IS NULL"),
	}
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.id = %s", opts.ID))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.batch_spec_id = %s", opts.BatchSpecID))
	}

	if opts.NamespaceUserID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.namespace_user_id = %s", opts.NamespaceUserID))
	}

	if opts.NamespaceOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.namespace_org_id = %s", opts.NamespaceOrgID))
	}

	if opts.Name != "" {
		preds = append(preds, sqlf.Sprintf("batch_changes.name = %s", opts.Name))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getBatchChangesQueryFmtstr,
		sqlf.Join(batchChangeColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

type GetBatchChangeDiffStatOpts struct {
	BatchChangeID int64
}

func (s *Store) GetBatchChangeDiffStat(ctx context.Context, opts GetBatchChangeDiffStatOpts) (stat *diff.Stat, err error) {
	ctx, endObservation := s.operations.getBatchChangeDiffStat.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchChangeID", int(opts.BatchChangeID)),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return nil, errors.Wrap(err, "GetBatchChangeDiffStat generating authz query conds")
	}
	q := getBatchChangeDiffStatQuery(opts, authzConds)

	var diffStat diff.Stat
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		return sc.Scan(&diffStat.Added, &diffStat.Changed, &diffStat.Deleted)
	})
	if err != nil {
		return nil, err
	}

	return &diffStat, nil
}

var getBatchChangeDiffStatQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:GetBatchChangeDiffStat
SELECT
	COALESCE(SUM(diff_stat_added), 0) AS added,
	COALESCE(SUM(diff_stat_changed), 0) AS changed,
	COALESCE(SUM(diff_stat_deleted), 0) AS deleted
FROM
	changesets
INNER JOIN repo ON changesets.repo_id = repo.id
WHERE
	changesets.batch_change_ids ? %s AND
	repo.deleted_at IS NULL AND
	-- authz conditions:
	%s
`

func getBatchChangeDiffStatQuery(opts GetBatchChangeDiffStatOpts, authzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(getBatchChangeDiffStatQueryFmtstr, strconv.Itoa(int(opts.BatchChangeID)), authzConds)
}

func (s *Store) GetRepoDiffStat(ctx context.Context, repoID api.RepoID) (stat *diff.Stat, err error) {
	ctx, endObservation := s.operations.getRepoDiffStat.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repoID", int(repoID)),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return nil, errors.Wrap(err, "GetRepoDiffStat generating authz query conds")
	}
	q := getRepoDiffStatQuery(int64(repoID), authzConds)

	var diffStat diff.Stat
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		return sc.Scan(&diffStat.Added, &diffStat.Changed, &diffStat.Deleted)
	})
	if err != nil {
		return nil, err
	}

	return &diffStat, nil
}

var getRepoDiffStatQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:GetRepoDiffStat
SELECT
	COALESCE(SUM(diff_stat_added), 0) AS added,
	COALESCE(SUM(diff_stat_changed), 0) AS changed,
	COALESCE(SUM(diff_stat_deleted), 0) AS deleted
FROM changesets
INNER JOIN repo ON changesets.repo_id = repo.id
WHERE
	changesets.repo_id = %s AND
	repo.deleted_at IS NULL AND
	-- authz conditions:
	%s
`

func getRepoDiffStatQuery(repoID int64, authzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(getRepoDiffStatQueryFmtstr, repoID, authzConds)
}

// ListBatchChangesOpts captures the query options needed for
// listing batches.
type ListBatchChangesOpts struct {
	LimitOpts
	ChangesetID int64
	Cursor      int64
	State       btypes.BatchChangeState

	InitialApplierID int32

	NamespaceUserID int32
	NamespaceOrgID  int32

	RepoID api.RepoID
}

// ListBatchChanges lists batch changes with the given filters.
func (s *Store) ListBatchChanges(ctx context.Context, opts ListBatchChangesOpts) (cs []*btypes.BatchChange, next int64, err error) {
	ctx, endObservation := s.operations.listBatchChanges.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repoAuthzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListBatchChanges generating authz query conds")
	}
	q := listBatchChangesQuery(&opts, repoAuthzConds)

	cs = make([]*btypes.BatchChange, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.BatchChange
		if err := scanBatchChange(&c, sc); err != nil {
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

var listBatchChangesQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:ListBatchChanges
SELECT %s FROM batch_changes
%s
WHERE %s
ORDER BY id DESC
`

func listBatchChangesQuery(opts *ListBatchChangesOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	joins := []*sqlf.Query{
		sqlf.Sprintf("LEFT JOIN users namespace_user ON batch_changes.namespace_user_id = namespace_user.id"),
		sqlf.Sprintf("LEFT JOIN orgs namespace_org ON batch_changes.namespace_org_id = namespace_org.id"),
	}
	preds := []*sqlf.Query{
		sqlf.Sprintf("namespace_user.deleted_at IS NULL"),
		sqlf.Sprintf("namespace_org.deleted_at IS NULL"),
	}

	if opts.Cursor != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.id <= %s", opts.Cursor))
	}

	if opts.ChangesetID != 0 {
		joins = append(joins, sqlf.Sprintf("INNER JOIN changesets ON changesets.batch_change_ids ? batch_changes.id::TEXT"))
		preds = append(preds, sqlf.Sprintf("changesets.id = %s", opts.ChangesetID))
	}

	switch opts.State {
	case btypes.BatchChangeStateOpen:
		preds = append(preds, sqlf.Sprintf("batch_changes.closed_at IS NULL"))
	case btypes.BatchChangeStateClosed:
		preds = append(preds, sqlf.Sprintf("batch_changes.closed_at IS NOT NULL"))
	}

	if opts.InitialApplierID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.initial_applier_id = %d", opts.InitialApplierID))
	}

	if opts.NamespaceUserID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.namespace_user_id = %s", opts.NamespaceUserID))
	}

	if opts.NamespaceOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_changes.namespace_org_id = %s", opts.NamespaceOrgID))
	}

	if opts.RepoID != 0 {
		preds = append(preds, sqlf.Sprintf(`EXISTS(
			SELECT * FROM changesets
			INNER JOIN repo ON changesets.repo_id = repo.id
			WHERE
				changesets.batch_change_ids ? batch_changes.id::TEXT AND
				changesets.repo_id = %s AND
				repo.deleted_at IS NULL AND
				-- authz conditions:
				%s
		)`, opts.RepoID, repoAuthzConds))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchChangesQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(batchChangeColumns, ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

func scanBatchChange(c *btypes.BatchChange, s dbutil.Scanner) error {
	return s.Scan(
		&c.ID,
		&c.Name,
		&dbutil.NullString{S: &c.Description},
		&dbutil.NullInt32{N: &c.InitialApplierID},
		&dbutil.NullInt32{N: &c.LastApplierID},
		&c.LastAppliedAt,
		&dbutil.NullInt32{N: &c.NamespaceUserID},
		&dbutil.NullInt32{N: &c.NamespaceOrgID},
		&c.CreatedAt,
		&c.UpdatedAt,
		&dbutil.NullTime{Time: &c.ClosedAt},
		&c.BatchSpecID,
	)
}

type ListExternalServicesOpts struct {
	LimitOpts
	Cursor        int64
	BatchChangeID int64
}

// ListExternalServices lists the external services that the given batch change
// has changesets published on.
//
// 🚨 SECURITY: Only site admins should have access to the list of external
// services. The raw results of this method MUST only be used in an API call if
// a site admin check has already occurred before invoking this method.
func (s *Store) ListExternalServices(ctx context.Context, opts ListExternalServicesOpts) (es []*types.ExternalService, next int64, err error) {
	ctx, endObservation := s.operations.listExternalServices.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	es, next, err = s.listExternalServices(ctx, &opts)
	return
}

func (s *Store) listExternalServices(ctx context.Context, opts *ListExternalServicesOpts) ([]*types.ExternalService, int64, error) {
	// We'll retrieve the external service IDs for the given batch change first,
	// then call ExternalServiceStore.List to actually hydrate the list of
	// returned external services.
	//
	// This involves two SELECTs, which would normally be a bit silly, but
	// there's a good reason for this: we need to apply the authz query
	// conditions at this level, since we have the repos here, and then
	// ExternalServiceStore.List has logic to handle decrypting the
	// configuration.
	//
	// The actual cost of this, in practice, is generally trivial: most batch
	// changes are going to ultimately only touch one external service (since
	// most customers only have one external service). We'll spend far more time
	// doing the joins and grouping required for the first query (to get the
	// IDs) -- which we'd have to do anyway, even if we replicated the
	// decryption logic here -- that we will making a SELECT of a single record
	// by its ID.

	repoAuthzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListExternalServices generating authz query conds")
	}
	q := listExternalServicesQuery(opts, repoAuthzConds)

	// Let's go get some external service IDs.
	ids := make([]int64, 0, opts.DBLimit())
	if err := s.query(ctx, q, func(sc scanner) error {
		var id int64
		if err := sc.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
		return nil
	}); err != nil {
		return nil, 0, errors.Wrap(err, "ListExternalServices querying external service IDs")
	}

	// ExternalServiceStore.List will treat an empty ID list as being a query to
	// retrieve _all_ external services, so we need to short circuit that here.
	if len(ids) == 0 {
		return []*types.ExternalService{}, 0, nil
	}

	// Calculate the next cursor, if any.
	var next int64
	if opts.Limit != 0 && len(ids) == opts.DBLimit() {
		next = ids[len(ids)-1]
		ids = ids[:len(ids)-1]
	}

	// Now we'll go retrieve the real ExternalService objects.
	es, err := database.ExternalServicesWith(s.Store).List(ctx, database.ExternalServicesListOptions{
		IDs: ids,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListExternalServices querying external services")
	}

	return es, next, nil
}

const listExternalServicesQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_changes.go:ListExternalService
SELECT
	DISTINCT external_service_repos.external_service_id
FROM
	external_service_repos
INNER JOIN
	repo ON external_service_repos.repo_id = repo.id
INNER JOIN
	changesets ON repo.id = changesets.repo_id
WHERE
	changesets.batch_change_ids ? %s AND
	repo.deleted_at IS NULL AND
	%s AND -- authz conditions
	%s -- cursor, if given
ORDER BY
	external_service_repos.external_service_id ASC
`

func listExternalServicesQuery(opts *ListExternalServicesOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	var cursor *sqlf.Query
	if opts.Cursor != 0 {
		cursor = sqlf.Sprintf("external_service_repos.external_service_id >= %s", opts.Cursor)
	} else {
		cursor = sqlf.Sprintf("TRUE")
	}

	return sqlf.Sprintf(
		listExternalServicesQueryFmtstr+opts.LimitOpts.ToDB(),
		fmt.Sprint(opts.BatchChangeID),
		repoAuthzConds,
		cursor,
	)
}

func (s *Store) CountExternalServices(ctx context.Context, batchChangeID int64) (count int64, err error) {
	ctx, endObservation := s.operations.listExternalServices.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, err = s.countExternalServices(ctx, batchChangeID)
	return
}

func (s *Store) countExternalServices(ctx context.Context, batchChangeID int64) (int64, error) {
	repoAuthzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return 0, errors.Wrap(err, "CountExternalServices generating authz query conds")
	}
	q := countExternalServicesQuery(batchChangeID, repoAuthzConds)

	var count int64
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "CountExternalServices executing query")
	}

	return count, nil
}

const countExternalServicesQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_changes.go:ListExternalService
SELECT
	COUNT(DISTINCT external_service_repos.external_service_id)
FROM
	external_service_repos
INNER JOIN
	repo ON external_service_repos.repo_id = repo.id
INNER JOIN
	changesets ON repo.id = changesets.repo_id
WHERE
	changesets.batch_change_ids ? %s AND
	repo.deleted_at IS NULL AND
	%s -- authz conditions
`

func countExternalServicesQuery(batchChangeID int64, repoAuthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		countExternalServicesQueryFmtstr,
		fmt.Sprint(batchChangeID),
		repoAuthzConds,
	)
}
