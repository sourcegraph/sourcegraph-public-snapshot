package migrate

import (
	"context"
	"errors"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	v0 "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/migrate/v0"
	v1 "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/migrate/v1"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
)

// ErrNoVersion occurs when there are no rows in the schema_version table.
var ErrNoVersion = errors.New("no rows in schema_version")

// MigrationFunc runs a migration on the given store. The given serializer is the one that is used by
// the current (most recent) version of the SQLite writer class. If a serializer output changes in a
// significant way, it may be necessary to inline the serializer behavior as new migrations are
// introduced.
type MigrationFunc func(ctx context.Context, s *store.Store, serializer serialization.Serializer) error

var migrations = []struct {
	Version       string
	MigrationFunc MigrationFunc
}{
	{"v00000", v0.Migrate},
	{"v00001", v1.Migrate},
}

var UnknownSchemaVersion = migrations[0].Version
var CurrentSchemaVersion = migrations[len(migrations)-1].Version

// Migrate determines the current schema version and runs any migrations necessary to transform it to
// the current schema version. Each migration is ran in an individual transaction. An error is returned
// if the current schema version is unknown or if a migration is unsuccessful.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	version, err := getVersion(ctx, s)
	if err != nil {
		return err
	}

	found := false
	for _, migration := range migrations {
		// This is our current version: all migrations _after_ this iteration should be applied
		if migration.Version == version {
			found = true
			continue
		}
		if !found {
			continue
		}

		if err := runMigration(ctx, s, serializer, migration.Version, migration.MigrationFunc); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("unrecognized schema version %s", version)
	}

	return nil
}

// runMigration applies a single migration function within a transaction. If the migration
// function is successful, the schema version will be reflected to update the new version.
func runMigration(ctx context.Context, store *store.Store, serializer serialization.Serializer, version string, migrationFunc MigrationFunc) (err error) {
	tx, err := store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	if err := migrationFunc(ctx, tx, serializer); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE schema_version SET version = %s", version)); err != nil {
		return err
	}

	return nil
}

// getVersion returns the current schema version of the store.
func getVersion(ctx context.Context, s *store.Store) (string, error) {
	// Determine if schema_version table exists
	_, exists, err := store.ScanFirstString(s.Query(ctx, sqlf.Sprintf("SELECT name FROM sqlite_master WHERE type = 'table' AND name = %s", "schema_version")))
	if err != nil {
		return "", err
	}
	if !exists {
		// We assume this database was created prior to this migration mechanism
		// and return the lowest migration version so that we apply all migrations.
		return UnknownSchemaVersion, nil
	}

	version, exists, err := store.ScanFirstString(s.Query(ctx, sqlf.Sprintf("SELECT version FROM schema_version LIMIT 1")))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrNoVersion
	}

	return version, nil
}
