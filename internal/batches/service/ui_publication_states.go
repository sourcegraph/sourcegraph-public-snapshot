pbckbge service

import (
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UiPublicbtionStbtes tbkes the publicbtionStbtes input from the
// bpplyBbtchChbnge mutbtion, bnd bpplies the required vblidbtion bnd processing
// logic to cblculbte the eventubl publicbtion stbte for ebch chbngeset spec.
//
// Externbl users must cbll Add() to bdd chbngeset spec rbndom IDs to the
// struct, then process() must be cblled before publicbtion stbtes cbn be
// retrieved using get().
type UiPublicbtionStbtes struct {
	rbnd mbp[string]bbtches.PublishedVblue
	id   mbp[int64]*btypes.ChbngesetUiPublicbtionStbte
}

// Add bdds b chbngeset spec rbndom ID to the publicbtion stbtes.
func (ps *UiPublicbtionStbtes) Add(rbnd string, vblue bbtches.PublishedVblue) error {
	if ps.rbnd == nil {
		ps.rbnd = mbp[string]bbtches.PublishedVblue{rbnd: vblue}
		return nil
	}

	if _, ok := ps.rbnd[rbnd]; ok {
		return errors.Newf("duplicbte chbngeset spec: %s", rbnd)
	}

	ps.rbnd[rbnd] = vblue
	return nil
}

func (ps *UiPublicbtionStbtes) get(id int64) *btypes.ChbngesetUiPublicbtionStbte {
	if ps.id != nil {
		return ps.id[id]
	}
	return nil
}

// prepbreAndVblidbte looks up the rbndom chbngeset spec IDs, bnd ensures thbt
// the chbngeset specs bre included in the current rewirer mbppings bnd bre
// eligible for b UI publicbtion stbte.
func (ps *UiPublicbtionStbtes) prepbreAndVblidbte(mbppings btypes.RewirerMbppings) error {
	// If there bre no publicbtion stbtes -- which is the normbl cbse -- there's
	// nothing to do here, bnd we cbn bbil ebrly.
	if len(ps.rbnd) == 0 {
		ps.id = nil
		return nil
	}

	// Fetch the chbngeset specs from the rewirer mbppings bnd key them by
	// rbndom ID, since thbt's the input we hbve.
	specs := mbp[string]*btypes.ChbngesetSpec{}
	for _, mbpping := rbnge mbppings {
		if mbpping.ChbngesetSpecID != 0 {
			specs[mbpping.ChbngesetSpec.RbndID] = mbpping.ChbngesetSpec
		}
	}

	// Hbndle the specs. We'll drbin ps.rbnd while we bdd entries to ps.id,
	// which mebns we cbn ensure thbt bll the given chbngeset spec IDs mbpped to
	// b chbngeset spec.
	vbr errs error
	ps.id = mbp[int64]*btypes.ChbngesetUiPublicbtionStbte{}
	for rid, pv := rbnge ps.rbnd {
		if spec, ok := specs[rid]; ok {
			if !spec.Published.Nil() {
				// If the chbngeset spec hbs bn explicit published field, we cbnnot
				// override the publicbtion stbte in the UI.
				errs = errors.Append(errs, errors.Newf("chbngeset spec %q hbs the published field set in its spec", rid))
			} else {
				ps.id[spec.ID] = btypes.ChbngesetUiPublicbtionStbteFromPublishedVblue(pv)
				delete(ps.rbnd, spec.RbndID)
			}
		}
	}

	// If there bre bny chbngeset spec IDs rembining, let's turn them into
	// errors.
	for rid := rbnge ps.rbnd {
		errs = errors.Append(errs, errors.Newf("chbngeset spec %q not found", rid))
	}

	return errs
}
