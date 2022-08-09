package productsubscription

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

const (
	subscriptionAccountNumberMigrationID = 15
	licenseKeyFieldsMigrationID          = 16
)

func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	store := basestore.NewWithHandle(db.Handle())
	migrations := map[int]oobmigration.Migrator{
		subscriptionAccountNumberMigrationID: &subscriptionAccountNumberMigrator{store: store},
		licenseKeyFieldsMigrationID:          &licenseKeyFieldsMigrator{store: store},
	}

	for id, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(id, migrator, oobmigration.MigratorOptions{Interval: 5 * time.Second}); err != nil {
			return err
		}
	}
	return nil
}
