package productsubscription

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	migrations := []interface {
		oobmigration.Migrator
		ID() int
		Interval() time.Duration
	}{
		NewSubscriptionAccountNumberMigrator(db),
		NewLicenseKeyFieldsMigrator(db),
	}
	for id, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(id, migrator, oobmigration.MigratorOptions{Interval: migrator.Interval()}); err != nil {
			return err
		}
	}
	return nil
}
