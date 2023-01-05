package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestFreeLicenseStore_Init(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	edb := NewEnterpriseDB(db)
	_, err := edb.FreeLicense().Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
