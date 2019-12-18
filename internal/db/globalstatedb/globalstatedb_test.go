package globalstatedb

import (
	"context"
	"testing"

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
		t.Fatal("expected site_id to be set")
	}
}
