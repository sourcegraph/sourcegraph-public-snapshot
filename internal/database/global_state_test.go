package database

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGlobalState_Get(t *testing.T) {
	ctx := context.Background()
	store := testGlobalStateStore(t)

	// Test pre-initialization
	config1, err := store.Get(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, config1.SiteID)
	require.False(t, config1.Initialized)
	require.Nil(t, config1.IsLicenseValid)

	// Test post-initialization
	_, err = store.EnsureInitialized(ctx)
	require.NoError(t, err)

	config2, err := store.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, config1.SiteID, config2.SiteID)
	require.True(t, config2.Initialized)
}

func TestGlobalState_SiteInitialized(t *testing.T) {
	ctx := context.Background()
	store := testGlobalStateStore(t)

	// Test pre-initialization
	siteInitialized, err := store.SiteInitialized(ctx)
	require.NoError(t, err)
	require.False(t, siteInitialized)

	// Test post-initialization
	_, err = store.EnsureInitialized(ctx)
	require.NoError(t, err)
	siteInitialized, err = store.SiteInitialized(ctx)
	require.NoError(t, err)
	require.True(t, siteInitialized)
}

func TestGlobalState_Update(t *testing.T) {
	ctx := context.Background()
	store := testGlobalStateStore(t)

	err := store.Update(ctx, true)
	require.NoError(t, err)

	state, err := store.Get(ctx)
	require.NoError(t, err)
	require.True(t, *state.IsLicenseValid)

	err = store.Update(ctx, false)
	require.NoError(t, err)
	state, err = store.Get(ctx)
	require.NoError(t, err)
	require.False(t, *state.IsLicenseValid)
}

func TestGlobalState_PrunesValues(t *testing.T) {
	ctx := context.Background()
	store := testGlobalStateStore(t)

	err := store.(*globalStateStore).Exec(ctx, sqlf.Sprintf(`
		INSERT INTO global_state(
			site_id,
			initialized
		)
		VALUES
			('00000000-0000-0000-0000-000000000000', false),
			('00000000-0000-0000-0000-000000000001', false),
			('00000000-0000-0000-0000-000000000010', false),
			('00000000-0000-0000-0000-000000000100', false),
			('00000000-0000-0000-0000-000000001000', false),
			('00000000-0000-0000-0000-000000010000', false),
			('00000000-0000-0000-0000-000000100000', false),
			('00000000-0000-0000-0000-000001000000', false),
			('00000000-0000-0000-0000-000010000000', true),
			('00000000-0000-0000-0000-000100000000', false),
			('00000000-0000-0000-0000-001000000000', false),
			('00000000-0000-0000-0000-010000000000', false),
			('00000000-0000-0000-0000-100000000000', false)
	`))
	require.NoError(t, err)

	config, err := store.Get(ctx)
	require.NoError(t, err)

	expectedSiteID := "00000000-0000-0000-0000-000000000000"
	require.Equal(t, expectedSiteID, config.SiteID)
	require.True(t, config.Initialized)
}

func testGlobalStateStore(t *testing.T) GlobalStateStore {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	return NewDB(logger, dbtest.NewDB(logger, t)).GlobalState()
}
