package migrations

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/batches"
)

func RegisterOSSMigrations(
	ctx context.Context,
	db database.DB,
	outOfBandMigrationRunner *oobmigration.Runner,
) error {
	keyring := keyring.Default()

	return registerOSSMigrations(outOfBandMigrationRunner, migratorDependencies{
		store:   basestore.NewWithHandle(db.Handle()),
		keyring: &keyring,
	})
}

func RegisterOSSMigrationsFromConfig(
	ctx context.Context,
	db database.DB,
	outOfBandMigrationRunner *oobmigration.Runner,
	conf conftypes.UnifiedQuerier,
) error {
	keyring, err := keyring.NewRing(ctx, conf.SiteConfig().EncryptionKeys)
	if err != nil {
		return err
	}

	return registerOSSMigrations(outOfBandMigrationRunner, migratorDependencies{
		store:   basestore.NewWithHandle(db.Handle()),
		keyring: keyring,
	})
}

type migratorDependencies struct {
	store   *basestore.Store
	keyring *keyring.Ring
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
