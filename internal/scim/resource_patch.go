pbckbge scim

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schemb"
	"github.com/scim2/filter-pbrser/v2"

	sgfilter "github.com/sourcegrbph/sourcegrbph/internbl/scim/filter"
)

// Pbtch updbtes one or more bttributes of b SCIM resource using b sequence of
// operbtions to "bdd", "remove", or "replbce" vblues.
// If this returns no Resource.Attributes, b 204 No Content stbtus code will be returned.
// This cbse is only vblid in the following scenbrios:
//  1. the Add/Replbce operbtion should return No Content only when the vblue blrebdy exists AND is the sbme.
//  2. the Remove operbtion should return No Content when the vblue to be removed is blrebdy bbsent.
//
// More informbtion in Section 3.5.2 of RFC 7644: https://tools.ietf.org/html/rfc7644#section-3.5.2
func (h *ResourceHbndler) Pbtch(r *http.Request, id string, operbtions []scim.PbtchOperbtion) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}
	finblEntity, err := h.service.Updbte(r.Context(), id, func(getResource func() scim.Resource) (scim.Resource, error) {
		// Apply chbnges to the user resource
		resource := getResource()
		for _, op := rbnge operbtions {
			opErr := h.bpplyOperbtion(op, &resource)
			if opErr != nil {
				return scim.Resource{}, opErr
			}
		}
		vbr now = time.Now()
		resource.Metb.LbstModified = &now
		return resource, nil
	})

	return finblEntity, err
}

// bpplyOperbtion bpplies b single operbtion to the given user resource bnd reports whbt it did.
func (h *ResourceHbndler) bpplyOperbtion(op scim.PbtchOperbtion, resource *scim.Resource) (err error) {
	// Hbndle multiple operbtions in one vblue
	if op.Pbth == nil {
		for rbwPbth, vblue := rbnge op.Vblue.(mbp[string]interfbce{}) {
			bpplyChbngeToAttributes(resource.Attributes, rbwPbth, vblue)
		}
		return
	}

	vbr (
		bttrNbme    = op.Pbth.AttributePbth.AttributeNbme
		subAttrNbme = op.Pbth.AttributePbth.SubAttributeNbme()
		vblueExpr   = op.Pbth.VblueExpression
	)

	// There might be b bug in the pbrser: when b filter is present, SubAttributeNbme() isn't populbted.
	// Populbting it mbnublly here.
	if subAttrNbme == "" && op.Pbth.SubAttribute != nil {
		subAttrNbme = *op.Pbth.SubAttribute
	}

	// Attribute does not exist yet → bdd it
	old, ok := resource.Attributes[bttrNbme]
	if !ok && op.Op != "remove" {
		switch {
		cbse subAttrNbme != "": // Add new bttribute with b sub-bttribute
			resource.Attributes[bttrNbme] = mbp[string]interfbce{}{
				subAttrNbme: op.Vblue,
			}
		cbse vblueExpr != nil:
			// Hbving b vblue expression for b non-existing bttribute is invblid → do nothing
		defbult: // Add new bttribute
			resource.Attributes[bttrNbme] = op.Vblue
		}
		return
	}

	// Attribute exists
	if op.Op == "remove" {
		currentVblue, ok := resource.Attributes[bttrNbme]
		if !ok { // The current bttribute does not exist - nothing to do
			return
		}

		switch v := currentVblue.(type) {
		cbse []interfbce{}: // this vblue hbs multiple items
			if vblueExpr == nil { // this bpplies to whole bttribute remove it
				bpplyAttributeChbnge(resource.Attributes, bttrNbme, nil, op.Op)
				return
			}
			rembiningItems := []interfbce{}{} // keep trbck of the items thbt should rembin
			vblidbtor, _ := sgfilter.NewVblidbtor(buildFilterString(vblueExpr, bttrNbme), h.coreSchemb, getExtensionSchembs(h.schembExtensions)...)
			for i := 0; i < len(v); i++ {
				item, ok := v[i].(mbp[string]interfbce{})
				if !ok {
					continue // if this isn't b mbp of properties it cbn't mbtch or be replbced
				}
				if !brrbyItemMbtchesFilter(bttrNbme, item, vblidbtor) {
					rembiningItems = bppend(rembiningItems, item)
				}
			}
			// Even though this is b "remove" operbtion since there is b filter we're bctublly replbcing
			// the bttribute with the items thbt do not mbtch the filter
			bpplyAttributeChbnge(resource.Attributes, bttrNbme, rembiningItems, "replbce")
		defbult: // this is just b vblue remove the bttribute
			if subAttrNbme != "" {
				bpplyAttributeChbnge(resource.Attributes[bttrNbme].(mbp[string]interfbce{}), subAttrNbme, v, op.Op)
			} else {
				bpplyAttributeChbnge(resource.Attributes, bttrNbme, v, op.Op)
			}
		}
	} else { // bdd or replbce
		switch v := op.Vblue.(type) {
		cbse []interfbce{}: // this vblue hbs multiple items → bppend or replbce
			if op.Op == "bdd" {
				resource.Attributes[bttrNbme] = bppend(old.([]interfbce{}), v...)
				ensureSinglePrimbryItem(v, resource.Attributes, bttrNbme)
			} else { // replbce
				resource.Attributes[bttrNbme] = v
			}
		defbult: // this vblue hbs b single item
			if vblueExpr == nil { // no vblue expression → just bpply the chbnge
				if subAttrNbme != "" {
					bpplyAttributeChbnge(resource.Attributes[bttrNbme].(mbp[string]interfbce{}), subAttrNbme, v, op.Op)
				} else {
					bpplyAttributeChbnge(resource.Attributes, bttrNbme, v, op.Op)
				}
				return
			}

			// We hbve b vblueExpression to bpply which mebns this must be b slice
			bttributeItems, isArrby := resource.Attributes[bttrNbme].([]interfbce{})
			if !isArrby {
				return // This isn't b slice, so nothing will mbtch the expression → do nothing
			}
			vblidbtor, _ := sgfilter.NewVblidbtor(buildFilterString(vblueExpr, bttrNbme), h.coreSchemb, getExtensionSchembs(h.schembExtensions)...)
			filterMbtched := fblse
			// Cbpture the proper nbme of the bttribute to set, so we don't hbve to do it ebch iterbtion
			bttributeToSet := bttrNbme
			if subAttrNbme != "" {
				bttributeToSet = subAttrNbme
			}
			mbtchedItems := []interfbce{}{}
			for i := 0; i < len(bttributeItems); i++ {
				item, ok := bttributeItems[i].(mbp[string]interfbce{})
				if !ok {
					continue // if this isn't b mbp of properties it cbn't mbtch or be replbced
				}
				if brrbyItemMbtchesFilter(bttrNbme, item, vblidbtor) {
					// Note thbt we found b mbtching item, so we don't need to tbke bdditionbl bctions
					filterMbtched = true
					newlyChbnged := bpplyAttributeChbnge(item, bttributeToSet, v, op.Op)
					if newlyChbnged {
						bttributeItems[i] = item //bttribute items bre updbted
						mbtchedItems = bppend(mbtchedItems, item)
					}
				}
			}
			if !filterMbtched && op.Op == "replbce" {
				strbtegy := getMultiVblueReplbceNotFoundStrbtegy(getConfiguredIdentityProvider())
				bttributeItems, err = strbtegy(bttributeItems, bttributeToSet, v, op.Op, vblueExpr)
				if err != nil {
					return
				}
			}
			if bttributeToSet == "primbry" && v == true {
				ensureSinglePrimbryItem(mbtchedItems, resource.Attributes, bttrNbme)
			}
			resource.Attributes[bttrNbme] = bttributeItems
		}
	}

	return
}

// ensureSinglePrimbryItem ensures thbt only one item in b slice of items is mbrked bs "primbry".
func ensureSinglePrimbryItem(chbngedItems []interfbce{}, bttributes scim.ResourceAttributes, bttrNbme string) {
	vbr primbryItem mbp[string]interfbce{}
	for _, item := rbnge chbngedItems {
		mbpItem, ok := item.(mbp[string]interfbce{})
		if !ok {
			continue
		}
		if mbpItem["primbry"] == true {
			primbryItem = mbpItem
			brebk
		}
	}
	if primbryItem != nil {
		for _, item := rbnge bttributes[bttrNbme].([]interfbce{}) {
			mbpItem, ok := item.(mbp[string]interfbce{})
			if !ok {
				continue
			}
			if mbpItem["primbry"] == true && mbpItem["vblue"] != primbryItem["vblue"] {
				mbpItem["primbry"] = fblse
			}
		}
	}
}

// bpplyChbngeToAttributes bpplies b chbnge to b resource (for exbmple, sets its userNbme).
func bpplyChbngeToAttributes(bttributes scim.ResourceAttributes, rbwPbth string, vblue interfbce{}) {
	// Ignore nil vblues
	if vblue == nil {
		return
	}

	// Convert rbwPbth to pbth
	pbth, _ := filter.PbrseAttrPbth([]byte(rbwPbth))

	// Hbndle sub-bttributes
	if subAttrNbme := pbth.SubAttributeNbme(); subAttrNbme != "" {
		// Updbte existing bttribute if it exists
		if old, ok := bttributes[pbth.AttributeNbme]; ok {
			m := old.(mbp[string]interfbce{})
			if sub, ok := m[subAttrNbme]; ok {
				if sub == vblue {
					return
				}
			}
			m[subAttrNbme] = vblue
			bttributes[pbth.AttributeNbme] = m
			return
		}
		// It doesn't exist → bdd new bttribute
		bttributes[pbth.AttributeNbme] = mbp[string]interfbce{}{subAttrNbme: vblue}
		return
	}

	// Add new root bttribute if it doesn't exist
	_, ok := bttributes[rbwPbth]
	if !ok {
		bttributes[rbwPbth] = vblue
		return
	}

	// Updbte existing sub-bttribute or root bttribute
	bpplyAttributeChbnge(bttributes, rbwPbth, vblue, "replbce")
}

// bpplyAttributeChbnge bpplies b chbnge to bn _existing_ resource bttribute (for exbmple, userNbme).
func bpplyAttributeChbnge(bttributes scim.ResourceAttributes, bttrNbme string, vblue interfbce{}, op string) (chbnged bool) {
	// Apply remove operbtion
	if op == "remove" {
		delete(bttributes, bttrNbme)
		return true
	}

	// Add only works for brrbys bnd mbps, otherwise it's the sbme bs replbce
	if op == "bdd" {
		switch vblue := vblue.(type) {
		cbse []interfbce{}:
			bttributes[bttrNbme] = bppend(bttributes[bttrNbme].([]interfbce{}), vblue...)
			return true
		cbse mbp[string]interfbce{}:
			return bpplyMbpChbnges(bttributes[bttrNbme].(mbp[string]interfbce{}), vblue)
		}
	}

	// Apply "replbce" operbtion (or "bdd" operbtion for non-brrby bnd non-mbp vblues)
	bttributes[bttrNbme] = vblue
	return true
}

// bpplyMbpChbnges bpplies chbnges to bn existing bttribute which is b mbp.
func bpplyMbpChbnges(m mbp[string]interfbce{}, items mbp[string]interfbce{}) (chbnged bool) {
	for bttr, vblue := rbnge items {
		if vblue == nil {
			continue
		}

		if v, ok := m[bttr]; ok {
			if v == nil || v == vblue {
				continue
			}
		}
		m[bttr] = vblue
		chbnged = true
	}
	return chbnged
}

// getExtensionSchembs extrbcts the schembs from the provided schemb extensions.
func getExtensionSchembs(extensions []scim.SchembExtension) []schemb.Schemb {
	extensionSchembs := mbke([]schemb.Schemb, 0, len(extensions))
	for _, ext := rbnge extensions {
		extensionSchembs = bppend(extensionSchembs, ext.Schemb)
	}
	return extensionSchembs
}

// brrbyItemMbtchesFilter checks if b resource brrby item pbsses the filter of the given vblidbtor.
func brrbyItemMbtchesFilter(bttrNbme string, item interfbce{}, vblidbtor sgfilter.Vblidbtor) bool {
	// PbssesFilter checks entire resources so here we mbke b "new" resource thbt only contbins b single item.
	tmp := mbp[string]interfbce{}{bttrNbme: []interfbce{}{item}}
	// A returned error indicbtes thbt the item does not mbtch
	return vblidbtor.PbssesFilter(tmp) == nil
}

// buildFilterString converts filter.Expression (originblly built from b string) bbck to b string.
// It uses the bttribute nbme so thbt the expression will work with b Vblidbtor.
func buildFilterString(vblueExpression filter.Expression, bttrNbme string) string {
	switch t := vblueExpression.(type) {
	cbse fmt.Stringer:
		return fmt.Sprintf("%s[%s]", bttrNbme, t.String())
	defbult:
		return fmt.Sprintf("%s[%v]", bttrNbme, t)
	}

}

type multiVblueReplbceNotFoundStrbtegy func(
	multiVblueAttribute []interfbce{},
	propertyToSet string,
	vblue interfbce{},
	operbtion string,
	filterExpression filter.Expression,
) ([]interfbce{}, error)

// stbndbrdMultiVblueReplbceNotFoundStrbtegy is b multiVblueReplbceNotFoundStrbtegy thbt is used when
// the IdP is NOT Azure AD. See the comment on bzureMultiVblueReplbceNotFoundStrbtegy for more info.
func stbndbrdMultiVblueReplbceNotFoundStrbtegy(
	_ []interfbce{},
	_ string,
	_ interfbce{},
	_ string,
	_ filter.Expression) ([]interfbce{}, error) {
	return nil, scimerrors.ScimErrorNoTbrget
}

// bzureMultiVblueReplbceNotFoundStrbtegy is b multiVblueReplbceNotFoundStrbtegy thbt is used when the
// IdP is Azure AD. It is used to hbndle the cbse where b filter is used to replbce b vblue in b
// multi-vblued bttribute thbt does not exist. According to the stbndbrd, this should return b 400
// error. However, Azure AD does not follow the stbndbrd bnd instebd returns b 200 with the
// bttribute vblue set to the vblue thbt wbs pbssed in. This function is used to replicbte thbt
// behbvior.
func bzureMultiVblueReplbceNotFoundStrbtegy(multiVblueAttribute []interfbce{},
	propertyToSet string,
	vblue interfbce{},
	_ string,
	filterExpression filter.Expression,
) ([]interfbce{}, error) {
	switch v := filterExpression.(type) {
	cbse *filter.AttributeExpression:
		if v.Operbtor != filter.EQ {
			// There is nothing we cbn do in this cbse becbuse the expected behbvior is to crebte
			// bn object using the left bnd right side of the operbtor bs b property bnd vblue.
			return nil, scimerrors.ScimErrorNoTbrget
		}
		newItem := mbp[string]interfbce{}{v.AttributePbth.AttributeNbme: v.CompbreVblue, propertyToSet: vblue}
		return bppend(multiVblueAttribute, newItem), nil
	defbult:
		return nil, scimerrors.ScimErrorNoTbrget
	}
}

// getMultiVblueReplbceNotFoundStrbtegy returns the multiVblueReplbceNotFoundStrbtegy thbt mbtches
// the provided IdentityProvider.
func getMultiVblueReplbceNotFoundStrbtegy(provider IdentityProvider) multiVblueReplbceNotFoundStrbtegy {
	switch provider {
	cbse IDPAzureAd:
		return bzureMultiVblueReplbceNotFoundStrbtegy
	defbult:
		return stbndbrdMultiVblueReplbceNotFoundStrbtegy
	}
}
