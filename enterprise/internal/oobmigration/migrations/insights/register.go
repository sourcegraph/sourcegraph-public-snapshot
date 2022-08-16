package insights

import (
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	var insightsMigrator oobmigration.Migrator
	if !insights.IsEnabled() {
		// This allows this migration to be "complete" even when insights is not enabled.
		insightsMigrator = NewMigratorNoOp()
	} else {
		insightsDB, err := insights.InitializeCodeInsightsDB("worker-oobmigrator")
		if err != nil {
			return err
		}
		insightsMigrator = NewMigrator(database.NewDBWith(log.Scoped("codeinsights-db", ""), insightsDB), db)
	}

	// This id (14) was defined arbitrarily in this migration file: 1528395945_settings_migration_out_of_band.up.sql.
	if err := outOfBandMigrationRunner.Register(14, insightsMigrator, oobmigration.MigratorOptions{Interval: 10 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to register settings migration job")
	}

	return nil
}
