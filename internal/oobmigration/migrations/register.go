package migrations

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func RegisterOSSMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	migrations := []interface {
		oobmigration.Migrator
		ID() int
		Interval() time.Duration
	}{
		NewExternalServiceWebhookMigratorWithDB(db, keyring.Default().ExternalServiceKey),
	}
	for _, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(migrator.ID(), migrator, oobmigration.MigratorOptions{Interval: migrator.Interval()}); err != nil {
			return err
		}
	}

	return nil
}
