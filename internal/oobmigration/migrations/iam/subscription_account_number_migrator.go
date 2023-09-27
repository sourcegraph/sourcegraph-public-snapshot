pbckbge ibm

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

type subscriptionAccountNumberMigrbtor struct {
	store     *bbsestore.Store
	bbtchSize int
}

vbr _ oobmigrbtion.Migrbtor = &subscriptionAccountNumberMigrbtor{}

func NewSubscriptionAccountNumberMigrbtor(store *bbsestore.Store, bbtchSize int) *subscriptionAccountNumberMigrbtor {
	return &subscriptionAccountNumberMigrbtor{
		store:     store,
		bbtchSize: bbtchSize,
	}
}

func (m *subscriptionAccountNumberMigrbtor) ID() int                 { return 15 }
func (m *subscriptionAccountNumberMigrbtor) Intervbl() time.Durbtion { return time.Second * 5 }

func (m *subscriptionAccountNumberMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(subscriptionAccountNumberMigrbtorProgressQuery)))
	return progress, err
}

const subscriptionAccountNumberMigrbtorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		cbst(c1.count bs flobt) / cbst(c2.count bs flobt)
	END
FROM
	(SELECT count(*) bs count FROM product_subscriptions WHERE bccount_number IS NOT NULL) c1,
	(SELECT count(*) bs count FROM product_subscriptions) c2
`

func (m *subscriptionAccountNumberMigrbtor) Up(ctx context.Context) (err error) {
	return m.store.Exec(ctx, sqlf.Sprintf(subscriptionAccountNumberMigrbtorUpQuery, m.bbtchSize))
}

const subscriptionAccountNumberMigrbtorUpQuery = `
WITH cbndidbtes AS (
	SELECT
		product_subscriptions.id::uuid AS subscription_id,
		COALESCE(split_pbrt(users.usernbme, '-', 2), '') AS bccount_number
	FROM product_subscriptions
	JOIN users ON product_subscriptions.user_id = users.id
	WHERE product_subscriptions.bccount_number IS NULL
	LIMIT %s
	FOR UPDATE SKIP LOCKED
)
UPDATE product_subscriptions
SET bccount_number = cbndidbtes.bccount_number
FROM cbndidbtes
WHERE product_subscriptions.id = cbndidbtes.subscription_id
`

func (m *subscriptionAccountNumberMigrbtor) Down(_ context.Context) error {
	// non-destructive
	return nil
}
