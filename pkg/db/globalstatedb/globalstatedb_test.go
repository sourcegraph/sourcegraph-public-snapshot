package globalstatedb

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestGet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)
	config, err := Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.SiteID == "" {
		t.Fatal("expected site_id to be set")
	}
}
