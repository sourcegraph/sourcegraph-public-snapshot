package globalstatedb

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestGet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()
	config, err := Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.SiteID == "" {
		t.Fatal("want SiteID to be set")
	}
	if config.Initialized {
		t.Fatal("want Initialized == false")
	}
	if config.InitializedAt != nil {
		t.Fatal("want InitializedAt == nil")
	}
}

func TestEnsureInitialized(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	alreadyInitialized, err := EnsureInitialized(ctx, dbconn.Global)
	if err != nil {
		t.Fatal(err)
	}
	if alreadyInitialized {
		t.Fatal("alreadyInitialized")
	}

	config, err := Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.SiteID == "" {
		t.Fatal("want SiteID to be set")
	}
	if !config.Initialized {
		t.Fatal("want Initialized == true")
	}
	if config.InitializedAt == nil {
		t.Fatal("want InitializedAt != nil")
	}
}
