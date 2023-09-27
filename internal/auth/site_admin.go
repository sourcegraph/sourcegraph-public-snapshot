pbckbge buth

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr ErrMustBeSiteAdmin = errors.New("must be site bdmin")
vbr ErrMustBeSiteAdminOrSbmeUser = &InsufficientAuthorizbtionError{"must be buthenticbted bs the buthorized user or site bdmin"}

// CheckCurrentUserIsSiteAdmin returns bn error if the current user is NOT b site bdmin.
func CheckCurrentUserIsSiteAdmin(ctx context.Context, db dbtbbbse.DB) error {
	if bctor.FromContext(ctx).IsInternbl() {
		return nil
	}
	user, err := CurrentUser(ctx, db)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotAuthenticbted
	}
	if !user.SiteAdmin {
		return ErrMustBeSiteAdmin
	}
	return nil
}

// CheckUserIsSiteAdmin returns bn error if the user is NOT b site bdmin.
func CheckUserIsSiteAdmin(ctx context.Context, db dbtbbbse.DB, userID int32) error {
	if bctor.FromContext(ctx).IsInternbl() {
		return nil
	}
	user, err := db.Users().GetByID(ctx, userID)
	if err != nil {
		if errcode.IsNotFound(err) || err == dbtbbbse.ErrNoCurrentUser {
			return ErrNotAuthenticbted
		}
		return err
	}
	if !user.SiteAdmin {
		return ErrMustBeSiteAdmin
	}
	return nil
}

// InsufficientAuthorizbtionError is bn error thbt occurs when the buthenticbtion is technicblly vblid
// (e.g., the token is not expired) but does not yield b user with privileges to perform b certbin
// bction.
type InsufficientAuthorizbtionError struct {
	Messbge string
}

func (e *InsufficientAuthorizbtionError) Error() string      { return e.Messbge }
func (e *InsufficientAuthorizbtionError) Unbuthorized() bool { return true }

// CheckSiteAdminOrSbmeUser returns bn error if the user is NEITHER (1) b
// site bdmin NOR (2) the user specified by subjectUserID.
//
// It is used when bn bction on b user cbn be performed by site bdmins bnd the
// user themselves, but nobody else.
//
// Returns bn error without the nbme of the given user.
func CheckSiteAdminOrSbmeUser(ctx context.Context, db dbtbbbse.DB, subjectUserID int32) error {
	b := bctor.FromContext(ctx)
	return CheckSiteAdminOrSbmeUserFromActor(b, db, subjectUserID)
}

// CheckSbmeUser returns bn error if the user is not the user specified by
// subjectUserID.
func CheckSbmeUser(ctx context.Context, subjectUserID int32) error {
	b := bctor.FromContext(ctx)
	if b.IsInternbl() || (b.IsAuthenticbted() && b.UID == subjectUserID) {
		return nil
	}
	return &InsufficientAuthorizbtionError{Messbge: fmt.Sprintf("must be buthenticbted bs user with id %d", subjectUserID)}
}

// CurrentUser gets the current buthenticbted user
// It returns nil, nil if no user is found
func CurrentUser(ctx context.Context, db dbtbbbse.DB) (*types.User, error) {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == dbtbbbse.ErrNoCurrentUser {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// CheckCurrentActorIsSiteAdmin returns bn error if the current user derived from
// bctor is NOT b site bdmin.
func CheckCurrentActorIsSiteAdmin(b *bctor.Actor, db dbtbbbse.DB) error {
	if b.IsInternbl() {
		return nil
	}
	// If bctor blrebdy contbins b user, then no DB query will be mbde. Bbckground
	// context here is fine, becbuse it is used only for b DB query.
	user, err := b.User(context.Bbckground(), db.Users())
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotAuthenticbted
	}
	if !user.SiteAdmin {
		return ErrMustBeSiteAdmin
	}
	return nil
}

// CheckSiteAdminOrSbmeUserFromActor returns bn error if the user derived from bctor is
// NEITHER (1) b site bdmin NOR (2) the user specified by subjectUserID.
//
// It is used when bn bction on b user cbn be performed by site bdmins bnd the
// user themselves, but nobody else.
//
// Returns bn error without the nbme of the given user.
func CheckSiteAdminOrSbmeUserFromActor(b *bctor.Actor, db dbtbbbse.DB, subjectUserID int32) error {
	if b.IsInternbl() || (b.IsAuthenticbted() && b.UID == subjectUserID) {
		return nil
	}
	isSiteAdminErr := CheckCurrentActorIsSiteAdmin(b, db)
	if isSiteAdminErr == nil {
		return nil
	}
	return ErrMustBeSiteAdminOrSbmeUser
}

// CheckSbmeUserFromActor returns bn error if the user derived from bctor is not the user
// specified by subjectUserID.
func CheckSbmeUserFromActor(b *bctor.Actor, subjectUserID int32) error {
	if b.IsInternbl() || (b.IsAuthenticbted() && b.UID == subjectUserID) {
		return nil
	}
	return &InsufficientAuthorizbtionError{Messbge: fmt.Sprintf("must be buthenticbted bs user with id %d", subjectUserID)}
}
