package batches

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	sshMigrator := NewSSHMigratorWithDB(db, keyring.Default().BatchChangesCredentialKey)
	migrations := map[int]oobmigration.Migrator{
		sshMigrator.ID(): sshMigrator,
	}

	for id, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(id, migrator, oobmigration.MigratorOptions{Interval: 5 * time.Second}); err != nil {
			return err
		}
	}

	return nil
}
