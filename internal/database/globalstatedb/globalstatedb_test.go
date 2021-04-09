package globalstatedb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func init() {
	dbtesting.DBNameSuffix = "globalstatedb"
}

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
		t.Fatal("expected site_id to be set")
	}
}

func TestEnsureInitialized_EmptyGlobalStateTable(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtesting.GetDB(t)
	ctx := context.Background()

	// Ensure that we are starting with an empty table.
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM global_state").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("Expected global_state rows count: 0, but got %d", count)
	}

	dbStore := basestore.NewWithDB(db, sql.TxOptions{})

	// First time that EnsureInitialized is invoked, it should write a single row and set
	// initialized=true.
	alreadyInitialized, err := EnsureInitialized(ctx, dbStore)
	if err != nil {
		t.Fatal(err)
	}

	if !alreadyInitialized {
		t.Errorf("Expected alreadyInitialized: true, but got %v", alreadyInitialized)
	}

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM global_state").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Expected global_state rows count: 1, but got %d", count)
	}

	var existingSiteID string
	err = db.QueryRowContext(ctx, "SELECT site_id FROM global_state").Scan(&existingSiteID)
	if err != nil {
		t.Fatal(err)
	}

	// Second time that EnsureInitialized is invoked, it should not perform any new writes.
	alreadyInitialized, err = EnsureInitialized(ctx, dbStore)
	if err != nil {
		t.Fatal(err)
	}

	if !alreadyInitialized {
		t.Errorf("Expected alreadyInitialized: true, but got %v", alreadyInitialized)
	}

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM global_state").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Expected global_state rows count: 1, but got %d", count)
	}

	var newSiteID string
	err = db.QueryRowContext(ctx, "SELECT site_id FROM global_state").Scan(&newSiteID)
	if err != nil {
		t.Fatal(err)
	}

	if existingSiteID != newSiteID {
		t.Fatalf("Expected site_id: %q, but got %q. site_id should not have changed on multiple invocations of EnsureInitiazlied", existingSiteID, newSiteID)
	}
}

func TestEnsureInitialized_NonEmptyGlobalStateTableInitializedFalse(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtesting.GetDB(t)
	ctx := context.Background()

	dbStore := basestore.NewWithDB(db, sql.TxOptions{})

	err := tryInsertNew(ctx, dbStore)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that we are starting with initialized=false.
	var initialized bool
	err = db.QueryRowContext(ctx, "SELECT initialized FROM global_state").Scan(&initialized)
	if err != nil {
		t.Fatal(err)
	}

	if initialized {
		t.Fatalf("Expected initialized in global_state: false, but got %t", initialized)
	}

	alreadyInitialized, err := EnsureInitialized(ctx, dbStore)
	if err != nil {
		t.Fatal(err)
	}

	if !alreadyInitialized {
		t.Errorf("Expected alreadyInitialized: true, but got %v", alreadyInitialized)
	}

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM global_state WHERE initialized=true").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Expected global_state rows count: 1, but got %d", count)
	}
}
