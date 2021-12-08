package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store
	migrationsTable string
	operations      *Operations
}

func NewWithDB(db dbutil.DB, migrationsTable string, operations *Operations) *Store {
	return &Store{
		Store:           basestore.NewWithDB(db, sql.TxOptions{}),
		migrationsTable: migrationsTable,
		operations:      operations,
	}
}

func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{
		Store:           s.Store.With(other),
		migrationsTable: s.migrationsTable,
		operations:      s.operations,
	}
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store:           txBase,
		migrationsTable: s.migrationsTable,
		operations:      s.operations,
	}, nil
}

func (s *Store) EnsureSchemaTable(ctx context.Context) (err error) {
	ctx, endObservation := s.operations.ensureSchemaTable.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.Exec(ctx, sqlf.Sprintf(`CREATE TABLE IF NOT EXISTS %s (version bigint NOT NULL PRIMARY KEY, dirty boolean NOT NULL)`, quote(s.migrationsTable)))
}

func (s *Store) Version(ctx context.Context) (version int, dirty bool, ok bool, err error) {
	ctx, endObservation := s.operations.version.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rows, err := s.Query(ctx, sqlf.Sprintf(`SELECT version, dirty FROM %s`, quote(s.migrationsTable)))
	if err != nil {
		return 0, false, false, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&version, &dirty); err != nil {
			return 0, false, false, err
		}

		return version, dirty, true, nil
	}

	return 0, false, false, nil
}

func (s *Store) Lock(ctx context.Context) (_ bool, _ func(err error) error, err error) {
	key := locker.StringKey(fmt.Sprintf("%s:migrations", s.migrationsTable))

	ctx, endObservation := s.operations.lock.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("key", key),
	}})
	defer endObservation(1, observation.Args{})

	if err := s.Exec(ctx, sqlf.Sprintf(`SELECT pg_advisory_lock(%s, %s)`, key, 0)); err != nil {
		return false, nil, err
	}

	close := func(err error) error {
		if unlockErr := s.Exec(ctx, sqlf.Sprintf(`SELECT pg_advisory_unlock(%s, %s)`, key, 0)); unlockErr != nil {
			err = multierror.Append(err, unlockErr)
		}

		return err
	}

	return true, close, nil
}

func (s *Store) Up(ctx context.Context, definition definition.Definition) (err error) {
	ctx, endObservation := s.operations.up.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if err := s.runMigrationQuery(ctx, definition, definition.ID-1, definition.UpQuery); err != nil {
		return err
	}

	return nil
}

func (s *Store) Down(ctx context.Context, definition definition.Definition) (err error) {
	ctx, endObservation := s.operations.down.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if err := s.runMigrationQuery(ctx, definition, definition.ID+1, definition.DownQuery); err != nil {
		return err
	}

	return nil
}

func (s *Store) runMigrationQuery(ctx context.Context, definition definition.Definition, expectedCurrentVersion int, query *sqlf.Query) error {
	if err := s.setVersion(ctx, expectedCurrentVersion, definition.ID); err != nil {
		return err
	}

	if err := s.Exec(ctx, query); err != nil {
		return err
	}

	if err := s.Exec(ctx, sqlf.Sprintf(`UPDATE %s SET dirty=false`, quote(s.migrationsTable))); err != nil {
		return err
	}

	return nil
}

func (s *Store) setVersion(ctx context.Context, expectedCurrentVersion, version int) (err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if currentVersion, dirty, ok, err := tx.Version(ctx); err != nil {
		return err
	} else if dirty {
		return errors.New("dirty database")
	} else if ok {
		if currentVersion != expectedCurrentVersion {
			return errors.New("wrong expected version")
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(`DELETE FROM %s`, quote(s.migrationsTable))); err != nil {
			return err
		}
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(`INSERT INTO %s (version, dirty) VALUES (%s, true)`, quote(s.migrationsTable), version)); err != nil {
		return err
	}

	return nil
}

var quote = sqlf.Sprintf
