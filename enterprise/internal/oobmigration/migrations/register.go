package migrations

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	batchesmigrations "github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/batches"
	codeintelmigrations "github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type TaggedMigrator interface {
	oobmigration.Migrator

	ID() int
	Interval() time.Duration
}

func RegisterEnterpriseMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	migrations := []TaggedMigrator{
		NewSubscriptionAccountNumberMigrator(db),
		NewLicenseKeyFieldsMigrator(db),
	}
	for _, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(migrator.ID(), migrator, oobmigration.MigratorOptions{Interval: migrator.Interval()}); err != nil {
			return err
		}
	}

	if err := batchesmigrations.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	if err := codeintelmigrations.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	if err := insights.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	return nil
}
