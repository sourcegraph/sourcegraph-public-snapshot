package migrators

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// ExternalServiceWebhookMigrator is a background job that calculates the
// has_webhooks field on external services based on the external service
// configuration.
type ExternalServiceWebhookMigrator struct {
	store     *basestore.Store
	BatchSize int
}

var _ oobmigration.Migrator = &ExternalServiceWebhookMigrator{}

func NewExternalServiceWebhookMigrator(store *basestore.Store) *ExternalServiceWebhookMigrator {
	// Batch size arbitrarily chosen to match ExternalServiceConfigMigrator.
	return &ExternalServiceWebhookMigrator{store: store, BatchSize: 50}
}

func NewExternalServiceWebhookMigratorWithDB(db dbutil.DB) *ExternalServiceWebhookMigrator {
	return NewExternalServiceWebhookMigrator(basestore.NewWithDB(db, sql.TxOptions{}))
}

// ID returns the migration row ID in the out_of_band_migrations table.
//
// This ID was defined in the migration:
// migrations/frontend/1528395921_add_has_webhooks.up.sql
func (m *ExternalServiceWebhookMigrator) ID() int {
	return 13
}

// Progress returns a value from 0 to 1 representing the percentage of external
// services that have had their has_webhooks field calculated.
func (m *ExternalServiceWebhookMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				CAST(c1.count AS float) / CAST(c2.count AS float)
			END
		FROM
			(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS NOT NULL) c1,
			(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL) c2
	`)))
	return progress, err
}

// Up loads BatchSize external services, locks them, and upserts them back into
// the database, which will calculate HasWebhooks along the way.
func (m *ExternalServiceWebhookMigrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	store := database.ExternalServicesWith(tx)

	svcs, err := store.List(ctx, database.ExternalServicesListOptions{
		OrderByDirection: "ASC",
		LimitOffset:      &database.LimitOffset{Limit: m.BatchSize},
		NoCachedWebhooks: true,
		ForUpdate:        true,
	})
	if err != nil {
		return err
	}

	err = store.Upsert(ctx, svcs...)
	return err
}

func (*ExternalServiceWebhookMigrator) Down(context.Context) error {
	// There's no sensible down migration here: if the SQL down migration has
	// been run, then the field no longer exists, and there's nothing to do.
	return nil
}
