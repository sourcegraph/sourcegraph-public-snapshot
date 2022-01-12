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

const currentMigrationLogSchemaVersion = 1

func (s *Store) EnsureSchemaTable(ctx context.Context) (err error) {
	ctx, endObservation := s.operations.ensureSchemaTable.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	queries := []*sqlf.Query{
		sqlf.Sprintf(`CREATE TABLE IF NOT EXISTS %s(version bigint NOT NULL PRIMARY KEY)`, quote(s.migrationsTable)),
		sqlf.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS dirty boolean NOT NULL`, quote(s.migrationsTable)),

		sqlf.Sprintf(`CREATE TABLE IF NOT EXISTS migration_logs(id SERIAL PRIMARY KEY)`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS migration_logs_schema_version integer NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS schema text NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS version integer NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS up bool NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS started_at timestamptz NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS finished_at timestamptz`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS success boolean`),
		sqlf.Sprintf(`ALTER TABLE migration_logs ADD COLUMN IF NOT EXISTS error_message text`),
	}

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	for _, query := range queries {
		if err := tx.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
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

// Lock creates and holds an advisory lock. This method returns a function that should be called
// once the lock should be released. This method accepts the current function's error output and
// wraps any additional errors that occur on close.
//
// Note that we don't use the internal/database/locker package here as that uses transactionally
// scoped advisory locks. We want to be able to hold locks outside of transactions for migrations.
func (s *Store) Lock(ctx context.Context) (_ bool, _ func(err error) error, err error) {
	key := s.lockKey()

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

// TryLock attempts to create hold an advisory lock. This method returns a function that should be
// called once the lock should be released. This method accepts the current function's error output
// and wraps any additional errors that occur on close. Calling this method when the lock was not
// acquired will return the given error without modification (no-op). If this method returns true,
// the lock was acquired and false if the lock is currently held by another process.
//
// Note that we don't use the internal/database/locker package here as that uses transactionally
// scoped advisory locks. We want to be able to hold locks outside of transactions for migrations.
func (s *Store) TryLock(ctx context.Context) (_ bool, _ func(err error) error, err error) {
	key := s.lockKey()

	ctx, endObservation := s.operations.tryLock.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("key", key),
	}})
	defer endObservation(1, observation.Args{})

	locked, _, err := basestore.ScanFirstBool(s.Query(ctx, sqlf.Sprintf(`SELECT pg_try_advisory_lock(%s, %s)`, key, 0)))
	if err != nil {
		return false, nil, err
	}

	close := func(err error) error {
		if locked {
			if unlockErr := s.Exec(ctx, sqlf.Sprintf(`SELECT pg_advisory_unlock(%s, %s)`, key, 0)); unlockErr != nil {
				err = multierror.Append(err, unlockErr)
			}
		}

		return err
	}

	return locked, close, nil
}

func (s *Store) lockKey() int32 {
	return locker.StringKey(fmt.Sprintf("%s:migrations", s.migrationsTable))
}

func (s *Store) Up(ctx context.Context, definition definition.Definition) (err error) {
	ctx, endObservation := s.operations.up.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if err := s.runMigrationQuery(ctx, definition.ID, true, definition.UpQuery); err != nil {
		return err
	}

	return nil
}

func (s *Store) Down(ctx context.Context, definition definition.Definition) (err error) {
	ctx, endObservation := s.operations.down.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if err := s.runMigrationQuery(ctx, definition.ID, false, definition.DownQuery); err != nil {
		return err
	}

	return nil
}

func (s *Store) runMigrationQuery(ctx context.Context, definitionVersion int, up bool, query *sqlf.Query) (err error) {
	targetVersion := definitionVersion
	expectedCurrentVersion := definitionVersion - 1
	if !up {
		targetVersion = definitionVersion - 1
		expectedCurrentVersion = definitionVersion
	}

	logID, err := s.setVersion(ctx, up, expectedCurrentVersion, targetVersion, definitionVersion)
	if err != nil {
		return err
	}

	defer func() {
		if execErr := s.Exec(ctx, sqlf.Sprintf(
			`UPDATE migration_logs SET finished_at = NOW(), success = %s, error_message = %s WHERE id = %d`,
			err == nil,
			strPtr(err),
			logID,
		)); execErr != nil {
			err = multierror.Append(err, execErr)
		}
	}()

	if err := s.Exec(ctx, query); err != nil {
		return err
	}

	if err := s.Exec(ctx, sqlf.Sprintf(`UPDATE %s SET dirty=false`, quote(s.migrationsTable))); err != nil {
		return err
	}

	return nil
}

func (s *Store) setVersion(ctx context.Context, up bool, expectedCurrentVersion, targetVersion, sourceVersion int) (_ int, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	assertionFailure := func(description string, args ...interface{}) error {
		cta := "This condition should not be reachable by normal use of the migration store via the runner and indicates a bug. Please report this issue."
		return errors.Errorf(description+"\n\n"+cta+"\n", args...)
	}

	if currentVersion, dirty, ok, err := tx.Version(ctx); err != nil {
		return 0, err
	} else if dirty {
		return 0, assertionFailure("dirty database")
	} else if ok {
		if currentVersion != expectedCurrentVersion {
			return 0, assertionFailure("expected schema to have version %d, but has version %d\n", expectedCurrentVersion, currentVersion)
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(`DELETE FROM %s`, quote(s.migrationsTable))); err != nil {
			return 0, err
		}
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(`INSERT INTO %s (version, dirty) VALUES (%s, true)`, quote(s.migrationsTable), targetVersion)); err != nil {
		return 0, err
	}

	id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		`
			INSERT INTO migration_logs (
				migration_logs_schema_version,
				schema,
				version,
				up,
				started_at
			) VALUES (%s, %s, %s, %s, NOW())
			RETURNING id
		`,
		currentMigrationLogSchemaVersion,
		s.migrationsTable,
		sourceVersion,
		up,
	)))
	if err != nil {
		return 0, err
	}

	return id, nil
}

var quote = sqlf.Sprintf

func strPtr(err error) *string {
	if err == nil {
		return nil
	}

	text := err.Error()
	return &text
}

type migrationLog struct {
	Schema  string
	Version int
	Up      bool
	Success *bool
}

// scanMigrationLogs scans a slice of migration logs from the return value of `*Store.query`.
func scanMigrationLogs(rows *sql.Rows, queryErr error) (_ []migrationLog, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var logs []migrationLog
	for rows.Next() {
		var log migrationLog

		if err := rows.Scan(
			&log.Schema,
			&log.Version,
			&log.Up,
			&log.Success,
		); err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	return logs, nil
}
