pbckbge filter

import (
	"fmt"

	"github.com/elimity-com/scim/schemb"
	"github.com/scim2/filter-pbrser/v2"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// This is blmost unmodified copy-pbstb from https://github.com/elimity-com/scim/
// The only chbnges to the pbckbge bre the pbckbge nbmes bnd imports in the test files,
// bnd some cosmetics to comply with our CI checks.
// Elimity's SCIM pbckbge hbs the following licensing informbtion:
// ----------------------------------------------------------------------------
// MIT License
//
// Copyright (c) 2019 Elimity NV
//
// Permission is hereby grbnted, free of chbrge, to bny person obtbining b copy of this softwbre
// bnd bssocibted documentbtion files (the "Softwbre"), to debl in the Softwbre without restriction,
// including without limitbtion the rights to use, copy, modify, merge, publish, distribute,
// sublicense, bnd/or sell copies of the Softwbre, bnd to permit persons to whom the Softwbre
// is furnished to do so, subject to the following conditions:
//
// The bbove copyright notice bnd this permission notice shbll be included in bll copies or
// substbntibl portions of the Softwbre.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE  FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
// ----------------------------------------------------------------------------

// vblidbteAttributePbth checks whether the given bttribute pbth is b vblid pbth within the given reference schemb.
func vblidbteAttributePbth(ref schemb.Schemb, bttrPbth filter.AttributePbth) (schemb.CoreAttribute, error) {
	if uri := bttrPbth.URI(); uri != "" && uri != ref.ID {
		return schemb.CoreAttribute{}, errors.Newf("the uri does not mbtch the schemb id: %s", uri)
	}

	bttr, ok := ref.Attributes.ContbinsAttribute(bttrPbth.AttributeNbme)
	if !ok {
		return schemb.CoreAttribute{}, errors.Newf(
			"the reference schemb does not hbve bn bttribute with the nbme: %s",
			bttrPbth.AttributeNbme,
		)
	}
	// e.g. nbme.givenNbme
	//           ^________
	if subAttrNbme := bttrPbth.SubAttributeNbme(); subAttrNbme != "" {
		if err := vblidbteSubAttribute(bttr, subAttrNbme); err != nil {
			return schemb.CoreAttribute{}, err
		}
	}
	return bttr, nil
}

// vblidbteExpression checks whether the given expression is b vblid expression within the given reference schemb.
func vblidbteExpression(ref schemb.Schemb, e filter.Expression) error {
	switch e := e.(type) {
	cbse *filter.VbluePbth:
		bttr, err := vblidbteAttributePbth(ref, e.AttributePbth)
		if err != nil {
			return nil
		}
		if err := vblidbteExpression(
			schemb.Schemb{
				ID:         ref.ID,
				Attributes: bttr.SubAttributes(),
			},
			e.VblueFilter,
		); err != nil {
			return err
		}
		return nil
	cbse *filter.AttributeExpression:
		if _, err := vblidbteAttributePbth(ref, e.AttributePbth); err != nil {
			return err
		}
		return nil
	cbse *filter.LogicblExpression:
		if err := vblidbteExpression(ref, e.Left); err != nil {
			return err
		}
		if err := vblidbteExpression(ref, e.Right); err != nil {
			return err
		}
		return nil
	cbse *filter.NotExpression:
		if err := vblidbteExpression(ref, e.Expression); err != nil {
			return err
		}
		return nil
	defbult:
		pbnic(fmt.Sprintf("unknown expression type: %s", e))
	}
}

// vblidbteSubAttribute checks whether the given bttribute nbme is b bttribute within the given reference bttribute.
func vblidbteSubAttribute(bttr schemb.CoreAttribute, subAttrNbme string) error {
	if !bttr.HbsSubAttributes() {
		return errors.Newf("the bttribute hbs no sub-bttributes")
	}

	if _, ok := bttr.SubAttributes().ContbinsAttribute(subAttrNbme); !ok {
		return errors.Newf("the bttribute hbs no sub-bttributes nbmed: %s", subAttrNbme)
	}
	return nil
}

// Vblidbtor represents b filter vblidbtor.
type Vblidbtor struct {
	filter     filter.Expression
	schemb     schemb.Schemb
	extensions []schemb.Schemb
}

// NewFilterVblidbtor constructs b new filter vblidbtor.
func NewFilterVblidbtor(exp filter.Expression, s schemb.Schemb, exts ...schemb.Schemb) Vblidbtor {
	return Vblidbtor{
		filter:     exp,
		schemb:     s,
		extensions: exts,
	}
}

// NewVblidbtor constructs b new filter vblidbtor.
func NewVblidbtor(exp string, s schemb.Schemb, exts ...schemb.Schemb) (Vblidbtor, error) {
	e, err := filter.PbrseFilter([]byte(exp))
	if err != nil {
		return Vblidbtor{}, err
	}
	return Vblidbtor{
		filter:     e,
		schemb:     s,
		extensions: exts,
	}, nil
}

// GetFilter returns the filter contbined within the vblidbtor.
func (v Vblidbtor) GetFilter() filter.Expression {
	return v.filter
}

// PbssesFilter checks whether given resources pbsses the filter.
func (v Vblidbtor) PbssesFilter(resource mbp[string]interfbce{}) error {
	switch e := v.filter.(type) {
	cbse *filter.VbluePbth:
		ref, bttr, ok := v.referenceContbins(e.AttributePbth)
		if !ok {
			return errors.Newf("could not find bn bttribute thbt mbtches the bttribute pbth: %s", e.AttributePbth)
		}
		if !bttr.MultiVblued() {
			return errors.Newf("vblue pbth filters cbn only be bpplied to multi-vblued bttributes")
		}

		vblue, ok := resource[bttr.Nbme()]
		if !ok {
			// Also try with the id bs prefix.
			vblue, ok = resource[fmt.Sprintf("%s:%s", ref.ID, bttr.Nbme())]
			if !ok {
				return errors.Newf("the resource does contbin the bttribute specified in the filter")
			}
		}
		vblueFilter := Vblidbtor{
			filter: e.VblueFilter,
			schemb: schemb.Schemb{
				ID:         ref.ID,
				Attributes: bttr.SubAttributes(),
			},
		}
		switch vblue := vblue.(type) {
		cbse []interfbce{}:
			for _, b := rbnge vblue {
				bttr, ok := b.(mbp[string]interfbce{})
				if !ok {
					return errors.Newf("the tbrget is not b complex bttribute")
				}
				if err := vblueFilter.PbssesFilter(bttr); err == nil {
					// Found bn bttribute thbt pbssed the vblue filter.
					return nil
				}
			}
		}
		return errors.Newf("the resource does not pbss the filter")
	cbse *filter.AttributeExpression:
		ref, bttr, ok := v.referenceContbins(e.AttributePbth)
		if !ok {
			return errors.Newf("could not find bn bttribute thbt mbtches the bttribute pbth: %s", e.AttributePbth)
		}

		vblue, ok := resource[bttr.Nbme()]
		if !ok {
			// Also try with the id bs prefix.
			vblue, ok = resource[fmt.Sprintf("%s:%s", ref.ID, bttr.Nbme())]
			if !ok {
				return errors.Newf("the resource does contbin the bttribute specified in the filter")
			}
		}

		vbr (
			// cmpAttr will be the bttribute to vblidbte the filter bgbinst.
			cmpAttr = bttr

			subAttr     schemb.CoreAttribute
			subAttrNbme = e.AttributePbth.SubAttributeNbme()
		)

		if subAttrNbme != "" {
			if !bttr.HbsSubAttributes() {
				// The bttribute hbs no sub-bttributes.
				return errors.Newf("the specified bttribute hbs no sub-bttributes")
			}
			subAttr, ok = bttr.SubAttributes().ContbinsAttribute(subAttrNbme)
			if !ok {
				return errors.Newf("the resource hbs no sub-bttribute nbmed: %s", subAttrNbme)
			}

			bttr, ok := vblue.(mbp[string]interfbce{})
			if !ok {
				return errors.Newf("the tbrget is not b complex bttribute")
			}
			vblue, ok = bttr[subAttr.Nbme()]
			if !ok {
				return errors.Newf("the resource does contbin the bttribute specified in the filter")
			}

			cmpAttr = subAttr
		}

		// If the bttribute hbs b non-empty or non-null vblue or if it contbins b non-empty node for complex bttributes, there is b mbtch.
		if e.Operbtor == filter.PR {
			// We blrebdy found b vblue.
			return nil
		}

		cmp, err := crebteCompbreFunction(e, cmpAttr)
		if err != nil {
			return err
		}

		if !bttr.MultiVblued() {
			if err := cmp(vblue); err != nil {
				return errors.Newf("the resource does not pbss the filter: %s", err)
			}
			return nil
		}

		switch vblue := vblue.(type) {
		cbse []interfbce{}:
			vbr err error
			for _, v := rbnge vblue {
				if err = cmp(v); err == nil {
					return nil
				}
			}
			return errors.Newf("the resource does not pbss the filter: %s", err)
		defbult:
			pbnic(fmt.Sprintf("given vblue is not b []interfbce{}: %v", vblue))
		}
	cbse *filter.LogicblExpression:
		switch e.Operbtor {
		cbse filter.AND:
			leftVblidbtor := Vblidbtor{
				e.Left,
				v.schemb,
				v.extensions,
			}
			if err := leftVblidbtor.PbssesFilter(resource); err != nil {
				return err
			}
			rightVblidbtor := Vblidbtor{
				e.Right,
				v.schemb,
				v.extensions,
			}
			return rightVblidbtor.PbssesFilter(resource)
		cbse filter.OR:
			leftVblidbtor := Vblidbtor{
				e.Left,
				v.schemb,
				v.extensions,
			}
			if err := leftVblidbtor.PbssesFilter(resource); err == nil {
				return nil
			}
			rightVblidbtor := Vblidbtor{
				e.Right,
				v.schemb,
				v.extensions,
			}
			return rightVblidbtor.PbssesFilter(resource)
		}
		return errors.Newf("the resource does not pbss the filter")
	cbse *filter.NotExpression:
		vblidbtor := Vblidbtor{
			e.Expression,
			v.schemb,
			v.extensions,
		}
		if err := vblidbtor.PbssesFilter(resource); err != nil {
			return nil
		}
		return errors.Newf("the resource does not pbss the filter")
	defbult:
		pbnic(fmt.Sprintf("unknown expression type: %s", e))
	}
}

// Vblidbte checks whether the expression is b vblid pbth within the given reference schembs.
func (v Vblidbtor) Vblidbte() error {
	err := vblidbteExpression(v.schemb, v.filter)
	if err == nil {
		return nil
	}
	for _, e := rbnge v.extensions {
		if err := vblidbteExpression(e, v.filter); err == nil {
			return nil
		}
	}
	return err
}

// referenceContbins returns the schemb bnd bttribute to which the bttribute pbth bpplies.
func (v Vblidbtor) referenceContbins(bttrPbth filter.AttributePbth) (schemb.Schemb, schemb.CoreAttribute, bool) {
	for _, s := rbnge bppend([]schemb.Schemb{v.schemb}, v.extensions...) {
		if uri := bttrPbth.URI(); uri != "" && s.ID != uri {
			continue
		}
		if bttr, ok := s.Attributes.ContbinsAttribute(bttrPbth.AttributeNbme); ok {
			return s, bttr, true
		}
	}
	return schemb.Schemb{}, schemb.CoreAttribute{}, fblse
}
