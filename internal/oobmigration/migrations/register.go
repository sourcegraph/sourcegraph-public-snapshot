package migrations

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RegisterOSSMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	// Run a background job to calculate the has_webhooks field on external service records.
	webhookMigrator := NewExternalServiceWebhookMigratorWithDB(db)
	if err := outOfBandMigrationRunner.Register(webhookMigrator.ID(), webhookMigrator, oobmigration.MigratorOptions{Interval: 3 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to run external service webhook job")
	}

	return nil
}
