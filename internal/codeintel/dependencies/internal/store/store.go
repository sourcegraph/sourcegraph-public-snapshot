package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store
	operations *operations
}

func newStore(db dbutil.DB, op *operations) *Store {
	return &Store{
		Store:      basestore.NewWithDB(db, sql.TxOptions{}),
		operations: op,
	}
}

func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{
		Store:      s.Store.With(other),
		operations: s.operations,
	}
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store:      txBase,
		operations: s.operations,
	}, nil
}

// LockfileDependencies returns package dependencies from a previous lockfiles result for
// the given repository and commit. It is assumed that the given commit is the canonical
// 40-character hash.
func (s *Store) LockfileDependencies(ctx context.Context, repoName, commit string) (deps []shared.PackageDependency, found bool, err error) {
	ctx, _, endObservation := s.operations.lockfileDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", repoName),
		log.String("commit", commit),
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
	defer func() { err = tx.Done(err) }()

	deps, err = scanPackageDependencies(tx.Query(ctx, sqlf.Sprintf(
		lockfileDependenciesQuery,
		repoName,
		dbutil.CommitBytea(commit),
	)))
	if err != nil {
		return nil, false, err
	}
	if len(deps) == 0 {
		// No dependencies were found, but we could have already written a record
		// that just had an empty references list. Check to see if this is the case
		// so we don't attempt to re-parse the lockfiles of this repo/commit from the
		// dependencies service.
		_, found, err = basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
			lockfileDependenciesExistsQuery,
			repoName,
			dbutil.CommitBytea(commit),
		)))

		return nil, found, err
	}

	return deps, true, nil
}

const lockfileDependenciesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:LockfileDependencies
SELECT
	repository_name,
	revspec,
	package_scheme,
	package_name,
	package_version
FROM codeintel_lockfile_references
WHERE id IN (
	SELECT DISTINCT unnest(codeintel_lockfile_reference_ids) AS id
	FROM codeintel_lockfiles
	WHERE repository_id = (SELECT id FROM repo WHERE name = %s) AND commit_bytea = %s
)
`

const lockfileDependenciesExistsQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:LockfileDependencies
SELECT 1
FROM codeintel_lockfiles
WHERE repository_id = (SELECT id FROM repo WHERE name = %s) AND commit_bytea = %s
`

// UpsertLockfileDependencies inserts the given package dependencies if they do not exist
// and inserts a new lockfiles result for the given repository and commit. It is assumed
// that the given commit is the canonical 40-character hash.
func (s *Store) UpsertLockfileDependencies(ctx context.Context, repoName, commit string, deps []shared.PackageDependency) (err error) {
	ctx, _, endObservation := s.operations.upsertLockfileDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoName", repoName),
		log.String("commit", commit),
		log.Int("numDeps", len(deps)),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(temporaryLockfileReferencesTableQuery)); err != nil {
		return err
	}

	if err := batch.InsertValues(
		ctx,
		tx.Handle().DB(),
		"t_codeintel_lockfile_references",
		batch.MaxNumPostgresParameters,
		[]string{"repository_name", "revspec", "package_scheme", "package_name", "package_version"},
		populatePackageDependencyChannel(deps),
	); err != nil {
		return err
	}

	ids, err := basestore.ScanInts(tx.Query(ctx, sqlf.Sprintf(upsertLockfileReferencesQuery)))
	if err != nil {
		return err
	}
	if ids == nil {
		ids = []int{}
	}
	idsArray := pq.Array(ids)

	return tx.Exec(ctx, sqlf.Sprintf(
		insertLockfilesQuery,
		dbutil.CommitBytea(commit),
		idsArray,
		repoName,
		idsArray,
	))
}

const temporaryLockfileReferencesTableQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpsertLockfileDependencies
CREATE TEMPORARY TABLE t_codeintel_lockfile_references (
	repository_name text NOT NULL,
	revspec text NOT NULL,
	package_scheme text NOT NULL,
	package_name text NOT NULL,
	package_version text NOT NULL
) ON COMMIT DROP
`

const upsertLockfileReferencesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpsertLockfileDependencies
WITH ins AS (
	INSERT INTO codeintel_lockfile_references (repository_name, revspec, package_scheme, package_name, package_version)
	SELECT repository_name, revspec, package_scheme, package_name, package_version FROM t_codeintel_lockfile_references
	ON CONFLICT DO NOTHING
	RETURNING id
),
duplicates AS (
	SELECT id
	FROM t_codeintel_lockfile_references t
	JOIN codeintel_lockfile_references r
	ON
		r.repository_name = t.repository_name AND
		r.revspec = t.revspec AND
		r.package_scheme = t.package_scheme AND
		r.package_name = t.package_name AND
		r.package_version = t.package_version
)
SELECT id FROM ins UNION
SELECT id FROM duplicates
ORDER BY id
`

const insertLockfilesQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:UpsertLockfileDependencies
INSERT INTO codeintel_lockfiles (
	repository_id,
	commit_bytea,
	codeintel_lockfile_reference_ids
)
SELECT id, %s, %s
FROM repo
WHERE name = %s
-- Last write wins
ON CONFLICT (repository_id, commit_bytea) DO UPDATE
SET codeintel_lockfile_reference_ids = %s
`

// populatePackageDependencyChannel populates a channel with the given dependencies for bulk insertion.
func populatePackageDependencyChannel(deps []shared.PackageDependency) <-chan []any {
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
			}
		}
	}()

	return ch
}

type ListDependencyReposOpts struct {
	Scheme      string
	Name        string
	After       int
	Limit       int
	NewestFirst bool
}

// ListDependencyRepos returns dependency repositories to be synced by gitserver.
func (s *Store) ListDependencyRepos(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shared.Repo, err error) {
	ctx, _, endObservation := s.operations.listDependencyRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("scheme", opts.Scheme),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRepos", len(dependencyRepos)),
		}})
	}()

	sortDirection := "ASC"
	if opts.NewestFirst {
		sortDirection = "DESC"
	}

	return scanDependencyRepos(s.Query(ctx, sqlf.Sprintf(
		listDependencyReposQuery,
		sqlf.Join(makeListDependencyReposConds(opts), "AND"),
		sqlf.Sprintf(sortDirection),
		makeLimit(opts.Limit),
	)))
}

const listDependencyReposQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:ListDependencyRepos
SELECT id, scheme, name, version
FROM lsif_dependency_repos
WHERE %s
ORDER BY id %s
%s
`

func makeListDependencyReposConds(opts ListDependencyReposOpts) []*sqlf.Query {
	conds := make([]*sqlf.Query, 0, 3)
	conds = append(conds, sqlf.Sprintf("scheme = %s", opts.Scheme))

	if opts.Name != "" {
		conds = append(conds, sqlf.Sprintf("name = %s", opts.Name))
	}
	if opts.After != 0 {
		if opts.NewestFirst {
			conds = append(conds, sqlf.Sprintf("id < %s", opts.After))
		} else {
			conds = append(conds, sqlf.Sprintf("id > %s", opts.After))
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

// UpsertDependencyRepos creates the given dependency repos if they don't yet exist. The values
// that did not exist previously are returned.
func (s *Store) UpsertDependencyRepos(ctx context.Context, deps []shared.Repo) (newDeps []shared.Repo, err error) {
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
		s.Handle().DB(),
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
func (s *Store) DeleteDependencyReposByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deleteDependencyReposByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numIDs", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil
	}

	return s.Exec(ctx, sqlf.Sprintf(deleteDependencyReposByIDQuery, pq.Array(ids)))
}

const deleteDependencyReposByIDQuery = `
-- source: internal/codeintel/dependencies/internal/store/store.go:DeleteDependencyReposByID
DELETE FROM lsif_dependency_repos
WHERE id = ANY(%s)
`
