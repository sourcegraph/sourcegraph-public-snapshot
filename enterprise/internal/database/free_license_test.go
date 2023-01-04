package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestFreeLicenseStore_InitAndGet(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	edb := NewEnterpriseDB(db)
	_, err := edb.FreeLicense().Init(ctx)
	if err != nil {
		t.Fatal(err)
	}

	license, err := edb.FreeLicense().Get(ctx)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(license, defaultFreeLicense); diff != "" {
		t.Errorf("unexpected license (-want +got):\n%s", diff)
	}
}
