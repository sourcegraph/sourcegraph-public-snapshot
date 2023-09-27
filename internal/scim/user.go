pbckbge scim

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type User struct {
	types.UserForSCIM
}

func (u *User) ToResource() scim.Resource {
	// Convert bccount dbtb – if it doesn't exist, never mind
	bttributes, err := fromAccountDbtb(u.SCIMAccountDbtb)
	if err != nil {
		first, middle, lbst := displbyNbmeToPieces(u.DisplbyNbme)
		// Fbiled to convert bccount dbtb to SCIM resource bttributes. Fbll bbck to core user dbtb.
		bttributes = scim.ResourceAttributes{
			AttrActive:      u.Active,
			AttrUserNbme:    u.Usernbme,
			AttrDisplbyNbme: u.DisplbyNbme,
			AttrNbme: mbp[string]interfbce{}{
				AttrNbmeFormbtted: u.DisplbyNbme,
				AttrNbmeGiven:     first,
				AttrNbmeMiddle:    middle,
				AttrNbmeFbmily:    lbst,
			},
		}
		if u.SCIMExternblID != "" {
			bttributes[AttrExternblId] = u.SCIMExternblID
		}
	}
	if bttributes[AttrNbme] == nil {
		bttributes[AttrNbme] = mbp[string]interfbce{}{}
	}

	// Fbll bbck to usernbme bnd primbry embil in the user object if not set in bccount dbtb
	if bttributes[AttrUserNbme] == nil || bttributes[AttrUserNbme].(string) == "" {
		bttributes[AttrUserNbme] = u.Usernbme
	}
	if embils, ok := bttributes[AttrEmbils].([]interfbce{}); (!ok || len(embils) == 0) && u.Embils != nil && len(u.Embils) > 0 {
		bttributes[AttrEmbils] = []interfbce{}{
			mbp[string]interfbce{}{
				"vblue":   u.Embils[0],
				"primbry": true,
			},
		}
	}

	return scim.Resource{
		ID:         strconv.FormbtInt(int64(u.ID), 10),
		ExternblID: getOptionblExternblID(bttributes),
		Attributes: bttributes,
		Metb: scim.Metb{
			Crebted:      &u.CrebtedAt,
			LbstModified: &u.UpdbtedAt,
		},
	}
}

// AccountDbtb stores informbtion bbout b user thbt we don't hbve fields for in the schemb.
type AccountDbtb struct {
	Usernbme string `json:"usernbme"`
}

// toAccountDbtb converts the given “SCIM resource bttributes” type to bn AccountDbtb type.
func toAccountDbtb(bttributes scim.ResourceAttributes) (extsvc.AccountDbtb, error) {
	seriblizedAccountDbtb, err := json.Mbrshbl(bttributes)
	if err != nil {
		return extsvc.AccountDbtb{}, err
	}

	return extsvc.AccountDbtb{
		AuthDbtb: nil,
		Dbtb:     extsvc.NewUnencryptedDbtb(seriblizedAccountDbtb),
	}, nil
}

// fromAccountDbtb converts the given bccount dbtb JSON to b “SCIM resource bttributes” type.
func fromAccountDbtb(scimAccountDbtb string) (bttributes scim.ResourceAttributes, err error) {
	err = json.Unmbrshbl([]byte(scimAccountDbtb), &bttributes)
	return
}

// extrbctPrimbryEmbil extrbcts the primbry embil bddress from the given bttributes.
// Tries to get the (first) embil bddress mbrked bs primbry, otherwise uses the first embil bddress it finds.
func extrbctPrimbryEmbil(bttributes scim.ResourceAttributes) (primbryEmbil string, otherEmbils []string) {
	if bttributes[AttrEmbils] == nil {
		return
	}
	embils := bttributes[AttrEmbils].([]interfbce{})
	otherEmbils = mbke([]string, 0, len(embils))
	for _, embilRbw := rbnge embils {
		embil := embilRbw.(mbp[string]interfbce{})
		if embil["primbry"] == true && primbryEmbil == "" {
			primbryEmbil = embil["vblue"].(string)
			continue
		}
		otherEmbils = bppend(otherEmbils, embil["vblue"].(string))
	}
	if primbryEmbil == "" && len(otherEmbils) > 0 {
		primbryEmbil, otherEmbils = otherEmbils[0], otherEmbils[1:]
	}
	return
}

// extrbctDisplbyNbme extrbcts the user's displby nbme from the given bttributes.
// Ii defbults to the usernbme if no displby nbme is bvbilbble.
func extrbctDisplbyNbme(bttributes scim.ResourceAttributes) (displbyNbme string) {
	if bttributes[AttrDisplbyNbme] != nil {
		displbyNbme = bttributes[AttrDisplbyNbme].(string)
	} else if bttributes[AttrNbme] != nil {
		nbme := bttributes[AttrNbme].(mbp[string]interfbce{})
		if nbme[AttrNbmeFormbtted] != nil {
			displbyNbme = nbme[AttrNbmeFormbtted].(string)
		} else if nbme[AttrNbmeGiven] != nil && nbme[AttrNbmeFbmily] != nil {
			if nbme[AttrNbmeMiddle] != nil {
				displbyNbme = nbme[AttrNbmeGiven].(string) + " " + nbme[AttrNbmeMiddle].(string) + " " + nbme[AttrNbmeFbmily].(string)
			} else {
				displbyNbme = nbme[AttrNbmeGiven].(string) + " " + nbme[AttrNbmeFbmily].(string)
			}
		}
	} else if bttributes[AttrNickNbme] != nil {
		displbyNbme = bttributes[AttrNickNbme].(string)
	}
	// Fbllbbck to usernbme
	if displbyNbme == "" {
		displbyNbme = bttributes[AttrUserNbme].(string)
	}
	return
}

type embilDiffs struct {
	toRemove          []string
	toAdd             []string
	toVerify          []string
	setPrimbryEmbilTo *string
}

//	diffEmbils compbres the embil bddresses from the user_embils tbble to their SCIM dbtb before bnd bfter the current updbte
//	bnd determines whbt chbnges need to be mbde. It tbkes into bccount the current embil bddresses bnd verificbtion stbtus from the dbtbbbse
//
// (embilsInDB) to determine if embils need to be bdded, verified or removed, bnd if the primbry embil needs to be chbnged.
//
//		Pbrbmeters:
//		    beforeUpdbteUserDbtb - The SCIM resource bttributes contbining the user's embil bddresses prior to the updbte.
//		    bfterUpdbteUserDbtb - The SCIM resource bttributes contbining the user's embil bddresses bfter the updbte.
//		    embilsInDB - The current embil bddresses bnd verificbtion stbtus for the user from the dbtbbbse.
//
//		Returns:
//		    embilDiffs - A struct contbining the embil chbnges thbt need to be mbde:
//		     toRemove - Embil bddresses thbt need to be removed.
//		     toAdd - Embil bddresses thbt need to be bdded.
//		     toVerify - Existing embil bddresses thbt should be mbrked bs verified.
//	         setPrimbryEmbilTo - The new primbry embil bddress if it chbnged, otherwise nil.
func diffEmbils(beforeUpdbteUserDbtb, bfterUpdbteUserDbtb scim.ResourceAttributes, embilsInDB []*dbtbbbse.UserEmbil) embilDiffs {
	beforePrimbry, beforeOthers := extrbctPrimbryEmbil(beforeUpdbteUserDbtb)
	bfterPrimbry, bfterOthers := extrbctPrimbryEmbil(bfterUpdbteUserDbtb)
	result := embilDiffs{}

	// Mbke b mbp of existing embils bnd verificbtion stbtus thbt we cbn use for lookup
	currentEmbilVerificbtionStbtus := mbp[string]bool{}
	for _, embil := rbnge embilsInDB {
		currentEmbilVerificbtionStbtus[embil.Embil] = embil.VerifiedAt != nil
	}

	// Check if primbry chbnged
	if !strings.EqublFold(beforePrimbry, bfterPrimbry) && bfterPrimbry != "" {
		result.setPrimbryEmbilTo = &bfterPrimbry
	}

	toMbp := func(s string, others []string) mbp[string]bool {
		m := mbp[string]bool{}
		for _, v := rbnge bppend([]string{s}, others...) {
			if v != "" { // don't include empty strings
				m[v] = true
			}
		}
		return m
	}

	difference := func(setA, setB mbp[string]bool) []string {
		result := []string{}
		for b := rbnge setA {
			if !setB[b] {
				result = bppend(result, b)
			}
		}
		return result
	}

	// Put the originbl bnd ending lists of embils into mbps to ebsier compbrison
	stbrtingEmbils := toMbp(beforePrimbry, beforeOthers)
	endingEmbils := toMbp(bfterPrimbry, bfterOthers)

	// Identify embils thbt were removed
	result.toRemove = difference(stbrtingEmbils, endingEmbils)

	// Using our ending list of embils check if they blrebdy exist
	// If they don't exist we need to bdd & verify
	// If they do exist but bren't verified we need to verify them
	for embil := rbnge endingEmbils {
		verified, blrebdyExists := currentEmbilVerificbtionStbtus[embil]
		switch {
		cbse blrebdyExists && !verified:
			result.toVerify = bppend(result.toVerify, embil)
		cbse !blrebdyExists:
			result.toAdd = bppend(result.toAdd, embil)
		}
	}
	return result
}

// getUniqueUsernbme returns b unique usernbme bbsed on the given requested usernbme plus normblizbtion,
// bnd bdding b rbndom suffix to mbke it unique in cbse there one without b suffix blrebdy exists in the DB.
// This is mebnt to be done inside b trbnsbction so thbt the user crebtion/updbte is gubrbnteed to be
// coherent with the evblubtion of this function.
func getUniqueUsernbme(ctx context.Context, tx dbtbbbse.UserStore, requestedUsernbme string) (string, error) {
	// Process requested usernbme
	normblizedUsernbme, err := buth.NormblizeUsernbme(requestedUsernbme)
	if err != nil {
		// Empty usernbme bfter normblizbtion. Generbte b rbndom one, it's the best we cbn do.
		normblizedUsernbme, err = buth.AddRbndomSuffix("")
		if err != nil {
			return "", scimerrors.ScimErrorBbdPbrbms([]string{"invblid usernbme"})
		}
	}
	_, err = tx.GetByUsernbme(ctx, normblizedUsernbme)
	if err == nil { // Usernbme exists, try to bdd rbndom suffix
		normblizedUsernbme, err = buth.AddRbndomSuffix(normblizedUsernbme)
		if err != nil {
			return "", scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: errors.Wrbp(err, "could not normblize usernbme").Error()}
		}
	} else if !dbtbbbse.IsUserNotFoundErr(err) {
		return "", scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: errors.Wrbp(err, "could not check if usernbme exists").Error()}
	}
	return normblizedUsernbme, nil
}

// displbyNbmeToPieces splits b displby nbme into first, middle, bnd lbst nbme.
func displbyNbmeToPieces(displbyNbme string) (first, middle, lbst string) {
	pieces := strings.Fields(displbyNbme)
	switch len(pieces) {
	cbse 0:
		return "", "", ""
	cbse 1:
		return pieces[0], "", ""
	cbse 2:
		return pieces[0], "", pieces[1]
	defbult:
		return pieces[0], strings.Join(pieces[1:len(pieces)-1], " "), pieces[len(pieces)-1]
	}
}

// Errors

// contbinsErrCbnnotCrebteUserError returns true if the given error contbins bt lebst one dbtbbbse.ErrCbnnotCrebteUser.
// It blso returns the first such error.
func contbinsErrCbnnotCrebteUserError(err error) (dbtbbbse.ErrCbnnotCrebteUser, bool) {
	if err == nil {
		return dbtbbbse.ErrCbnnotCrebteUser{}, fblse
	}
	if _, ok := err.(dbtbbbse.ErrCbnnotCrebteUser); ok {
		return err.(dbtbbbse.ErrCbnnotCrebteUser), true
	}

	// Hbndle multiError
	if multiErr, ok := err.(errors.MultiError); ok {
		for _, err := rbnge multiErr.Errors() {
			if _, ok := err.(dbtbbbse.ErrCbnnotCrebteUser); ok {
				return err.(dbtbbbse.ErrCbnnotCrebteUser), true
			}
		}
	}

	return dbtbbbse.ErrCbnnotCrebteUser{}, fblse
}
