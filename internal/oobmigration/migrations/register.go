package migrations

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/batches"
)

func RegisterOSSMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	return registerOSSMigrations(outOfBandMigrationRunner, migratorDependencies{
		store:   basestore.NewWithHandle(db.Handle()),
		keyring: keyring.Default(),
	})
}

func RegisterOSSMigrationsFromConfig(db database.DB, outOfBandMigrationRunner *oobmigration.Runner, conf conftypes.UnifiedQuerier) error {
	return registerOSSMigrations(outOfBandMigrationRunner, migratorDependencies{
		store:   basestore.NewWithHandle(db.Handle()),
		keyring: keyring.Default(), // TODO - get from config
	})
}

type migratorDependencies struct {
	store   *basestore.Store
	keyring keyring.Ring
}

func registerOSSMigrations(outOfBandMigrationRunner *oobmigration.Runner, deps migratorDependencies) error {
	return RegisterAll(outOfBandMigrationRunner, []TaggedMigrator{
		batches.NewExternalServiceWebhookMigratorWithDB(deps.store, deps.keyring.ExternalServiceKey, 50),
	})
}

type TaggedMigrator interface {
	oobmigration.Migrator
	ID() int
	Interval() time.Duration
}

func RegisterAll(outOfBandMigrationRunner *oobmigration.Runner, migrators []TaggedMigrator) error {
	for _, migrator := range migrators {
		if err := outOfBandMigrationRunner.Register(
			migrator.ID(),
			migrator,
			oobmigration.MigratorOptions{Interval: migrator.Interval()},
		); err != nil {
			return err
		}
	}

	return nil
}
