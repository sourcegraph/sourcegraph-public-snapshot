package background

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

const (
	// BatchChangesSSHMigrationID is the ID of row holding the ssh migration. It
	// is defined in `1528395788_campaigns_ssh_key_migration.up`.
	BatchChangesSSHMigrationID = 2

	// BatchChangesUserCredentialMigrationID is the ID of the row holding the
	// user credential migration. It is defined in
	// `1528395818_oob_credential_encryption_up.sql`.
	BatchChangesUserCredentialMigrationID = 9
)

// RegisterMigrations registers all currently implemented out of band migrations
// by batch changes with the migration runner.
func RegisterMigrations(cstore *store.Store, outOfBandMigrationRunner *oobmigration.Runner) error {
	migrations := map[int]oobmigration.Migrator{
		BatchChangesSSHMigrationID:            &sshMigrator{store: cstore},
		BatchChangesUserCredentialMigrationID: &userCredentialMigrator{store: cstore},
	}

	for id, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(id, migrator, oobmigration.MigratorOptions{Interval: 5 * time.Second}); err != nil {
			return err
		}
	}

	return nil
}
