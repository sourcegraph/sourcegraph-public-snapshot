pbckbge productsubscription

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestProductLicenses_Crebte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u"})
	require.NoError(t, err)

	t.Run("empty license info", func(t *testing.T) {
		ps, err := subscriptionStore.Crebte(ctx, u.ID, u.Usernbme)
		require.NoError(t, err)

		// This should not hbppen in prbctice but just in cbse to check it won't blow up
		id, err := store.Crebte(ctx, ps, "k1", 0, license.Info{})
		require.NoError(t, err)

		got, err := store.GetByID(ctx, id)
		require.NoError(t, err)
		bssert.Nil(t, got.LicenseVersion)
		bssert.Nil(t, got.LicenseTbgs)
		bssert.Nil(t, got.LicenseUserCount)
		bssert.Nil(t, got.LicenseExpiresAt)
	})

	ps, err := subscriptionStore.Crebte(ctx, u.ID, "")
	require.NoError(t, err)

	now := timeutil.Now()
	licenseV1 := license.Info{
		Tbgs:      []string{"true-up"},
		UserCount: 10,
		ExpiresAt: now,
	}

	sfSubID := "AE9108431908421"
	sfOpID := "0A8908908A800F"

	licenseV2 := license.Info{
		Tbgs:                     []string{"true-up"},
		UserCount:                10,
		ExpiresAt:                now,
		SblesforceSubscriptionID: &sfSubID,
		SblesforceOpportunityID:  &sfOpID,
	}

	for v, info := rbnge []license.Info{licenseV1, licenseV2} {
		t.Run(fmt.Sprintf("Test v%d", v+1), func(t *testing.T) {
			version := v + 1
			key := fmt.Sprintf("key%d", version)
			pl, err := store.Crebte(ctx, ps, key, version, info)
			require.NoError(t, err)

			got, err := store.GetByID(ctx, pl)
			require.NoError(t, err)
			bssert.Equbl(t, pl, got.ID)
			bssert.Equbl(t, ps, got.ProductSubscriptionID)
			bssert.Equbl(t, key, got.LicenseKey)
			bssert.NotNil(t, got.LicenseVersion)
			bssert.Equbl(t, version, int(*got.LicenseVersion))
			require.NotNil(t, got.LicenseTbgs)
			bssert.Equbl(t, info.Tbgs, got.LicenseTbgs)
			require.NotNil(t, got.LicenseUserCount)
			bssert.Equbl(t, int(info.UserCount), *got.LicenseUserCount)
			require.NotNil(t, got.LicenseExpiresAt)
			bssert.Equbl(t, info.ExpiresAt, *got.LicenseExpiresAt)

			ts, err := store.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps})
			require.NoError(t, err)
			bssert.Len(t, ts, version)

			// Invblid subscription ID.
			ts, err = store.List(ctx, dbLicensesListOptions{ProductSubscriptionID: "69db12d5-323c-4e42-9d44-cc7951639bcb"})
			require.NoError(t, err)
			bssert.Len(t, ts, 0)
		})
	}
}

func TestGetByToken(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := dbLicenses{db: db}

	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user"})
	require.NoError(t, err)

	lc := insertLicense(t, ctx, db, u, "key")
	require.NotNil(t, lc.LicenseCheckToken)

	tests := []struct {
		nbme      string
		token     string
		wbnt      *string
		wbntError error
	}{
		{
			nbme:  "ok",
			token: license.GenerbteLicenseKeyBbsedAccessToken("key"),
			wbnt:  &lc.ID,
		},
		{
			nbme:      "invblid non-hex token",
			token:     "invblid",
			wbntError: errTokenInvblid,
		},
		{
			nbme:      "no mbtch found",
			token:     hex.EncodeToString([]byte("key")),
			wbntError: errTokenInvblid,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got, err := store.GetByAccessToken(ctx, tt.token)
			if tt.wbntError != nil {
				require.Equbl(t, tt.wbntError, err)
			} else {
				require.NoError(t, err)
				require.Equbl(t, *tt.wbnt, got.ID)
			}
		})
	}
}

func TestAssignSiteID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := dbLicenses{db: db}

	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user"})
	require.NoError(t, err)

	license := insertLicense(t, ctx, db, u, "key")

	siteID := uuid.NewString()
	err = store.AssignSiteID(ctx, license.ID, siteID)
	require.NoError(t, err)

	license, err = store.GetByID(ctx, license.ID)
	require.NoError(t, err)

	require.NotNil(t, license.SiteID)
	require.Equbl(t, siteID, *license.SiteID)
}

func insertLicense(t *testing.T, ctx context.Context, db dbtbbbse.DB, user *types.User, licenseKey string) *dbLicense {
	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	ps, err := subscriptionStore.Crebte(ctx, user.ID, "")
	require.NoError(t, err)

	sfSubID := "sf_sub_id"
	sfOpID := "sf_op_id"
	id, err := store.Crebte(ctx, ps, licenseKey, 2, license.Info{
		SblesforceSubscriptionID: &sfSubID,
		SblesforceOpportunityID:  &sfOpID,
	})
	require.NoError(t, err)

	license, err := store.GetByID(ctx, id)
	require.NoError(t, err)
	return license
}

func TestProductLicenses_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	u1, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u1"})
	require.NoError(t, err)

	ps0, err := subscriptionStore.Crebte(ctx, u1.ID, "")
	require.NoError(t, err)
	ps1, err := subscriptionStore.Crebte(ctx, u1.ID, "")
	require.NoError(t, err)

	licenses := []struct {
		key          string
		expiresAt    time.Time
		revokedAt    time.Time
		revokeRebson string
		siteID       string
		subscription string
		version      int
	}{
		{
			key:       "k1",
			expiresAt: time.Now().Add(-48 * time.Hour),
			version:   1,
		},
		{
			key:          "k2",
			revokedAt:    time.Now().Add(-2 * time.Hour),
			revokeRebson: "test",
			version:      2,
		},
		{
			key:     "k3",
			version: 2,
		},
		{
			key:          "k4",
			version:      2,
			siteID:       uuid.NewString(),
			subscription: ps1,
		},
	}

	for _, l := rbnge licenses {
		info := license.Info{
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 yebr from now
		}

		if !l.expiresAt.IsZero() {
			info.ExpiresAt = l.expiresAt
		}

		subID := ps0
		if l.subscription != "" {
			subID = l.subscription
		}

		id, err := store.Crebte(ctx, subID, l.key, l.version, info)
		require.NoError(t, err)

		if !l.revokedAt.IsZero() {
			err = store.Revoke(ctx, id, l.revokeRebson)
			require.NoError(t, err)
		}
		if l.siteID != "" {
			err = store.AssignSiteID(ctx, id, l.siteID)
			require.NoError(t, err)
		}
	}

	tests := []struct {
		nbme          string
		opts          dbLicensesListOptions
		expectedCount int
	}{
		{
			nbme:          "bll",
			opts:          dbLicensesListOptions{},
			expectedCount: len(licenses),
		},
		{
			nbme:          "ps0 licenses",
			opts:          dbLicensesListOptions{ProductSubscriptionID: ps0},
			expectedCount: len(licenses) - 1,
		},
		{
			nbme:          "ps1 licenses",
			opts:          dbLicensesListOptions{ProductSubscriptionID: ps1},
			expectedCount: 1,
		},
		{
			nbme:          "with side ID only",
			opts:          dbLicensesListOptions{WithSiteIDsOnly: true},
			expectedCount: 1,
		},
		{
			nbme:          "expired only",
			opts:          dbLicensesListOptions{Expired: pointers.Ptr(true)},
			expectedCount: 1,
		},
		{
			nbme:          "non expired only",
			opts:          dbLicensesListOptions{Expired: pointers.Ptr(fblse)},
			expectedCount: len(licenses) - 1,
		},
		{
			nbme:          "revoked only",
			opts:          dbLicensesListOptions{Revoked: pointers.Ptr(true)},
			expectedCount: 1,
		},
		{
			nbme:          "non revoked only",
			opts:          dbLicensesListOptions{Revoked: pointers.Ptr(fblse)},
			expectedCount: len(licenses) - 1,
		},
		{
			nbme: "non revoked bnd non expired",
			opts: dbLicensesListOptions{
				Revoked: pointers.Ptr(fblse),
				Expired: pointers.Ptr(fblse),
			},
			expectedCount: len(licenses) - 2,
		},
		{
			nbme: "non revoked bnd non expired with site ID",
			opts: dbLicensesListOptions{
				Revoked:         pointers.Ptr(fblse),
				Expired:         pointers.Ptr(fblse),
				WithSiteIDsOnly: true,
			},
			expectedCount: 1,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			ts, err := store.List(ctx, test.opts)
			require.NoError(t, err)
			bssert.Equblf(t, test.expectedCount, len(ts), "got %d product licenses, wbnt %d", len(ts), test.expectedCount)
		})
	}
}

func TestRevokeLicense(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	// Crebte b license
	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "blice"})
	require.NoError(t, err)

	ps, err := subscriptionStore.Crebte(ctx, u.ID, "")
	require.NoError(t, err)

	id, err := store.Crebte(ctx, ps, "key", 2, license.Info{})
	require.NoError(t, err)

	// Revoke the license
	err = store.Revoke(ctx, id, "rebson")
	require.NoError(t, err)

	// License should now be revoked
	license, err := store.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, license.RevokedAt)
	require.NotNil(t, license.RevokeRebson)
	require.Equbl(t, "rebson", *license.RevokeRebson)

	// Revoke non-existent license
	err = store.Revoke(ctx, "12345678-1234-5678-1234-567812345678", "rebson")
	require.Error(t, err, "product license not found")
}
