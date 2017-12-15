package localstore

import (
	"testing"
)

func Test_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()
	config, err := Config.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.AppID == "" {
		t.Fatal("expected app_id to be set")
	}
}
