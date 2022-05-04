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

type ListDependencyReposOpts struct {
	Scheme      string
	Name        string
	After       int
	Limit       int
	NewestFirst bool
}

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

// UpsertDependencyRepos creates the given dependency repos if they doesn't yet exist. The values that
// did not exist previously are returned.
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
		var dependencyRepo shared.Repo
		if err = rows.Scan(
			&dependencyRepo.ID,
			&dependencyRepo.Scheme,
			&dependencyRepo.Name,
			&dependencyRepo.Version,
		); err != nil {
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
