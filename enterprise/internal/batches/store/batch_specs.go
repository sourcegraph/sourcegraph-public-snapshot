package store

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

// batchSpecColumns are used by the batchSpec related Store methods to insert,
// update and query batch specs.
var batchSpecColumns = []*sqlf.Query{
	sqlf.Sprintf("batch_specs.id"),
	sqlf.Sprintf("batch_specs.rand_id"),
	sqlf.Sprintf("batch_specs.raw_spec"),
	sqlf.Sprintf("batch_specs.spec"),
	sqlf.Sprintf("batch_specs.namespace_user_id"),
	sqlf.Sprintf("batch_specs.namespace_org_id"),
	sqlf.Sprintf("batch_specs.user_id"),
	sqlf.Sprintf("batch_specs.created_from_raw"),
	sqlf.Sprintf("batch_specs.allow_unsupported"),
	sqlf.Sprintf("batch_specs.allow_ignored"),
	sqlf.Sprintf("batch_specs.created_at"),
	sqlf.Sprintf("batch_specs.updated_at"),
}

// batchSpecInsertColumns is the list of batch_specs columns that are
// modified when updating/inserting batch specs.
var batchSpecInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("rand_id"),
	sqlf.Sprintf("raw_spec"),
	sqlf.Sprintf("spec"),
	sqlf.Sprintf("namespace_user_id"),
	sqlf.Sprintf("namespace_org_id"),
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("created_from_raw"),
	sqlf.Sprintf("allow_unsupported"),
	sqlf.Sprintf("allow_ignored"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

const batchSpecInsertColsFmt = `(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`

// CreateBatchSpec creates the given BatchSpec.
func (s *Store) CreateBatchSpec(ctx context.Context, c *btypes.BatchSpec) (err error) {
	ctx, endObservation := s.operations.createBatchSpec.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q, err := s.createBatchSpecQuery(c)
	if err != nil {
		return err
	}
	return s.query(ctx, q, func(sc dbutil.Scanner) error { return scanBatchSpec(c, sc) })
}

var createBatchSpecQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:CreateBatchSpec
INSERT INTO batch_specs (%s)
VALUES ` + batchSpecInsertColsFmt + `
RETURNING %s`

func (s *Store) createBatchSpecQuery(c *btypes.BatchSpec) (*sqlf.Query, error) {
	spec, err := jsonbColumn(c.Spec)
	if err != nil {
		return nil, err
	}

	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	if c.RandID == "" {
		if c.RandID, err = RandomID(); err != nil {
			return nil, errors.Wrap(err, "creating RandID failed")
		}
	}

	return sqlf.Sprintf(
		createBatchSpecQueryFmtstr,
		sqlf.Join(batchSpecInsertColumns, ", "),
		c.RandID,
		c.RawSpec,
		spec,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		nullInt32Column(c.UserID),
		c.CreatedFromRaw,
		c.AllowUnsupported,
		c.AllowIgnored,
		c.CreatedAt,
		c.UpdatedAt,
		sqlf.Join(batchSpecColumns, ", "),
	), nil
}

// UpdateBatchSpec updates the given BatchSpec.
func (s *Store) UpdateBatchSpec(ctx context.Context, c *btypes.BatchSpec) (err error) {
	ctx, endObservation := s.operations.updateBatchSpec.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(c.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q, err := s.updateBatchSpecQuery(c)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanBatchSpec(c, sc)
	})
}

var updateBatchSpecQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:UpdateBatchSpec
UPDATE batch_specs
SET (%s) = ` + batchSpecInsertColsFmt + `
WHERE id = %s
RETURNING %s`

func (s *Store) updateBatchSpecQuery(c *btypes.BatchSpec) (*sqlf.Query, error) {
	spec, err := jsonbColumn(c.Spec)
	if err != nil {
		return nil, err
	}

	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateBatchSpecQueryFmtstr,
		sqlf.Join(batchSpecInsertColumns, ", "),
		c.RandID,
		c.RawSpec,
		spec,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		nullInt32Column(c.UserID),
		c.CreatedFromRaw,
		c.AllowUnsupported,
		c.AllowIgnored,
		c.CreatedAt,
		c.UpdatedAt,
		c.ID,
		sqlf.Join(batchSpecColumns, ", "),
	), nil
}

// DeleteBatchSpec deletes the BatchSpec with the given ID.
func (s *Store) DeleteBatchSpec(ctx context.Context, id int64) (err error) {
	ctx, endObservation := s.operations.deleteBatchSpec.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteBatchSpecQueryFmtstr, id))
}

var deleteBatchSpecQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:DeleteBatchSpec
DELETE FROM batch_specs WHERE id = %s
`

// CountBatchSpecsOpts captures the query options needed for
// counting batch specs.
type CountBatchSpecsOpts struct {
	BatchChangeID int64
}

// CountBatchSpecs returns the number of code mods in the database.
func (s *Store) CountBatchSpecs(ctx context.Context, opts CountBatchSpecsOpts) (count int, err error) {
	ctx, endObservation := s.operations.countBatchSpecs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := countBatchSpecsQuery(opts)

	return s.queryCount(ctx, q)
}

var countBatchSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:CountBatchSpecs
SELECT COUNT(id)
FROM batch_specs
-- Joins go here:
%s
WHERE %s
`

func countBatchSpecsQuery(opts CountBatchSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	joins := []*sqlf.Query{}

	if opts.BatchChangeID != 0 {
		joins = append(joins, sqlf.Sprintf(`INNER JOIN batch_changes
ON
	batch_changes.name = batch_specs.spec->>'name'
	AND
	batch_changes.namespace_user_id IS NOT DISTINCT FROM batch_specs.namespace_user_id
	AND
	batch_changes.namespace_org_id IS NOT DISTINCT FROM batch_specs.namespace_org_id`))
		preds = append(preds, sqlf.Sprintf("batch_changes.id = %s", opts.BatchChangeID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		countBatchSpecsQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

// GetBatchSpecOpts captures the query options needed for getting a BatchSpec
type GetBatchSpecOpts struct {
	ID     int64
	RandID string
}

// GetBatchSpec gets a BatchSpec matching the given options.
func (s *Store) GetBatchSpec(ctx context.Context, opts GetBatchSpecOpts) (spec *btypes.BatchSpec, err error) {
	ctx, endObservation := s.operations.getBatchSpec.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
		log.String("randID", opts.RandID),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecQuery(&opts)

	var c btypes.BatchSpec
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpec(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:GetBatchSpec
SELECT %s FROM batch_specs
WHERE %s
LIMIT 1
`

func getBatchSpecQuery(opts *GetBatchSpecOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.RandID != "" {
		preds = append(preds, sqlf.Sprintf("rand_id = %s", opts.RandID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getBatchSpecsQueryFmtstr,
		sqlf.Join(batchSpecColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// GetNewestBatchSpecOpts captures the query options needed to get the latest
// batch spec for the given parameters. One of the namespace fields and all
// the others must be defined.
type GetNewestBatchSpecOpts struct {
	NamespaceUserID int32
	NamespaceOrgID  int32
	UserID          int32
	Name            string
}

// GetNewestBatchSpec returns the newest batch spec that matches the given
// options.
func (s *Store) GetNewestBatchSpec(ctx context.Context, opts GetNewestBatchSpecOpts) (spec *btypes.BatchSpec, err error) {
	ctx, endObservation := s.operations.getNewestBatchSpec.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := getNewestBatchSpecQuery(&opts)

	var c btypes.BatchSpec
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpec(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

const getNewestBatchSpecQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:GetNewestBatchSpec
SELECT %s FROM batch_specs
WHERE %s
ORDER BY id DESC
LIMIT 1
`

func getNewestBatchSpecQuery(opts *GetNewestBatchSpecOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("user_id = %s", opts.UserID),
		sqlf.Sprintf("spec->>'name' = %s", opts.Name),
	}

	if opts.NamespaceUserID != 0 {
		preds = append(preds, sqlf.Sprintf(
			"namespace_user_id = %s",
			opts.NamespaceUserID,
		))
	}

	if opts.NamespaceOrgID != 0 {
		preds = append(preds, sqlf.Sprintf(
			"namespace_org_id = %s",
			opts.NamespaceOrgID,
		))
	}

	return sqlf.Sprintf(
		getNewestBatchSpecQueryFmtstr,
		sqlf.Join(batchSpecColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBatchSpecsOpts captures the query options needed for
// listing batch specs.
type ListBatchSpecsOpts struct {
	LimitOpts
	Cursor        int64
	BatchChangeID int64
}

// ListBatchSpecs lists BatchSpecs with the given filters.
func (s *Store) ListBatchSpecs(ctx context.Context, opts ListBatchSpecsOpts) (cs []*btypes.BatchSpec, next int64, err error) {
	ctx, endObservation := s.operations.listBatchSpecs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecsQuery(&opts)

	cs = make([]*btypes.BatchSpec, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.BatchSpec
		if err := scanBatchSpec(&c, sc); err != nil {
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

var listBatchSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:ListBatchSpecs
SELECT %s FROM batch_specs
-- Joins go here:
%s
WHERE %s
ORDER BY id ASC
`

func listBatchSpecsQuery(opts *ListBatchSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("batch_specs.id >= %s", opts.Cursor),
	}
	joins := []*sqlf.Query{}

	if opts.BatchChangeID != 0 {
		joins = append(joins, sqlf.Sprintf(`INNER JOIN batch_changes
ON
	batch_changes.name = batch_specs.spec->>'name'
	AND
	batch_changes.namespace_user_id IS NOT DISTINCT FROM batch_specs.namespace_user_id
	AND
	batch_changes.namespace_org_id IS NOT DISTINCT FROM batch_specs.namespace_org_id`))
		preds = append(preds, sqlf.Sprintf("batch_changes.id = %s", opts.BatchChangeID))
	}

	return sqlf.Sprintf(
		listBatchSpecsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(batchSpecColumns, ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

// DeleteExpiredBatchSpecs deletes BatchSpecs that have not been attached
// to a Batch change within BatchSpecTTL.
func (s *Store) DeleteExpiredBatchSpecs(ctx context.Context) (err error) {
	ctx, endObservation := s.operations.deleteExpiredBatchSpecs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	expirationTime := s.now().Add(-btypes.BatchSpecTTL)
	q := sqlf.Sprintf(deleteExpiredBatchSpecsQueryFmtstr, expirationTime)

	return s.Store.Exec(ctx, q)
}

func (s *Store) GetBatchSpecStats(ctx context.Context, ids []int64) (stats map[int64]btypes.BatchSpecStats, err error) {
	stats = make(map[int64]btypes.BatchSpecStats)
	q := getBatchSpecStatsQuery(ids)
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var (
			s  btypes.BatchSpecStats
			id int64
		)
		if err := sc.Scan(
			&id,
			&s.Workspaces,
			&dbutil.NullTime{Time: &s.StartedAt},
			&dbutil.NullTime{Time: &s.FinishedAt},
			&s.Executions,
			&s.Completed,
			&s.Processing,
			&s.Queued,
			&s.Failed,
			&s.Canceled,
			&s.Canceling,
		); err != nil {
			return err
		}
		stats[id] = s
		return nil
	})
	return stats, err
}

func getBatchSpecStatsQuery(ids []int64) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("batch_specs.id = ANY(%s)", pq.Array(ids)),
	}

	return sqlf.Sprintf(
		getBatchSpecStatsFmtstr,
		sqlf.Join(preds, " AND "),
	)
}

const getBatchSpecStatsFmtstr = `
-- source: enterprise/internal/batches/store/batch_specs.go:GetBatchSpecStats
SELECT
	batch_specs.id AS batch_spec_id,
	COUNT(ws.id) AS workspaces,
	MIN(jobs.started_at) AS started_at,
	MAX(jobs.finished_at) AS finished_at,
	COUNT(jobs.id) AS executions,
	COUNT(jobs.id) FILTER (WHERE jobs.state = 'completed') AS completed,
	COUNT(jobs.id) FILTER (WHERE jobs.state = 'processing' AND jobs.cancel = FALSE) AS processing,
	COUNT(jobs.id) FILTER (WHERE jobs.state = 'queued') AS queued,
	COUNT(jobs.id) FILTER (WHERE jobs.state = 'failed' AND jobs.cancel = FALSE) AS failed,
	COUNT(jobs.id) FILTER (WHERE jobs.state = 'failed' AND jobs.cancel = TRUE) AS canceled,
	COUNT(jobs.id) FILTER (WHERE jobs.state = 'processing' AND jobs.cancel = TRUE) AS canceling
FROM batch_specs
LEFT JOIN batch_spec_workspaces ws ON ws.batch_spec_id = batch_specs.id
LEFT JOIN batch_spec_workspace_execution_jobs jobs ON jobs.batch_spec_workspace_id = ws.id
WHERE
	%s
GROUP BY batch_specs.id
`

var deleteExpiredBatchSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store.go:DeleteExpiredBatchSpecs
DELETE FROM
  batch_specs
WHERE
  created_at < %s
AND NOT EXISTS (
  SELECT 1 FROM batch_changes WHERE batch_spec_id = batch_specs.id
)
AND NOT EXISTS (
  SELECT 1 FROM changeset_specs WHERE batch_spec_id = batch_specs.id
)
`

func scanBatchSpec(c *btypes.BatchSpec, s dbutil.Scanner) error {
	var spec json.RawMessage

	err := s.Scan(
		&c.ID,
		&c.RandID,
		&c.RawSpec,
		&spec,
		&dbutil.NullInt32{N: &c.NamespaceUserID},
		&dbutil.NullInt32{N: &c.NamespaceOrgID},
		&dbutil.NullInt32{N: &c.UserID},
		&c.CreatedFromRaw,
		&c.AllowUnsupported,
		&c.AllowIgnored,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, "scanning batch spec")
	}

	var batchSpec batcheslib.BatchSpec
	if err = json.Unmarshal(spec, &batchSpec); err != nil {
		return errors.Wrap(err, "scanBatchSpec: failed to unmarshal spec")
	}
	c.Spec = &batchSpec

	return nil
}
