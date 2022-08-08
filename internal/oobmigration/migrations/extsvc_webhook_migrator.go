package migrations

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type ExternalServiceWebhookMigrator struct {
	logger    log.Logger
	store     *basestore.Store
	BatchSize int
}

var _ oobmigration.Migrator = &ExternalServiceWebhookMigrator{}

func NewExternalServiceWebhookMigratorWithDB(db database.DB) *ExternalServiceWebhookMigrator {
	return &ExternalServiceWebhookMigrator{
		logger:    log.Scoped("ExternalServiceWebhookMigrator", ""),
		store:     basestore.NewWithHandle(db.Handle()),
		BatchSize: 50,
	}
}

func (m *ExternalServiceWebhookMigrator) ID() int {
	return 13
}

// Progress returns the percentage (ranged [0, 1]) of external services with a
// populated has_webhooks column.
func (m *ExternalServiceWebhookMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(externalServiceWebhookMigratorProgressQuery)))
	return progress, err
}

const externalServiceWebhookMigratorProgressQuery = `
-- source: internal/oobmigration/migrations/extsvc_webhook_migrator.go:Progress
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST(c1.count AS float) / CAST(c2.count AS float)
	END
FROM
	(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS NOT NULL) c1,
	(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL) c2
`

// Up loads a set of external services without a populated has_webhooks column and
// updates that value by looking at that external service's configuration values.
func (m *ExternalServiceWebhookMigrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	store := database.ExternalServicesWith(m.logger, tx)

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
	// non-destructive
	return nil
}
