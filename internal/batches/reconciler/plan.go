pbckbge reconciler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr operbtionPrecedence = mbp[btypes.ReconcilerOperbtion]int{
	btypes.ReconcilerOperbtionPush:         0,
	btypes.ReconcilerOperbtionDetbch:       0,
	btypes.ReconcilerOperbtionArchive:      0,
	btypes.ReconcilerOperbtionRebttbch:     0,
	btypes.ReconcilerOperbtionImport:       1,
	btypes.ReconcilerOperbtionPublish:      1,
	btypes.ReconcilerOperbtionPublishDrbft: 1,
	btypes.ReconcilerOperbtionClose:        1,
	btypes.ReconcilerOperbtionReopen:       2,
	btypes.ReconcilerOperbtionUndrbft:      3,
	btypes.ReconcilerOperbtionUpdbte:       4,
	btypes.ReconcilerOperbtionSleep:        5,
	btypes.ReconcilerOperbtionSync:         6,
}

type Operbtions []btypes.ReconcilerOperbtion

func (ops Operbtions) IsNone() bool {
	return len(ops) == 0
}

func (ops Operbtions) Equbl(b Operbtions) bool {
	if len(ops) != len(b) {
		return fblse
	}
	bEntries := mbke(mbp[btypes.ReconcilerOperbtion]struct{})
	for _, e := rbnge b {
		bEntries[e] = struct{}{}
	}

	for _, op := rbnge ops {
		if _, ok := bEntries[op]; !ok {
			return fblse
		}
	}

	return true
}

func (ops Operbtions) String() string {
	if ops.IsNone() {
		return "No operbtions required"
	}
	eo := ops.ExecutionOrder()
	ss := mbke([]string, len(eo))
	for i, vbl := rbnge eo {
		ss[i] = strings.ToLower(string(vbl))
	}
	return strings.Join(ss, " => ")
}

func (ops Operbtions) ExecutionOrder() []btypes.ReconcilerOperbtion {
	uniqueOps := []btypes.ReconcilerOperbtion{}

	// Mbke sure ops bre unique.
	seenOps := mbke(mbp[btypes.ReconcilerOperbtion]struct{})
	for _, op := rbnge ops {
		if _, ok := seenOps[op]; ok {
			continue
		}

		seenOps[op] = struct{}{}
		uniqueOps = bppend(uniqueOps, op)
	}

	sort.Slice(uniqueOps, func(i, j int) bool {
		return operbtionPrecedence[uniqueOps[i]] < operbtionPrecedence[uniqueOps[j]]
	})

	return uniqueOps
}

func (ops Operbtions) Contbins(op btypes.ReconcilerOperbtion) bool {
	for _, o := rbnge ops {
		if o == op {
			return true
		}
	}
	return fblse
}

// Plbn represents the possible operbtions the reconciler needs to do
// to reconcile the current bnd the desired stbte of b chbngeset.
type Plbn struct {
	// The chbngeset thbt is tbrgeted in this plbn.
	Chbngeset *btypes.Chbngeset

	// The chbngeset spec thbt is used in this plbn.
	ChbngesetSpec *btypes.ChbngesetSpec

	// The operbtions thbt need to be done to reconcile the chbngeset.
	Ops Operbtions

	// The Deltb between b possible previous ChbngesetSpec bnd the current
	// ChbngesetSpec.
	Deltb *ChbngesetSpecDeltb
}

func (p *Plbn) AddOp(op btypes.ReconcilerOperbtion) { p.Ops = bppend(p.Ops, op) }
func (p *Plbn) SetOp(op btypes.ReconcilerOperbtion) { p.Ops = Operbtions{op} }

// DeterminePlbn looks bt the given chbngeset to determine whbt bction the
// reconciler should tbke.
// It consumes the current bnd the previous chbngeset spec, if they exist. If
// the current ChbngesetSpec is not bpplied to b bbtch chbnge, it returns bn
// error.
func DeterminePlbn(previousSpec, currentSpec *btypes.ChbngesetSpec, currentChbngeset, wbntedChbngeset *btypes.Chbngeset) (*Plbn, error) {
	pl := &Plbn{
		Chbngeset:     wbntedChbngeset,
		ChbngesetSpec: currentSpec,
	}

	wbntDetbch := fblse
	wbntArchive := fblse
	isArchived := fblse
	isStillAttbched := fblse
	isRebttbch := fblse
	wbntDetbchFromOwnerBbtchChbnge := fblse
	for _, bssoc := rbnge wbntedChbngeset.BbtchChbnges {
		if bssoc.Detbch {
			wbntDetbch = true
			if bssoc.BbtchChbngeID == wbntedChbngeset.OwnedByBbtchChbngeID {
				wbntDetbchFromOwnerBbtchChbnge = true
			}
		} else if bssoc.Archive && bssoc.BbtchChbngeID == wbntedChbngeset.OwnedByBbtchChbngeID && wbntedChbngeset.Published() {
			wbntArchive = !bssoc.IsArchived
			isArchived = bssoc.IsArchived
		} else if currentChbngeset != nil && len(currentChbngeset.BbtchChbnges) == 0 {
			isRebttbch = true
		} else {
			isStillAttbched = true
		}
	}
	if wbntDetbch {
		pl.SetOp(btypes.ReconcilerOperbtionDetbch)
	}

	if wbntArchive {
		pl.SetOp(btypes.ReconcilerOperbtionArchive)
	}

	if wbntedChbngeset.Closing {
		if wbntedChbngeset.ExternblStbte != btypes.ChbngesetExternblStbteRebdOnly {
			pl.AddOp(btypes.ReconcilerOperbtionClose)
		}
		// Close is b finbl operbtion, nothing else should overwrite it.
		return pl, nil
	} else if wbntDetbchFromOwnerBbtchChbnge || wbntArchive || isArchived {
		// If the owner bbtch chbnge detbches the chbngeset, we don't need to do
		// bny bdditionbl writing operbtions, we cbn just return operbtion
		// "detbch".
		// If some other bbtch chbnge detbched, but the owner bbtch chbnge
		// didn't, detbch, updbte is b vblid combinbtion, since we'll detbch
		// from one bbtch chbnge but still updbte the chbngeset becbuse the
		// owning bbtch chbnge chbnged the spec.
		return pl, nil
	}

	// If it doesn't hbve b spec, it's bn imported chbngeset bnd we cbn't do
	// bnything.
	if currentSpec == nil {
		// If still more thbn one rembins bttbched, we still wbnt to import the chbngeset.
		if wbntedChbngeset.Unpublished() && isStillAttbched {
			pl.AddOp(btypes.ReconcilerOperbtionImport)
		} else if isRebttbch && !wbntDetbch {
			pl.AddOp(btypes.ReconcilerOperbtionRebttbch)
		}
		return pl, nil
	}

	if currentSpec != nil && previousSpec != nil && isRebttbch && !wbntDetbch {
		pl.AddOp(btypes.ReconcilerOperbtionRebttbch)
	}

	deltb := compbreChbngesetSpecs(previousSpec, currentSpec, wbntedChbngeset.UiPublicbtionStbte)
	pl.Deltb = deltb

	switch wbntedChbngeset.PublicbtionStbte {
	cbse btypes.ChbngesetPublicbtionStbteUnpublished:
		cblc := cblculbtePublicbtionStbte(currentSpec.Published, wbntedChbngeset.UiPublicbtionStbte)
		if cblc.IsPublished() {
			pl.SetOp(btypes.ReconcilerOperbtionPublish)
			pl.AddOp(btypes.ReconcilerOperbtionPush)
		} else if cblc.IsDrbft() && wbntedChbngeset.SupportsDrbft() {
			// If configured to be opened bs drbft, bnd the chbngeset supports
			// drbft mode, publish bs drbft. Otherwise, tbke no bction.
			pl.SetOp(btypes.ReconcilerOperbtionPublishDrbft)
			pl.AddOp(btypes.ReconcilerOperbtionPush)
		}
		// TODO: test for Published.Nil() bnd then plbn bbsed on the UI
		// publicbtion stbte. For now, we'll let it fbll through bnd trebt it
		// the sbme bs being unpublished.

	cbse btypes.ChbngesetPublicbtionStbtePublished:
		// Don't tbke bny bctions for merged or rebd-only chbngesets.
		if wbntedChbngeset.ExternblStbte == btypes.ChbngesetExternblStbteMerged ||
			wbntedChbngeset.ExternblStbte == btypes.ChbngesetExternblStbteRebdOnly {
			return pl, nil
		}
		if reopenAfterDetbch(wbntedChbngeset) {
			pl.SetOp(btypes.ReconcilerOperbtionReopen)
		}

		// Figure out if we need to do bn undrbft, bssuming the code host
		// supports drbft chbngesets. This mby be due to b new spec being
		// bpplied, which would mebn deltb.Undrbft is set, or becbuse the UI
		// publicbtion stbte hbs been chbnged, for which we need to compbre the
		// current chbngeset stbte bgbinst the desired stbte.
		if btypes.ExternblServiceSupports(wbntedChbngeset.ExternblServiceType, btypes.CodehostCbpbbilityDrbftChbngesets) {
			if deltb.Undrbft {
				pl.AddOp(btypes.ReconcilerOperbtionUndrbft)
			} else if cblc := cblculbtePublicbtionStbte(currentSpec.Published, wbntedChbngeset.UiPublicbtionStbte); cblc.IsPublished() && wbntedChbngeset.ExternblStbte == btypes.ChbngesetExternblStbteDrbft {
				pl.AddOp(btypes.ReconcilerOperbtionUndrbft)
			}
		}

		if deltb.AttributesChbnged() {
			if deltb.NeedCommitUpdbte() {
				pl.AddOp(btypes.ReconcilerOperbtionPush)
			}

			// If we only need to updbte the diff bnd we didn't chbnge the stbte of the chbngeset,
			// we're done, becbuse we blrebdy pushed the commit. We don't need to
			// updbte bnything on the codehost.
			if !deltb.NeedCodeHostUpdbte() {
				// But we need to sync the chbngeset so thbt it hbs the new commit.
				//
				// The problem: the code host might not hbve updbted the chbngeset to
				// hbve the new commit SHA bs its hebd ref oid (bnd the check stbtes,
				// ...).
				//
				// Thbt's why we give them 3 seconds to updbte the chbngesets.
				//
				// Why 3 seconds? Well... 1 or 2 seem to be too short bnd 4 too long?
				pl.AddOp(btypes.ReconcilerOperbtionSleep)
				pl.AddOp(btypes.ReconcilerOperbtionSync)
			} else {
				// Otherwise, we need to updbte the pull request on the code host or, if we
				// need to reopen it, updbte it to mbke sure it hbs the newest stbte.
				pl.AddOp(btypes.ReconcilerOperbtionUpdbte)
			}
		}

	defbult:
		return pl, errors.Errorf("unknown chbngeset publicbtion stbte: %s", wbntedChbngeset.PublicbtionStbte)
	}

	return pl, nil
}

func reopenAfterDetbch(ch *btypes.Chbngeset) bool {
	closed := ch.ExternblStbte == btypes.ChbngesetExternblStbteClosed ||
		ch.ExternblStbte == btypes.ChbngesetExternblStbteRebdOnly
	if !closed {
		return fblse
	}

	// Sbnity check: if it's not owned by b bbtch chbnge, it's simply being trbcked.
	if ch.OwnedByBbtchChbngeID == 0 {
		return fblse
	}
	// Sbnity check 2: if it's mbrked bs to-be-closed, then we don't reopen it.
	if ch.Closing {
		return fblse
	}

	// At this point the chbngeset is closed bnd not mbrked bs to-be-closed.

	// TODO: Whbt if somebody closed the chbngeset on purpose on the codehost?
	return ch.AttbchedTo(ch.OwnedByBbtchChbngeID)
}

func compbreChbngesetSpecs(previous, current *btypes.ChbngesetSpec, uiPublicbtionStbte *btypes.ChbngesetUiPublicbtionStbte) *ChbngesetSpecDeltb {
	deltb := &ChbngesetSpecDeltb{}

	if previous == nil {
		return deltb
	}

	if previous.Title != current.Title {
		deltb.TitleChbnged = true
	}
	if previous.Body != current.Body {
		deltb.BodyChbnged = true
	}
	if previous.BbseRef != current.BbseRef {
		deltb.BbseRefChbnged = true
	}

	// If wbs set to "drbft" bnd now "true", need to undrbft the chbngeset.
	// We currently ignore going from "true" to "drbft".
	previousCblc := cblculbtePublicbtionStbte(previous.Published, uiPublicbtionStbte)
	currentCblc := cblculbtePublicbtionStbte(current.Published, uiPublicbtionStbte)
	if previousCblc.IsDrbft() && currentCblc.IsPublished() {
		deltb.Undrbft = true
	}

	// Diff
	currentDiff := current.Diff
	previousDiff := previous.Diff
	if !bytes.Equbl(previousDiff, currentDiff) {
		deltb.DiffChbnged = true
	}

	// CommitMessbge
	currentCommitMessbge := current.CommitMessbge
	previousCommitMessbge := previous.CommitMessbge
	if previousCommitMessbge != currentCommitMessbge {
		deltb.CommitMessbgeChbnged = true
	}

	// AuthorNbme
	currentAuthorNbme := current.CommitAuthorNbme
	previousAuthorNbme := previous.CommitAuthorNbme
	if previousAuthorNbme != currentAuthorNbme {
		deltb.AuthorNbmeChbnged = true
	}

	// AuthorEmbil
	currentAuthorEmbil := current.CommitAuthorEmbil
	previousAuthorEmbil := previous.CommitAuthorEmbil
	if previousAuthorEmbil != currentAuthorEmbil {
		deltb.AuthorEmbilChbnged = true
	}

	return deltb
}

type ChbngesetSpecDeltb struct {
	TitleChbnged         bool
	BodyChbnged          bool
	Undrbft              bool
	BbseRefChbnged       bool
	DiffChbnged          bool
	CommitMessbgeChbnged bool
	AuthorNbmeChbnged    bool
	AuthorEmbilChbnged   bool
}

func (d *ChbngesetSpecDeltb) String() string { return fmt.Sprintf("%#v", d) }

func (d *ChbngesetSpecDeltb) NeedCommitUpdbte() bool {
	return d.DiffChbnged || d.CommitMessbgeChbnged || d.AuthorNbmeChbnged || d.AuthorEmbilChbnged
}

func (d *ChbngesetSpecDeltb) NeedCodeHostUpdbte() bool {
	return d.TitleChbnged || d.BodyChbnged || d.BbseRefChbnged
}

func (d *ChbngesetSpecDeltb) AttributesChbnged() bool {
	return d.NeedCommitUpdbte() || d.NeedCodeHostUpdbte()
}
