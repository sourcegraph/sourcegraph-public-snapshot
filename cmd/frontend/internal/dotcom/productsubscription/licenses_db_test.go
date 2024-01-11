package productsubscription

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestProductLicenses_Create(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	u, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	require.NoError(t, err)

	t.Run("empty license info", func(t *testing.T) {
		ps, err := subscriptionStore.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)

		// This should not happen in practice but just in case to check it won't blow up
		id, err := store.Create(ctx, ps, "k1", 0, license.Info{})
		require.NoError(t, err)

		got, err := store.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, got.LicenseVersion)
		assert.Nil(t, got.LicenseTags)
		assert.Nil(t, got.LicenseUserCount)
		assert.Nil(t, got.LicenseExpiresAt)
	})

	ps, err := subscriptionStore.Create(ctx, u.ID, "")
	require.NoError(t, err)

	now := timeutil.Now()
	licenseV1 := license.Info{
		Tags:      []string{"true-up"},
		UserCount: 10,
		ExpiresAt: now,
	}

	sfSubID := "AE9108431908421"
	sfOpID := "0A8908908A800F"

	licenseV2 := license.Info{
		Tags:                     []string{"true-up"},
		UserCount:                10,
		ExpiresAt:                now,
		SalesforceSubscriptionID: &sfSubID,
		SalesforceOpportunityID:  &sfOpID,
	}

	for v, info := range []license.Info{licenseV1, licenseV2} {
		t.Run(fmt.Sprintf("Test v%d", v+1), func(t *testing.T) {
			version := v + 1
			key := fmt.Sprintf("key%d", version)
			pl, err := store.Create(ctx, ps, key, version, info)
			require.NoError(t, err)

			got, err := store.GetByID(ctx, pl)
			require.NoError(t, err)
			assert.Equal(t, pl, got.ID)
			assert.Equal(t, ps, got.ProductSubscriptionID)
			assert.Equal(t, key, got.LicenseKey)
			assert.NotNil(t, got.LicenseVersion)
			assert.Equal(t, version, int(*got.LicenseVersion))
			require.NotNil(t, got.LicenseTags)
			assert.Equal(t, info.Tags, got.LicenseTags)
			require.NotNil(t, got.LicenseUserCount)
			assert.Equal(t, int(info.UserCount), *got.LicenseUserCount)
			require.NotNil(t, got.LicenseExpiresAt)
			assert.Equal(t, info.ExpiresAt, *got.LicenseExpiresAt)

			ts, err := store.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps})
			require.NoError(t, err)
			assert.Len(t, ts, version)

			// Invalid subscription ID.
			ts, err = store.List(ctx, dbLicensesListOptions{ProductSubscriptionID: "69da12d5-323c-4e42-9d44-cc7951639bca"})
			require.NoError(t, err)
			assert.Len(t, ts, 0)
		})
	}
}

func TestGetByToken(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	store := dbLicenses{db: db}

	u, err := db.Users().Create(ctx, database.NewUser{Username: "user"})
	require.NoError(t, err)

	lc := insertLicense(t, ctx, db, u, "key")
	require.NotNil(t, lc.LicenseCheckToken)

	tests := []struct {
		name      string
		token     string
		want      *string
		wantError error
	}{
		{
			name:  "ok",
			token: license.GenerateLicenseKeyBasedAccessToken("key"),
			want:  &lc.ID,
		},
		{
			name:      "invalid non-hex token",
			token:     "invalid",
			wantError: errTokenInvalid,
		},
		{
			name:      "no match found",
			token:     hex.EncodeToString([]byte("key")),
			wantError: errTokenInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByAccessToken(ctx, tt.token)
			if tt.wantError != nil {
				require.Equal(t, tt.wantError, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, *tt.want, got.ID)
			}
		})
	}
}

func TestAssignSiteID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	store := dbLicenses{db: db}

	u, err := db.Users().Create(ctx, database.NewUser{Username: "user"})
	require.NoError(t, err)

	license := insertLicense(t, ctx, db, u, "key")

	siteID := uuid.NewString()
	_, err = store.AssignSiteID(ctx, license, siteID)
	require.NoError(t, err)

	license, err = store.GetByID(ctx, license.ID)
	require.NoError(t, err)

	require.NotNil(t, license.SiteID)
	require.Equal(t, siteID, *license.SiteID)
}

func insertLicense(t *testing.T, ctx context.Context, db database.DB, user *types.User, licenseKey string) *dbLicense {
	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	ps, err := subscriptionStore.Create(ctx, user.ID, "")
	require.NoError(t, err)

	sfSubID := "sf_sub_id"
	sfOpID := "sf_op_id"
	id, err := store.Create(ctx, ps, licenseKey, 2, license.Info{
		SalesforceSubscriptionID: &sfSubID,
		SalesforceOpportunityID:  &sfOpID,
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
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	u1, err := db.Users().Create(ctx, database.NewUser{Username: "u1"})
	require.NoError(t, err)

	ps0, err := subscriptionStore.Create(ctx, u1.ID, "")
	require.NoError(t, err)
	ps1, err := subscriptionStore.Create(ctx, u1.ID, "")
	require.NoError(t, err)

	licenses := []struct {
		key          string
		expiresAt    time.Time
		revokedAt    time.Time
		revokeReason string
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
			revokeReason: "test",
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

	for _, l := range licenses {
		info := license.Info{
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 year from now
		}

		if !l.expiresAt.IsZero() {
			info.ExpiresAt = l.expiresAt
		}

		subID := ps0
		if l.subscription != "" {
			subID = l.subscription
		}

		id, err := store.Create(ctx, subID, l.key, l.version, info)
		require.NoError(t, err)

		if !l.revokedAt.IsZero() {
			err = store.Revoke(ctx, id, l.revokeReason)
			require.NoError(t, err)
		}
		if l.siteID != "" {
			_, err = store.AssignSiteID(ctx, &dbLicense{ID: id}, l.siteID)
			require.NoError(t, err)
		}
	}

	tests := []struct {
		name          string
		opts          dbLicensesListOptions
		expectedCount int
	}{
		{
			name:          "all",
			opts:          dbLicensesListOptions{},
			expectedCount: len(licenses),
		},
		{
			name:          "ps0 licenses",
			opts:          dbLicensesListOptions{ProductSubscriptionID: ps0},
			expectedCount: len(licenses) - 1,
		},
		{
			name:          "ps1 licenses",
			opts:          dbLicensesListOptions{ProductSubscriptionID: ps1},
			expectedCount: 1,
		},
		{
			name:          "with side ID only",
			opts:          dbLicensesListOptions{WithSiteIDsOnly: true},
			expectedCount: 1,
		},
		{
			name:          "expired only",
			opts:          dbLicensesListOptions{Expired: pointers.Ptr(true)},
			expectedCount: 1,
		},
		{
			name:          "non expired only",
			opts:          dbLicensesListOptions{Expired: pointers.Ptr(false)},
			expectedCount: len(licenses) - 1,
		},
		{
			name:          "revoked only",
			opts:          dbLicensesListOptions{Revoked: pointers.Ptr(true)},
			expectedCount: 1,
		},
		{
			name:          "non revoked only",
			opts:          dbLicensesListOptions{Revoked: pointers.Ptr(false)},
			expectedCount: len(licenses) - 1,
		},
		{
			name: "non revoked and non expired",
			opts: dbLicensesListOptions{
				Revoked: pointers.Ptr(false),
				Expired: pointers.Ptr(false),
			},
			expectedCount: len(licenses) - 2,
		},
		{
			name: "non revoked and non expired with site ID",
			opts: dbLicensesListOptions{
				Revoked:         pointers.Ptr(false),
				Expired:         pointers.Ptr(false),
				WithSiteIDsOnly: true,
			},
			expectedCount: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts, err := store.List(ctx, test.opts)
			require.NoError(t, err)
			assert.Equalf(t, test.expectedCount, len(ts), "got %d product licenses, want %d", len(ts), test.expectedCount)
		})
	}
}

func TestRevokeLicense(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	// Create a license
	u, err := db.Users().Create(ctx, database.NewUser{Username: "alice"})
	require.NoError(t, err)

	ps, err := subscriptionStore.Create(ctx, u.ID, "")
	require.NoError(t, err)

	id, err := store.Create(ctx, ps, "key", 2, license.Info{})
	require.NoError(t, err)

	// Revoke the license
	err = store.Revoke(ctx, id, "reason")
	require.NoError(t, err)

	// License should now be revoked
	license, err := store.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, license.RevokedAt)
	require.NotNil(t, license.RevokeReason)
	require.Equal(t, "reason", *license.RevokeReason)

	// Revoke non-existent license
	err = store.Revoke(ctx, "12345678-1234-5678-1234-567812345678", "reason")
	require.Error(t, err, "product license not found")
}

func TestRenderLicenseCreationSlackMessage(t *testing.T) {
	staticExpiresAt, err := time.Parse(time.RFC3339, "2023-02-24T14:48:30Z")
	require.NoError(t, err)

	// Typical case is that license expires in the future, so emulate now to
	// be some time before that
	staticNow := staticExpiresAt.Add(-72 * time.Hour)

	message := renderLicenseCreationSlackMessage(
		staticNow,
		&types.User{},
		"1234", 123,
		&staticExpiresAt,
		license.Info{})
	autogold.Expect("\nA new license was created by ** for subscription <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/1234|1234>:\n\n• *License version*: 123\n• *Expiration (UTC)*: Feb 24, 2023 2:48pm UTC (3.0 days remaining)\n• *Expiration (PT)*: Feb 24, 2023 6:48am PST\n• *User count*: 0\n• *License tags*: ``\n• *Salesforce subscription ID*: unknown\n• *Salesforce opportunity ID*: <https://sourcegraph2020.lightning.force.com/lightning/r/Opportunity/unknown/view|unknown>\n\nReply with a :approved_stamp: when this is approved\nReply with a :white_check_mark: when this has been sent to the customer\n").Equal(t, message)
}
