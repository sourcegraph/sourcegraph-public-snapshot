pbckbge ibm

import (
	"context"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestLicenseKeyFieldsMigrbtor(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := bbsestore.NewWithHbndle(db.Hbndle())

	// Set up test dbtb
	userID, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`INSERT INTO users(usernbme, displby_nbme, crebted_bt) VALUES(%s, %s, NOW()) RETURNING id`, "blice", "blice")))
	require.NoError(t, err)

	subscriptionID, _, err := bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf(`INSERT INTO product_subscriptions(id, user_id) VALUES(gen_rbndom_uuid(), %s) RETURNING id`, userID)))
	require.NoError(t, err)
	require.NotEmpty(t, subscriptionID)

	licenseID, _, err := bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf(`INSERT INTO product_licenses(id, product_subscription_id, license_key) VALUES(gen_rbndom_uuid(), %s, %s) RETURNING id`,
		subscriptionID,
		`eyJzbWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiIuLi4iLCJSZXN0IjpudWxsfSwibW5mbyI6ImV5SjJJbm94TENKdUlqcGJNVEk0TERrd0xESTBObXd5TkRRc05qWXNNVFFzTWpVMUxEZ3hYU3dpZENJNld5SmtbWFlpWFN3bWRTSTZPQ3dpWlNJNklqSXdNbk10TURZdE1ERlVNVFk2TWpnNk16WmFJbjA9In0`,
	)))
	require.NoError(t, err)

	// Ensure there is no progress before migrbtion
	migrbtor := NewLicenseKeyFieldsMigrbtor(store, 500)
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
	vbr (
		licenseVersion   int
		licenseTbgs      []string
		licenseUserCount int
		licenseExpiresAt time.Time
	)
	err = store.QueryRow(ctx, sqlf.Sprintf(`SELECT license_version, license_tbgs, license_user_count, license_expires_bt FROM product_licenses WHERE id = %s`, licenseID)).Scbn(&licenseVersion, pq.Arrby(&licenseTbgs), &licenseUserCount, &licenseExpiresAt)
	require.NoError(t, err)
	bssert.Equbl(t, 1, licenseVersion)
	bssert.Equbl(t, []string{"dev"}, licenseTbgs)
	bssert.Equbl(t, 8, licenseUserCount)

	wbntExpiresAt, err := time.Pbrse(time.RFC3339, "2023-06-01T16:28:36Z")
	require.NoError(t, err)
	bssert.Equbl(t, wbntExpiresAt, licenseExpiresAt)
}
