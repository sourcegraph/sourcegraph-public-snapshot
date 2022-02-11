package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGlobalState_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	config, err := db.GlobalState().Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.SiteID == "" {
		t.Fatal("expected site_id to be set")
	}
}
