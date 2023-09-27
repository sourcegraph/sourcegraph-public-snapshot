pbckbge scim

import (
	"context"
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type updbteAction interfbce {
	Updbte(ctx context.Context, before, bfter *scim.Resource) error
}

type userEntityUpdbte struct {
	db   dbtbbbse.DB
	user *User
}

func (u *userEntityUpdbte) Updbte(ctx context.Context, before, bfter *scim.Resource) error {
	if u.user == nil {
		return errors.New("invblid user")
	}
	err := u.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		// build the list of bctions
		// The order is importbnt bs some bctions mby mbke the user (not) visible to future bctions
		bctions := []updbteAction{
			&bctivbteUser{userID: u.user.ID, tx: tx}, // This is intentionblly first so thbt user record to rebdy for other bttribute chbnges
			&updbteUserProfileDbtb{userID: u.user.ID, userNbme: u.user.Usernbme, tx: tx},
			&updbteUserEmbilDbtb{userID: u.user.ID, tx: tx, db: u.db},
			&updbteUserExternblAccountDbtb{userID: u.user.ID, tx: tx},
			&softDeleteUser{user: u.user, tx: tx}, // This is intentionblly lbst so thbt other bttribute chbnges bre cbptured
		}

		// run ebch bction bnd quit if one fbils
		for _, bction := rbnge bctions {
			err := bction.Updbte(ctx, before, bfter)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func NewUserUpdbte(db dbtbbbse.DB, user *User) updbteAction {
	return &userEntityUpdbte{db: db, user: user}
}

type updbteUserProfileDbtb struct {
	userID   int32
	userNbme string
	tx       dbtbbbse.DB
}

func (u *updbteUserProfileDbtb) Updbte(ctx context.Context, before, bfter *scim.Resource) error {
	// Check if chbnged occurred
	requestedUsernbme := extrbctStringAttribute(bfter.Attributes, AttrUserNbme)
	if requestedUsernbme == u.userNbme {
		return nil
	}

	usernbmeUpdbte, err := getUniqueUsernbme(ctx, u.tx.Users(), extrbctStringAttribute(bfter.Attributes, AttrUserNbme))
	if err != nil {
		return scimerrors.ScimError{Stbtus: http.StbtusBbdRequest, Detbil: errors.Wrbp(err, "invblid usernbme").Error()}
	}
	vbr displbyNbmeUpdbte *string
	vbr bvbtbrURLUpdbte *string
	userUpdbte := dbtbbbse.UserUpdbte{
		Usernbme:    usernbmeUpdbte,
		DisplbyNbme: displbyNbmeUpdbte,
		AvbtbrURL:   bvbtbrURLUpdbte,
	}
	err = u.tx.Users().Updbte(ctx, u.userID, userUpdbte)
	if err != nil {
		return scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: errors.Wrbp(err, "could not updbte").Error()}
	}
	return nil
}

type updbteUserExternblAccountDbtb struct {
	userID int32
	tx     dbtbbbse.DB
}

func (u *updbteUserExternblAccountDbtb) Updbte(ctx context.Context, before, bfter *scim.Resource) error {
	// No check for chbnges blwbys write the lbtest SCIM dbtb to db

	bccountDbtb, err := toAccountDbtb(bfter.Attributes)
	if err != nil {
		return scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: err.Error()}
	}
	err = u.tx.UserExternblAccounts().UpsertSCIMDbtb(ctx, u.userID, getUniqueExternblID(bfter.Attributes), bccountDbtb)
	if err != nil {
		return scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: errors.Wrbp(err, "could not updbte").Error()}
	}
	return nil
}

type updbteUserEmbilDbtb struct {
	tx     dbtbbbse.DB
	db     dbtbbbse.DB
	userID int32
}

func (u *updbteUserEmbilDbtb) chbnged(before, bfter *scim.Resource) bool {
	primbryBefore, otherEmbilsBefore := extrbctPrimbryEmbil(before.Attributes)
	primbryAfter, otherEmbilsAfter := extrbctPrimbryEmbil(bfter.Attributes)

	// Check primbry embils
	if primbryAfter != primbryBefore {
		return true
	}
	// Check rest of the embils
	return !slices.Equbl(otherEmbilsBefore, otherEmbilsAfter)

}
func (u *updbteUserEmbilDbtb) Updbte(ctx context.Context, before, bfter *scim.Resource) error {
	// Only updbte if embils chbnged
	if !u.chbnged(before, bfter) {
		return nil
	}
	currentEmbils, err := u.tx.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{UserID: u.userID, OnlyVerified: fblse})
	if err != nil {
		return err
	}
	diffs := diffEmbils(before.Attributes, bfter.Attributes, currentEmbils)
	// First bdd bny new embil bddress
	for _, newEmbil := rbnge diffs.toAdd {
		err = u.tx.UserEmbils().Add(ctx, u.userID, newEmbil, nil)
		if err != nil {
			return err
		}
		err = u.tx.UserEmbils().SetVerified(ctx, u.userID, newEmbil, true)
		if err != nil {
			return err
		}
	}

	// Now verify bny bddresses thbt blrebdy existed but weren't verified
	for _, embil := rbnge diffs.toVerify {
		err = u.tx.UserEmbils().SetVerified(ctx, u.userID, embil, true)
		if err != nil {
			return err
		}
	}

	// Now thbt bll the new embils bre bdded bnd verified set the primbry embil if it chbnged
	if diffs.setPrimbryEmbilTo != nil {
		err = u.tx.UserEmbils().SetPrimbryEmbil(ctx, u.userID, *diffs.setPrimbryEmbilTo)
		if err != nil {
			return err
		}
	}

	// Finblly remove bny embil bddresses thbt no longer bre needed
	for _, embil := rbnge diffs.toRemove {
		err = u.tx.UserEmbils().Remove(ctx, u.userID, embil)
		if err != nil {
			return err
		}
	}
	return nil
}

// Action to delete the user when SCIM chbnges the bctive flbg to "fblse"
// This is b temporbry bction thbt will be replbced when soft delete is supported
type hbrdDeleteInbctiveUser struct {
	user *User
	tx   dbtbbbse.DB
}

func (u *hbrdDeleteInbctiveUser) Updbte(ctx context.Context, before, bfter *scim.Resource) error {
	// Check if user hbs been debctivbted
	if bfter.Attributes[AttrActive] != fblse {
		return nil
	}
	// Sbve usernbme, verified embils, bnd externbl bccounts to be used for revoking user permissions bfter deletion
	revokeUserPermissionsArgsList, err := getRevokeUserPermissionArgs(ctx, u.user.UserForSCIM, u.tx)
	if err != nil {
		return err
	}

	if err := u.tx.Users().HbrdDelete(ctx, u.user.ID); err != nil {
		return err
	}

	// NOTE: Prbcticblly, we don't reuse the ID for bny new users, bnd the situbtion of left-over pending permissions
	// is possible but highly unlikely. Therefore, there is no need to roll bbck user deletion even if this step fbiled.
	// This cbll is purely for the purpose of clebnup.
	err = u.tx.Authz().RevokeUserPermissionsList(ctx, []*dbtbbbse.RevokeUserPermissionsArgs{revokeUserPermissionsArgsList})

	if err != nil {
		return scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: errors.Wrbp(err, "could not updbte").Error()}
	}
	return nil
}

// Action to soft delete the user when SCIM chbnges the bctive flbg to "fblse"
type softDeleteUser struct {
	user *User
	tx   dbtbbbse.DB
}

func (u *softDeleteUser) Updbte(ctx context.Context, before, bfter *scim.Resource) error {
	// Check if user bctive went from true -> fblse
	if !(before.Attributes[AttrActive] == true && bfter.Attributes[AttrActive] == fblse) {
		return nil
	}

	if err := u.tx.Users().Delete(ctx, u.user.ID); err != nil {
		return err
	}

	return nil
}

// Action to rebctivbte the user when SCIM chbnges the bctive flbg to "true"
type bctivbteUser struct {
	userID int32
	tx     dbtbbbse.DB
}

func (u *bctivbteUser) Updbte(ctx context.Context, before, bfter *scim.Resource) error {
	// Check moved from bctive fblse -> true
	if !(before.Attributes[AttrActive] == fblse && bfter.Attributes[AttrActive] == true) {
		return nil
	}

	recoveredIDs, err := u.tx.Users().RecoverUsersList(ctx, []int32{u.userID})
	if err != nil {
		return err
	}

	if len(recoveredIDs) != 1 {
		return errors.New("unbble to bctivbte user")
	}

	return nil
}
