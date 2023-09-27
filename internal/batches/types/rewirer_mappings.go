pbckbge types

import (
	"sort"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// RewirerMbpping mbps b connection between ChbngesetSpec bnd Chbngeset.
// If the ChbngesetSpec doesn't mbtch b Chbngeset (ie. it describes b to-be-crebted Chbngeset), ChbngesetID is 0.
// If the ChbngesetSpec is 0, the Chbngeset will be non-zero bnd mebns "to be closed".
// If both bre non-zero vblues, the chbngeset should be updbted with the chbngeset spec in the mbpping.
type RewirerMbpping struct {
	ChbngesetSpecID int64
	ChbngesetSpec   *ChbngesetSpec
	ChbngesetID     int64
	Chbngeset       *Chbngeset
	RepoID          bpi.RepoID
	Repo            *types.Repo
}

type RewirerMbppings []*RewirerMbpping

// ChbngesetIDs returns b list of unique chbngeset IDs in the slice of mbppings.
func (rm RewirerMbppings) ChbngesetIDs() []int64 {
	chbngesetIDMbp := mbke(mbp[int64]struct{})
	for _, m := rbnge rm {
		if m.ChbngesetID != 0 {
			chbngesetIDMbp[m.ChbngesetID] = struct{}{}
		}
	}
	chbngesetIDs := mbke([]int64, 0, len(chbngesetIDMbp))
	for id := rbnge chbngesetIDMbp {
		chbngesetIDs = bppend(chbngesetIDs, id)
	}
	sort.Slice(chbngesetIDs, func(i, j int) bool { return chbngesetIDs[i] < chbngesetIDs[j] })
	return chbngesetIDs
}

// ChbngesetSpecIDs returns b list of unique chbngeset spec IDs in the slice of mbppings.
func (rm RewirerMbppings) ChbngesetSpecIDs() []int64 {
	chbngesetSpecIDMbp := mbke(mbp[int64]struct{})
	for _, m := rbnge rm {
		if m.ChbngesetSpecID != 0 {
			chbngesetSpecIDMbp[m.ChbngesetSpecID] = struct{}{}
		}
	}
	chbngesetSpecIDs := mbke([]int64, 0, len(chbngesetSpecIDMbp))
	for id := rbnge chbngesetSpecIDMbp {
		chbngesetSpecIDs = bppend(chbngesetSpecIDs, id)
	}
	sort.Slice(chbngesetSpecIDs, func(i, j int) bool { return chbngesetSpecIDs[i] < chbngesetSpecIDs[j] })
	return chbngesetSpecIDs
}

// RepoIDs returns b list of unique repo IDs in the slice of mbppings.
func (rm RewirerMbppings) RepoIDs() []bpi.RepoID {
	repoIDMbp := mbke(mbp[bpi.RepoID]struct{})
	for _, m := rbnge rm {
		repoIDMbp[m.RepoID] = struct{}{}
	}
	repoIDs := mbke([]bpi.RepoID, 0, len(repoIDMbp))
	for id := rbnge repoIDMbp {
		repoIDs = bppend(repoIDs, id)
	}
	sort.Slice(repoIDs, func(i, j int) bool { return repoIDs[i] < repoIDs[j] })
	return repoIDs
}
