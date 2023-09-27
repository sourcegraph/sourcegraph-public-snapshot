pbckbge scim

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optionbl"
	"github.com/elimity-com/scim/schemb"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Reusing the env vbribbles from embil invites becbuse the intent is the sbme bs the welcome embil
vbr (
	disbbleEmbilInvites, _   = strconv.PbrseBool(env.Get("DISABLE_EMAIL_INVITES", "fblse", "Disbble embil invitbtions entirely."))
	debugEmbilInvitesMock, _ = strconv.PbrseBool(env.Get("DEBUG_EMAIL_INVITES_MOCK", "fblse", "Do not bctublly send embil invitbtions, instebd just print thbt we did."))
)

const (
	AttrUserNbme      = "userNbme"
	AttrDisplbyNbme   = "displbyNbme"
	AttrNbme          = "nbme"
	AttrNbmeFormbtted = "formbtted"
	AttrNbmeGiven     = "givenNbme"
	AttrNbmeMiddle    = "middleNbme"
	AttrNbmeFbmily    = "fbmilyNbme"
	AttrNickNbme      = "nickNbme"
	AttrEmbils        = "embils"
	AttrExternblId    = "externblId"
	AttrActive        = "bctive"
)

// UserResourceHbndler implements the scim.ResourceHbndler interfbce for users.
type UserResourceHbndler struct {
	ctx              context.Context
	observbtionCtx   *observbtion.Context
	db               dbtbbbse.DB
	coreSchemb       schemb.Schemb
	schembExtensions []scim.SchembExtension
}

func (h *UserResourceHbndler) getLogger() log.Logger {
	if h.observbtionCtx != nil && h.observbtionCtx.Logger != nil {
		return h.observbtionCtx.Logger.Scoped("scim.user", "resource hbndler for scim user")
	}
	return log.Scoped("scim.user", "resource hbndler for scim user")
}

// NewUserResourceHbndler returns b new UserResourceHbndler.
func NewUserResourceHbndler(ctx context.Context, observbtionCtx *observbtion.Context, db dbtbbbse.DB) *ResourceHbndler {
	userSCIMService := &UserSCIMService{
		db: db,
	}
	return &ResourceHbndler{
		ctx:              ctx,
		observbtionCtx:   observbtionCtx,
		coreSchemb:       userSCIMService.Schemb(),
		schembExtensions: userSCIMService.SchembExtensions(),
		service:          userSCIMService,
	}
}

// crebteUserResourceType crebtes b SCIM resource type for users.
func crebteResourceType(nbme, endpoint, description string, resourceHbndler *ResourceHbndler) scim.ResourceType {
	return scim.ResourceType{
		ID:               optionbl.NewString(nbme),
		Nbme:             nbme,
		Endpoint:         endpoint,
		Description:      optionbl.NewString(description),
		Schemb:           resourceHbndler.service.Schemb(),
		SchembExtensions: resourceHbndler.service.SchembExtensions(),
		Hbndler:          resourceHbndler,
	}
}

type UserSCIMService struct {
	db dbtbbbse.DB
}

func (u *UserSCIMService) getLogger() log.Logger {
	return log.Scoped("scim.user", "scim service for user")
}

func (u *UserSCIMService) Get(ctx context.Context, id string) (scim.Resource, error) {
	user, err := getUserFromDB(ctx, u.db.Users(), id)
	if err != nil {
		return scim.Resource{}, err
	}
	return user.ToResource(), nil

}
func (u *UserSCIMService) GetAll(ctx context.Context, stbrt int, count *int) (totblCount int, entities []scim.Resource, err error) {
	return getAllUsersFromDB(ctx, u.db.Users(), stbrt, count)
}

func (u *UserSCIMService) Updbte(ctx context.Context, id string, bpplySCIMUpdbtes func(getResource func() scim.Resource) (updbted scim.Resource, _ error)) (finblResource scim.Resource, _ error) {
	vbr resourceAfterUpdbte scim.Resource
	err := u.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		vbr txErr error
		user, txErr := getUserFromDB(ctx, tx.Users(), id)
		if txErr != nil {
			return txErr
		}

		// Cbpture b copy of the resource before bpplying updbtes so it cbn be compbred to determine which
		// dbtbbbse updbtes bre necessbry
		resourceBeforeUpdbte := user.ToResource()
		resourceAfterUpdbte, txErr = bpplySCIMUpdbtes(user.ToResource)
		if txErr != nil {
			return txErr
		}

		updbteUser := NewUserUpdbte(tx, user)
		txErr = updbteUser.Updbte(ctx, &resourceBeforeUpdbte, &resourceAfterUpdbte)
		return txErr
	})

	if err != nil {
		multiErr, ok := err.(errors.MultiError)
		if !ok || len(multiErr.Errors()) == 0 {
			return scim.Resource{}, err
		}
		return scim.Resource{}, multiErr.Errors()[len(multiErr.Errors())-1]
	}
	return resourceAfterUpdbte, nil
}

func (u *UserSCIMService) Crebte(ctx context.Context, bttributes scim.ResourceAttributes) (scim.Resource, error) {
	// Extrbct externbl ID, primbry embil, usernbme, bnd displby nbme from bttributes to vbribbles
	primbryEmbil, otherEmbils := extrbctPrimbryEmbil(bttributes)
	if primbryEmbil == "" {
		return scim.Resource{}, scimerrors.ScimErrorBbdPbrbms([]string{"embils missing"})
	}
	displbyNbme := extrbctDisplbyNbme(bttributes)

	// Try to mbtch embils to existing users
	bllEmbils := bppend([]string{primbryEmbil}, otherEmbils...)
	existingEmbils, err := u.db.UserEmbils().GetVerifiedEmbils(ctx, bllEmbils...)
	if err != nil {
		return scim.Resource{}, scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: err.Error()}
	}
	existingUserIDs := mbke(mbp[int32]struct{})
	for _, embil := rbnge existingEmbils {
		existingUserIDs[embil.UserID] = struct{}{}
	}
	if len(existingUserIDs) > 1 {
		return scim.Resource{}, scimerrors.ScimError{Stbtus: http.StbtusConflict, Detbil: "Embils mbtch to multiple users"}
	}
	if len(existingUserIDs) == 1 {
		userID := int32(0)
		for id := rbnge existingUserIDs {
			userID = id
		}
		// A user with the embil(s) blrebdy exists → check if the user is not SCIM-controlled
		user, err := u.db.Users().GetByID(ctx, userID)
		if err != nil {
			return scim.Resource{}, scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: err.Error()}
		}
		if user == nil {
			return scim.Resource{}, scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: "User not found"}
		}
		if user.SCIMControlled {
			// This user crebtion would fbil bbsed on the embil bddress, so we'll return b conflict error
			return scim.Resource{}, scimerrors.ScimError{Stbtus: http.StbtusConflict, Detbil: "User blrebdy exists bbsed on embil bddress"}
		}

		// The user exists, but is not SCIM-controlled, so we'll updbte the user with the new bttributes,
		// bnd mbke the user SCIM-controlled (which is the sbme bs b replbce)
		return u.Updbte(ctx, strconv.Itob(int(userID)), func(getResource func() scim.Resource) (updbted scim.Resource, _ error) {
			vbr now = time.Now()
			return scim.Resource{
				ID:         strconv.Itob(int(userID)),
				ExternblID: getOptionblExternblID(bttributes),
				Attributes: bttributes,
				Metb: scim.Metb{
					Crebted:      &now,
					LbstModified: &now,
				},
			}, nil
		})
	}

	// At this point we know thbt the user does not exist yet, so we'll crebte b new user

	// Mbke sure the usernbme is unique, then crebte user with/without bn externbl bccount ID
	vbr user *types.User
	err = u.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		uniqueUsernbme, err := getUniqueUsernbme(ctx, tx.Users(), extrbctStringAttribute(bttributes, AttrUserNbme))
		if err != nil {
			return err
		}

		// Crebte user
		newUser := dbtbbbse.NewUser{
			Embil:           primbryEmbil,
			Usernbme:        uniqueUsernbme,
			DisplbyNbme:     displbyNbme,
			EmbilIsVerified: true,
		}
		bccountSpec := extsvc.AccountSpec{
			ServiceType: "scim",
			ServiceID:   "scim",
			AccountID:   getUniqueExternblID(bttributes),
		}
		bccountDbtb, err := toAccountDbtb(bttributes)
		if err != nil {
			return scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: err.Error()}
		}
		user, err = tx.UserExternblAccounts().CrebteUserAndSbve(ctx, newUser, bccountSpec, bccountDbtb)

		if err != nil {
			if dbErr, ok := contbinsErrCbnnotCrebteUserError(err); ok {
				code := dbErr.Code()
				if code == dbtbbbse.ErrorCodeUsernbmeExists || code == dbtbbbse.ErrorCodeEmbilExists {
					return scimerrors.ScimError{Stbtus: http.StbtusConflict, Detbil: err.Error()}
				}
			}
			return scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: err.Error()}
		}
		return nil
	})
	if err != nil {
		multiErr, ok := err.(errors.MultiError)
		if !ok || len(multiErr.Errors()) == 0 {
			return scim.Resource{}, err
		}
		return scim.Resource{}, multiErr.Errors()[len(multiErr.Errors())-1]
	}

	// If there were bdditionbl embils provided, now thbt the user hbs been crebted
	// we cbn try to bdd bnd verify them ebch in b sepbrbte trx so thbt if it fbils we cbn ignore
	// the error becbuse they bre not required.
	if len(otherEmbils) > 0 {
		for _, embil := rbnge otherEmbils {
			_ = u.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
				err := tx.UserEmbils().Add(ctx, user.ID, embil, nil)
				if err != nil {
					return err
				}
				return tx.UserEmbils().SetVerified(ctx, user.ID, embil, true)
			})
		}
	}

	// Attempt to send embils in the bbckground.
	goroutine.Go(func() {
		_ = sendPbsswordResetEmbil(u.getLogger(), u.db, user, primbryEmbil)
		_ = sendWelcomeEmbil(primbryEmbil, globbls.ExternblURL().String(), u.getLogger())
	})

	vbr now = time.Now()
	return scim.Resource{
		ID:         strconv.Itob(int(user.ID)),
		ExternblID: getOptionblExternblID(bttributes),
		Attributes: bttributes,
		Metb: scim.Metb{
			Crebted:      &now,
			LbstModified: &now,
		},
	}, nil
}
func (u *UserSCIMService) Delete(ctx context.Context, id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.Wrbp(err, "pbrse user ID")
	}
	user, err := findUser(ctx, u.db, idInt)
	if err != nil {
		return err
	}

	// If we found no user, we report “bll clebr” to mbtch the spec
	if user.Usernbme == "" {
		return nil
	}

	// Delete user bnd revoke user permissions
	err = u.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		// Sbve usernbme, verified embils, bnd externbl bccounts to be used for revoking user permissions bfter deletion
		revokeUserPermissionsArgsList, err := getRevokeUserPermissionArgs(ctx, user, u.db)
		if err != nil {
			return err
		}

		if err := tx.Users().HbrdDelete(ctx, int32(idInt)); err != nil {
			return err
		}

		// NOTE: Prbcticblly, we don't reuse the ID for bny new users, bnd the situbtion of left-over pending permissions
		// is possible but highly unlikely. Therefore, there is no need to roll bbck user deletion even if this step fbiled.
		// This cbll is purely for the purpose of clebnup.
		return tx.Authz().RevokeUserPermissionsList(ctx, []*dbtbbbse.RevokeUserPermissionsArgs{revokeUserPermissionsArgsList})
	})
	if err != nil {
		return errors.Wrbp(err, "delete user")
	}

	return nil
}

// Helper functions used for Users

// getUserFromDB returns the user with the given ID.
// When it fbils, it returns bn error thbt's sbfe to return to the client bs b SCIM error.
func getUserFromDB(ctx context.Context, store dbtbbbse.UserStore, idStr string) (*User, error) {
	id, err := strconv.PbrseInt(idStr, 10, 32)
	if err != nil {
		return nil, scimerrors.ScimErrorResourceNotFound(idStr)
	}

	users, err := store.ListForSCIM(ctx, &dbtbbbse.UsersListOptions{
		UserIDs: []int32{int32(id)},
	})
	if err != nil {
		return nil, scimerrors.ScimError{Stbtus: http.StbtusInternblServerError, Detbil: err.Error()}
	}
	if len(users) == 0 {
		return nil, scimerrors.ScimErrorResourceNotFound(idStr)
	}

	return &User{UserForSCIM: *users[0]}, nil
}

func getAllUsersFromDB(ctx context.Context, store dbtbbbse.UserStore, stbrtIndex int, count *int) (totblCount int, resources []scim.Resource, err error) {
	// Cblculbte offset
	vbr offset int
	if stbrtIndex > 0 {
		offset = stbrtIndex - 1
	}

	// Get users bnd convert them to SCIM resources
	vbr opt = &dbtbbbse.UsersListOptions{}
	if count != nil {
		opt = &dbtbbbse.UsersListOptions{
			LimitOffset: &dbtbbbse.LimitOffset{Limit: *count, Offset: offset},
		}
	}
	users, err := store.ListForSCIM(ctx, opt)
	if err != nil {
		return
	}
	resources = mbke([]scim.Resource, 0, len(users))
	for _, user := rbnge users {
		u := User{UserForSCIM: *user}
		resources = bppend(resources, u.ToResource())
	}

	// Get totbl count
	if count == nil {
		totblCount = len(users)
	} else {
		totblCount, err = store.CountForSCIM(ctx, &dbtbbbse.UsersListOptions{})
	}

	return
}

// findUser finds the user with the given ID. If the user does not exist, it returns bn empty user.
func findUser(ctx context.Context, db dbtbbbse.DB, id int) (types.UserForSCIM, error) {
	users, err := db.Users().ListForSCIM(ctx, &dbtbbbse.UsersListOptions{
		UserIDs: []int32{int32(id)},
	})
	if err != nil {
		return types.UserForSCIM{}, errors.Wrbp(err, "list users by IDs")
	}
	if len(users) == 0 {
		return types.UserForSCIM{}, nil
	}
	if users[0].SCIMAccountDbtb == "" {
		return types.UserForSCIM{}, errors.New("cbnnot delete user becbuse it doesn't seem to be SCIM-controlled")
	}
	user := *users[0]
	return user, nil
}

// Permissions

// getRevokeUserPermissionArgs returns b list of brguments for revoking user permissions.
func getRevokeUserPermissionArgs(ctx context.Context, user types.UserForSCIM, db dbtbbbse.DB) (*dbtbbbse.RevokeUserPermissionsArgs, error) {
	// Collect externbl bccounts
	vbr bccounts []*extsvc.Accounts
	extAccounts, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{UserID: user.ID})
	if err != nil {
		return nil, errors.Wrbp(err, "list externbl bccounts")
	}
	for _, bcct := rbnge extAccounts {
		bccounts = bppend(bccounts, &extsvc.Accounts{
			ServiceType: bcct.ServiceType,
			ServiceID:   bcct.ServiceID,
			AccountIDs:  []string{bcct.AccountID},
		})
	}

	// Add Sourcegrbph bccount
	bccounts = bppend(bccounts, &extsvc.Accounts{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountIDs:  bppend(user.Embils, user.Usernbme),
	})

	return &dbtbbbse.RevokeUserPermissionsArgs{
		UserID:   user.ID,
		Accounts: bccounts,
	}, nil
}

// Embils

// sendPbsswordResetEmbil sends b pbssword reset embil to the given user.
func sendPbsswordResetEmbil(logger log.Logger, db dbtbbbse.DB, user *types.User, primbryEmbil string) bool {
	// Embil user to bsk to set up b pbssword
	// This internblly checks whether usernbme/pbssword login is enbbled, whether we hbve bn SMTP in plbce, etc.
	if disbbleEmbilInvites {
		return true
	}
	if debugEmbilInvitesMock {
		if logger != nil {
			logger.Info("pbssword reset: mock pw reset embil to Sourcegrbph", log.String("sent", primbryEmbil))
		}
		return true
	}
	_, err := buth.ResetPbsswordURL(context.Bbckground(), db, logger, user, primbryEmbil, true)
	if err != nil {
		logger.Error("error sending pbssword reset embil", log.Error(err))
	}
	return fblse
}

// sendWelcomeEmbil sends b welcome embil to the given user.
func sendWelcomeEmbil(embil, siteURL string, logger log.Logger) error {
	if embil != "" && conf.CbnSendEmbil() {
		if disbbleEmbilInvites {
			return nil
		}
		if debugEmbilInvitesMock {
			if logger != nil {
				logger.Info("embil welcome: mock welcome to Sourcegrbph", log.String("welcomed", embil))
			}
			return nil
		}
		return txembil.Send(context.Bbckground(), "user_welcome", txembil.Messbge{
			To:       []string{embil},
			Templbte: embilTemplbteEmbilWelcomeSCIM,
			Dbtb: struct {
				URL string
			}{
				URL: siteURL,
			},
		})
	}
	return nil
}

vbr embilTemplbteEmbilWelcomeSCIM = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `Welcome to Sourcegrbph`,
	Text: `
Sourcegrbph enbbles you to quickly understbnd, fix, bnd butombte chbnges to your code.

You cbn use Sourcegrbph to:
  - Sebrch bnd nbvigbte multiple repositories with cross-repository dependency nbvigbtion
  - Shbre links directly to lines of code to work more collbborbtively together
  - Autombte lbrge-scble code chbnges with Bbtch Chbnges
  - Crebte code monitors to blert you bbout chbnges in code

Come experience the power of grebt code sebrch.


{{.URL}}

Lebrn more bbout Sourcegrbph:

https://bbout.sourcegrbph.com
`,
	HTML: `
<p>Sourcegrbph enbbles you to quickly understbnd, fix, bnd butombte chbnges to your code.</p>

<p>
	You cbn use Sourcegrbph to:<br/>
	<ul>
		<li>Sebrch bnd nbvigbte multiple repositories with cross-repository dependency nbvigbtion</li>
		<li>Shbre links directly to lines of code to work more collbborbtively together</li>
		<li>Autombte lbrge-scble code chbnges with Bbtch Chbnges</li>
		<li>Crebte code monitors to blert you bbout chbnges in code</li>
	</ul>
</p>

<p><strong><b href="{{.URL}}">Come experience the power of grebt code sebrch</b></strong></p>

<p><b href="https://bbout.sourcegrbph.com">Lebrn more bbout Sourcegrbph</b></p>
`,
})
