package testing

import (
	"context"
	"testing"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

type CreateBatchSpecer interface {
	CreateBatchSpec(ctx context.Context, batchSpec *btypes.BatchSpec) error
}

func CreateBatchSpec(t *testing.T, ctx context.Context, store CreateBatchSpecer, name string, userID int32, bcID int64) *btypes.BatchSpec {
	t.Helper()

	s := &btypes.BatchSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: &batcheslib.BatchSpec{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: &batcheslib.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
		BatchChangeID: bcID,
	}

	if err := store.CreateBatchSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}
