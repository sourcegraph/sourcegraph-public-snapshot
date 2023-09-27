pbckbge reconciler

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

type mockMissingErr struct {
	mockNbme string
}

func (e mockMissingErr) Error() string {
	return fmt.Sprintf("FbkeStore is missing mock for %s", e.mockNbme)
}

type FbkeStore struct {
	GetBbtchChbngeMock func(context.Context, store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error)
}

func (fs *FbkeStore) GetBbtchChbnge(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
	if fs.GetBbtchChbngeMock != nil {
		return fs.GetBbtchChbngeMock(ctx, opts)
	}
	return nil, mockMissingErr{"GetBbtchChbnge"}
}
