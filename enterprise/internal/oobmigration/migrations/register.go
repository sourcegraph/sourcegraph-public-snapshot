package migrations

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	batchesmigrations "github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/batches"
	codeintelmigrations "github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func RegisterEnterpriseMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	if err := batchesmigrations.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	if err := codeintelmigrations.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	if err := insights.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	if err := productsubscription.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	return nil
}
