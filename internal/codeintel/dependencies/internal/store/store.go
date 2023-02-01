package store

import (
	"context"

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
	ListPackageRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shared.PackageRepoReference, total int, err error)
	InsertPackageRepoRefs(ctx context.Context, deps []shared.MinimalPackageRepoRef) (newDeps []shared.PackageRepoReference, newVersions []shared.PackageRepoRefVersion, err error)
	DeletePackageRepoRefsByID(ctx context.Context, ids ...int) (err error)
	DeletePackageRepoRefVersionsByID(ctx context.Context, ids ...int) (err error)
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

// ListDependencyReposOpts are options for listing dependency repositories.
type ListDependencyReposOpts struct {
	Scheme              string
	Name                reposource.PackageName
	ExactNameOnly       bool
	After               int
	Limit               int
	MostRecentlyUpdated bool
}

// ListDependencyRepos returns dependency repositories to be synced by gitserver.
func (s *store) ListPackageRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shared.PackageRepoReference, total int, err error) {
	ctx, _, endObservation := s.operations.listDependencyRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("scheme", opts.Scheme),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRepos", len(dependencyRepos)),
		}})
	}()

	sortExpr := "ORDER BY lr.id ASC"
	if opts.MostRecentlyUpdated {
		sortExpr = "ORDER BY prv.id DESC"
	}

	selectColumns := sqlf.Sprintf("lr.id, lr.scheme, lr.name, prv.id, prv.package_id, prv.version")

	depReposMap := basestore.NewOrderedMap[int, shared.PackageRepoReference]()
	scanner := basestore.NewKeyedCollectionScanner[*basestore.OrderedMap[int, shared.PackageRepoReference], int, shared.PackageRepoReference, shared.PackageRepoReference](depReposMap, func(s dbutil.Scanner) (int, shared.PackageRepoReference, error) {
		dep, err := scanDependencyRepo(s)
		return dep.ID, dep, err
	}, dependencyVersionsReducer{})

	query := sqlf.Sprintf(
		listDependencyReposQuery,
		selectColumns,
		sqlf.Join(makeListDependencyReposConds(opts), "AND"),
		makeLimit(opts.Limit),
		sqlf.Sprintf(sortExpr),
	)
	err = scanner(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, errors.Wrap(err, "error listing dependency repos")
	}

	query = sqlf.Sprintf(
		listDependencyReposQuery,
		sqlf.Sprintf("COUNT(lr.id)"),
		sqlf.Join(makeListDependencyReposConds(opts), "AND"),
		sqlf.Sprintf(""), sqlf.Sprintf(""),
	)
	totalCount, _, err := basestore.ScanFirstInt(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, errors.Wrap(err, "error counting dependency repos")
	}

	dependencyRepos = depReposMap.Values()
	return dependencyRepos, totalCount, err
}

type dependencyVersionsReducer struct{}

func (dependencyVersionsReducer) Create() shared.PackageRepoReference {
	return shared.PackageRepoReference{}
}

func (dependencyVersionsReducer) Reduce(collection shared.PackageRepoReference, value shared.PackageRepoReference) shared.PackageRepoReference {
	value.Versions = append(collection.Versions, value.Versions...)
	collection, value = value, collection
	return collection
}

const listDependencyReposQuery = `
SELECT %s
FROM (
	SELECT id, scheme, name
	FROM lsif_dependency_repos
	WHERE %s
	%s
) lr
JOIN package_repo_versions prv
ON lr.id = prv.package_id
%s
`

func makeListDependencyReposConds(opts ListDependencyReposOpts) []*sqlf.Query {
	conds := make([]*sqlf.Query, 0, 3)
	conds = append(conds, sqlf.Sprintf("scheme = %s", opts.Scheme))

	if opts.Name != "" && opts.ExactNameOnly {
		conds = append(conds, sqlf.Sprintf("name = %s", opts.Name))
	} else if opts.Name != "" {
		conds = append(conds, sqlf.Sprintf("name LIKE ('%%%%' || %s || '%%%%')", opts.Name))
	}

	switch {
	case opts.MostRecentlyUpdated && opts.After > 0:
		conds = append(conds, sqlf.Sprintf("id < %s", opts.After))
	case !opts.MostRecentlyUpdated && opts.After > 0:
		conds = append(conds, sqlf.Sprintf("id > %s", opts.After))
	}

	return conds
}

func makeLimit(limit int) *sqlf.Query {
	if limit == 0 {
		return sqlf.Sprintf("")
	}

	return sqlf.Sprintf("LIMIT %s", limit)
}

// InsertDependencyRepos creates the given dependency repos if they don't yet exist. The values that did not exist previously are returned.
// [{npm, @types/nodejs, [v0.0.1]}, {npm, @types/nodejs, [v0.0.2]}] will be collapsed into [{npm, @types/nodejs, [v0.0.1, v0.0.2]}]
func (s *store) InsertPackageRepoRefs(ctx context.Context, deps []shared.MinimalPackageRepoRef) (newDeps []shared.PackageRepoReference, newVersions []shared.PackageRepoRefVersion, err error) {
	ctx, _, endObservation := s.operations.upsertDependencyRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numInputDeps", len(deps)),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("newDependencies", len(newDeps)),
			log.Int("newVersion", len(newVersions)),
			log.Int("numDedupedDeps", len(deps)),
		}})
	}()

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

	db, err := s.db.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		err = db.Done(err)
	}()

	newDeps = make([]shared.PackageRepoReference, 0, len(deps))
	dependencyScanner := func(rows dbutil.Scanner) error {
		var dep shared.PackageRepoReference
		if err := rows.Scan(&dep.ID, &dep.Scheme, &dep.Name); err != nil {
			return err
		}
		newDeps = append(newDeps, dep)
		return nil
	}

	err = batch.WithInserterWithReturn(
		ctx,
		db.Handle(),
		"lsif_dependency_repos",
		batch.MaxNumPostgresParameters,
		[]string{"scheme", "name", "version"},
		"ON CONFLICT DO NOTHING",
		[]string{"id", "scheme", "name"},
		dependencyScanner,
		func(inserter *batch.Inserter) error {
			for _, pkg := range deps {
				// temporary sentinel value so ON CONFLICT still works
				if err := inserter.Insert(ctx, pkg.Scheme, pkg.Name, "👁️ temporary_sentintel_value 👁️"); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to insert package repos")
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

		allIDsWindow, err := basestore.ScanInts(db.Query(ctx, query))
		if err != nil {
			return nil, nil, err
		}
		allIDs = append(allIDs, allIDsWindow...)
	}

	// rough estimate of 1 version per dependency
	newVersions = make([]shared.PackageRepoRefVersion, 0, len(deps))
	versionScanner := func(rows dbutil.Scanner) error {
		var version shared.PackageRepoRefVersion
		if err := rows.Scan(&version.ID, &version.PackageRefID, &version.Version); err != nil {
			return err
		}
		newVersions = append(newVersions, version)
		return nil
	}

	err = batch.WithInserterWithReturn(
		ctx,
		db.Handle(),
		"package_repo_versions",
		batch.MaxNumPostgresParameters,
		[]string{"package_id", "version"},
		"ON CONFLICT DO NOTHING",
		[]string{"id", "package_id", "version"},
		versionScanner,
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
		return nil, nil, errors.Wrapf(err, "failed to insert package repo versions")
	}

	return newDeps, newVersions, err
}

const getAttemptedInsertDependencyReposQuery = `
SELECT id FROM lsif_dependency_repos
WHERE (scheme, name) IN (VALUES %s)
ORDER BY scheme, name
`

// DeleteDependencyReposByID removes the dependency repos with the given ids, if they exist.
func (s *store) DeletePackageRepoRefsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deleteDependencyReposByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	if len(ids) == 0 {
		return nil
	}

	return s.db.Exec(ctx, sqlf.Sprintf(deleteDependencyRepoVersionsByID, pq.Array(ids)))
}

const deleteDependencyRepoVersionsByID = `
DELETE FROM package_repo_versions
WHERE id = ANY(%s)
`
