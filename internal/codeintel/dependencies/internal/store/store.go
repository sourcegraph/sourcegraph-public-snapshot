package store

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for package dependencies storage.
type Store interface {
	WithTransact(context.Context, func(Store) error) error

	ListPackageRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shared.PackageRepoReference, total int, err error)
	InsertPackageRepoRefs(ctx context.Context, deps []shared.MinimalPackageRepoRef) (newDeps []shared.PackageRepoReference, newVersions []shared.PackageRepoRefVersion, err error)
	DeletePackageRepoRefsByID(ctx context.Context, ids ...int) (err error)
	DeletePackageRepoRefVersionsByID(ctx context.Context, ids ...int) (err error)

	ListPackageRepoRefFilters(ctx context.Context, opts ListPackageRepoRefFiltersOpts) ([]shared.PackageFilter, error)
	CreatePackageRepoFilter(ctx context.Context, filter shared.MinimalPackageFilter) (err error)
	UpdatePackageRepoFilter(ctx context.Context, filter shared.PackageFilter) (err error)
	DeletePacakgeRepoFilter(ctx context.Context, id int) (err error)

	IsPackageRepoVersionAllowed(ctx context.Context, scheme string, pkg reposource.PackageName, version string) (allowed bool, err error)
	IsPackageRepoAllowed(ctx context.Context, scheme string, pkg reposource.PackageName) (allowed bool, err error)
	PackagesOrVersionsMatchingFilter(ctx context.Context, filter shared.MinimalPackageFilter, limit, after int) (_ []shared.PackageRepoReference, _ int, err error)
	ExistsPackageRepoRefLastCheckedBefore(ctx context.Context, before time.Time) (exists bool, err error)
	ApplyPackageFilters(ctx context.Context) (pkgsAffected, versionsAffected int, err error)
}

// store manages the database tables for package dependencies.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new store.
func New(op *observation.Context, db database.DB) *store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(op),
	}
}

func (s *store) WithTransact(ctx context.Context, f func(tx Store) error) error {
	return s.db.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&store{
			db:         tx,
			operations: s.operations,
		})
	})
}

const (
	groupedVersionedPackageReposColumns = `
	lr.id,
	lr.scheme,
	lr.name,
	lr.blocked,
	lr.last_checked_at,
	array_agg(prv.id) as vid,
	array_agg(prv.version) as version,
	array_agg(prv.blocked) as vers_blocked,
	array_agg(prv.last_checked_at) as vers_last_checked_at
`
	packageReposColumns = "lr.id, lr.scheme, lr.name, lr.blocked, lr.last_checked_at"
)

// ListDependencyReposOpts are options for listing dependency repositories.
type ListDependencyReposOpts struct {
	Scheme         string
	Name           reposource.PackageName
	ExactNameOnly  bool
	After          int
	Limit          int
	IncludeBlocked bool
}

// ListDependencyRepos returns dependency repositories to be synced by gitserver.
func (s *store) ListPackageRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shared.PackageRepoReference, total int, err error) {
	ctx, _, endObservation := s.operations.listPackageRepoRefs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("scheme", opts.Scheme),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRepos", len(dependencyRepos)),
		}})
	}()

	query := sqlf.Sprintf(
		listDependencyReposQuery,
		sqlf.Sprintf(groupedVersionedPackageReposColumns),
		sqlf.Join([]*sqlf.Query{makeListDependencyReposConds(opts), makeOffset(opts.After)}, "AND"),
		sqlf.Sprintf("GROUP BY lr.id, lr.scheme, lr.name"),
		sqlf.Sprintf("ORDER BY lr.id ASC"),
		makeLimit(opts.Limit),
	)
	dependencyRepos, err = basestore.NewSliceScanner(scanDependencyRepoWithVersions)(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, errors.Wrap(err, "error listing dependency repos")
	}

	query = sqlf.Sprintf(
		listDependencyReposQuery,
		sqlf.Sprintf("COUNT(DISTINCT(lr.id))"),
		makeListDependencyReposConds(opts),
		sqlf.Sprintf(""),
		sqlf.Sprintf(""),
		sqlf.Sprintf("LIMIT ALL"),
	)
	totalCount, _, err := basestore.ScanFirstInt(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, errors.Wrap(err, "error counting dependency repos")
	}

	return dependencyRepos, totalCount, err
}

const listDependencyReposQuery = `
SELECT %s
FROM lsif_dependency_repos lr
JOIN package_repo_versions prv ON lr.id = prv.package_id
WHERE %s
%s -- group by
%s -- order by
%s -- limit
`

func makeListDependencyReposConds(opts ListDependencyReposOpts) *sqlf.Query {
	conds := make([]*sqlf.Query, 0, 4)

	if opts.Scheme != "" {
		conds = append(conds, sqlf.Sprintf("scheme = %s", opts.Scheme))
	}

	if opts.Name != "" && opts.ExactNameOnly {
		conds = append(conds, sqlf.Sprintf("name = %s", opts.Name))
	} else if opts.Name != "" {
		conds = append(conds, sqlf.Sprintf("name LIKE ('%%%%' || %s || '%%%%')", opts.Name))
	}

	if !opts.IncludeBlocked {
		conds = append(conds, sqlf.Sprintf("lr.blocked <> true AND prv.blocked <> true"))
	}

	if len(conds) > 0 {
		return sqlf.Sprintf("%s", sqlf.Join(conds, "AND"))
	}

	return sqlf.Sprintf("TRUE")
}

func makeLimit(limit int) *sqlf.Query {
	if limit == 0 {
		return sqlf.Sprintf("LIMIT ALL")
	}
	return sqlf.Sprintf("LIMIT %s", limit)
}

func makeOffset(id int) *sqlf.Query {
	if id > 0 {
		return sqlf.Sprintf("lr.id > %s", id)
	}

	return sqlf.Sprintf("TRUE")
}

// InsertDependencyRepos creates the given dependency repos if they don't yet exist. The values that did not exist previously are returned.
// [{npm, @types/nodejs, [v0.0.1]}, {npm, @types/nodejs, [v0.0.2]}] will be collapsed into [{npm, @types/nodejs, [v0.0.1, v0.0.2]}]
func (s *store) InsertPackageRepoRefs(ctx context.Context, deps []shared.MinimalPackageRepoRef) (newDeps []shared.PackageRepoReference, newVersions []shared.PackageRepoRefVersion, err error) {
	ctx, _, endObservation := s.operations.insertPackageRepoRefs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numInputDeps", len(deps)),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("newDependencies", len(newDeps)),
			log.Int("newVersion", len(newVersions)),
			log.Int("numDedupedDeps", len(deps)),
		}})
	}()

	if len(deps) == 0 {
		return
	}

	slices.SortStableFunc(deps, func(a, b shared.MinimalPackageRepoRef) bool {
		if a.Scheme != b.Scheme {
			return a.Scheme < b.Scheme
		}

		return a.Name < b.Name
	})

	// first reduce
	var lastCommon int
	for i, dep := range deps[1:] {
		if dep.Name == deps[lastCommon].Name && dep.Scheme == deps[lastCommon].Scheme {
			deps[lastCommon].Versions = append(deps[lastCommon].Versions, dep.Versions...)
			deps[i+1] = shared.MinimalPackageRepoRef{}
		} else {
			lastCommon = i + 1
		}
	}

	// then collapse
	nonDupes := deps[:0]
	for _, dep := range deps {
		if dep.Name != "" && dep.Scheme != "" {
			nonDupes = append(nonDupes, dep)
		}
	}
	// replace the originals :wave
	deps = nonDupes

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	for _, tempTableQuery := range []string{temporaryPackageRepoRefsTableQuery, temporaryPackageRepoRefVersionsTableQuery} {
		if err := tx.Exec(ctx, sqlf.Sprintf(tempTableQuery)); err != nil {
			return nil, nil, errors.Wrap(err, "failed to create temporary tables")
		}
	}

	err = batch.WithInserter(
		ctx,
		tx.Handle(),
		"t_package_repo_refs",
		batch.MaxNumPostgresParameters,
		[]string{"scheme", "name"},
		func(inserter *batch.Inserter) error {
			for _, pkg := range deps {
				if err := inserter.Insert(ctx, pkg.Scheme, pkg.Name); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to insert package repos in temporary table")
	}

	newDeps, err = basestore.NewSliceScanner(func(rows dbutil.Scanner) (dep shared.PackageRepoReference, err error) {
		err = rows.Scan(&dep.ID, &dep.Scheme, &dep.Name)
		return
	})(tx.Query(ctx, sqlf.Sprintf(transferPackageRepoRefsQuery)))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to transfer package repos from temporary table")
	}

	// we need the IDs of all newly inserted and already existing package repo references
	// for all of the references in `deps`, so that we have the package repo reference ID that
	// we need for the package repo reference versions table.
	// We already have the IDs of newly inserted ones (in `newDeps`), but for simplicity we'll
	// just search based on (scheme, name) tuple in `deps`.

	// we slice into `deps`, which will continuously shrink as we batch based on the amount of
	// postgres parameters we can fit. Divide by 2 because for each entry in the batch, we need 2 free params
	const maxBatchSize = batch.MaxNumPostgresParameters / 2
	remainingDeps := deps

	allIDs := make([]int, 0, len(deps))

	for len(remainingDeps) > 0 {
		// avoid slice out of bounds nonsense
		var batch []shared.MinimalPackageRepoRef
		if len(remainingDeps) <= maxBatchSize {
			batch, remainingDeps = remainingDeps, nil
		} else {
			batch, remainingDeps = remainingDeps[:maxBatchSize], remainingDeps[maxBatchSize:]
		}

		// dont over-allocate
		max := maxBatchSize
		if len(remainingDeps) < maxBatchSize {
			max = len(remainingDeps)
		}
		params := make([]*sqlf.Query, 0, max)
		for _, dep := range batch {
			params = append(params, sqlf.Sprintf("(%s, %s)", dep.Scheme, dep.Name))
		}

		query := sqlf.Sprintf(
			getAttemptedInsertDependencyReposQuery,
			sqlf.Join(params, ", "),
		)

		allIDsWindow, err := basestore.ScanInts(tx.Query(ctx, query))
		if err != nil {
			return nil, nil, err
		}
		allIDs = append(allIDs, allIDsWindow...)
	}

	err = batch.WithInserter(
		ctx,
		tx.Handle(),
		"t_package_repo_versions",
		batch.MaxNumPostgresParameters,
		[]string{"package_id", "version"},
		func(inserter *batch.Inserter) error {
			for i, dep := range deps {
				for _, version := range dep.Versions {
					if err := inserter.Insert(ctx, allIDs[i], version); err != nil {
						return err
					}
				}
			}
			return nil
		})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to insert package repo versions in temporary table")
	}

	newVersions, err = basestore.NewSliceScanner(func(rows dbutil.Scanner) (version shared.PackageRepoRefVersion, err error) {
		err = rows.Scan(&version.ID, &version.PackageRefID, &version.Version)
		return
	})(tx.Query(ctx, sqlf.Sprintf(transferPackageRepoRefVersionsQuery)))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to transfer package repos from temporary table")
	}

	return newDeps, newVersions, err
}

const temporaryPackageRepoRefsTableQuery = `
CREATE TEMPORARY TABLE t_package_repo_refs (
	scheme TEXT NOT NULL,
	name TEXT NOT NULL
) ON COMMIT DROP
`

const temporaryPackageRepoRefVersionsTableQuery = `
CREATE TEMPORARY TABLE t_package_repo_versions (
	package_id BIGINT NOT NULL,
	version TEXT NOT NULL
) ON COMMIT DROP
`

const transferPackageRepoRefsQuery = `
INSERT INTO lsif_dependency_repos (scheme, name)
SELECT scheme, name
FROM t_package_repo_refs t
WHERE NOT EXISTS (
	SELECT scheme, name
	FROM lsif_dependency_repos
	WHERE scheme = t.scheme AND
	name = t.name
)
ORDER BY name
RETURNING id, scheme, name
`

const transferPackageRepoRefVersionsQuery = `
INSERT INTO package_repo_versions (package_id, version)
-- we dont reduce package repo versions,
-- so DISTINCT here to avoid conflict
SELECT DISTINCT package_id, version
FROM t_package_repo_versions t
WHERE NOT EXISTS (
	SELECT package_id, version
	FROM package_repo_versions
	WHERE package_id = t.package_id AND
	version = t.version
)
-- unit tests rely on a certain order
ORDER BY package_id, version
RETURNING id, package_id, version
`

const getAttemptedInsertDependencyReposQuery = `
SELECT id FROM lsif_dependency_repos
WHERE (scheme, name) IN (VALUES %s)
ORDER BY (scheme, name)
`

// DeleteDependencyReposByID removes the dependency repos with the given ids, if they exist.
func (s *store) DeletePackageRepoRefsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoRefsByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numIDs", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil
	}

	return s.db.Exec(ctx, sqlf.Sprintf(deleteDependencyReposByIDQuery, pq.Array(ids)))
}

const deleteDependencyReposByIDQuery = `
DELETE FROM lsif_dependency_repos
WHERE id = ANY(%s)
`

func (s *store) DeletePackageRepoRefVersionsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoRefVersionsByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numIDs", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil
	}

	return s.db.Exec(ctx, sqlf.Sprintf(deleteDependencyRepoVersionsByID, pq.Array(ids)))
}

const deleteDependencyRepoVersionsByID = `
DELETE FROM package_repo_versions
WHERE id = ANY(%s)
`

type ListPackageRepoRefFiltersOpts struct {
	IDs           []int
	PackageScheme string
	After         int
	Limit         int
}

func (s *store) ListPackageRepoRefFilters(ctx context.Context, opts ListPackageRepoRefFiltersOpts) (_ []shared.PackageFilter, err error) {
	ctx, _, endObservation := s.operations.listPackageRepoFilters.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numPackageRepoFilterIDs", len(opts.IDs)),
		log.String("packageScheme", opts.PackageScheme),
		log.Int("after", opts.After),
		log.Int("limit", opts.Limit),
	}})
	defer endObservation(1, observation.Args{})

	conds := make([]*sqlf.Query, 0, 4)
	// we never want to surface filters, and the table should be small enough that a
	// handful of old entries doesnt matter.
	conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))

	if len(opts.IDs) != 0 {
		conds = append(conds, sqlf.Sprintf("id = ANY(%s)", pq.Array(opts.IDs)))
	}

	if opts.PackageScheme != "" {
		conds = append(conds, sqlf.Sprintf("scheme = %s", opts.PackageScheme))
	}

	if opts.After != 0 {
		conds = append(conds, sqlf.Sprintf("id > %s", opts.After))
	}

	limit := sqlf.Sprintf("")
	if opts.Limit != 0 {
		limit = sqlf.Sprintf("LIMIT %s", opts.Limit)
	}

	filters, err := basestore.NewSliceScanner(scanPackageFilter)(
		s.db.Query(ctx, sqlf.Sprintf(
			listPackageRepoRefFiltersQuery,
			sqlf.Join(conds, "AND"),
			limit,
		)),
	)

	return filters, err
}

const listPackageRepoRefFiltersQuery = `
SELECT id, behaviour, scheme, matcher, deleted_at, updated_at
FROM package_repo_filters
-- filter
%s
-- limit
%s
`

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *store) CreatePackageRepoFilter(ctx context.Context, filter shared.MinimalPackageFilter) (err error) {
	ctx, _, endObservation := s.operations.createPackageRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", filter.PackageScheme),
		log.String("behaviour", deref(filter.Behaviour)),
		log.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		log.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})

	var matcherJSON driver.Value
	if filter.NameFilter != nil {
		matcherJSON, err = json.Marshal(filter.NameFilter)
		err = errors.Wrapf(err, "error marshalling %+v", filter.NameFilter)
	} else if filter.VersionFilter != nil {
		matcherJSON, err = json.Marshal(filter.VersionFilter)
		err = errors.Wrapf(err, "error marshalling %+v", filter.VersionFilter)
	}
	if err != nil {
		return err
	}

	result, err := s.db.ExecResult(ctx, sqlf.Sprintf(createPackageRepoFilter, filter.Behaviour, filter.PackageScheme, matcherJSON))
	if err != nil {
		return errors.Wrap(err, "error inserting package repo filter")
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return errors.New("package repo filter already exists")
	}
	return nil
}

const createPackageRepoFilter = `
INSERT INTO package_repo_filters (behaviour, scheme, matcher)
VALUES (%s, %s, %s)
ON CONFLICT DO NOTHING
`

func (s *store) UpdatePackageRepoFilter(ctx context.Context, filter shared.PackageFilter) (err error) {
	ctx, _, endObservation := s.operations.updatePackageRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", filter.ID),
		log.String("packageScheme", filter.PackageScheme),
		log.String("behaviour", filter.Behaviour),
		log.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		log.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})

	var matcherJSON driver.Value
	if filter.NameFilter != nil {
		matcherJSON, err = json.Marshal(filter.NameFilter)
		err = errors.Wrapf(err, "error marshalling %+v", filter.NameFilter)
	} else if filter.VersionFilter != nil {
		matcherJSON, err = json.Marshal(filter.VersionFilter)
		err = errors.Wrapf(err, "error marshalling %+v", filter.VersionFilter)
	}
	if err != nil {
		return err
	}

	result, err := s.db.ExecResult(ctx, sqlf.Sprintf(updatePackageRepoFilterQuery, filter.Behaviour, filter.PackageScheme, matcherJSON, filter.ID))
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return errors.Newf("no package repo filters for ID %d", filter.ID)
	}
	return nil
}

const updatePackageRepoFilterQuery = `
UPDATE package_repo_filters
SET
	behaviour = %s,
	scheme = %s,
	matcher = %s
WHERE id = %s`

func (s *store) DeletePacakgeRepoFilter(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	result, err := s.db.ExecResult(ctx, sqlf.Sprintf(deletePackagRepoFilterQuery, id))
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return errors.Newf("no package repo filters for ID %d", id)
	}
	return nil
}

const deletePackagRepoFilterQuery = `
UPDATE package_repo_filters
SET deleted_at = now()
WHERE id = %s`

func (s *store) ExistsPackageRepoRefLastCheckedBefore(ctx context.Context, before time.Time) (exists bool, err error) {
	ctx, _, endObservation := s.operations.existsPackageRepoRefLastCheckedBefore.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("beforeTime", string(pq.FormatTimestamp(before))),
	}})
	defer endObservation(1, observation.Args{})

	_, exists, err = basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(existsPackageRepoRefLastCheckedBeforeQuery, before, before)))
	return
}

const existsPackageRepoRefLastCheckedBeforeQuery = `
SELECT 1
WHERE EXISTS (
	SELECT 1
	FROM lsif_dependency_repos
	WHERE
		last_checked_at IS NULL
		OR last_checked_at < %s
	LIMIT 1
) OR EXISTS (
	SELECT 1
	FROM package_repo_versions
	WHERE
		last_checked_at IS NULL
		OR last_checked_at < %s
	LIMIT 1
)
`

func (s *store) ApplyPackageFilters(ctx context.Context) (pkgsAffected, versionsAffected int, err error) {
	ctx, _, endObservation := s.operations.applyPackageFilters.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return pkgsAffected, versionsAffected, basestore.NewCallbackScanner(func(s dbutil.Scanner) (bool, error) {
		return false, s.Scan(&pkgsAffected, &versionsAffected)
	})(s.db.Query(ctx, sqlf.Sprintf(applyAllFiltersToAllPackageRepoRefsQuery)))
}

const applyAllFiltersToAllPackageRepoRefsQuery = `
WITH
apply_unversioned_package_filters AS (
	UPDATE lsif_dependency_repos lr1
	SET
		blocked = NOT(is_unversioned_package_allowed(lr1.name, lr1.scheme)),
		last_checked_at = statement_timestamp()
	FROM lsif_dependency_repos lr2
	WHERE lr1.id = lr2.id
	RETURNING lr1.blocked <> lr2.blocked AS changed
),
apply_versioned_package_filters AS (
	UPDATE package_repo_versions prv1
	SET
		blocked = NOT(is_versioned_package_allowed(lr.name, prv1.version, lr.scheme)),
		last_checked_at = statement_timestamp()
	FROM package_repo_versions prv2
	JOIN lsif_dependency_repos lr
	ON prv2.package_id = lr.id
	WHERE prv1.package_id = lr.id
	RETURNING prv1.blocked <> prv2.blocked AS changed
)
SELECT (
	SELECT COUNT(*) FILTER (WHERE changed)
	FROM apply_unversioned_package_filters
) AS packages_changed, (
	SELECT COUNT(*) FILTER (WHERE changed)
	FROM apply_versioned_package_filters
) AS versions_changed
`

func (s *store) IsPackageRepoVersionAllowed(ctx context.Context, scheme string, pkg reposource.PackageName, version string) (allowed bool, err error) {
	ctx, _, endObservation := s.operations.isPackageRepoVersionAllowed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", scheme),
		log.String("name", string(pkg)),
		log.String("version", version),
	}})
	defer endObservation(1, observation.Args{})

	allowed, _, err = basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf("SELECT is_versioned_package_allowed(%s, %s, %s)", pkg, version, scheme)))
	return
}

func (s *store) IsPackageRepoAllowed(ctx context.Context, scheme string, pkg reposource.PackageName) (allowed bool, err error) {
	ctx, _, endObservation := s.operations.isPackageRepoAllowed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", scheme),
		log.String("name", string(pkg)),
	}})
	defer endObservation(1, observation.Args{})

	allowed, _, err = basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf("SELECT is_unversioned_package_allowed(%s, %s)", pkg, scheme)))
	return
}

func (s *store) PackagesOrVersionsMatchingFilter(ctx context.Context, filter shared.MinimalPackageFilter, limit, after int) (_ []shared.PackageRepoReference, _ int, err error) {
	ctx, _, endObservation := s.operations.pkgsOrVersionsMatchingFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", filter.PackageScheme),
		log.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		log.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})

	offsetExpr := sqlf.Sprintf("TRUE")
	if after > 0 {
		offsetExpr = sqlf.Sprintf("lr.id > %s", after)
	}

	limitExpr := sqlf.Sprintf("ALL")
	if limit > 0 {
		limitExpr = sqlf.Sprintf("%s", limit)
	}

	// name filter case
	if filter.NameFilter != nil {
		matcherJSON, err := json.Marshal(filter.NameFilter)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "error marshalling %+v", filter.NameFilter)
		}

		pkgs, err := basestore.NewSliceScanner(scanDependencyRepo)(s.db.Query(ctx, sqlf.Sprintf(
			packagesMatchingFilterQuery,
			sqlf.Sprintf(packageReposColumns),
			driver.Value(matcherJSON),
			filter.PackageScheme,
			offsetExpr,
			limitExpr,
		)))
		if err != nil {
			return nil, 0, err
		}

		count, _, err := basestore.ScanFirstInt(
			s.db.Query(ctx, sqlf.Sprintf(
				packagesMatchingFilterQuery,
				sqlf.Sprintf("COUNT(*)"),
				driver.Value(matcherJSON),
				filter.PackageScheme,
				sqlf.Sprintf("TRUE"),
				sqlf.Sprintf("ALL"),
			)),
		)

		return pkgs, count, err
	}

	// version filter case
	matcherJSON, err := json.Marshal(filter.VersionFilter)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "error marshalling %+v", filter.VersionFilter)
	}

	pkgs, err := basestore.NewSliceScanner(scanDependencyRepoWithVersions)(s.db.Query(ctx, sqlf.Sprintf(
		packageVersionsMatchingFilterQuery,
		sqlf.Sprintf(groupedVersionedPackageReposColumns),
		driver.Value(matcherJSON),
		filter.PackageScheme,
		offsetExpr,
		limitExpr,
	)))
	if err != nil {
		return nil, 0, err
	}

	count, _, err := basestore.ScanFirstInt(
		s.db.Query(ctx, sqlf.Sprintf(
			packageVersionsMatchingFilterQuery,
			sqlf.Sprintf("COUNT(*)"),
			driver.Value(matcherJSON),
			filter.PackageScheme,
			sqlf.Sprintf("TRUE"),
			sqlf.Sprintf("ALL"),
		)),
	)
	return pkgs, count, err
}

const packageVersionsMatchingFilterQuery = `
WITH parsed_matcher AS (
	SELECT
		*,
		CASE
			WHEN matcher ? 'VersionGlob' THEN glob_to_regex(matcher->>'VersionGlob')
			WHEN matcher ? 'PackageGlob' THEN glob_to_regex(matcher->>'PackageGlob')
		END AS regex
	FROM (
		SELECT %s::jsonb AS matcher
	) AS z
)
SELECT %s
FROM lsif_dependency_repos lr
JOIN package_repo_versions prv
ON lr.id = prv.package_id
WHERE lr.scheme = %s
AND
(
	(
		parsed_matcher.matcher ? 'PackageGlob'
		AND package ~ parsed_matcher.regex
	) OR (
		parsed_matcher.matcher->>'PackageName' = package
		AND version ~ parsed_matcher.regex
	)
)
AND %s
LIMIT %s
`

const packagesMatchingFilterQuery = `
WITH parsed_matcher AS (
	SELECT
		*,
		CASE
			WHEN matcher ? 'VersionGlob' THEN glob_to_regex(matcher->>'VersionGlob')
			WHEN matcher ? 'PackageGlob' THEN glob_to_regex(matcher->>'PackageGlob')
		END AS regex
	FROM (
		SELECT %s::jsonb AS matcher
	) AS z
)
SELECT %s
FROM lsif_dependency_repos lr
WHERE lr.scheme = %s
AND
(
	(
		parsed_matcher.matcher ? 'PackageGlob'
		AND package ~ parsed_matcher.regex
	) OR (
		parsed_matcher.matcher->>'PackageName' = package
		AND parsed_matcher.matcher->>'VersionGlob' = '*'
	)
)
AND %s
LIMIT %s
`
