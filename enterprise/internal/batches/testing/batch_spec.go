package testing

import (
	"context"
	"testing"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

type CreateBatchSpecer interface {
	CreateBatchSpec(ctx context.Context, batchSpec *btypes.BatchSpec) error
}

func CreateBatchSpec(t *testing.T, ctx context.Context, store CreateBatchSpecer, name string, userID int32) *btypes.BatchSpec {
	t.Helper()

	s := &btypes.BatchSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: btypes.BatchSpecFields{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: btypes.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
	}

	if err := store.CreateBatchSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}
