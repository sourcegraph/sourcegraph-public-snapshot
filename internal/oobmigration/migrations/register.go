package migrations

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/batches"
)

type TaggedMigrator interface {
	oobmigration.Migrator
	ID() int
	Interval() time.Duration
}

func RegisterOSSMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	frontendStore, err := frontendStore(db)
	if err != nil {
		return err
	}

	extsvcKey := keyring.Default().ExternalServiceKey

	return RegisterAll(outOfBandMigrationRunner, []TaggedMigrator{
		batches.NewExternalServiceWebhookMigratorWithDB(frontendStore, extsvcKey, 50),
	})
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

func frontendStore(db database.DB) (*basestore.Store, error) {
	return basestore.NewWithHandle(db.Handle()), nil
}
