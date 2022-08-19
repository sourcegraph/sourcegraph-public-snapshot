package migrations

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
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
	extsvcKey := keyring.Default().ExternalServiceKey

	return RegisterAll(outOfBandMigrationRunner, []TaggedMigrator{
		batches.NewExternalServiceWebhookMigratorWithDB(db, extsvcKey, 50),
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
