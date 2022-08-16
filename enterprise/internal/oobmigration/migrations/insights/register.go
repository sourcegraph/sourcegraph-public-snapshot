package insights

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	insightsMigrator, err := getMigrator(db)
	if err != nil {
		return err
	}

	// This id (14) was defined arbitrarily in this migration file: 1528395945_settings_migration_out_of_band.up.sql.
	if err := outOfBandMigrationRunner.Register(14, insightsMigrator, oobmigration.MigratorOptions{Interval: 10 * time.Second}); err != nil {
		return errors.Wrap(err, "failed to register settings migration job")
	}

	return nil
}

func getMigrator(db database.DB) (oobmigration.Migrator, error) {
	if !insights.IsEnabled() {
		// This allows this migration to be "complete" even when insights is not enabled.
		return noop, nil
	}

	insightsDB, err := insights.InitializeCodeInsightsDB("worker-oobmigrator")
	if err != nil {
		return nil, err
	}
	return NewMigrator(database.NewDBWith(log.Scoped("codeinsights-db", ""), insightsDB), db), nil
}

type noopMigrator struct{}

var noop oobmigration.Migrator = &noopMigrator{}

func (m *noopMigrator) Progress(ctx context.Context) (float64, error) { return 1, nil }
func (m *noopMigrator) Up(ctx context.Context) (err error)            { return nil }
func (m *noopMigrator) Down(ctx context.Context) (err error)          { return nil }
