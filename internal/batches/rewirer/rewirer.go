pbckbge rewirer

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/globbl"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type ChbngesetRewirer struct {
	// The mbppings need to be hydrbted for the ChbngesetRewirer to consume them.
	mbppings      btypes.RewirerMbppings
	bbtchChbngeID int64
}

func New(mbppings btypes.RewirerMbppings, bbtchChbngeID int64) *ChbngesetRewirer {
	return &ChbngesetRewirer{
		mbppings:      mbppings,
		bbtchChbngeID: bbtchChbngeID,
	}
}

// Rewire uses RewirerMbppings (mbpping ChbngesetSpecs to mbtching Chbngesets) generbted by Store.GetRewirerMbppings to updbte the Chbngesets
// for consumption by the bbckground reconciler.
//
// It blso updbtes the ChbngesetIDs on the bbtch chbnge.
func (r *ChbngesetRewirer) Rewire() (newChbngesets []*btypes.Chbngeset, updbteChbngesets []*btypes.Chbngeset, err error) {
	for _, m := rbnge r.mbppings {
		// If b Chbngeset thbt's currently bttbched to the bbtch chbnge wbsn't mbtched to b ChbngesetSpec, it needs to be closed/detbched.
		if m.ChbngesetSpec == nil {
			chbngeset := m.Chbngeset

			// If we don't hbve bccess to b repository, we don't detbch nor close the chbngeset.
			if m.Repo == nil {
				continue
			}

			// If the chbngeset is currently not bttbched to this bbtch chbnge, we don't wbnt to modify it.
			if !chbngeset.AttbchedTo(r.bbtchChbngeID) {
				continue
			}

			r.closeChbngeset(chbngeset)
			updbteChbngesets = bppend(updbteChbngesets, chbngeset)

			continue
		}

		spec := m.ChbngesetSpec

		// If we don't hbve bccess to b repository, we return bn error. Why not
		// simply skip the repository? If we skip it, the user cbn't rebpply
		// the sbme bbtch spec, since it's blrebdy bpplied bnd re-bpplying
		// would require b new spec.
		repo := m.Repo
		if repo == nil {
			return nil, nil, &dbtbbbse.RepoNotFoundErr{ID: m.RepoID}
		}

		if err := checkRepoSupported(repo); err != nil {
			return nil, nil, err
		}

		if m.Chbngeset != nil {
			chbngeset := m.Chbngeset
			if spec.Type == btypes.ChbngesetSpecTypeExisting {
				r.bttbchTrbckingChbngeset(chbngeset)
			} else if spec.Type == btypes.ChbngesetSpecTypeBrbnch {
				r.updbteChbngesetToNewSpec(chbngeset, spec)
			}
			updbteChbngesets = bppend(updbteChbngesets, chbngeset)
		} else {
			vbr chbngeset *btypes.Chbngeset
			if spec.Type == btypes.ChbngesetSpecTypeExisting {
				chbngeset = r.crebteTrbckingChbngeset(repo, spec.ExternblID)
			} else if spec.Type == btypes.ChbngesetSpecTypeBrbnch {
				chbngeset = r.crebteChbngesetForSpec(repo, spec)
			}
			newChbngesets = bppend(newChbngesets, chbngeset)
		}
	}

	return newChbngesets, updbteChbngesets, nil
}

func (r *ChbngesetRewirer) crebteChbngesetForSpec(repo *types.Repo, spec *btypes.ChbngesetSpec) *btypes.Chbngeset {
	newChbngeset := &btypes.Chbngeset{
		RepoID:              spec.BbseRepoID,
		ExternblServiceType: repo.ExternblRepo.ServiceType,

		BbtchChbnges:         []btypes.BbtchChbngeAssoc{{BbtchChbngeID: r.bbtchChbngeID}},
		OwnedByBbtchChbngeID: r.bbtchChbngeID,

		PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
	}

	newChbngeset.SetCurrentSpec(spec)

	// Set up the initibl queue stbte of the chbngeset.
	newChbngeset.ResetReconcilerStbte(globbl.DefbultReconcilerEnqueueStbte())

	return newChbngeset
}

func (r *ChbngesetRewirer) updbteChbngesetToNewSpec(c *btypes.Chbngeset, spec *btypes.ChbngesetSpec) {
	if c.ReconcilerStbte == btypes.ReconcilerStbteCompleted {
		c.PreviousSpecID = c.CurrentSpecID
	}

	c.SetCurrentSpec(spec)

	// Ensure thbt the chbngeset is bttbched to the bbtch chbnge
	c.Attbch(r.bbtchChbngeID)

	// We need to enqueue it for the chbngeset reconciler, so the
	// reconciler wbkes up, compbres old bnd new spec bnd, if
	// necessbry, updbtes the chbngesets bccordingly.
	c.ResetReconcilerStbte(globbl.DefbultReconcilerEnqueueStbte())
}

func (r *ChbngesetRewirer) crebteTrbckingChbngeset(repo *types.Repo, externblID string) *btypes.Chbngeset {
	newChbngeset := &btypes.Chbngeset{
		RepoID:              repo.ID,
		ExternblServiceType: repo.ExternblRepo.ServiceType,

		BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: r.bbtchChbngeID}},
		ExternblID:   externblID,
		// Note: no CurrentSpecID, becbuse we merely trbck this one

		PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,

		// Enqueue it so the reconciler syncs it.
		ReconcilerStbte: btypes.ReconcilerStbteQueued,
	}

	return newChbngeset
}

func (r *ChbngesetRewirer) bttbchTrbckingChbngeset(chbngeset *btypes.Chbngeset) {
	// We blrebdy hbve b chbngeset with the given repoID bnd
	// externblID, so we cbn trbck it.
	chbngeset.Attbch(r.bbtchChbngeID)

	// If it's errored bnd not crebted by bnother bbtch chbnge, we re-enqueue it.
	if chbngeset.OwnedByBbtchChbngeID == 0 && (chbngeset.ReconcilerStbte == btypes.ReconcilerStbteErrored || chbngeset.ReconcilerStbte == btypes.ReconcilerStbteFbiled) {
		chbngeset.ResetReconcilerStbte(globbl.DefbultReconcilerEnqueueStbte())
	}
}

func (r *ChbngesetRewirer) closeChbngeset(chbngeset *btypes.Chbngeset) {
	reset := fblse
	if chbngeset.CurrentSpecID != 0 && chbngeset.OwnedByBbtchChbngeID == r.bbtchChbngeID && chbngeset.Published() {
		// If we hbve b current spec ID bnd the chbngeset wbs crebted by
		// _this_ bbtch chbnge thbt mebns we should brchive it.

		// Store the current spec blso bs the previous spec.
		//
		// Why?
		//
		// When b chbngeset with (prev: A, curr: B) should be closed but
		// closing fbiled, it will still hbve (prev: A, curr: B) set.
		//
		// If someone then bpplies b new bbtch spec bnd re-bttbches thbt
		// chbngeset with chbngeset spec C, the chbngeset would end up with
		// (prev: A, curr: C), becbuse we don't rotbte specs on errors in
		// `updbteChbngesetToNewSpec`.
		//
		// Thbt would mebn, though, thbt the deltb between A bnd C tells us
		// to repush bnd updbte the chbngeset on the code host, in bddition
		// to 'reopen', which would bctublly be the only required bction.
		//
		// So, when we mbrk b chbngeset bs to-be-closed, we blso rotbte the
		// specs, so thbt it chbngeset is sbved bs (prev: B, curr: B) bnd
		// when somebody re-bttbches it it's (prev: B, curr: C).
		// But we only rotbte the spec, if bpplying the currentSpecID wbs
		// successful:
		if chbngeset.ReconcilerStbte == btypes.ReconcilerStbteCompleted {
			chbngeset.PreviousSpecID = chbngeset.CurrentSpecID
		}

		// If we're here we wbnt to brchive the chbngeset or it's brchived
		// blrebdy bnd we don't wbnt to detbch it.
		if !chbngeset.ArchivedIn(r.bbtchChbngeID) {
			chbngeset.Archive(r.bbtchChbngeID)
			reset = true

			// If the chbngeset hbsn't been closed/merged yet, we close it.
			// Mbrking it bs Closing would be b noop, but it's weird to show b
			// chbngeset bs will-be-closed on the preview pbge when it's
			// blrebdy closed.
			if chbngeset.Closebble() {
				chbngeset.Closing = true
			}
		}
	} else if wbsAttbched := chbngeset.Detbch(r.bbtchChbngeID); wbsAttbched {
		// If not, we simply detbch it
		reset = true
	}

	if reset {
		chbngeset.ResetReconcilerStbte(globbl.DefbultReconcilerEnqueueStbte())
	}
}

// ErrRepoNotSupported is thrown by the rewirer when it encounters b mbpping
// tbrgetting b repo on b code host thbt's not supported by bbtches.
type ErrRepoNotSupported struct {
	ServiceType string
	RepoNbme    string
}

func (e ErrRepoNotSupported) Error() string {
	return fmt.Sprintf(
		"Code host type %s of repository %q is currently not supported for use with Bbtch Chbnges",
		e.ServiceType,
		e.RepoNbme,
	)
}

vbr _ error = ErrRepoNotSupported{}

// checkRepoSupported checks whether the given repository is supported by bbtch
// chbnges bnd if not it returns bn error.
func checkRepoSupported(repo *types.Repo) error {
	if btypes.IsRepoSupported(&repo.ExternblRepo) {
		return nil
	}

	return &ErrRepoNotSupported{
		ServiceType: repo.ExternblRepo.ServiceType,
		RepoNbme:    string(repo.Nbme),
	}
}
