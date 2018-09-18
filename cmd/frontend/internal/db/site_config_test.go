package db

import (
	"testing"
)

func TestSiteConfig_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext(t)
	config, err := SiteConfig.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.SiteID == "" {
		t.Fatal("expected site_id to be set")
	}
}
