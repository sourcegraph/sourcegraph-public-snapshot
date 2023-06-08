package productsubscription

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestProductLicenses_Create(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	store := dbLicenses{db: db}

	u, err := db.Users().Create(ctx, database.NewUser{Username: "user"})
	require.NoError(t, err)

	license := insertLicense(t, ctx, db, u, "key")
	require.NotNil(t, license.LicenseCheckToken)

	tests := []struct {
		name      string
		token     string
		want      *string
		wantError error
	}{
		{
			name:  "ok",
			token: hex.EncodeToString(*license.LicenseCheckToken),
			want:  &license.ID,
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
			got, err := store.GetByToken(ctx, tt.token)
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	store := dbLicenses{db: db}

	u, err := db.Users().Create(ctx, database.NewUser{Username: "user"})
	require.NoError(t, err)

	license := insertLicense(t, ctx, db, u, "key")

	siteID, err := uuid.NewV4()
	require.NoError(t, err)
	err = store.AssignSiteID(ctx, license.ID, siteID.String())
	require.NoError(t, err)

	license, err = store.GetByID(ctx, license.ID)
	require.NoError(t, err)

	require.NotNil(t, license.SiteID)
	require.Equal(t, siteID.String(), *license.SiteID)
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	subscriptionStore := dbSubscriptions{db: db}
	store := dbLicenses{db: db}

	u1, err := db.Users().Create(ctx, database.NewUser{Username: "u1"})
	require.NoError(t, err)

	ps0, err := subscriptionStore.Create(ctx, u1.ID, "")
	require.NoError(t, err)
	ps1, err := subscriptionStore.Create(ctx, u1.ID, "")
	require.NoError(t, err)

	_, err = store.Create(ctx, ps0, "k1", 1, license.Info{})
	require.NoError(t, err)
	_, err = store.Create(ctx, ps0, "n1", 1, license.Info{})
	require.NoError(t, err)

	{
		// List all product licenses.
		ts, err := store.List(ctx, dbLicensesListOptions{})
		require.NoError(t, err)
		assert.Equalf(t, 2, len(ts), "got %d product licenses, want 2", len(ts))
		count, err := store.Count(ctx, dbLicensesListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	}

	{
		// List ps0's product licenses.
		ts, err := store.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps0})
		require.NoError(t, err)
		assert.Equalf(t, 2, len(ts), "got %d product licenses, want 2", len(ts))
	}

	{
		// List ps1's product licenses.
		ts, err := store.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps1})
		require.NoError(t, err)
		assert.Equalf(t, 0, len(ts), "got %d product licenses, want 0", len(ts))
	}
}

func TestRevokeLicense(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
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
	if err := store.Revoke(ctx, id, "reason"); err != nil {
		t.Fatal(err)
	}

	// License should now be revoked
	license, err := store.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, license.RevokedAt)
	require.NotNil(t, license.RevokeReason)
	require.Equal(t, "reason", *license.RevokeReason)
}
