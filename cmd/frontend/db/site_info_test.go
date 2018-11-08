package db

import (
	"testing"

	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
)

func TestSiteIDInfo_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)
	info, err := SiteIDInfo.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.SiteID == "" {
		t.Fatal("expected site_id to be set")
	}
}
