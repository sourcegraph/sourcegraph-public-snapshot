package connections

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
)

// memoryStore implements runner.Store but writes to migration metadata to any
// underlying persistence layer.
type memoryStore struct {
	db         *sql.DB
	version    int
	versionSet bool
	dirty      bool
}

func newMemoryStore(db *sql.DB) runner.Store {
	return &memoryStore{
		db: db,
	}
}

func (s *memoryStore) Version(ctx context.Context) (int, bool, bool, error) {
	return s.version, s.dirty, s.versionSet, nil
}

func (s *memoryStore) Lock(ctx context.Context) (bool, func(err error) error, error) {
	return true, func(err error) error { return err }, nil
}

func (s *memoryStore) TryLock(ctx context.Context) (bool, func(err error) error, error) {
	return true, func(err error) error { return err }, nil
}

func (s *memoryStore) Up(ctx context.Context, migration definition.Definition) error {
	return s.exec(ctx, migration, migration.UpQuery)
}

func (s *memoryStore) Down(ctx context.Context, migration definition.Definition) error {
	return s.exec(ctx, migration, migration.DownQuery)
}

func (s *memoryStore) exec(ctx context.Context, migration definition.Definition, query *sqlf.Query) error {
	_, err := s.db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	s.version, s.dirty, s.versionSet = migration.ID, s.dirty || err != nil, true
	return err
}
