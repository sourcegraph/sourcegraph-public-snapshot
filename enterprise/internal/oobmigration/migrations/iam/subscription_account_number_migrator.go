package iam

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type subscriptionAccountNumberMigrator struct {
	store     *basestore.Store
	batchSize int
}

var _ oobmigration.Migrator = &subscriptionAccountNumberMigrator{}

func NewSubscriptionAccountNumberMigrator(store *basestore.Store, batchSize int) *subscriptionAccountNumberMigrator {
	return &subscriptionAccountNumberMigrator{
		store:     store,
		batchSize: batchSize,
	}
}

func (m *subscriptionAccountNumberMigrator) ID() int                 { return 15 }
func (m *subscriptionAccountNumberMigrator) Interval() time.Duration { return time.Second * 5 }

func (m *subscriptionAccountNumberMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(subscriptionAccountNumberMigratorProgressQuery)))
	return progress, err
}

const subscriptionAccountNumberMigratorProgressQuery = `
-- source: enterprise/internal/productsubscription/subscription_account_number_migrator.go:Progress
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		cast(c1.count as float) / cast(c2.count as float)
	END
FROM
	(SELECT count(*) as count FROM product_subscriptions WHERE account_number IS NOT NULL) c1,
	(SELECT count(*) as count FROM product_subscriptions) c2
`

func (m *subscriptionAccountNumberMigrator) Up(ctx context.Context) (err error) {
	return m.store.Exec(ctx, sqlf.Sprintf(subscriptionAccountNumberMigratorUpQuery, m.batchSize))
}

const subscriptionAccountNumberMigratorUpQuery = `
-- source: enterprise/internal/productsubscription/subscription_account_number_migrator.go:Up
WITH candidates AS (
	SELECT
		product_subscriptions.id::uuid AS subscription_id,
		COALESCE(split_part(users.username, '-', 2), '') AS account_number
	FROM product_subscriptions
	JOIN users ON product_subscriptions.user_id = users.id
	WHERE product_subscriptions.account_number IS NULL
	LIMIT %s
	FOR UPDATE SKIP LOCKED
)
UPDATE product_subscriptions
SET account_number = candidates.account_number
FROM candidates
WHERE product_subscriptions.id = candidates.subscription_id
`

func (m *subscriptionAccountNumberMigrator) Down(_ context.Context) error {
	// non-destructive
	return nil
}
