package database

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGlobalState_Get(t *testing.T) {
	ctx := context.Background()
	store := testGlobalStateStore(t)

	// Test pre-initialization
	config1, err := store.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config1.SiteID == "" {
		t.Fatal("expected site_id to be set")
	}
	if config1.Initialized {
		t.Fatal("site expected to be uninitialized")
	}

	// Test post-initialization
	if _, err := store.EnsureInitialized(ctx); err != nil {
		t.Fatal(err)
	}

	config2, err := store.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config2.SiteID != config1.SiteID {
		t.Fatalf("unexpected site id. want=%s have=%s", config1.SiteID, config2.SiteID)
	}
	if !config2.Initialized {
		t.Fatal("site expected to be initialized")
	}
}

func TestGlobalState_SiteInitialized(t *testing.T) {
	ctx := context.Background()
	store := testGlobalStateStore(t)

	// Test pre-initialization
	siteInitialized, err := store.SiteInitialized(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if siteInitialized {
		t.Fatal("site expected to be uninitialized")
	}

	// Test post-initialization
	if _, err := store.EnsureInitialized(ctx); err != nil {
		t.Fatal(err)
	}
	siteInitialized, err = store.SiteInitialized(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !siteInitialized {
		t.Fatal("site expected to be initialized")
	}
}

func TestGlobalState_PrunesValues(t *testing.T) {
	ctx := context.Background()
	store := testGlobalStateStore(t)

	if err := store.(*globalStateStore).Exec(ctx, sqlf.Sprintf(`
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
	`)); err != nil {
		t.Fatal(err)
	}

	config, err := store.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expectedSiteID := "00000000-0000-0000-0000-000000000000"
	if config.SiteID != expectedSiteID {
		t.Fatalf("unexpected site-id. want=%s have=%s", expectedSiteID, config.SiteID)
	}
	if !config.Initialized {
		t.Fatal("expected site to be initialized")
	}
}

func testGlobalStateStore(t *testing.T) GlobalStateStore {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	return NewDB(logger, dbtest.NewDB(t)).GlobalState()
}
