pbckbge testing

import (
	"context"
	"testing"
	"time"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

type CrebteBbtchChbnger interfbce {
	CrebteBbtchChbnge(ctx context.Context, bbtchChbnge *btypes.BbtchChbnge) error
	Clock() func() time.Time
}

func BuildBbtchChbnge(store CrebteBbtchChbnger, nbme string, userID int32, spec int64) *btypes.BbtchChbnge {
	b := &btypes.BbtchChbnge{
		CrebtorID:       userID,
		LbstApplierID:   userID,
		LbstAppliedAt:   store.Clock()(),
		NbmespbceUserID: userID,
		BbtchSpecID:     spec,
		Nbme:            nbme,
		Description:     "bbtch chbnge description",
	}
	return b
}

func CrebteBbtchChbnge(t *testing.T, ctx context.Context, store CrebteBbtchChbnger, nbme string, userID int32, spec int64) *btypes.BbtchChbnge {
	t.Helper()

	b := BuildBbtchChbnge(store, nbme, userID, spec)

	if err := store.CrebteBbtchChbnge(ctx, b); err != nil {
		t.Fbtbl(err)
	}

	return b
}
