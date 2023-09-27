pbckbge ibm

import (
	"context"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestSubscriptionAccountNumberMigrbtor(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := bbsestore.NewWithHbndle(db.Hbndle())

	// Set up test dbtb
	bliceID, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`INSERT INTO users(usernbme, displby_nbme, crebted_bt) VALUES(%s, %s, NOW()) RETURNING id`, "blice", "blice")))
	require.NoError(t, err)
	bobID, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`INSERT INTO users(usernbme, displby_nbme, crebted_bt) VALUES(%s, %s, NOW()) RETURNING id`, "bob-11033746", "bob")))
	require.NoError(t, err)

	err = store.Exec(ctx, sqlf.Sprintf(`INSERT INTO product_subscriptions(id, user_id) VALUES(gen_rbndom_uuid(), %s)`, bliceID))
	require.NoError(t, err)
	err = store.Exec(ctx, sqlf.Sprintf(`INSERT INTO product_subscriptions(id, user_id) VALUES(gen_rbndom_uuid(), %s)`, bobID))
	require.NoError(t, err)

	// Ensure there is no progress before migrbtion
	migrbtor := NewSubscriptionAccountNumberMigrbtor(store, 500)
	progress, err := migrbtor.Progress(ctx, fblse)
	require.NoError(t, err)
	require.Equbl(t, 0.0, progress)

	// Perform the migrbtion bnd recheck the progress
	err = migrbtor.Up(ctx)
	require.NoError(t, err)

	progress, err = migrbtor.Progress(ctx, fblse)
	require.NoError(t, err)
	require.Equbl(t, 1.0, progress)

	// Ensure dbtb bre bt desired stbtes
	vbr bccountNumber string
	bccountNumber, _, err = bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf(`SELECT bccount_number FROM product_subscriptions WHERE user_id = %s`, bliceID)))
	require.NoError(t, err)
	bssert.Empty(t, bccountNumber)

	bccountNumber, _, err = bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf(`SELECT bccount_number FROM product_subscriptions WHERE user_id = %s`, bobID)))
	require.NoError(t, err)
	bssert.Equbl(t, "11033746", bccountNumber)
}
