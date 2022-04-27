package migrations

import (
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

const (
	// BatchChangesSSHMigrationID is the ID of row holding the ssh migration. It
	// is defined in `1528395788_campaigns_ssh_key_migration.up`.
	BatchChangesSSHMigrationID = 2

	// BatchChangesUserCredentialMigrationID is the ID of the row holding the
	// user credential migration. It is defined in
	// `1528395819_oob_credential_encryption_up.sql`.
	BatchChangesUserCredentialMigrationID = 9

	// BatchChangesSiteCredentialMigrationID is the ID of the row holding the
	// site credential migration. It is defined in
	// `1528395821_oob_site_credential_encryption_up.sql`.
	BatchChangesSiteCredentialMigrationID = 10
)

func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	observationContext := &observation.Context{
		Logger:     log.Scoped("migrations", ""),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Initialize store.
	cstore := store.New(db, observationContext, keyring.Default().BatchChangesCredentialKey)

	// Register Batch Changes OOB migrations.
	return Register(cstore, outOfBandMigrationRunner)
}

// Register registers all currently implemented out of band migrations
// by batch changes with the migration runner.
func Register(cstore *store.Store, outOfBandMigrationRunner *oobmigration.Runner) error {
	allowDecrypt := os.Getenv("ALLOW_DECRYPT_MIGRATION") == "true"

	migrations := map[int]oobmigration.Migrator{
		BatchChangesSSHMigrationID: &sshMigrator{store: cstore},
		BatchChangesUserCredentialMigrationID: &userCredentialMigrator{
			store:        cstore,
			allowDecrypt: allowDecrypt,
		},
		BatchChangesSiteCredentialMigrationID: &siteCredentialMigrator{
			store:        cstore,
			allowDecrypt: allowDecrypt,
		},
	}

	for id, migrator := range migrations {
		if err := outOfBandMigrationRunner.Register(id, migrator, oobmigration.MigratorOptions{Interval: 5 * time.Second}); err != nil {
			return err
		}
	}

	return nil
}
