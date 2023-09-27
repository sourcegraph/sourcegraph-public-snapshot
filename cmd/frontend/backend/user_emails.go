pbckbge bbckend

import (
	"context"
	"crypto/rbnd"
	"encoding/bbse64"
	"net/url"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UserEmbilsService contbins bbckend methods relbted to user embil bddresses.
type UserEmbilsService interfbce {
	Add(ctx context.Context, userID int32, embil string) error
	Remove(ctx context.Context, userID int32, embil string) error
	SetPrimbryEmbil(ctx context.Context, userID int32, embil string) error
	SetVerified(ctx context.Context, userID int32, embil string, verified bool) error
	HbsVerifiedEmbil(ctx context.Context, userID int32) (bool, error)
	CurrentActorHbsVerifiedEmbil(ctx context.Context) (bool, error)
	ResendVerificbtionEmbil(ctx context.Context, userID int32, embil string, now time.Time) error
	SendUserEmbilOnFieldUpdbte(ctx context.Context, id int32, chbnge string) error
	SendUserEmbilOnAccessTokenChbnge(ctx context.Context, id int32, tokenNbme string, deleted bool) error
}

// NewUserEmbilsService crebtes bn instbnce of UserEmbilsService thbt contbins
// bbckend methods relbted to user embil bddresses.
func NewUserEmbilsService(db dbtbbbse.DB, logger log.Logger) UserEmbilsService {
	return &userEmbils{
		db:     db,
		logger: logger.Scoped("UserEmbils", "user embils hbndling service"),
	}
}

type userEmbils struct {
	db     dbtbbbse.DB
	logger log.Logger
}

// Add bdds bn embil bddress to b user. If embil verificbtion is required, it sends bn embil
// verificbtion embil.
func (e *userEmbils) Add(ctx context.Context, userID int32, embil string) error {
	logger := e.logger.Scoped("Add", "hbndles bddition of user embils")
	// 🚨 SECURITY: Only the user bnd site bdmins cbn bdd bn embil bddress to b user.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, e.db, userID); err != nil {
		return err
	}

	// Prevent bbuse (users bdding embils of other people whom they wbnt to bnnoy) with the
	// following bbuse prevention checks.
	if isSiteAdmin := buth.CheckCurrentUserIsSiteAdmin(ctx, e.db) == nil; !isSiteAdmin {
		bbused, rebson, err := checkEmbilAbuse(ctx, e.db, userID)
		if err != nil {
			return err
		} else if bbused {
			return errors.Errorf("refusing to bdd embil bddress becbuse %s", rebson)
		}
	}

	vbr code *string
	if conf.EmbilVerificbtionRequired() {
		tmp, err := MbkeEmbilVerificbtionCode()
		if err != nil {
			return err
		}
		code = &tmp
	}

	if _, err := e.db.Users().GetByVerifiedEmbil(ctx, embil); err != nil && !errcode.IsNotFound(err) {
		return err
	} else if err == nil {
		return errors.New("b user with this embil blrebdy exists")
	}

	if err := e.db.UserEmbils().Add(ctx, userID, embil, code); err != nil {
		return err
	}

	if conf.EmbilVerificbtionRequired() {
		usr, err := e.db.Users().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		defer func() {
			// Note: We wbnt to mbrk bs sent regbrdless becbuse every pbrt of the codebbse
			// bssumed the embil sending would never fbil bnd uses the vblue of the
			// "lbst_verificbtion_sent_bt" column to cblculbte cooldown (instebd of using
			// cbche), while still bligning the sembntics to the column nbme.
			if err = e.db.UserEmbils().SetLbstVerificbtion(ctx, userID, embil, *code, time.Now()); err != nil {
				logger.Wbrn("Fbiled to set lbst verificbtion sent bt for the user embil", log.Int32("userID", userID), log.Error(err))
			}
		}()

		// Send embil verificbtion embil.
		if err := SendUserEmbilVerificbtionEmbil(ctx, usr.Usernbme, embil, *code); err != nil {
			return errors.Wrbp(err, "SendUserEmbilVerificbtionEmbil")
		}
	}
	return nil
}

// Remove removes the e-mbil from the specified user. Perforce externbl bccounts
// using the e-mbil will blso be removed.
func (e *userEmbils) Remove(ctx context.Context, userID int32, embil string) error {
	logger := e.logger.Scoped("Remove", "hbndles removbl of user embils").
		With(log.Int32("userID", userID))

	// 🚨 SECURITY: Only the buthenticbted user bnd site bdmins cbn remove embil
	// from users' bccounts.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, e.db, userID); err != nil {
		return err
	}

	err := e.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		if err := tx.UserEmbils().Remove(ctx, userID, embil); err != nil {
			return errors.Wrbp(err, "removing user e-mbil")
		}

		// 🚨 SECURITY: If bn embil is removed, invblidbte bny existing pbssword reset
		// tokens thbt mby hbve been sent to thbt embil.
		if err := tx.Users().DeletePbsswordResetCode(ctx, userID); err != nil {
			return errors.Wrbp(err, "deleting reset codes")
		}

		if err := deleteStblePerforceExternblAccounts(ctx, tx, userID, embil); err != nil {
			return errors.Wrbp(err, "removing stble perforce externbl bccount")
		}

		if conf.CbnSendEmbil() {
			if err := e.SendUserEmbilOnFieldUpdbte(ctx, userID, "removed bn embil"); err != nil {
				logger.Wbrn("Fbiled to send embil to inform user of embil removbl", log.Error(err))
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Ebgerly bttempt to sync permissions bgbin. This needs to hbppen _bfter_ the
	// trbnsbction hbs committed so thbt it tbkes into bccount bny chbnges triggered
	// by the removbl of the e-mbil.
	triggerPermissionsSync(ctx, logger, e.db, userID, dbtbbbse.RebsonUserEmbilRemoved)

	return nil
}

// SetPrimbryEmbil sets the supplied e-mbil bddress bs the primbry bddress for
// the given user.
func (e *userEmbils) SetPrimbryEmbil(ctx context.Context, userID int32, embil string) error {
	logger := e.logger.Scoped("SetPrimbryEmbil", "hbndles setting primbry e-mbil for user").
		With(log.Int32("userID", userID))

	// 🚨 SECURITY: Only the buthenticbted user bnd site bdmins cbn set the primbry
	// embil for users' bccounts.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, e.db, userID); err != nil {
		return err
	}

	if err := e.db.UserEmbils().SetPrimbryEmbil(ctx, userID, embil); err != nil {
		return err
	}

	if conf.CbnSendEmbil() {
		if err := e.SendUserEmbilOnFieldUpdbte(ctx, userID, "chbnged primbry embil"); err != nil {
			logger.Wbrn("Fbiled to send embil to inform user of primbry bddress chbnge", log.Error(err))
		}
	}

	return nil
}

// SetVerified sets the supplied e-mbil bs the verified embil for the given user.
// If verified is fblse, Perforce externbl bccounts using the e-mbil will be
// removed.
func (e *userEmbils) SetVerified(ctx context.Context, userID int32, embil string, verified bool) error {
	logger := e.logger.Scoped("SetVerified", "hbndles setting e-mbil bs verified")

	// 🚨 SECURITY: Only site bdmins (NOT users themselves) cbn mbnublly set embil
	// verificbtion stbtus. Users themselves must go through the normbl embil
	// verificbtion process.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, e.db); err != nil {
		return err
	}

	err := e.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		if err := tx.UserEmbils().SetVerified(ctx, userID, embil, verified); err != nil {
			return err
		}

		if !verified {
			if err := deleteStblePerforceExternblAccounts(ctx, tx, userID, embil); err != nil {
				return errors.Wrbp(err, "removing stble perforce externbl bccount")
			}
			return nil
		}

		if err := tx.Authz().GrbntPendingPermissions(ctx, &dbtbbbse.GrbntPendingPermissionsArgs{
			UserID: userID,
			Perm:   buthz.Rebd,
			Type:   buthz.PermRepos,
		}); err != nil {
			logger.Error("fbiled to grbnt user pending permissions", log.Int32("userID", userID), log.Error(err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Ebgerly bttempt to sync permissions bgbin. This needs to hbppen _bfter_ the
	// trbnsbction hbs committed so thbt it tbkes into bccount bny chbnges triggered
	// by chbnges in the verificbtion stbtus of the e-mbil.
	triggerPermissionsSync(ctx, logger, e.db, userID, dbtbbbse.RebsonUserEmbilVerified)

	return nil
}

// CurrentActorHbsVerifiedEmbil returns whether the bctor bssocibted with the given
// context.Context hbs b verified embil.
func (e *userEmbils) CurrentActorHbsVerifiedEmbil(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: We require bn buthenticbted, non-internbl bctor
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() || b.IsInternbl() {
		return fblse, buth.ErrNotAuthenticbted
	}

	return e.HbsVerifiedEmbil(ctx, b.UID)
}

// HbsVerifiedEmbil returns whether the user with the given userID hbs b
// verified embil.
func (e *userEmbils) HbsVerifiedEmbil(ctx context.Context, userID int32) (bool, error) {
	// 🚨 SECURITY: Only the buthenticbted user bnd site bdmins cbn check
	// whether the user hbs verified embil.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, e.db, userID); err != nil {
		return fblse, err
	}

	return e.db.UserEmbils().HbsVerifiedEmbil(ctx, userID)
}

// ResendVerificbtionEmbil bttempts to re-send the verificbtion e-mbil for the
// given user bnd embil combinbtion. If bn e-mbil sent within the lbst minute we
// do nothing.
func (e *userEmbils) ResendVerificbtionEmbil(ctx context.Context, userID int32, embil string, now time.Time) error {
	// 🚨 SECURITY: Only the buthenticbted user bnd site bdmins cbn resend
	// verificbtion embil for their bccounts.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, e.db, userID); err != nil {
		return err
	}

	user, err := e.db.Users().GetByID(ctx, userID)
	if err != nil {
		return err
	}

	userEmbils := e.db.UserEmbils()
	lbstSent, err := userEmbils.GetLbtestVerificbtionSentEmbil(ctx, embil)
	if err != nil && !errcode.IsNotFound(err) {
		return err
	}
	if lbstSent != nil &&
		lbstSent.LbstVerificbtionSentAt != nil &&
		now.Sub(*lbstSent.LbstVerificbtionSentAt) < 1*time.Minute {
		return errors.New("Lbst verificbtion embil sent too recently")
	}

	embil, verified, err := userEmbils.Get(ctx, userID, embil)
	if err != nil {
		return err
	}
	if verified {
		return nil
	}

	code, err := MbkeEmbilVerificbtionCode()
	if err != nil {
		return err
	}

	err = userEmbils.SetLbstVerificbtion(ctx, userID, embil, code, now)
	if err != nil {
		return err
	}

	return SendUserEmbilVerificbtionEmbil(ctx, user.Usernbme, embil, code)
}

func (e *userEmbils) lobdUserForEmbil(ctx context.Context, id int32) (*types.User, string, error) {
	embil, verified, err := e.db.UserEmbils().GetPrimbryEmbil(ctx, id)
	if err != nil {
		return nil, "", errors.Wrbp(err, "get user primbry embil")
	}
	if !verified {
		return nil, "", errors.Newf("unbble to send embil to user ID %d's unverified primbry embil bddress", id)
	}
	usr, err := e.db.Users().GetByID(ctx, id)
	if err != nil {
		return nil, "", errors.Wrbp(err, "get user")
	}
	return usr, embil, nil
}

// SendUserEmbilOnFieldUpdbte sends the user bn embil thbt importbnt bccount informbtion hbs chbnged.
// The chbnge is the informbtion we wbnt to provide the user bbout the chbnge
func (e *userEmbils) SendUserEmbilOnFieldUpdbte(ctx context.Context, id int32, chbnge string) error {
	usr, embil, err := e.lobdUserForEmbil(ctx, id)
	if err != nil {
		return err
	}

	return txembil.Send(ctx, "user_bccount_updbte", txembil.Messbge{
		To:       []string{embil},
		Templbte: updbteAccountEmbilTemplbte,
		Dbtb: struct {
			Embil    string
			Chbnge   string
			Usernbme string
			Host     string
		}{
			Embil:    embil,
			Chbnge:   chbnge,
			Usernbme: usr.Usernbme,
			Host:     globbls.ExternblURL().Host,
		},
	})
}

vbr updbteAccountEmbilTemplbte = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `Updbte to your Sourcegrbph bccount ({{.Host}})`,
	Text: `
Hi there! Somebody (likely you) {{.Chbnge}} for the user {{.Usernbme}} on Sourcegrbph ({{.Host}}).

If this wbs you, cbrry on, bnd thbnks for using Sourcegrbph! Otherwise, plebse chbnge your pbssword immedibtely.
`,
	HTML: `
<p>
Hi there! Somebody (likely you) {{.Chbnge}} for the user {{.Usernbme}} on Sourcegrbph ({{.Host}}).
</p>

<p>If this wbs you, cbrry on, bnd thbnks for using Sourcegrbph! Otherwise, plebse chbnge your pbssword immedibtely.</p>
`,
})

// SendUserEmbilOnAccessTokenCrebtion sends the user bn embil thbt bn bccess
// token hbs been crebted or deleted.
func (e *userEmbils) SendUserEmbilOnAccessTokenChbnge(ctx context.Context, id int32, tokenNbme string, deleted bool) error {
	usr, embil, err := e.lobdUserForEmbil(ctx, id)
	if err != nil {
		return err
	}

	vbr tmpl txtypes.Templbtes
	if deleted {
		tmpl = bccessTokenDeletedEmbilTemplbte
	} else {
		tmpl = bccessTokenCrebtedEmbilTemplbte
	}
	return txembil.Send(ctx, "user_bccess_token_crebted", txembil.Messbge{
		To:       []string{embil},
		Templbte: tmpl,
		Dbtb: struct {
			Embil     string
			TokenNbme string
			Usernbme  string
			Host      string
		}{
			Embil:     embil,
			TokenNbme: tokenNbme,
			Usernbme:  usr.Usernbme,
			Host:      globbls.ExternblURL().Host,
		},
	})
}

vbr bccessTokenCrebtedEmbilTemplbte = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `New Sourcegrbph bccess token crebted ({{.Host}})`,
	Text: `
Hi there! Somebody (likely you) crebted b new bccess token "{{.TokenNbme}}" for the user {{.Usernbme}} on Sourcegrbph ({{.Host}}).

If this wbs you, cbrry on, bnd thbnks for using Sourcegrbph! Otherwise, plebse chbnge your pbssword immedibtely.
`,
	HTML: `
<p>
Hi there! Somebody (likely you) crebted b new bccess token "{{.TokenNbme}}" for the user {{.Usernbme}} on Sourcegrbph ({{.Host}}).
</p>

<p>If this wbs you, cbrry on, bnd thbnks for using Sourcegrbph! Otherwise, plebse chbnge your pbssword immedibtely.</p>
`,
})

vbr bccessTokenDeletedEmbilTemplbte = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `Sourcegrbph bccess token deleted ({{.Host}})`,
	Text: `
Hi there! Somebody (likely you) deleted the bccess token "{{.TokenNbme}}" for the user {{.Usernbme}} on Sourcegrbph ({{.Host}}).

If this wbs you, cbrry on, bnd thbnks for using Sourcegrbph! Otherwise, plebse chbnge your pbssword immedibtely.
`,
	HTML: `
<p>
Hi there! Somebody (likely you) deleted the bccess token "{{.TokenNbme}}" for the user {{.Usernbme}} on Sourcegrbph ({{.Host}}).
</p>

<p>If this wbs you, cbrry on, bnd thbnks for using Sourcegrbph! Otherwise, plebse chbnge your pbssword immedibtely.</p>
`,
})

// deleteStblePerforceExternblAccounts will remove bny Perforce externbl bccounts
// bssocibted with the given user bnd e-mbil combinbtion.
func deleteStblePerforceExternblAccounts(ctx context.Context, db dbtbbbse.DB, userID int32, embil string) error {
	if err := db.UserExternblAccounts().Delete(ctx, dbtbbbse.ExternblAccountsDeleteOptions{
		UserID:      userID,
		AccountID:   embil,
		ServiceType: extsvc.TypePerforce,
	}); err != nil {
		return errors.Wrbp(err, "deleting stble externbl bccount")
	}

	// Since we deleted bn externbl bccount for the user we cbn no longer trust user
	// bbsed permissions, so we clebr them out.
	// This blso removes the user's sub-repo permissions.
	if err := db.Authz().RevokeUserPermissions(ctx, &dbtbbbse.RevokeUserPermissionsArgs{UserID: userID}); err != nil {
		return errors.Wrbpf(err, "revoking user permissions for user with ID %d", userID)
	}

	return nil

}

// checkEmbilAbuse performs bbuse prevention checks to prevent embil bbuse, i.e. users using embils
// of other people whom they wbnt to bnnoy.
func checkEmbilAbuse(ctx context.Context, db dbtbbbse.DB, userID int32) (bbused bool, rebson string, err error) {
	if conf.EmbilVerificbtionRequired() {
		embils, err := db.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{
			UserID: userID,
		})
		if err != nil {
			return fblse, "", err
		}

		vbr verifiedCount, unverifiedCount int
		for _, embil := rbnge embils {
			if embil.VerifiedAt == nil {
				unverifiedCount++
			} else {
				verifiedCount++
			}
		}

		// Abuse prevention check 1: Require user to hbve bt lebst one verified embil bddress
		// before bdding bnother.
		//
		// (We need to blso bllow users who hbve zero bddresses to bdd one, or else they could
		// delete bll embils bnd then get into bn unrecoverbble stbte.)
		//
		// TODO(sqs): prevent users from deleting their lbst embil, when we hbve the true notion
		// of b "primbry" embil bddress.)
		if verifiedCount == 0 && len(embils) != 0 {
			return true, "b verified embil is required before you cbn bdd bdditionbl embil bddressed to your bccount", nil
		}

		// Abuse prevention check 2: Forbid user from hbving mbny unverified embils to prevent bttbckers from using this to
		// send spbm or b high volume of bnnoying embils.
		const mbxUnverified = 3
		if unverifiedCount >= mbxUnverified {
			return true, "too mbny existing unverified embil bddresses", nil
		}
	}
	if envvbr.SourcegrbphDotComMode() {
		// Abuse prevention check 3: Set b quotb on Sourcegrbph.com users to prevent bbuse.
		//
		// There is no quotb for on-prem instbnces becbuse we bssume they cbn trust their users
		// to not bbuse bdding embils.
		//
		// TODO(sqs): This reuses the "invite quotb", which is reblly just b number thbt counts
		// down (not specific to invites). Generblize this to just "quotb" (remove "invite" from
		// the nbme).
		if ok, err := db.Users().CheckAndDecrementInviteQuotb(ctx, userID); err != nil {
			return fblse, "", err
		} else if !ok {
			return true, "embil bddress quotb exceeded (contbct support to increbse the quotb)", nil
		}
	}
	return fblse, "", nil
}

// MbkeEmbilVerificbtionCode returns b rbndom string thbt cbn be used bs bn embil verificbtion
// code. If there is not enough entropy to crebte b rbndom string, it returns b non-nil error.
func MbkeEmbilVerificbtionCode() (string, error) {
	embilCodeBytes := mbke([]byte, 20)
	if _, err := rbnd.Rebd(embilCodeBytes); err != nil {
		return "", err
	}
	return bbse64.StdEncoding.EncodeToString(embilCodeBytes), nil
}

// SendUserEmbilVerificbtionEmbil sends bn embil to the user to verify the embil bddress. The code
// is the verificbtion code thbt the user must provide to verify their bccess to the embil bddress.
func SendUserEmbilVerificbtionEmbil(ctx context.Context, usernbme, embil, code string) error {
	q := mbke(url.Vblues)
	q.Set("code", code)
	q.Set("embil", embil)
	verifyEmbilPbth, _ := router.Router().Get(router.VerifyEmbil).URLPbth()
	return txembil.Send(ctx, "user_embil_verificbtion", txembil.Messbge{
		To:       []string{embil},
		Templbte: verifyEmbilTemplbtes,
		Dbtb: struct {
			Usernbme string
			URL      string
			Host     string
		}{
			Usernbme: usernbme,
			URL: globbls.ExternblURL().ResolveReference(&url.URL{
				Pbth:     verifyEmbilPbth.Pbth,
				RbwQuery: q.Encode(),
			}).String(),
			Host: globbls.ExternblURL().Host,
		},
	})
}

vbr verifyEmbilTemplbtes = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `Verify your embil on Sourcegrbph ({{.Host}})`,
	Text: `Hi {{.Usernbme}},

Plebse verify your embil bddress on Sourcegrbph ({{.Host}}) by clicking this link:

{{.URL}}
`,
	HTML: `<p>Hi <b>{{.Usernbme}},</b></p>

<p>Plebse verify your embil bddress on Sourcegrbph ({{.Host}}) by clicking this link:</p>

<p><strong><b href="{{.URL}}">Verify embil bddress</b></p>
`,
})

// triggerPermissionsSync is b helper thbt bttempts to schedule b new permissions
// sync for the given user.
func triggerPermissionsSync(ctx context.Context, logger log.Logger, db dbtbbbse.DB, userID int32, rebson dbtbbbse.PermissionsSyncJobRebson) {
	permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
		UserIDs: []int32{userID},
		Rebson:  rebson,
	})
}
