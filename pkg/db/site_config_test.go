package db

import (
	"testing"
)

func TestSiteConfig_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()
	config, err := SiteConfig.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.AppID == "" {
		t.Fatal("expected app_id to be set")
	}
}
