package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for package dependencies storage.
type Store interface {
	PreciseDependencies(ctx context.Context, repoName, commit string) (deps map[api.RepoName]types.RevSpecSet, err error)
	PreciseDependents(ctx context.Context, repoName, commit string) (deps map[api.RepoName]types.RevSpecSet, err error)
	LockfileDependencies(ctx context.Context, opts LockfileDependenciesOpts) (deps []shared.PackageDependency, found bool, err error)
	UpsertLockfileGraph(ctx context.Context, repoName, commit, lockfile string, deps []shared.PackageDependency, graph shared.DependencyGraph) (err error)
	SelectRepoRevisionsToResolve(ctx context.Context, batchSize int, minimumCheckInterval time.Duration) (_ map[string][]string, err error)
	UpdateResolvedRevisions(ctx context.Context, repoRevsToResolvedRevs map[string]map[string]string) (err error)
	LockfileDependents(ctx context.Context, repoName, commit string) (deps []api.RepoCommit, err error)
	ListDependencyRepos(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shared.Repo, err error)
	UpsertDependencyRepos(ctx context.Context, deps []shared.Repo) (newDeps []shared.Repo, err error)
	DeleteDependencyReposByID(ctx context.Context, ids ...int) (err error)
	ListLockfileIndexes(ctx context.Context, opts ListLockfileIndexesOpts) (indexes []shared.LockfileIndex, totalCount int, err error)
	GetLockfileIndex(ctx context.Context, opts GetLockfileIndexOpts) (index shared.LockfileIndex, err error)
	DeleteLockfileIndexByID(ctx context.Context, id int) (err error)
}

// store manages the database tables for package dependencies.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new store.
func New(db database.DB, op *observation.Context) *store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(op),
	}
}

// PreciseDependencies returns package dependencies from precise indexes. It is assumed that
// the given commit is the canonical 40-character hash.
func (s *store) PreciseDependencies(ctx context.Context, repoName, commit string) (deps map[api.RepoName]types.RevSpecSet, err error) {
	ctx, _, endObservation := s.operations.preciseDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", repoName),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	return scanRepoRevSpecSets(s.db.Query(ctx, sqlf.Sprintf(preciseDependenciesQuery, repoName, commit)))
}

const preciseDependenciesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:PreciseDependencies
SELECT pr.name, pu.commit
FROM lsif_packages lp
JOIN lsif_uploads pu ON pu.id = lp.dump_id
JOIN repo pr ON pr.id = pu.repository_id
JOIN lsif_references lr ON lr.scheme = lp.scheme AND lr.name = lp.name AND lr.version = lp.version
JOIN lsif_uploads ru ON ru.id = lr.dump_id
JOIN repo rr ON rr.id = ru.repository_id
WHERE rr.name = %s AND ru.commit = %s
`

// PreciseDependents returns package dependents from precise indexes. It is assumed that
// the given commit is the canonical 40-character hash.
func (s *store) PreciseDependents(ctx context.Context, repoName, commit string) (deps map[api.RepoName]types.RevSpecSet, err error) {
	ctx, _, endObservation := s.operations.preciseDependents.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", repoName),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	return scanRepoRevSpecSets(s.db.Query(ctx, sqlf.Sprintf(preciseDependentsQuery, repoName, commit)))
}

const preciseDependentsQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:PreciseDependents
SELECT rr.name, ru.commit
FROM lsif_packages lp
JOIN lsif_uploads pu ON pu.id = lp.dump_id
JOIN repo pr ON pr.id = pu.repository_id
JOIN lsif_references lr ON lr.scheme = lp.scheme AND lr.name = lp.name AND lr.version = lp.version
JOIN lsif_uploads ru ON ru.id = lr.dump_id
JOIN repo rr ON rr.id = ru.repository_id
WHERE pr.name = %s AND pu.commit = %s
`

type LockfileDependenciesOpts struct {
	RepoName string
	Commit   string

	// IncludeTransitive determines whether transitive dependencies are included in the result.
	// NOTE: if a lockfile doesn't allow us to distinguish between
	// transitive/direct all of the dependencies are persisted as direct
	// dependencies.
	IncludeTransitive bool

	// Lockfile, if specified, causes only dependencies from that lockfile to
	// be included.
	Lockfile string
}

// LockfileDependencies returns package dependencies from a previous lockfiles result for
// the given repository and commit. It is assumed that the given commit is the canonical
// 40-character hash.
func (s *store) LockfileDependencies(ctx context.Context, opts LockfileDependenciesOpts) (deps []shared.PackageDependency, found bool, err error) {
	ctx, _, endObservation := s.operations.lockfileDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", opts.RepoName),
		log.String("commit", opts.Commit),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Bool("found", found),
			log.Int("numDeps", len(deps)),
		}})
	}()

	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() { err = tx.db.Done(err) }()

	deps, err = scanPackageDependencies(tx.db.Query(ctx, lockfileDependenciesQuery(opts)))
	if err != nil {
		return nil, false, err
	}
	if len(deps) == 0 {
		// No dependencies were found, but we could have already written a record
		// that just had an empty references list. Check to see if this is the case
		// so we don't attempt to re-parse the lockfiles of this repo/commit from the
		// dependencies service.
		_, found, err = basestore.ScanFirstInt(tx.db.Query(ctx, sqlf.Sprintf(
			lockfileDependenciesExistsQuery,
			opts.RepoName,
			dbutil.CommitBytea(opts.Commit),
		)))

		return nil, found, err
	}

	return deps, true, nil
}

func lockfileDependenciesQuery(opts LockfileDependenciesOpts) *sqlf.Query {
	maxDependencyLevel := 0
	if opts.IncludeTransitive {
		// TODO: We should improve SQL here to falsify instead of using this limit
		maxDependencyLevel = 9999
	}

	// predicates to find the row in codeintel_lockfiles from which we get the dependencies
	lockfilesPreds := []*sqlf.Query{
		sqlf.Sprintf("repository_id = (SELECT id FROM repo WHERE name = %s)", opts.RepoName),
		sqlf.Sprintf("commit_bytea = %s", dbutil.CommitBytea(opts.Commit)),
	}

	if opts.Lockfile != "" {
		lockfilesPreds = append(lockfilesPreds, sqlf.Sprintf("lockfile = %s", opts.Lockfile))
	}

	return sqlf.Sprintf(
		lockfileDependenciesQueryFmtStr,
		maxDependencyLevel,
		sqlf.Join(lockfilesPreds, "\n AND "),
	)
}

const lockfileDependenciesQueryFmtStr = `
-- source: internal/codeintel/dependencies/internal/store/store.go:LockfileDependencies
WITH RECURSIVE dependencies(id, resolution_repository_id, resolution_commit_bytea, resolution_lockfile, depends_on, level, max_level) AS (
  SELECT
    id, resolution_repository_id, resolution_commit_bytea, resolution_lockfile, depends_on, 0 AS level, %s::int AS max_level
  FROM
    codeintel_lockfile_references
  WHERE
    id IN (
      SELECT
        unnest(codeintel_lockfile_reference_ids)
      FROM
        codeintel_lockfiles
      WHERE
	    %s -- lockfilePreds
    )

  UNION ALL

  SELECT
    lr.id, lr.resolution_repository_id, lr.resolution_commit_bytea, lr.resolution_lockfile, lr.depends_on, (d.level+1) AS level, d.max_level
  FROM
    codeintel_lockfile_references lr
  JOIN dependencies d ON (
	  lr.id = ANY (d.depends_on) AND
	  lr.resolution_repository_id = d.resolution_repository_id AND
	  lr.resolution_commit_bytea = d.resolution_commit_bytea AND
	  lr.resolution_lockfile = d.resolution_lockfile
  )
  WHERE
    level < d.max_level
)
SELECT
  -- We could also select dependencies.level here
  lr.repository_name,
  lr.revspec,
  lr.package_scheme,
  lr.package_name,
  lr.package_version
FROM
  dependencies, codeintel_lockfile_references lr
WHERE
  dependencies.id = lr.id
ORDER BY lr.package_name
`

const lockfileDependenciesExistsQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:LockfileDependencies
SELECT 1
FROM codeintel_lockfiles
WHERE repository_id = (SELECT id FROM repo WHERE name = %s) AND commit_bytea = %s
`

// populatePackageDependencyChannel populates a channel with the given dependencies for bulk insertion.
func populatePackageDependencyChannel(deps []shared.PackageDependency, lockfile, commit string) <-chan []any {
	ch := make(chan []any, len(deps))

	go func() {
		defer close(ch)

		for _, dep := range deps {
			ch <- []any{
				dep.RepoName(),
				dep.GitTagFromVersion(),
				dep.Scheme(),
				dep.PackageSyntax(),
				dep.PackageVersion(),
				pq.Array([]int{}),
				lockfile,
				dbutil.CommitBytea(commit),
			}
		}
	}()

	return ch
}

// UpsertLockfileGraph insert the given `deps` as `codeintel_lockfile_references`
// and creates an entry in `codeintel_lockfiles` with the given `repoName`,
// `commit`, `lockfile` that references the inserted `deps`.
//
// If `graph` is not nil, only the direct dependencies are referenced in the
// `codeintel_lockfiles` entry and the full graph is represented in
// `codeintel_lockfile_references` as edges in the `depends_on` column.
func (s *store) UpsertLockfileGraph(ctx context.Context, repoName, commit, lockfile string, deps []shared.PackageDependency, graph shared.DependencyGraph) (err error) {
	return s.upsertLockfileGraphAt(ctx, repoName, commit, lockfile, deps, graph, time.Now())
}

func (s *store) upsertLockfileGraphAt(ctx context.Context, repoName, commit, lockfile string, deps []shared.PackageDependency, graph shared.DependencyGraph, now time.Time) (err error) {
	ctx, _, endObservation := s.operations.upsertLockfileGraph.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", repoName),
		log.String("commit", commit),
		log.String("lockfile", lockfile),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.db.Done(err) }()

	if err := tx.db.Exec(ctx, sqlf.Sprintf(temporaryLockfileReferencesTableQuery)); err != nil {
		return err
	}

	//
	// Step 1: Insert all packages into codeintel_lockfile_references table,
	//         return their names and IDs.
	//
	if err := batch.InsertValues(
		ctx,
		tx.db.Handle(),
		"t_codeintel_lockfile_references",
		batch.MaxNumPostgresParameters,
		[]string{
			"repository_name",
			"revspec",
			"package_scheme",
			"package_name",
			"package_version",
			"depends_on",
			"resolution_lockfile",
			"resolution_commit_bytea",
			// resolution_repository_id missing because we don't insert that
			// into the temp table and instead do a sub-select to get the repo
			// ID.
		},
		populatePackageDependencyChannel(deps, lockfile, commit),
	); err != nil {
		return err
	}

	// Get IDs and name->ID mapping for upserted packages
	nameIDs, ids, err := scanIdNames(tx.db.Query(ctx, sqlf.Sprintf(upsertLockfileReferencesQuery, repoName, repoName)))
	if err != nil {
		return err
	}

	// If we don't have a graph, we insert all of the dependencies as direct
	// dependencies and return.
	if graph == nil {
		idsArray := pq.Array(ids)
		return tx.db.Exec(ctx, sqlf.Sprintf(
			insertLockfilesQuery,
			dbutil.CommitBytea(commit),
			idsArray,
			lockfile,
			shared.IndexFidelityFlat,
			now,
			now,
			repoName,
			idsArray,
			now,
		))
	}

	//
	// Step 2: Collect all the dependencies (i.e. A depends on B, C, D;
	//         B depends on E, F) and map them to database IDs.
	//
	dependencies := make(map[int][]int)
	for _, edge := range graph.AllEdges() {
		sourceName, targetName := edge[0].PackageSyntax(), edge[1].PackageSyntax()

		sourceID, ok := nameIDs[sourceName]
		if !ok {
			return errors.Newf("id for source %s not found", sourceName)
		}

		targetID, ok := nameIDs[targetName]
		if !ok {
			return errors.Newf("id for target %s not found", sourceName)
		}

		if ids, ok := dependencies[sourceID]; !ok {
			dependencies[sourceID] = []int{targetID}
		} else {
			dependencies[sourceID] = append(ids, targetID)
		}
	}

	// Insert edges into DB. TODO: We could/should batch this
	for source, targets := range dependencies {
		if err := tx.db.Exec(ctx, sqlf.Sprintf(
			insertLockfilesEdgesQuery,
			pq.Array(targets),
			source,
			repoName,
			dbutil.CommitBytea(commit),
			lockfile,
		)); err != nil {
			return err
		}
	}

	//
	// Step 3: Insert codeintel_lockfile entry, pointing to the rootIDs of the
	//         graph (i.e. direct dependencies)
	//
	var (
		roots, rootsUndeterminable = graph.Roots()
		rootIDs                    = make([]int, len(roots))
	)
	for i, r := range roots {
		name := r.PackageSyntax()
		id, ok := nameIDs[name]
		if !ok {
			return errors.Newf("id for root %s not found", name)
		}
		rootIDs[i] = id
	}

	fidelity := shared.IndexFidelityGraph
	if rootsUndeterminable {
		fidelity = shared.IndexFidelityCircular
	}

	idsArray := pq.Array(rootIDs)
	return tx.db.Exec(ctx, sqlf.Sprintf(
		insertLockfilesQuery,
		dbutil.CommitBytea(commit),
		idsArray,
		lockfile,
		fidelity,
		now,
		now,
		repoName,
		idsArray,
		now,
	))
}

const temporaryLockfileReferencesTableQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpsertLockfileGraph
CREATE TEMPORARY TABLE t_codeintel_lockfile_references (
	repository_name text NOT NULL,
	revspec text NOT NULL,
	package_scheme text NOT NULL,
	package_name text NOT NULL,
	package_version text NOT NULL,
	depends_on integer[] NOT NULL,
	resolution_lockfile text NOT NULL,
	resolution_commit_bytea bytea NOT NULL
) ON COMMIT DROP
`

const upsertLockfileReferencesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpsertLockfileGraph
WITH ins AS (
	INSERT INTO codeintel_lockfile_references (repository_name, revspec, package_scheme, package_name, package_version, depends_on, resolution_lockfile, resolution_repository_id, resolution_commit_bytea)
	SELECT repository_name, revspec, package_scheme, package_name, package_version, depends_on, resolution_lockfile, (SELECT id FROM repo WHERE name = %s), resolution_commit_bytea
	FROM t_codeintel_lockfile_references
	ON CONFLICT DO NOTHING
	RETURNING id, package_name
),
duplicates AS (
	SELECT r.id, r.package_name
	FROM t_codeintel_lockfile_references t
	JOIN codeintel_lockfile_references r
	ON
		r.repository_name = t.repository_name AND
		r.revspec = t.revspec AND
		r.package_scheme = t.package_scheme AND
		r.package_name = t.package_name AND
		r.package_version = t.package_version AND
		r.resolution_lockfile = t.resolution_lockfile AND
		r.resolution_repository_id = (SELECT id FROM repo WHERE name = %s) AND
		r.resolution_commit_bytea = t.resolution_commit_bytea
		-- We ignore depends_on since that is updated in a second query and we can't use it to compare
)
SELECT id, package_name FROM ins UNION
SELECT id, package_name FROM duplicates
ORDER BY id
`

const insertLockfilesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpsertLockfileGraph
INSERT INTO codeintel_lockfiles (
	repository_id,
	commit_bytea,
	codeintel_lockfile_reference_ids,
	lockfile,
	fidelity,
	updated_at,
	created_at
)
SELECT id, %s, %s, %s, %s, %s, %s
FROM repo
WHERE name = %s
-- Last write wins
ON CONFLICT (repository_id, commit_bytea, lockfile) DO UPDATE
SET codeintel_lockfile_reference_ids = %s, updated_at = %s
`

const insertLockfilesEdgesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpsertLockfileGraph
UPDATE codeintel_lockfile_references
SET depends_on = %s
WHERE
	id = %s
AND resolution_repository_id = (SELECT id FROM repo WHERE name = %s)
AND resolution_commit_bytea = %s
AND resolution_lockfile = %s
`

// SelectRepoRevisionsToResolve selects the references lockfile packages to
// possibly resolve them to repositories on the Sourcegraph instance.
func (s *store) SelectRepoRevisionsToResolve(ctx context.Context, batchSize int, minimumCheckInterval time.Duration) (_ map[string][]string, err error) {
	return s.selectRepoRevisionsToResolve(ctx, batchSize, minimumCheckInterval, time.Now())
}

func (s *store) selectRepoRevisionsToResolve(ctx context.Context, batchSize int, minimumCheckInterval time.Duration, now time.Time) (_ map[string][]string, err error) {
	var count int
	ctx, _, endObservation := s.operations.selectRepoRevisionsToResolve.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{
		LogFields: []log.Field{
			log.Int("count", count),
		},
	})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(selectRepoRevisionsToResolveQuery, now, int64(minimumCheckInterval/time.Hour), batchSize, now))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	m := map[string][]string{}
	for rows.Next() {
		var repositoryName, commit string
		if err := rows.Scan(&repositoryName, &commit); err != nil {
			return nil, err
		}

		count++
		m[repositoryName] = append(m[repositoryName], commit)
	}

	return m, nil
}

const selectRepoRevisionsToResolveQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:SelectRepoRevisionsToResolve
WITH candidates AS (
	SELECT
		repository_name,
		revspec
	FROM codeintel_lockfile_references
	WHERE
		last_check_at IS NULL OR
		%s - last_check_at >= (%s * '1 hour'::interval)
	GROUP BY repository_name, revspec
	ORDER BY repository_name, revspec
	-- TODO - select for update to reduce contention
	LIMIT %s
),
updated AS (
	UPDATE codeintel_lockfile_references
	SET last_check_at = %s
	WHERE (repository_name, revspec) IN (SELECT * FROM candidates)
)
SELECT * FROM candidates
`

// UpdateResolvedRevisions updates the lockfile packages that were resolved to
// repositories/revisions pairs on the Sourcegraph instance.
func (s *store) UpdateResolvedRevisions(ctx context.Context, repoRevsToResolvedRevs map[string]map[string]string) (err error) {
	ctx, _, endObservation := s.operations.updateResolvedRevisions.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	for repoName, resolvedRevs := range repoRevsToResolvedRevs {
		for commit, resolvedCommit := range resolvedRevs {
			// TODO - batch these updates
			if err := s.db.Exec(ctx, sqlf.Sprintf(
				updateResolvedRevisionsQuery,
				repoName,
				dbutil.CommitBytea(resolvedCommit),
				repoName,
				commit,
			)); err != nil {
				return err
			}
		}
	}

	return nil
}

const updateResolvedRevisionsQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpdateResolvedRevisions
UPDATE
	codeintel_lockfile_references
SET
	repository_id = (SELECT id FROM repo WHERE name = %s),
	commit_bytea = %s
WHERE
	repository_name = %s AND
	revspec = %s
-- TODO - order before update to reduce contention
`

// LockfileDependents returns the set of repositories that have lockfile results pointing to the
// given repo and commit (related to a particular resolved repo/commit of a lockfile reference).
func (s *store) LockfileDependents(ctx context.Context, repoName, commit string) (deps []api.RepoCommit, err error) {
	ctx, _, endObservation := s.operations.lockfileDependents.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", repoName),
		log.String("commit", commit),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencies", len(deps)),
		}})
	}()

	return scanRepoCommits(s.db.Query(ctx, sqlf.Sprintf(lockfileDependentsQuery, repoName, dbutil.CommitBytea(commit))))
}

// TODO: This only returns direct dependents
const lockfileDependentsQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:LockfileDependents
SELECT r.name, encode(lf.commit_bytea, 'hex') AS commit
FROM codeintel_lockfile_references lr
JOIN codeintel_lockfiles lf ON lf.codeintel_lockfile_reference_ids @> ARRAY [lr.id]
JOIN repo r ON r.id = lf.repository_id
JOIN repo rr ON rr.id = lr.repository_id
WHERE rr.name = %s AND lr.commit_bytea = %s
ORDER BY r.name, lf.commit_bytea
`

// ListDependencyReposOpts are options for listing dependency repositories.
type ListDependencyReposOpts struct {
	Scheme          string
	Name            reposource.PackageName
	After           any
	Limit           int
	NewestFirst     bool
	ExcludeVersions bool
}

// ListDependencyRepos returns dependency repositories to be synced by gitserver.
func (s *store) ListDependencyRepos(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shared.Repo, err error) {
	ctx, _, endObservation := s.operations.listDependencyRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("scheme", opts.Scheme),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRepos", len(dependencyRepos)),
		}})
	}()

	sortExpr := "id ASC"
	switch {
	case opts.NewestFirst && !opts.ExcludeVersions:
		sortExpr = "id DESC"
	case opts.ExcludeVersions:
		sortExpr = "name ASC"
	}

	selectCols := sqlf.Sprintf("id, scheme, name, version")
	if opts.ExcludeVersions {
		// id is likely not stable here, so no one should actually use it. Should we set it to 0?
		selectCols = sqlf.Sprintf("DISTINCT ON(name) id, scheme, name, '' AS version")
	}

	return scanDependencyRepos(s.db.Query(ctx, sqlf.Sprintf(
		listDependencyReposQuery,
		selectCols,
		sqlf.Join(makeListDependencyReposConds(opts), "AND"),
		sqlf.Sprintf(sortExpr),
		makeLimit(opts.Limit),
	)))
}

const listDependencyReposQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:ListDependencyRepos
SELECT %s
FROM lsif_dependency_repos
WHERE %s
ORDER BY %s
%s
`

func makeListDependencyReposConds(opts ListDependencyReposOpts) []*sqlf.Query {
	conds := make([]*sqlf.Query, 0, 3)
	conds = append(conds, sqlf.Sprintf("scheme = %s", opts.Scheme))

	if opts.Name != "" {
		conds = append(conds, sqlf.Sprintf("name = %s", opts.Name))
	}

	switch after := opts.After.(type) {
	case nil:
		break
	case int:
		switch {
		case opts.ExcludeVersions:
			panic("cannot set ExcludeVersions and pass ID-based offset")
		case opts.NewestFirst && after > 0:
			conds = append(conds, sqlf.Sprintf("id < %s", opts.After))
		case !opts.NewestFirst && after > 0:
			conds = append(conds, sqlf.Sprintf("id > %s", opts.After))
		}
	case string, reposource.PackageName:
		switch {
		case opts.NewestFirst:
			panic("cannot set NewestFirst and pass name-based offset")
		case opts.ExcludeVersions && after != "":
			conds = append(conds, sqlf.Sprintf("name > %s", opts.After))
		}
	}

	return conds
}

func makeLimit(limit int) *sqlf.Query {
	if limit == 0 {
		return sqlf.Sprintf("")
	}

	return sqlf.Sprintf("LIMIT %s", limit)
}

// ListLockfileIndexesOpts are options for listing lockfile indexes.
type ListLockfileIndexesOpts struct {
	RepoName string
	Commit   string
	Lockfile string

	After int
	Limit int
}

// ListLockfileIndexes returns lockfile indexes.
func (s *store) ListLockfileIndexes(ctx context.Context, opts ListLockfileIndexesOpts) (indexes []shared.LockfileIndex, totalCount int, err error) {
	ctx, _, endObservation := s.operations.listLockfileIndexes.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", opts.RepoName),
		log.String("commit", opts.Commit),
		log.String("lockfile", opts.Commit),
		log.Int("after", opts.After),
		log.Int("limit", opts.Limit),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numIndexes", len(indexes)),
			log.Int("totalCount", totalCount),
		}})
	}()

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = tx.Done(err) }()

	totalCount, err = basestore.ScanInt(tx.QueryRow(ctx, sqlf.Sprintf(
		countLockfileIndexesQuery,
		sqlf.Join(makeListLockfileIndexesConds(opts, true), "AND"),
	)))
	if err != nil {
		return nil, 0, err
	}

	indexes, err = scanLockfileIndexes(tx.Query(ctx, sqlf.Sprintf(
		listLockfileIndexesQuery,
		sqlf.Join(makeListLockfileIndexesConds(opts, false), "AND"),
		makeLimit(opts.Limit),
	)))
	return indexes, totalCount, err
}

const listLockfileIndexesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:ListLockfileIndexes
SELECT id, repository_id, commit_bytea, codeintel_lockfile_reference_ids, lockfile, fidelity, updated_at, created_at
FROM codeintel_lockfiles
WHERE %s
ORDER BY id ASC
%s
`

const countLockfileIndexesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:CountLockfileIndexes
SELECT COUNT(1)
FROM codeintel_lockfiles
WHERE %s
`

func makeListLockfileIndexesConds(opts ListLockfileIndexesOpts, forCount bool) []*sqlf.Query {
	conds := make([]*sqlf.Query, 0, 2)

	if opts.RepoName != "" {
		conds = append(conds, sqlf.Sprintf("repository_id IN (SELECT id FROM repo WHERE name = %s)", opts.RepoName))
	}

	if opts.Commit != "" {
		conds = append(conds, sqlf.Sprintf("commit_bytea = %s", dbutil.CommitBytea(opts.Commit)))
	}

	if opts.Lockfile != "" {
		conds = append(conds, sqlf.Sprintf("lockfile = %s", opts.Lockfile))
	}

	if opts.After != 0 && !forCount {
		conds = append(conds, sqlf.Sprintf("id > %s", opts.After))
	}

	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	return conds
}

// GetLockfileIndexOpts are options for loading a lockfile index from database.
type GetLockfileIndexOpts struct {
	ID       int
	RepoName string
	Commit   string
	Lockfile string
}

var ErrLockfileIndexNotFound = errors.New("lockfile index not found")

// GetLockfileIndex returns a lockfile index.
func (s *store) GetLockfileIndex(ctx context.Context, opts GetLockfileIndexOpts) (index shared.LockfileIndex, err error) {
	ctx, _, endObservation := s.operations.getLockfileIndex.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", opts.ID),
		log.String("repoName", opts.RepoName),
		log.String("commit", opts.Commit),
		log.String("lockfile", opts.Commit),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", index.ID),
		}})
	}()

	conds, err := makeGetLockfileIndexConds(opts)
	if err != nil {
		return index, err
	}

	index, err = scanLockfileIndex(s.db.QueryRow(ctx, sqlf.Sprintf(
		getLockfileIndexQuery,
		sqlf.Join(conds, "AND"),
	)))
	if errors.Is(err, sql.ErrNoRows) {
		return index, ErrLockfileIndexNotFound
	}
	return index, err
}

const getLockfileIndexQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:GetLockfileIndex
SELECT id, repository_id, commit_bytea, codeintel_lockfile_reference_ids, lockfile, fidelity, updated_at, created_at
FROM codeintel_lockfiles
WHERE %s
ORDER BY id
`

func makeGetLockfileIndexConds(opts GetLockfileIndexOpts) ([]*sqlf.Query, error) {
	var conds []*sqlf.Query

	if opts.ID != 0 {
		conds = append(conds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.RepoName != "" {
		conds = append(conds, sqlf.Sprintf("repository_id IN (SELECT id FROM repo WHERE name = %s)", opts.RepoName))
	}

	if opts.Commit != "" {
		conds = append(conds, sqlf.Sprintf("commit_bytea = %s", dbutil.CommitBytea(opts.Commit)))
	}

	if opts.Lockfile != "" {
		conds = append(conds, sqlf.Sprintf("lockfile = %s", opts.Lockfile))
	}

	if len(conds) == 0 {
		return nil, errors.New("not enough conditions given to query lockfile index")
	}

	return conds, nil
}

// DeleteLockfileIndexByID deletes the lockfile index with the given ID.
func (s *store) DeleteLockfileIndexByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.getLockfileIndex.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{}})
	}()

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	var (
		repoID   int
		commit   dbutil.CommitBytea
		lockfile string
	)
	err = tx.QueryRow(ctx, sqlf.Sprintf(deleteLockfileIndexByIDQuery, id)).Scan(&repoID, &commit, &lockfile)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLockfileIndexNotFound
		}
		return err
	}

	return tx.Exec(ctx, sqlf.Sprintf(deleteLockfileReferencesQuery, repoID, commit, lockfile))
}

const deleteLockfileIndexByIDQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:DeleteLockfileIndexByID
DELETE FROM codeintel_lockfiles
WHERE id = %s
RETURNING repository_id, commit_bytea, lockfile
`

const deleteLockfileReferencesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:DeleteLockfileIndexByID
DELETE FROM codeintel_lockfile_references
WHERE
	resolution_repository_id = %s
AND
	resolution_commit_bytea = %s
AND
	resolution_lockfile = %s
`

// UpsertDependencyRepos creates the given dependency repos if they don't yet exist. The values
// that did not exist previously are returned.
func (s *store) UpsertDependencyRepos(ctx context.Context, deps []shared.Repo) (newDeps []shared.Repo, err error) {
	ctx, _, endObservation := s.operations.upsertDependencyRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numDeps", len(deps)),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numNewDeps", len(newDeps)),
		}})
	}()

	callback := func(inserter *batch.Inserter) error {
		for _, dep := range deps {
			if err := inserter.Insert(ctx, dep.Scheme, dep.Name, dep.Version); err != nil {
				return err
			}
		}

		return nil
	}

	returningScanner := func(rows dbutil.Scanner) error {
		dependencyRepo, err := scanDependencyRepo(rows)
		if err != nil {
			return err
		}

		newDeps = append(newDeps, dependencyRepo)
		return nil
	}

	err = batch.WithInserterWithReturn(
		ctx,
		s.db.Handle(),
		"lsif_dependency_repos",
		batch.MaxNumPostgresParameters,
		[]string{"scheme", "name", "version"},
		"ON CONFLICT DO NOTHING",
		[]string{"id", "scheme", "name", "version"},
		returningScanner,
		callback,
	)
	return newDeps, err
}

// DeleteDependencyReposByID removes the dependency repos with the given ids, if they exist.
func (s *store) DeleteDependencyReposByID(ctx context.Context, ids ...int) (err error) {
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
-- source: internal/codeintel/dependencies/internal/store/store.go:DeleteDependencyReposByID
DELETE FROM lsif_dependency_repos
WHERE id = ANY(%s)
`

// Transact returns a store in a transaction.
func (s *store) Transact(ctx context.Context) (*store, error) {
	txBase, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		db:         txBase,
		operations: s.operations,
	}, nil
}
