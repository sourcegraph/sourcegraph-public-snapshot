package batches

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

const (
	// BatchChangesSSHMigrationID is the ID of row holding the ssh migration. It
	// is defined in `1528395788_campaigns_ssh_key_migration.up`.
	BatchChangesSSHMigrationID = 2
)

func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	migrations := map[int]oobmigration.Migrator{
		BatchChangesSSHMigrationID: NewSSHMigratorWithDB(db, keyring.Default().BatchChangesCredentialKey),
	}

	for id, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(id, migrator, oobmigration.MigratorOptions{Interval: 5 * time.Second}); err != nil {
			return err
		}
	}

	return nil
}
