package codeintel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"

	stores "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSCIPMigrator(t *testing.T) {
	logger := logtest.Scoped(t)
	rawDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, rawDB)
	codeIntelDB := stores.NewCodeIntelDB(logger, rawDB)
	store := basestore.NewWithHandle(db.Handle())
	codeIntelStore := basestore.NewWithHandle(codeIntelDB.Handle())
	migrator := NewSCIPMigrator(store, codeIntelStore)
	ctx := context.Background()

	contents, err := os.ReadFile("./testdata/lsif.sql")
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}
	if _, err := codeIntelDB.ExecContext(ctx, string(contents)); err != nil {
		t.Fatalf("unexpected error executing test file: %s", err)
	}

	fmt.Printf("YUP!!\n") // TODO
	_ = migrator
}
