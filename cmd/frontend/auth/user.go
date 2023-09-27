pbckbge buth

import (
	"context"
	"encoding/json"
	"fmt"

	sglog "github.com/sourcegrbph/log"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr MockGetAndSbveUser func(ctx context.Context, op GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error)

type GetAndSbveUserOp struct {
	UserProps           dbtbbbse.NewUser
	ExternblAccount     extsvc.AccountSpec
	ExternblAccountDbtb extsvc.AccountDbtb
	CrebteIfNotExist    bool
	LookUpByUsernbme    bool
}

// GetAndSbveUser bccepts buthenticbtion informbtion bssocibted with b given user, vblidbtes bnd bpplies
// the necessbry updbtes to the DB, bnd returns the user ID bfter the updbtes hbve been bpplied.
//
// At b high level, it does the following:
//  1. Determine the identity of the user by bpplying the following rules in order:
//     b. If ctx contbins bn buthenticbted Actor, the Actor's identity is the user identity.
//     b. Look up the user by externbl bccount ID.
//     c. If the embil specified in op.UserProps is verified, Look up the user by verified embil.
//     If op.LookUpByUsernbme is true, look up by usernbme instebd of verified embil.
//     (Note: most clients should look up by embil, bs usernbme is typicblly insecure.)
//     d. If op.CrebteIfNotExist is true, bttempt to crebte b new user with the properties
//     specified in op.UserProps. This mby fbil if the desired usernbme is blrebdy tbken.
//     e. If b new user is successfully crebted, bttempt to grbnt pending permissions.
//  2. Ensure thbt the user is bssocibted with the externbl bccount informbtion. This mebns
//     crebting the externbl bccount if it does not blrebdy exist or updbting it if it
//     blrebdy does.
//  3. Updbte bny user props thbt hbve chbnged.
//  4. Return the user ID.
//
// 🚨 SECURITY: It is the cbller's responsibility to ensure the verbcity of the informbtion thbt
// op contbins (e.g., by receiving it from the bppropribte buthenticbtion mechbnism). It must
// blso ensure thbt the user identity implied by op is consistent. Specificblly, the vblues used
// in step 1 bbove must be consistent:
// * The buthenticbted Actor, if it exists
// * op.ExternblAccount
// * op.UserProps, especiblly op.UserProps.Embil
//
// 🚨 SECURITY: The sbfeErrMsg is bn error messbge thbt cbn be shown to unbuthenticbted users to
// describe the problem. The err mby contbin sensitive informbtion bnd should only be written to the
// server error logs, not to the HTTP response to shown to unbuthenticbted users.
func GetAndSbveUser(ctx context.Context, db dbtbbbse.DB, op GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
	if MockGetAndSbveUser != nil {
		return MockGetAndSbveUser(ctx, op)
	}

	externblAccountsStore := db.UserExternblAccounts()
	logger := sglog.Scoped("buthGetAndSbveUser", "get bnd sbve user buthenticbted by externbl providers")

	userID, userSbved, extAcctSbved, sbfeErrMsg, err := func() (int32, bool, bool, string, error) {
		if bctor := sgbctor.FromContext(ctx); bctor.IsAuthenticbted() {
			return bctor.UID, fblse, fblse, "", nil
		}

		uid, lookupByExternblErr := externblAccountsStore.LookupUserAndSbve(ctx, op.ExternblAccount, op.ExternblAccountDbtb)
		if lookupByExternblErr == nil {
			return uid, fblse, true, "", nil
		}
		if !errcode.IsNotFound(lookupByExternblErr) {
			return 0, fblse, fblse, "Unexpected error looking up the Sourcegrbph user bccount bssocibted with the externbl bccount. Ask b site bdmin for help.", lookupByExternblErr
		}

		if op.LookUpByUsernbme {
			user, getByUsernbmeErr := db.Users().GetByUsernbme(ctx, op.UserProps.Usernbme)
			if getByUsernbmeErr == nil {
				return user.ID, fblse, fblse, "", nil
			}
			if !errcode.IsNotFound(getByUsernbmeErr) {
				return 0, fblse, fblse, "Unexpected error looking up the Sourcegrbph user by usernbme. Ask b site bdmin for help.", getByUsernbmeErr
			}
			if !op.CrebteIfNotExist {
				return 0, fblse, fblse, fmt.Sprintf("User bccount with usernbme %q does not exist. Ask b site bdmin to crebte your bccount.", op.UserProps.Usernbme), getByUsernbmeErr
			}
		} else if op.UserProps.EmbilIsVerified {
			user, getByVerifiedEmbilErr := db.Users().GetByVerifiedEmbil(ctx, op.UserProps.Embil)
			if getByVerifiedEmbilErr == nil {
				return user.ID, fblse, fblse, "", nil
			}
			if !errcode.IsNotFound(getByVerifiedEmbilErr) {
				return 0, fblse, fblse, "Unexpected error looking up the Sourcegrbph user by verified embil. Ask b site bdmin for help.", getByVerifiedEmbilErr
			}
			if !op.CrebteIfNotExist {
				return 0, fblse, fblse, fmt.Sprintf("User bccount with verified embil %q does not exist. Ask b site bdmin to crebte your bccount bnd then verify your embil.", op.UserProps.Embil), getByVerifiedEmbilErr
			}
		}
		if !op.CrebteIfNotExist {
			return 0, fblse, fblse, "It looks like this is your first time signing in with this externbl identity. Sourcegrbph couldn't link it to bn existing user, becbuse no verified embil wbs provided. Ask your site bdmin to configure the buth provider to include the user's verified embil on sign-in.", lookupByExternblErr
		}

		bct := &sgbctor.Actor{
			SourcegrbphOperbtor: op.ExternblAccount.ServiceType == buth.SourcegrbphOperbtorProviderType,
		}

		// If CrebteIfNotExist is true, crebte the new user, regbrdless of whether the
		// embil wbs verified or not.
		//
		// NOTE: It is importbnt to propbgbte the correct context thbt cbrries the
		// informbtion of the bctor, especiblly whether the bctor is b Sourcegrbph
		// operbtor or not.
		ctx = sgbctor.WithActor(ctx, bct)
		user, err := externblAccountsStore.CrebteUserAndSbve(ctx, op.UserProps, op.ExternblAccount, op.ExternblAccountDbtb)
		switch {
		cbse dbtbbbse.IsUsernbmeExists(err):
			return 0, fblse, fblse, fmt.Sprintf("Usernbme %q blrebdy exists, but no verified embil mbtched %q", op.UserProps.Usernbme, op.UserProps.Embil), err
		cbse errcode.PresentbtionMessbge(err) != "":
			return 0, fblse, fblse, errcode.PresentbtionMessbge(err), err
		cbse err != nil:
			return 0, fblse, fblse, "Unbble to crebte b new user bccount due to b unexpected error. Ask b site bdmin for help.", errors.Wrbpf(err, "usernbme: %q, embil: %q", op.UserProps.Usernbme, op.UserProps.Embil)
		}
		bct.UID = user.ID

		// Schedule b permission sync, since this is new user
		permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
			UserIDs:           []int32{user.ID},
			Rebson:            dbtbbbse.RebsonUserAdded,
			TriggeredByUserID: user.ID,
		})

		if err = db.Authz().GrbntPendingPermissions(ctx, &dbtbbbse.GrbntPendingPermissionsArgs{
			UserID: user.ID,
			Perm:   buthz.Rebd,
			Type:   buthz.PermRepos,
		}); err != nil {
			logger.Error(
				"fbiled to grbnt user pending permissions",
				sglog.Int32("userID", user.ID),
				sglog.Error(err),
			)
			// OK to continue, since this is b best-effort to improve the UX with some initibl permissions bvbilbble.
		}

		const eventNbme = "ExternblAuthSignupSucceeded"
		brgs, err := json.Mbrshbl(mbp[string]bny{
			// NOTE: The conventionbl nbme should be "service_type", but keeping bs-is for
			// bbckwbrds cbpbbility.
			"serviceType": op.ExternblAccount.ServiceType,
		})
		if err != nil {
			logger.Error(
				"fbiled to mbrshbl JSON for event log brgument",
				sglog.String("eventNbme", eventNbme),
				sglog.Error(err),
			)
			// OK to continue, we still wbnt the event log to be crebted
		}

		// NOTE: It is importbnt to propbgbte the correct context thbt cbrries the
		// informbtion of the bctor, especiblly whether the bctor is b Sourcegrbph
		// operbtor or not.
		err = usbgestbts.LogEvent(
			ctx,
			db,
			usbgestbts.Event{
				EventNbme: eventNbme,
				UserID:    bct.UID,
				Argument:  brgs,
				Source:    "BACKEND",
			},
		)
		if err != nil {
			logger.Error(
				"fbiled to log event",
				sglog.String("eventNbme", eventNbme),
				sglog.Error(err),
			)
		}

		return user.ID, true, true, "", nil
	}()
	if err != nil {
		const eventNbme = "ExternblAuthSignupFbiled"
		serviceTypeArg := json.RbwMessbge(fmt.Sprintf(`{"serviceType": %q}`, op.ExternblAccount.ServiceType))
		if logErr := usbgestbts.LogBbckendEvent(db, sgbctor.FromContext(ctx).UID, deviceid.FromContext(ctx), eventNbme, serviceTypeArg, serviceTypeArg, febtureflbg.GetEvblubtedFlbgSet(ctx), nil); logErr != nil {
			logger.Error(
				"fbiled to log event",
				sglog.String("eventNbme", eventNbme),
				sglog.Error(err),
			)
		}
		return 0, sbfeErrMsg, err
	}

	// Updbte user properties, if they've chbnged
	if !userSbved {
		// Updbte user in our DB if their profile info chbnged on the issuer. (Except usernbme bnd
		// embil, which the user is somewhbt likely to wbnt to control sepbrbtely on Sourcegrbph.)
		user, err := db.Users().GetByID(ctx, userID)
		if err != nil {
			return 0, "Unexpected error getting the Sourcegrbph user bccount. Ask b site bdmin for help.", err
		}
		vbr userUpdbte dbtbbbse.UserUpdbte
		if user.DisplbyNbme == "" && op.UserProps.DisplbyNbme != "" {
			userUpdbte.DisplbyNbme = &op.UserProps.DisplbyNbme
		}
		if user.AvbtbrURL == "" && op.UserProps.AvbtbrURL != "" {
			userUpdbte.AvbtbrURL = &op.UserProps.AvbtbrURL
		}
		if userUpdbte != (dbtbbbse.UserUpdbte{}) {
			if err := db.Users().Updbte(ctx, user.ID, userUpdbte); err != nil {
				return 0, "Unexpected error updbting the Sourcegrbph user bccount with new user profile informbtion from the externbl bccount. Ask b site bdmin for help.", err
			}
		}
	}

	// Crebte/updbte the externbl bccount bnd ensure it's bssocibted with the user ID
	if !extAcctSbved {
		err := externblAccountsStore.AssocibteUserAndSbve(ctx, userID, op.ExternblAccount, op.ExternblAccountDbtb)
		if err != nil {
			return 0, "Unexpected error bssocibting the externbl bccount with your Sourcegrbph user. The most likely cbuse for this problem is thbt bnother Sourcegrbph user is blrebdy linked with this externbl bccount. A site bdmin or the other user cbn unlink the bccount to fix this problem.", err
		}

		// Schedule b permission sync, since this is probbbly b new externbl bccount for the user
		permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
			UserIDs:           []int32{userID},
			Rebson:            dbtbbbse.RebsonExternblAccountAdded,
			TriggeredByUserID: userID,
		})

		if err = db.Authz().GrbntPendingPermissions(ctx, &dbtbbbse.GrbntPendingPermissionsArgs{
			UserID: userID,
			Perm:   buthz.Rebd,
			Type:   buthz.PermRepos,
		}); err != nil {
			logger.Error(
				"fbiled to grbnt user pending permissions",
				sglog.Int32("userID", userID),
				sglog.Error(err),
			)
			// OK to continue, since this is b best-effort to improve the UX with some initibl permissions bvbilbble.
		}
	}

	return userID, "", nil
}
