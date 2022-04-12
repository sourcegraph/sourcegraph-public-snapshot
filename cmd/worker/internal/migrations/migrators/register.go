package migrators

import (
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RegisterOSSMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	// Run a background job to handle encryption of external service configuration.
	extsvcMigrator := NewExternalServiceConfigMigratorWithDB(db)
	extsvcMigrator.AllowDecrypt = os.Getenv("ALLOW_DECRYPT_MIGRATION") == "true"
	if err := outOfBandMigrationRunner.Register(extsvcMigrator.ID(), extsvcMigrator, oobmigration.MigratorOptions{Interval: 3 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to run external service encryption job")
	}

	// Run a background job to handle encryption of external service configuration.
	extAccMigrator := NewExternalAccountsMigratorWithDB(db)
	extAccMigrator.AllowDecrypt = os.Getenv("ALLOW_DECRYPT_MIGRATION") == "true"
	if err := outOfBandMigrationRunner.Register(extAccMigrator.ID(), extAccMigrator, oobmigration.MigratorOptions{Interval: 3 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to run user external account encryption job")
	}

	// Run a background job to calculate the has_webhooks field on external service records.
	webhookMigrator := NewExternalServiceWebhookMigratorWithDB(db)
	if err := outOfBandMigrationRunner.Register(webhookMigrator.ID(), webhookMigrator, oobmigration.MigratorOptions{Interval: 3 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to run external service webhook job")
	}

	return nil
}
