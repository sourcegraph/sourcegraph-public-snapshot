package productsubscription

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestProductLicenses_Create(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	u, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	require.NoError(t, err)

	t.Run("empty license info", func(t *testing.T) {
		ps, err := dbSubscriptions{db: db}.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)

		// This should not happen in practice but just in case to check it won't blow up
		pl, err := dbLicenses{db: db}.Create(ctx, ps, "k1", 0, license.Info{})
		require.NoError(t, err)

		got, err := dbLicenses{db: db}.GetByID(ctx, pl)
		require.NoError(t, err)
		assert.Nil(t, got.LicenseVersion)
		assert.Nil(t, got.LicenseTags)
		assert.Nil(t, got.LicenseUserCount)
		assert.Nil(t, got.LicenseExpiresAt)
	})

	ps, err := dbSubscriptions{db: db}.Create(ctx, u.ID, "")
	require.NoError(t, err)

	now := timeutil.Now()
	info := license.Info{
		Tags:      []string{"true-up"},
		UserCount: 10,
		ExpiresAt: now,
	}
	pl, err := dbLicenses{db: db}.Create(ctx, ps, "k2", 1, info)
	require.NoError(t, err)

	got, err := dbLicenses{db: db}.GetByID(ctx, pl)
	require.NoError(t, err)
	assert.Equal(t, pl, got.ID)
	assert.Equal(t, ps, got.ProductSubscriptionID)
	assert.Equal(t, "k2", got.LicenseKey)

	require.NotNil(t, got.LicenseVersion)
	assert.Equal(t, 1, *got.LicenseVersion)
	require.NotNil(t, got.LicenseTags)
	assert.Equal(t, info.Tags, got.LicenseTags)
	require.NotNil(t, got.LicenseUserCount)
	assert.Equal(t, int(info.UserCount), *got.LicenseUserCount)
	require.NotNil(t, got.LicenseExpiresAt)
	assert.Equal(t, info.ExpiresAt, *got.LicenseExpiresAt)

	ts, err := dbLicenses{db: db}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps})
	require.NoError(t, err)
	assert.Len(t, ts, 1)

	ts, err = dbLicenses{db: db}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: "69da12d5-323c-4e42-9d44-cc7951639bca" /* invalid */})
	require.NoError(t, err)
	assert.Len(t, ts, 0)
}

func TestProductLicenses_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	u1, err := db.Users().Create(ctx, database.NewUser{Username: "u1"})
	if err != nil {
		t.Fatal(err)
	}

	ps0, err := dbSubscriptions{db: db}.Create(ctx, u1.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	ps1, err := dbSubscriptions{db: db}.Create(ctx, u1.ID, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = dbLicenses{db: db}.Create(ctx, ps0, "k1", 1, license.Info{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = dbLicenses{db: db}.Create(ctx, ps0, "n1", 1, license.Info{})
	if err != nil {
		t.Fatal(err)
	}

	{
		// List all product licenses.
		ts, err := dbLicenses{db: db}.List(ctx, dbLicensesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d product licenses, want %d", len(ts), want)
		}
		count, err := dbLicenses{db: db}.Count(ctx, dbLicensesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List ps0's product licenses.
		ts, err := dbLicenses{db: db}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps0})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d product licenses, want %d", len(ts), want)
		}
	}

	{
		// List ps1's product licenses.
		ts, err := dbLicenses{db: db}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps1})
		if err != nil {
			t.Fatal(err)
		}
		if want := 0; len(ts) != want {
			t.Errorf("got %d product licenses, want %d", len(ts), want)
		}
	}
}
