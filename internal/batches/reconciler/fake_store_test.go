package reconciler

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
)

type mockMissingErr struct {
	mockName string
}

func (e mockMissingErr) Error() string {
	return fmt.Sprintf("FakeStore is missing mock for %s", e.mockName)
}

type FakeStore struct {
	GetBatchChangeMock func(context.Context, store.GetBatchChangeOpts) (*btypes.BatchChange, error)
}

func (fs *FakeStore) GetBatchChange(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error) {
	if fs.GetBatchChangeMock != nil {
		return fs.GetBatchChangeMock(ctx, opts)
	}
	return nil, mockMissingErr{"GetBatchChange"}
}
