package lsifstore

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestIDsWithMeta(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t))
	store := New(codeIntelDB, &observation.TestContext)
	ctx := context.Background()

	// TODO - setup test

	_, err := store.IDsWithMeta(ctx, []int{})
	if err != nil {
		t.Fatalf("failed to find upload IDs with metadata: %s", err)
	}

	// TODO - assert
	t.Fail()
}

func TestReconcileCandidates(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t))
	store := New(codeIntelDB, &observation.TestContext)
	ctx := context.Background()

	// TODO - setup test

	_, err := store.ReconcileCandidates(ctx, 50)
	if err != nil {
		t.Fatalf("failed to get candidate IDs for reconciliation: %s", err)
	}

	// TODO - assert
	t.Fail()
}
