package store

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/search"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// changesetSpecInsertColumns is the list of changeset_specs columns that are
// modified when inserting or updating a changeset spec.
var changesetSpecInsertColumns = []string{
	"rand_id",
	"batch_spec_id",
	"repo_id",
	"user_id",
	"diff_stat_added",
	"diff_stat_changed",
	"diff_stat_deleted",
	"created_at",
	"updated_at",
	"fork_namespace",

	"external_id",
	"head_ref",
	"title",
	"base_rev",
	"base_ref",
	"body",
	"published",
	"diff",
	"commit_message",
	"commit_author_name",
	"commit_author_email",
	"type",
}

// changesetSpecColumns are used by the changeset spec related Store methods to
// insert, update and query changeset specs.
var changesetSpecColumns = SQLColumns{
	"changeset_specs.id",
	"changeset_specs.rand_id",
	"changeset_specs.batch_spec_id",
	"changeset_specs.repo_id",
	"changeset_specs.user_id",
	"changeset_specs.diff_stat_added",
	"changeset_specs.diff_stat_changed",
	"changeset_specs.diff_stat_deleted",
	"changeset_specs.created_at",
	"changeset_specs.updated_at",
	"changeset_specs.fork_namespace",
	"changeset_specs.external_id",
	"changeset_specs.head_ref",
	"changeset_specs.title",
	"changeset_specs.base_rev",
	"changeset_specs.base_ref",
	"changeset_specs.body",
	"changeset_specs.published",
	"changeset_specs.diff",
	"changeset_specs.commit_message",
	"changeset_specs.commit_author_name",
	"changeset_specs.commit_author_email",
	"changeset_specs.type",
}

var oneGigabyte = 1000000000

// CreateChangesetSpec creates the given ChangesetSpecs.
func (s *Store) CreateChangesetSpec(ctx context.Context, cs ...*btypes.ChangesetSpec) (err error) {
	ctx, _, endObservation := s.operations.createChangesetSpec.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("Count", len(cs)),
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

			if c.RandID == "" {
				if c.RandID, err = RandomID(); err != nil {
					return errors.Wrap(err, "creating RandID failed")
				}
			}

			var published []byte
			if c.Published.Val != nil {
				published, err = json.Marshal(c.Published)
				if err != nil {
					return err
				}
			}

			// We check if the resulting diff is greater than 1GB, since the limit
			// for the diff column (which is bytea) is 1GB
			if len(c.Diff) > oneGigabyte {
				link := "https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference#transformchanges"
				return errors.Errorf("The changeset patch generated is over the size limit. You can make use of [transformChanges](%s) to break down the changesets into smaller pieces.", link)
			}

			if err := inserter.Insert(
				ctx,
				c.RandID,
				nullInt64Column(c.BatchSpecID),
				c.BaseRepoID,
				nullInt32Column(c.UserID),
				c.DiffStatAdded,
				c.DiffStatChanged,
				c.DiffStatDeleted,
				c.CreatedAt,
				c.UpdatedAt,
				c.ForkNamespace,
				dbutil.NewNullString(c.ExternalID),
				dbutil.NewNullString(c.HeadRef),
				dbutil.NewNullString(c.Title),
				dbutil.NewNullString(c.BaseRev),
				dbutil.NewNullString(c.BaseRef),
				dbutil.NewNullString(c.Body),
				published,
				c.Diff,
				dbutil.NewNullString(c.CommitMessage),
				dbutil.NewNullString(c.CommitAuthorName),
				dbutil.NewNullString(c.CommitAuthorEmail),
				c.Type,
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
		"changeset_specs",
		batch.MaxNumPostgresParameters,
		changesetSpecInsertColumns,
		"",
		changesetSpecColumns,
		func(rows dbutil.Scanner) error {
			i++
			return scanChangesetSpec(cs[i], rows)
		},
		inserter,
	)
}

// UpdateChangesetSpecBatchSpecID updates the given ChangesetSpecs to be owned by the given batch spec.
func (s *Store) UpdateChangesetSpecBatchSpecID(ctx context.Context, cs []int64, batchSpec int64) (err error) {
	ctx, _, endObservation := s.operations.updateChangesetSpecBatchSpecID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("Count", len(cs)),
	}})
	defer endObservation(1, observation.Args{})

	q := s.updateChangesetSpecQuery(cs, batchSpec)

	return s.Exec(ctx, q)
}

var updateChangesetSpecBatchSpecIDQueryFmtstr = `
-- source: enterprise/internal/batches/store_changeset_specs.go:UpdateChangesetSpecBatchSpecID
UPDATE changeset_specs
SET batch_spec_id = %s
WHERE id = ANY (%s)
`

func (s *Store) updateChangesetSpecQuery(cs []int64, batchSpec int64) *sqlf.Query {
	return sqlf.Sprintf(
		updateChangesetSpecBatchSpecIDQueryFmtstr,
		batchSpec,
		pq.Array(cs),
	)
}

// DeleteChangesetSpec deletes the ChangesetSpec with the given ID.
func (s *Store) DeleteChangesetSpec(ctx context.Context, id int64) (err error) {
	ctx, _, endObservation := s.operations.deleteChangesetSpec.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteChangesetSpecQueryFmtstr, id))
}

var deleteChangesetSpecQueryFmtstr = `
-- source: enterprise/internal/batches/store_changeset_specs.go:DeleteChangesetSpec
DELETE FROM changeset_specs WHERE id = %s
`

// CountChangesetSpecsOpts captures the query options needed for counting
// ChangesetSpecs.
type CountChangesetSpecsOpts struct {
	BatchSpecID int64
	Type        batcheslib.ChangesetSpecDescriptionType
}

// CountChangesetSpecs returns the number of changeset specs in the database.
func (s *Store) CountChangesetSpecs(ctx context.Context, opts CountChangesetSpecsOpts) (count int, err error) {
	ctx, _, endObservation := s.operations.countChangesetSpecs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSpecID", int(opts.BatchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	return s.queryCount(ctx, countChangesetSpecsQuery(&opts))
}

var countChangesetSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store_changeset_specs.go:CountChangesetSpecs
SELECT COUNT(changeset_specs.id)
FROM changeset_specs
INNER JOIN repo ON repo.id = changeset_specs.repo_id
WHERE %s
`

func countChangesetSpecsQuery(opts *CountChangesetSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.BatchSpecID != 0 {
		cond := sqlf.Sprintf("changeset_specs.batch_spec_id = %s", opts.BatchSpecID)
		preds = append(preds, cond)
	}

	if opts.Type != "" {
		if opts.Type == batcheslib.ChangesetSpecDescriptionTypeExisting {
			// Check that externalID is not empty.
			preds = append(preds, sqlf.Sprintf("changeset_specs.external_id IS NOT NULL"))
		} else {
			// Check that externalID is empty.
			preds = append(preds, sqlf.Sprintf("changeset_specs.external_id IS NULL"))
		}
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countChangesetSpecsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetChangesetSpecOpts captures the query options needed for getting a ChangesetSpec
type GetChangesetSpecOpts struct {
	ID     int64
	RandID string
}

// GetChangesetSpec gets a changeset spec matching the given options.
func (s *Store) GetChangesetSpec(ctx context.Context, opts GetChangesetSpecOpts) (spec *btypes.ChangesetSpec, err error) {
	ctx, _, endObservation := s.operations.getChangesetSpec.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
		log.String("randID", opts.RandID),
	}})
	defer endObservation(1, observation.Args{})

	q := getChangesetSpecQuery(&opts)

	var c btypes.ChangesetSpec
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanChangesetSpec(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

// GetChangesetSpecByID gets a changeset spec with the given ID.
func (s *Store) GetChangesetSpecByID(ctx context.Context, id int64) (*btypes.ChangesetSpec, error) {
	return s.GetChangesetSpec(ctx, GetChangesetSpecOpts{ID: id})
}

var getChangesetSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store_changeset_specs.go:GetChangesetSpec
SELECT %s FROM changeset_specs
INNER JOIN repo ON repo.id = changeset_specs.repo_id
WHERE %s
LIMIT 1
`

func getChangesetSpecQuery(opts *GetChangesetSpecOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_specs.id = %s", opts.ID))
	}

	if opts.RandID != "" {
		preds = append(preds, sqlf.Sprintf("changeset_specs.rand_id = %s", opts.RandID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getChangesetSpecsQueryFmtstr,
		sqlf.Join(changesetSpecColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListChangesetSpecsOpts captures the query options needed for
// listing code mods.
type ListChangesetSpecsOpts struct {
	LimitOpts
	Cursor int64

	BatchSpecID int64
	RandIDs     []string
	IDs         []int64
	Type        batcheslib.ChangesetSpecDescriptionType
}

// ListChangesetSpecs lists ChangesetSpecs with the given filters.
func (s *Store) ListChangesetSpecs(ctx context.Context, opts ListChangesetSpecsOpts) (cs btypes.ChangesetSpecs, next int64, err error) {
	ctx, _, endObservation := s.operations.listChangesetSpecs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listChangesetSpecsQuery(&opts)

	cs = make(btypes.ChangesetSpecs, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.ChangesetSpec
		if err := scanChangesetSpec(&c, sc); err != nil {
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

var listChangesetSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store_changeset_specs.go:ListChangesetSpecs
SELECT %s FROM changeset_specs
INNER JOIN repo ON repo.id = changeset_specs.repo_id
WHERE %s
ORDER BY changeset_specs.id ASC
`

func listChangesetSpecsQuery(opts *ListChangesetSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("changeset_specs.id >= %s", opts.Cursor),
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_specs.batch_spec_id = %d", opts.BatchSpecID))
	}

	if len(opts.RandIDs) != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_specs.rand_id = ANY (%s)", pq.Array(opts.RandIDs)))
	}

	if len(opts.IDs) != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_specs.id = ANY (%s)", pq.Array(opts.IDs)))
	}

	if opts.Type != "" {
		if opts.Type == batcheslib.ChangesetSpecDescriptionTypeExisting {
			// Check that externalID is not empty.
			preds = append(preds, sqlf.Sprintf("changeset_specs.external_id IS NOT NULL"))
		} else {
			// Check that externalID is empty.
			preds = append(preds, sqlf.Sprintf("changeset_specs.external_id IS NULL"))
		}
	}

	return sqlf.Sprintf(
		listChangesetSpecsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(changesetSpecColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

type ChangesetSpecHeadRefConflict struct {
	RepoID  api.RepoID
	HeadRef string
	Count   int
}

var listChangesetSpecsWithConflictingHeadQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_specs.go:ListChangesetSpecsWithConflictingHeadRef
SELECT
	repo_id,
	head_ref,
	COUNT(*) AS count
FROM
	changeset_specs
WHERE
	batch_spec_id = %s
AND
	head_ref IS NOT NULL
GROUP BY
	repo_id, head_ref
HAVING COUNT(*) > 1
ORDER BY repo_id ASC, head_ref ASC
`

func (s *Store) ListChangesetSpecsWithConflictingHeadRef(ctx context.Context, batchSpecID int64) (conflicts []ChangesetSpecHeadRefConflict, err error) {
	ctx, _, endObservation := s.operations.listChangesetSpecsWithConflictingHeadRef.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(listChangesetSpecsWithConflictingHeadQueryFmtstr, batchSpecID)

	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c ChangesetSpecHeadRefConflict
		if err := sc.Scan(&c.RepoID, &c.HeadRef, &c.Count); err != nil {
			return errors.Wrap(err, "scanning head ref conflict")
		}
		conflicts = append(conflicts, c)
		return nil
	})

	return conflicts, err
}

// DeleteUnattachedExpiredChangesetSpecs deletes each ChangesetSpec that has not been
// attached to a BatchSpec within ChangesetSpecTTL.
func (s *Store) DeleteUnattachedExpiredChangesetSpecs(ctx context.Context) (err error) {
	ctx, _, endObservation := s.operations.deleteUnattachedExpiredChangesetSpecs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	changesetSpecTTLExpiration := s.now().Add(-btypes.ChangesetSpecTTL)
	q := sqlf.Sprintf(deleteUnattachedExpiredChangesetSpecsQueryFmtstr, changesetSpecTTLExpiration)
	return s.Store.Exec(ctx, q)
}

var deleteUnattachedExpiredChangesetSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_specs.go:DeleteUnattachedExpiredChangesetSpecs
DELETE FROM
  changeset_specs
WHERE
  -- The spec is older than the ChangesetSpecTTL
  created_at < %s
  AND
  -- and it was never attached to a batch_spec
  batch_spec_id IS NULL
`

// DeleteExpiredChangesetSpecs deletes each ChangesetSpec that is attached
// to a BatchSpec that is not applied and is not attached to a Changeset
// within BatchSpecTTL, and that hasn't been created by SSBC.
func (s *Store) DeleteExpiredChangesetSpecs(ctx context.Context) (err error) {
	ctx, _, endObservation := s.operations.deleteExpiredChangesetSpecs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	batchSpecTTLExpiration := s.now().Add(-btypes.BatchSpecTTL)
	q := sqlf.Sprintf(deleteExpiredChangesetSpecsQueryFmtstr, batchSpecTTLExpiration)
	return s.Store.Exec(ctx, q)
}

var deleteExpiredChangesetSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_specs.go:DeleteExpiredChangesetSpecs
WITH candidates AS (
	SELECT cs.id
	FROM changeset_specs cs
	JOIN batch_specs bs ON bs.id = cs.batch_spec_id
	LEFT JOIN batch_changes bc ON bs.id = bc.batch_spec_id
	LEFT JOIN changesets c ON (c.current_spec_id = cs.id OR c.previous_spec_id = cs.id)
	WHERE
		-- The spec is older than the BatchSpecTTL
		cs.created_at < %s
		-- and it is not created by SSBC
		AND NOT bs.created_from_raw
		-- and the batch spec it is attached to is not applied to a batch change
		AND bc.id IS NULL
		-- and it is not attached to a changeset
		AND c.id IS NULL
	FOR UPDATE OF cs
)
DELETE FROM changeset_specs
WHERE
	id IN (SELECT id FROM candidates)
`

type DeleteChangesetSpecsOpts struct {
	BatchSpecID int64
	IDs         []int64
}

// DeleteChangesetSpecs deletes the ChangesetSpecs matching the given options.
func (s *Store) DeleteChangesetSpecs(ctx context.Context, opts DeleteChangesetSpecsOpts) (err error) {
	ctx, _, endObservation := s.operations.deleteChangesetSpecs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSpecID", int(opts.BatchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	if opts.BatchSpecID == 0 && len(opts.IDs) == 0 {
		return errors.New("BatchSpecID is 0 and no IDs given")
	}

	q := deleteChangesetSpecsQuery(&opts)
	return s.Store.Exec(ctx, q)
}

var deleteChangesetSpecsQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_specs.go:DeleteChangesetSpecs
DELETE FROM changeset_specs
WHERE
  %s
`

func deleteChangesetSpecsQuery(opts *DeleteChangesetSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_specs.batch_spec_id = %s", opts.BatchSpecID))
	}

	if len(opts.IDs) != 0 {
		preds = append(preds, sqlf.Sprintf("changeset_specs.id = ANY(%s)", pq.Array(opts.IDs)))
	}

	return sqlf.Sprintf(deleteChangesetSpecsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

func scanChangesetSpec(c *btypes.ChangesetSpec, s dbutil.Scanner) error {
	var published []byte
	var typ string
	err := s.Scan(
		&c.ID,
		&c.RandID,
		&dbutil.NullInt64{N: &c.BatchSpecID},
		&c.BaseRepoID,
		&dbutil.NullInt32{N: &c.UserID},
		&c.DiffStatAdded,
		&c.DiffStatChanged,
		&c.DiffStatDeleted,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.ForkNamespace,
		&dbutil.NullString{S: &c.ExternalID},
		&dbutil.NullString{S: &c.HeadRef},
		&dbutil.NullString{S: &c.Title},
		&dbutil.NullString{S: &c.BaseRev},
		&dbutil.NullString{S: &c.BaseRef},
		&dbutil.NullString{S: &c.Body},
		&published,
		&c.Diff,
		&dbutil.NullString{S: &c.CommitMessage},
		&dbutil.NullString{S: &c.CommitAuthorName},
		&dbutil.NullString{S: &c.CommitAuthorEmail},
		&typ,
	)
	if err != nil {
		return errors.Wrap(err, "scanning changeset spec")
	}

	c.Type = btypes.ChangesetSpecType(typ)

	if len(published) != 0 {
		if err := json.Unmarshal(published, &c.Published); err != nil {
			return err
		}
	}

	return nil
}

type GetRewirerMappingsOpts struct {
	BatchSpecID   int64
	BatchChangeID int64

	LimitOffset  *database.LimitOffset
	TextSearch   []search.TextSearchTerm
	CurrentState *btypes.ChangesetState
}

// GetRewirerMappings returns RewirerMappings between changeset specs and changesets.
//
// We have two imaginary lists, the current changesets in the batch change and the new changeset specs:
//
// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
// │Changeset 1 | Repo A | #111 | run-gofmt│   │  Spec 1 | Repo A | run-gofmt  │
// └───────────────────────────────────────┘   └───────────────────────────────┘
// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
// │Changeset 2 | Repo B |      | run-gofmt│   │  Spec 2 | Repo B | run-gofmt  │
// └───────────────────────────────────────┘   └───────────────────────────────┘
// ┌───────────────────────────────────────┐   ┌───────────────────────────────────┐
// │Changeset 3 | Repo C | #222 | run-gofmt│   │  Spec 3 | Repo C | run-goimports  │
// └───────────────────────────────────────┘   └───────────────────────────────────┘
// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
// │Changeset 4 | Repo C | #333 | older-pr │   │    Spec 4 | Repo C | #333     │
// └───────────────────────────────────────┘   └───────────────────────────────┘
//
// We need to:
// 1. Find out whether our new specs should _update_ an existing
//    changeset (ChangesetSpec != 0, Changeset != 0), or whether we need to create a new one.
// 2. Since we can have multiple changesets per repository, we need to match
//    based on repo and external ID for imported changesets and on repo and head_ref for 'branch' changesets.
// 3. If a changeset wasn't published yet, it doesn't have an external ID nor does it have an external head_ref.
//    In that case, we need to check whether the branch on which we _might_
//    push the commit (because the changeset might not be published
//    yet) is the same or compare the external IDs in the current and new specs.
//
// What we want:
//
// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
// │Changeset 1 | Repo A | #111 | run-gofmt│───▶│  Spec 1 | Repo A | run-gofmt  │
// └───────────────────────────────────────┘    └───────────────────────────────┘
// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
// │Changeset 2 | Repo B |      | run-gofmt│───▶│  Spec 2 | Repo B | run-gofmt  │
// └───────────────────────────────────────┘    └───────────────────────────────┘
// ┌───────────────────────────────────────┐
// │Changeset 3 | Repo C | #222 | run-gofmt│
// └───────────────────────────────────────┘
// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
// │Changeset 4 | Repo C | #333 | older-pr │───▶│    Spec 4 | Repo C | #333     │
// └───────────────────────────────────────┘    └───────────────────────────────┘
// ┌───────────────────────────────────────┐    ┌───────────────────────────────────┐
// │Changeset 5 | Repo C | | run-goimports │───▶│  Spec 3 | Repo C | run-goimports  │
// └───────────────────────────────────────┘    └───────────────────────────────────┘
//
// Spec 1 should be attached to Changeset 1 and (possibly) update its title/body/diff. (ChangesetSpec = 1, Changeset = 1)
// Spec 2 should be attached to Changeset 2 and publish it on the code host. (ChangesetSpec = 2, Changeset = 2)
// Spec 3 should get a new Changeset, since its branch doesn't match Changeset 3's branch. (ChangesetSpec = 3, Changeset = 0)
// Spec 4 should be attached to Changeset 4, since it tracks PR #333 in Repo C. (ChangesetSpec = 4, Changeset = 4)
// Changeset 3 doesn't have a matching spec and should be detached from the batch change (and closed) (ChangesetSpec == 0, Changeset = 3).
func (s *Store) GetRewirerMappings(ctx context.Context, opts GetRewirerMappingsOpts) (mappings btypes.RewirerMappings, err error) {
	ctx, _, endObservation := s.operations.getRewirerMappings.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSpecID", int(opts.BatchSpecID)),
		log.Int("batchChangeID", int(opts.BatchChangeID)),
	}})
	defer endObservation(1, observation.Args{})

	q, err := getRewirerMappingsQuery(opts)
	if err != nil {
		return nil, err
	}

	if err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.RewirerMapping
		if err := sc.Scan(&c.ChangesetSpecID, &c.ChangesetID, &c.RepoID); err != nil {
			return err
		}
		mappings = append(mappings, &c)
		return nil
	}); err != nil {
		return nil, err
	}

	// Hydrate the rewirer mappings:
	changesetsByID := map[int64]*btypes.Changeset{}
	changesetSpecsByID := map[int64]*btypes.ChangesetSpec{}

	changesetSpecIDs := mappings.ChangesetSpecIDs()
	if len(changesetSpecIDs) > 0 {
		changesetSpecs, _, err := s.ListChangesetSpecs(ctx, ListChangesetSpecsOpts{
			IDs: changesetSpecIDs,
		})
		if err != nil {
			return nil, err
		}
		for _, c := range changesetSpecs {
			changesetSpecsByID[c.ID] = c
		}
	}

	changesetIDs := mappings.ChangesetIDs()
	if len(changesetIDs) > 0 {
		changesets, _, err := s.ListChangesets(ctx, ListChangesetsOpts{IDs: changesetIDs})
		if err != nil {
			return nil, err
		}
		for _, c := range changesets {
			changesetsByID[c.ID] = c
		}
	}

	accessibleReposByID, err := s.Repos().GetReposSetByIDs(ctx, mappings.RepoIDs()...)
	if err != nil {
		return nil, err
	}

	for _, m := range mappings {
		if m.ChangesetID != 0 {
			m.Changeset = changesetsByID[m.ChangesetID]
		}
		if m.ChangesetSpecID != 0 {
			m.ChangesetSpec = changesetSpecsByID[m.ChangesetSpecID]
		}
		if m.RepoID != 0 {
			// This can be nil, but that's okay. It just means the ctx actor has no access to the repo.
			m.Repo = accessibleReposByID[m.RepoID]
		}
	}

	return mappings, err
}

func getRewirerMappingsQuery(opts GetRewirerMappingsOpts) (*sqlf.Query, error) {
	// If there's a text search, we want to add the appropriate WHERE clauses to
	// the query. Note that we need to use different WHERE clauses depending on
	// which part of the big UNION below we're in; more detail on that is
	// documented in getRewirerMappingsTextSearch().
	detachTextSearch, viewTextSearch := getRewirerMappingTextSearch(opts.TextSearch)

	// Happily, current state is simpler. Less happily, it can error.
	currentState := sqlf.Sprintf("")
	if opts.CurrentState != nil {
		currentState = sqlf.Sprintf("AND computed_state = %s", *opts.CurrentState)
	}

	return sqlf.Sprintf(
		getRewirerMappingsQueryFmtstr,
		opts.BatchSpecID,
		viewTextSearch,
		currentState,
		opts.BatchChangeID,
		opts.BatchSpecID,
		viewTextSearch,
		currentState,
		opts.BatchSpecID,
		opts.BatchChangeID,
		opts.BatchSpecID,
		strconv.Itoa(int(opts.BatchChangeID)),
		strconv.Itoa(int(opts.BatchChangeID)),
		detachTextSearch,
		currentState,
		opts.LimitOffset.SQL(),
	), nil
}

func getRewirerMappingTextSearch(terms []search.TextSearchTerm) (detachTextSearch, viewTextSearch *sqlf.Query) {
	// This gets a little tricky: we want to search both the changeset name and
	// the repository name. These are exposed somewhat differently depending on
	// which subquery we're adding the clause to in the big UNION query that's
	// going to get run: the two views expose changeset_name and repo_name
	// fields, whereas the detached changeset subquery has to query the fields
	// directly, since it's just a simple JOIN. As a result, we need two sets of
	// everything.
	if len(terms) > 0 {
		detachSearches := make([]*sqlf.Query, len(terms))
		viewSearches := make([]*sqlf.Query, len(terms))

		for i, term := range terms {
			detachSearches[i] = textSearchTermToClause(
				term,
				sqlf.Sprintf("changesets.external_title"),
				sqlf.Sprintf("repo.name"),
			)

			viewSearches[i] = textSearchTermToClause(
				term,
				sqlf.Sprintf("COALESCE(changeset_name, '')"),
				sqlf.Sprintf("repo_name"),
			)
		}

		detachTextSearch = sqlf.Sprintf("AND %s", sqlf.Join(detachSearches, " AND "))
		viewTextSearch = sqlf.Sprintf("AND %s", sqlf.Join(viewSearches, " AND "))
	} else {
		detachTextSearch = sqlf.Sprintf("")
		viewTextSearch = sqlf.Sprintf("")
	}

	return detachTextSearch, viewTextSearch
}

var getRewirerMappingsQueryFmtstr = `
-- source: enterprise/internal/batches/store_changeset_specs.go:GetRewirerMappings

SELECT mappings.changeset_spec_id, mappings.changeset_id, mappings.repo_id FROM (
	-- Fetch all changeset specs in the batch spec that want to import/track an ChangesetSpecDescriptionTypeExisting changeset.
	-- Match the entries to changesets in the target batch change by external ID and repo.
	SELECT
		changeset_spec_id, changeset_id, repo_id
	FROM
		tracking_changeset_specs_and_changesets
	WHERE
		batch_spec_id = %s
		%s -- text search query, if provided
		%s -- current state, if provided

	UNION ALL

	-- Fetch all changeset specs in the batch spec that are of type ChangesetSpecDescriptionTypeBranch.
	-- Match the entries to changesets in the target batch change by head ref and repo.
	SELECT
		changeset_spec_id, MAX(CASE WHEN owner_batch_change_id = %s THEN changeset_id ELSE 0 END), repo_id
	FROM
		branch_changeset_specs_and_changesets
	WHERE
		batch_spec_id = %s
		%s -- text search query, if provided
		%s -- current state, if provided
	GROUP BY changeset_spec_id, repo_id

	UNION ALL

	-- Finally, fetch all changesets that didn't match a changeset spec in the batch spec and that aren't part of tracked_mappings and branch_mappings. Those are to be closed or detached.
	SELECT 0 as changeset_spec_id, changesets.id as changeset_id, changesets.repo_id as repo_id
	FROM changesets
	INNER JOIN repo ON changesets.repo_id = repo.id
	WHERE
		repo.deleted_at IS NULL AND
		changesets.id NOT IN (
				SELECT
					changeset_id
				FROM
					tracking_changeset_specs_and_changesets
				WHERE
					batch_spec_id = %s
			UNION
				SELECT
					MAX(CASE WHEN owner_batch_change_id = %s THEN changeset_id ELSE 0 END)
				FROM
					branch_changeset_specs_and_changesets
				WHERE
					batch_spec_id = %s
				GROUP BY changeset_spec_id, repo_id
		) AND
		changesets.batch_change_ids ? %s
		AND
		NOT COALESCE((changesets.batch_change_ids->%s->>'isArchived')::bool, false)
		%s -- text search query, if provided
		%s -- current state, if provided
) AS mappings
ORDER BY mappings.changeset_spec_id ASC, mappings.changeset_id ASC
-- LIMIT, OFFSET
%s
`
