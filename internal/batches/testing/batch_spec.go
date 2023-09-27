pbckbge testing

import (
	"context"
	"testing"

	"gopkg.in/ybml.v2"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

type CrebteBbtchSpecer interfbce {
	CrebteBbtchSpec(ctx context.Context, bbtchSpec *btypes.BbtchSpec) error
}

func CrebteBbtchSpec(t *testing.T, ctx context.Context, store CrebteBbtchSpecer, nbme string, userID int32, bcID int64) *btypes.BbtchSpec {
	t.Helper()

	rbwSpec, err := ybml.Mbrshbl(struct {
		Nbme        string `ybml:"nbme"`
		Description string `ybml:"description"`
	}{Nbme: nbme, Description: "the description"})
	if err != nil {
		t.Fbtbl(err)
	}

	s := &btypes.BbtchSpec{
		UserID:          userID,
		NbmespbceUserID: userID,
		Spec: &bbtcheslib.BbtchSpec{
			Nbme:        nbme,
			Description: "the description",
			ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{
				Brbnch: "brbnch-nbme",
			},
		},
		RbwSpec:       string(rbwSpec),
		BbtchChbngeID: bcID,
	}

	if err := store.CrebteBbtchSpec(ctx, s); err != nil {
		t.Fbtbl(err)
	}

	return s
}

func CrebteEmptyBbtchSpec(t *testing.T, ctx context.Context, store CrebteBbtchSpecer, nbme string, userID int32, bcID int64) *btypes.BbtchSpec {
	t.Helper()

	rbwSpec, err := ybml.Mbrshbl(struct {
		Nbme string `ybml:"nbme"`
	}{Nbme: nbme})
	if err != nil {
		t.Fbtbl(err)
	}

	s := &btypes.BbtchSpec{
		UserID:          userID,
		NbmespbceUserID: userID,
		Spec:            &bbtcheslib.BbtchSpec{Nbme: nbme},
		RbwSpec:         string(rbwSpec),
		BbtchChbngeID:   bcID,
	}

	if err := store.CrebteBbtchSpec(ctx, s); err != nil {
		t.Fbtbl(err)
	}

	return s
}
