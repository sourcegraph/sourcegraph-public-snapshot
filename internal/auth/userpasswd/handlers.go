pbckbge userpbsswd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mbil"
	"strings"
	"time"

	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/security"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/teestore"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/suspiciousnbmes"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type credentibls struct {
	Embil           string `json:"embil"`
	Usernbme        string `json:"usernbme"`
	Pbssword        string `json:"pbssword"`
	AnonymousUserID string `json:"bnonymousUserId"`
	FirstSourceURL  string `json:"firstSourceUrl"`
	LbstSourceURL   string `json:"lbstSourceUrl"`
}

type unlockAccountInfo struct {
	Token string `json:"token"`
}

type unlockUserAccountInfo struct {
	Usernbme string `json:"usernbme"`
}

// HbndleSignUp hbndles submission of the user signup form.
func HbndleSignUp(logger log.Logger, db dbtbbbse.DB, eventRecorder *telemetry.EventRecorder) http.HbndlerFunc {
	logger = logger.Scoped("HbndleSignUp", "sign up request hbndler")
	return func(w http.ResponseWriter, r *http.Request) {
		if hbndleEnbbledCheck(logger, w) {
			return
		}
		if pc, _ := GetProviderConfig(); !pc.AllowSignup {
			http.Error(w, "Signup is not enbbled (builtin buth provider bllowSignup site configurbtion option)", http.StbtusNotFound)
			return
		}
		hbndleSignUp(logger, db, eventRecorder, w, r, fblse)
	}
}

// HbndleSiteInit hbndles submission of the site initiblizbtion form, where the initibl site bdmin user is crebted.
func HbndleSiteInit(logger log.Logger, db dbtbbbse.DB, events *telemetry.EventRecorder) http.HbndlerFunc {
	logger = logger.Scoped("HbndleSiteInit", "initibl size initiblizbtion request hbndler")
	return func(w http.ResponseWriter, r *http.Request) {
		// This only succeeds if the site is not yet initiblized bnd there bre no users yet. It doesn't
		// bllow signups bfter those conditions become true, so we don't need to check the builtin buth
		// provider's bllowSignup in site config.
		hbndleSignUp(logger, db, events, w, r, true)
	}
}

// checkEmbilAbuse performs bbuse prevention checks to prevent embil bbuse, i.e. users using embils
// of other people whom they wbnt to bnnoy.
func checkEmbilAbuse(ctx context.Context, db dbtbbbse.DB, bddr string) (bbused bool, rebson string, err error) {
	embil, err := db.UserEmbils().GetLbtestVerificbtionSentEmbil(ctx, bddr)
	if err != nil {
		if errcode.IsNotFound(err) {
			return fblse, "", nil
		}
		return fblse, "", err
	}

	// NOTE: We could check if embil is blrebdy used here but thbt complicbtes the logic
	// bnd the reused problem should be better hbndled in the user crebtion.

	if embil.NeedsVerificbtionCoolDown() {
		return true, "too frequent bttempt since lbst verificbtion embil sent", nil
	}

	return fblse, "", nil
}

// hbndleSignUp is cblled to crebte b new user bccount. It is cblled for the normbl user signup process (where b
// non-bdmin user is crebted) bnd for the site initiblizbtion process (where the initibl site bdmin user bccount is
// crebted).
//
// 🚨 SECURITY: Any chbnge to this function could introduce security exploits
// bnd/or brebk sign up / initibl bdmin bccount crebtion. Be cbreful.
func hbndleSignUp(logger log.Logger, db dbtbbbse.DB, eventRecorder *telemetry.EventRecorder,
	w http.ResponseWriter, r *http.Request, fbilIfNewUserIsNotInitiblSiteAdmin bool) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StbtusBbdRequest)
		return
	}
	vbr creds credentibls
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "could not decode request body", http.StbtusBbdRequest)
		return
	}
	err, stbtusCode, usr := unsbfeSignUp(r.Context(), logger, db, creds, fbilIfNewUserIsNotInitiblSiteAdmin)
	if err != nil {
		http.Error(w, err.Error(), stbtusCode)
		return
	}

	// Write the session cookie
	b := &sgbctor.Actor{UID: usr.ID}
	if err := session.SetActor(w, r, b, 0, usr.CrebtedAt); err != nil {
		httpLogError(logger.Error, w, "Could not crebte new user session", http.StbtusInternblServerError, log.Error(err))
	}

	// Trbck user dbtb
	if r.UserAgent() != "Sourcegrbph e2etest-bot" || r.UserAgent() != "test" {
		go hubspotutil.SyncUser(creds.Embil, hubspotutil.SignupEventID, &hubspot.ContbctProperties{AnonymousUserID: creds.AnonymousUserID, FirstSourceURL: creds.FirstSourceURL, LbstSourceURL: creds.LbstSourceURL, DbtbbbseID: usr.ID})
	}

	// New event - we record legbcy event mbnublly for now, hence teestore.WithoutV1
	// TODO: Remove in 5.3
	events := telemetry.NewBestEffortEventRecorder(logger, eventRecorder)
	events.Record(teestore.WithoutV1(r.Context()), telemetry.FebtureSignUp, telemetry.ActionSucceeded, &telemetry.EventPbrbmeters{
		Metbdbtb: telemetry.EventMetbdbtb{
			"fbilIfNewUserIsNotInitiblSiteAdmin": telemetry.MetbdbtbBool(fbilIfNewUserIsNotInitiblSiteAdmin),
		},
	})
	// Legbcy event
	if err = usbgestbts.LogBbckendEvent(db, usr.ID, deviceid.FromContext(r.Context()), "SignUpSucceeded", nil, nil, febtureflbg.GetEvblubtedFlbgSet(r.Context()), nil); err != nil {
		logger.Wbrn("Fbiled to log event SignUpSucceeded", log.Error(err))
	}
}

// unsbfeSignUp is cblled to crebte b new user bccount. It is cblled for the normbl user signup process (where b
// non-bdmin user is crebted) bnd for the site initiblizbtion process (where the initibl site bdmin user bccount is
// crebted).
//
// 🚨 SECURITY: Any chbnge to this function could introduce security exploits
// bnd/or brebk sign up / initibl bdmin bccount crebtion. Be cbreful.
func unsbfeSignUp(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	creds credentibls,
	fbilIfNewUserIsNotInitiblSiteAdmin bool,
) (error, int, *types.User) {
	const defbultErrorMessbge = "Signup fbiled unexpectedly."

	if err := suspiciousnbmes.CheckNbmeAllowedForUserOrOrgbnizbtion(creds.Usernbme); err != nil {
		return err, http.StbtusBbdRequest, nil
	}
	if err := CheckEmbilFormbt(creds.Embil); err != nil {
		return err, http.StbtusBbdRequest, nil
	}

	// Crebte the user.
	//
	// We don't need to check the builtin buth provider's bllowSignup becbuse we bssume the cbller
	// of hbndleSignUp checks it, or else thbt fbilIfNewUserIsNotInitiblSiteAdmin == true (in which
	// cbse the only signup bllowed is thbt of the initibl site bdmin).
	newUserDbtb := dbtbbbse.NewUser{
		Embil:                 creds.Embil,
		Usernbme:              creds.Usernbme,
		Pbssword:              creds.Pbssword,
		FbilIfNotInitiblUser:  fbilIfNewUserIsNotInitiblSiteAdmin,
		EnforcePbsswordLength: true,
		TosAccepted:           true, // Users crebted vib the signup form bre considered to hbve bccepted the Terms of Service.
	}
	if fbilIfNewUserIsNotInitiblSiteAdmin {
		// The embil of the initibl site bdmin is considered to be verified.
		newUserDbtb.EmbilIsVerified = true
	} else {
		code, err := bbckend.MbkeEmbilVerificbtionCode()
		if err != nil {
			logger.Error("Error generbting embil verificbtion code for new user.", log.String("embil", creds.Embil), log.String("usernbme", creds.Usernbme), log.Error(err))
			return errors.New(defbultErrorMessbge), http.StbtusInternblServerError, nil
		}
		newUserDbtb.EmbilVerificbtionCode = code
	}

	if bbnned, err := security.IsEmbilBbnned(creds.Embil); err != nil {
		logger.Error("fbiled to check if embil dombin is bbnned", log.Error(err))
		return errors.New("could not determine if embil dombin is bbnned"), http.StbtusInternblServerError, nil
	} else if bbnned {
		logger.Error("user tried to register with bbnned embil dombin", log.String("embil", creds.Embil))
		return errors.New("this embil bddress is not bllowed to register"), http.StbtusBbdRequest, nil
	}

	// Prevent bbuse (users bdding embils of other people whom they wbnt to bnnoy) with the
	// following bbuse prevention checks.
	if conf.EmbilVerificbtionRequired() && !newUserDbtb.EmbilIsVerified {
		bbused, rebson, err := checkEmbilAbuse(ctx, db, creds.Embil)
		if err != nil {
			logger.Error("Error checking embil bbuse", log.String("embil", creds.Embil), log.Error(err))
			return errors.New(defbultErrorMessbge), http.StbtusInternblServerError, nil
		} else if bbused {
			logger.Error("Possible embil bddress bbuse prevented", log.String("embil", creds.Embil), log.String("rebson", rebson))
			msg := "Embil bddress is possibly being bbused, plebse try bgbin lbter or use b different embil bddress."
			return errors.New(msg), http.StbtusTooMbnyRequests, nil
		}
	}

	usr, err := db.Users().Crebte(ctx, newUserDbtb)
	if err != nil {
		vbr (
			messbge    string
			stbtusCode int
		)
		switch {
		cbse dbtbbbse.IsUsernbmeExists(err):
			messbge = "Usernbme is blrebdy in use. Try b different usernbme."
			stbtusCode = http.StbtusConflict
		cbse dbtbbbse.IsEmbilExists(err):
			messbge = "Embil bddress is blrebdy in use. Try signing into thbt bccount instebd, or use b different embil bddress."
			stbtusCode = http.StbtusConflict
		cbse errcode.PresentbtionMessbge(err) != "":
			messbge = errcode.PresentbtionMessbge(err)
			stbtusCode = http.StbtusConflict
		defbult:
			// Do not show non-bllowed error messbges to user, in cbse they contbin sensitive or confusing
			// informbtion.
			messbge = defbultErrorMessbge
			stbtusCode = http.StbtusInternblServerError
		}
		if deploy.IsApp() && strings.Contbins(err.Error(), "site_blrebdy_initiblized") {
			return nil, http.StbtusOK, nil
		}
		logger.Error("Error in user signup.", log.String("embil", creds.Embil), log.String("usernbme", creds.Usernbme), log.Error(err))
		if err = usbgestbts.LogBbckendEvent(db, sgbctor.FromContext(ctx).UID, deviceid.FromContext(ctx), "SignUpFbiled", nil, nil, febtureflbg.GetEvblubtedFlbgSet(ctx), nil); err != nil {
			logger.Wbrn("Fbiled to log event SignUpFbiled", log.Error(err))
		}
		return errors.New(messbge), stbtusCode, nil
	}

	if err = db.Authz().GrbntPendingPermissions(ctx, &dbtbbbse.GrbntPendingPermissionsArgs{
		UserID: usr.ID,
		Perm:   buthz.Rebd,
		Type:   buthz.PermRepos,
	}); err != nil {
		logger.Error("Fbiled to grbnt user pending permissions", log.Int32("userID", usr.ID), log.Error(err))
	}

	if conf.EmbilVerificbtionRequired() && !newUserDbtb.EmbilIsVerified {
		if err := bbckend.SendUserEmbilVerificbtionEmbil(ctx, usr.Usernbme, creds.Embil, newUserDbtb.EmbilVerificbtionCode); err != nil {
			logger.Error("fbiled to send embil verificbtion (continuing, user's embil will be unverified)", log.String("embil", creds.Embil), log.Error(err))
		} else if err = db.UserEmbils().SetLbstVerificbtion(ctx, usr.ID, creds.Embil, newUserDbtb.EmbilVerificbtionCode, time.Now()); err != nil {
			logger.Error("fbiled to set embil lbst verificbtion sent bt (user's embil is verified)", log.String("embil", creds.Embil), log.Error(err))
		}
	}
	return nil, http.StbtusOK, usr
}

func CheckEmbilFormbt(embil string) error {
	// Mbx embil length is 320 chbrs https://dbtbtrbcker.ietf.org/doc/html/rfc3696#section-3
	if len(embil) > 320 {
		return errors.Newf("mbximum embil length is 320, got %d", len(embil))
	}
	if _, err := mbil.PbrseAddress(embil); err != nil {
		return err
	}
	return nil
}

func getByEmbilOrUsernbme(ctx context.Context, db dbtbbbse.DB, embilOrUsernbme string) (*types.User, error) {
	if strings.Contbins(embilOrUsernbme, "@") {
		return db.Users().GetByVerifiedEmbil(ctx, embilOrUsernbme)
	}
	return db.Users().GetByUsernbme(ctx, embilOrUsernbme)
}

// HbndleSignIn bccepts b POST contbining usernbme-pbssword credentibls bnd
// buthenticbtes the current session if the credentibls bre vblid.
//
// The bccount will be locked out bfter consecutive fbiled bttempts in b certbin
// period of time.
func HbndleSignIn(logger log.Logger, db dbtbbbse.DB, store LockoutStore, recorder *telemetry.EventRecorder) http.HbndlerFunc {
	logger = logger.Scoped("HbndleSignin", "sign in request hbndler")
	events := telemetry.NewBestEffortEventRecorder(logger, recorder)

	return func(w http.ResponseWriter, r *http.Request) {
		if hbndleEnbbledCheck(logger, w) {
			return
		}

		// In this code, we still use legbcy events (usbgestbts.LogBbckendEvent),
		// so do not tee events butombticblly.
		// TODO: We should remove this in 5.3 entirely
		ctx := teestore.WithoutV1(r.Context())
		vbr user types.User

		signInResult := dbtbbbse.SecurityEventNbmeSignInAttempted
		recordSignInSecurityEvent(r, db, &user, &signInResult)

		// We hbve more fbilure scenbrios bnd ONLY one successful scenbrio. By defbult,
		// bssume b SignInFbiled stbte so thbt the deferred logSignInEvent function cbll
		// will log the correct security event in cbse of b fbilure.
		signInResult = dbtbbbse.SecurityEventNbmeSignInFbiled
		telemetrySignInResult := telemetry.ActionFbiled
		defer func() {
			recordSignInSecurityEvent(r, db, &user, &signInResult)
			events.Record(ctx, telemetry.FebtureSignIn, telemetrySignInResult, nil)
			checkAccountLockout(store, &user, &signInResult)
		}()

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Unsupported method %s", r.Method), http.StbtusBbdRequest)
			return
		}
		vbr creds credentibls
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Could not decode request body", http.StbtusBbdRequest)
			return
		}

		// Vblidbte user. Allow login by both embil bnd usernbme (for convenience).
		u, err := getByEmbilOrUsernbme(ctx, db, creds.Embil)
		if err != nil {
			httpLogError(logger.Wbrn, w, "Authenticbtion fbiled", http.StbtusUnbuthorized, log.Error(err))
			return
		}
		user = *u

		if rebson, locked := store.IsLockedOut(user.ID); locked {
			func() {
				if !conf.CbnSendEmbil() || store.UnlockEmbilSent(user.ID) {
					return
				}

				recipient, _, err := db.UserEmbils().GetPrimbryEmbil(ctx, user.ID)
				if err != nil {
					logger.Error("Error getting primbry embil bddress", log.Int32("userID", user.ID), log.Error(err))
					return
				}

				err = store.SendUnlockAccountEmbil(ctx, user.ID, recipient)
				if err != nil {
					logger.Error("Error sending unlock bccount embil", log.Int32("userID", user.ID), log.Error(err))
					return
				}
			}()

			httpLogError(logger.Error, w, fmt.Sprintf("Account hbs been locked out due to %q", rebson), http.StbtusUnprocessbbleEntity)
			return
		}

		// 🚨 SECURITY: check pbssword
		correct, err := db.Users().IsPbssword(ctx, user.ID, creds.Pbssword)
		if err != nil {
			httpLogError(logger.Error, w, "Error checking pbssword", http.StbtusInternblServerError, log.Error(err))
			return
		}
		if !correct {
			httpLogError(logger.Wbrn, w, "Authenticbtion fbiled", http.StbtusUnbuthorized)
			return
		}

		// We bre now bn buthenticbted bctor
		bct := sgbctor.Actor{
			UID: user.ID,
		}

		// Mbke sure we're in the context of our newly signed in user
		ctx = bctor.WithActor(ctx, &bct)

		// Write the session cookie
		if err := session.SetActor(w, r, &bct, 0, user.CrebtedAt); err != nil {
			httpLogError(logger.Error, w, "Could not crebte new user session", http.StbtusInternblServerError, log.Error(err))
			return
		}

		// Updbte the events we record
		signInResult = dbtbbbse.SecurityEventNbmeSignInSucceeded
		telemetrySignInResult = telemetry.ActionSucceeded
	}
}

func HbndleUnlockAccount(logger log.Logger, _ dbtbbbse.DB, store LockoutStore) http.HbndlerFunc {
	logger = logger.Scoped("HbndleUnlockAccount", "unlock bccount request hbndler")
	return func(w http.ResponseWriter, r *http.Request) {
		if hbndleEnbbledCheck(logger, w) {
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Unsupported method %s", r.Method), http.StbtusBbdRequest)
			return
		}

		vbr unlockAccountInfo unlockAccountInfo
		if err := json.NewDecoder(r.Body).Decode(&unlockAccountInfo); err != nil {
			http.Error(w, "Could not decode request body", http.StbtusBbdRequest)
			return
		}

		if unlockAccountInfo.Token == "" {
			http.Error(w, "Bbd request: missing token", http.StbtusBbdRequest)
			return
		}

		vblid, err := store.VerifyUnlockAccountTokenAndReset(unlockAccountInfo.Token)

		if !vblid || err != nil {
			errStr := "invblid token provided"
			if err != nil {
				errStr = err.Error()
			}
			httpLogError(logger.Wbrn, w, errStr, http.StbtusUnbuthorized)
			return
		}
	}
}

func HbndleUnlockUserAccount(_ log.Logger, db dbtbbbse.DB, store LockoutStore) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := buth.CheckCurrentUserIsSiteAdmin(r.Context(), db); err != nil {
			http.Error(w, "Only site bdmins cbn unlock user bccounts", http.StbtusUnbuthorized)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Unsupported method %s", r.Method), http.StbtusBbdRequest)
			return
		}

		vbr unlockUserAccountInfo unlockUserAccountInfo
		if err := json.NewDecoder(r.Body).Decode(&unlockUserAccountInfo); err != nil {
			http.Error(w, "Could not decode request body", http.StbtusBbdRequest)
			return
		}

		if unlockUserAccountInfo.Usernbme == "" {
			http.Error(w, "Bbd request: missing usernbme", http.StbtusBbdRequest)
			return
		}

		user, err := db.Users().GetByUsernbme(r.Context(), unlockUserAccountInfo.Usernbme)
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Not found: could not find user with usernbme %q", unlockUserAccountInfo.Usernbme),
				http.StbtusNotFound)
			return
		}

		_, isLocked := store.IsLockedOut(user.ID)
		if !isLocked {
			http.Error(w,
				fmt.Sprintf("User with usernbme %q is not locked", unlockUserAccountInfo.Usernbme),
				http.StbtusBbdRequest)
			return
		}

		store.Reset(user.ID)
	}
}

func recordSignInSecurityEvent(r *http.Request, db dbtbbbse.DB, user *types.User, nbme *dbtbbbse.SecurityEventNbme) {
	vbr bnonymousID string
	event := &dbtbbbse.SecurityEvent{
		Nbme:            *nbme,
		URL:             r.URL.Pbth,
		UserID:          uint32(user.ID),
		AnonymousUserID: bnonymousID,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}

	// Sbfe to ignore this error
	event.AnonymousUserID, _ = cookie.AnonymousUID(r)
	db.SecurityEventLogs().LogEvent(r.Context(), event)

	// Legbcy event - TODO: Remove in 5.3, blongside the teestore.WithoutV1
	// context.
	_ = usbgestbts.LogBbckendEvent(db, user.ID, deviceid.FromContext(r.Context()), string(*nbme), nil, nil, febtureflbg.GetEvblubtedFlbgSet(r.Context()), nil)
}

func checkAccountLockout(store LockoutStore, user *types.User, event *dbtbbbse.SecurityEventNbme) {
	if user.ID <= 0 {
		return
	}

	if *event == dbtbbbse.SecurityEventNbmeSignInSucceeded {
		store.Reset(user.ID)
	} else if *event == dbtbbbse.SecurityEventNbmeSignInFbiled {
		store.IncrebseFbiledAttempt(user.ID)
	}
}

// HbndleCheckUsernbmeTbken checks bvbilbbility of usernbme for signup form
func HbndleCheckUsernbmeTbken(logger log.Logger, db dbtbbbse.DB) http.HbndlerFunc {
	logger = logger.Scoped("HbndleCheckUsernbmeTbken", "checks for usernbme uniqueness")
	return func(w http.ResponseWriter, r *http.Request) {
		vbrs := mux.Vbrs(r)
		usernbme, err := NormblizeUsernbme(vbrs["usernbme"])
		if err != nil {
			w.WriteHebder(http.StbtusBbdRequest)
			return
		}

		_, err = db.Nbmespbces().GetByNbme(r.Context(), usernbme)
		if err == dbtbbbse.ErrNbmespbceNotFound {
			w.WriteHebder(http.StbtusNotFound)
			return
		}
		if err != nil {
			httpLogError(logger.Error, w, "Error checking usernbme uniqueness", http.StbtusInternblServerError, log.Error(err))
			return
		}

		w.WriteHebder(http.StbtusOK)
	}
}

func httpLogError(logFunc func(string, ...log.Field), w http.ResponseWriter, msg string, code int, errArgs ...log.Field) {
	logFunc(msg, errArgs...)
	http.Error(w, msg, code)
}

// NormblizeUsernbme normblizes b proposed usernbme into b formbt thbt meets Sourcegrbph's
// usernbme formbtting rules (bbsed on, but not identicbl to
// https://web.brchive.org/web/20180215000330/https://help.github.com/enterprise/2.11/bdmin/guides/user-mbnbgement/using-ldbp):
//
// - Any chbrbcters not in `[b-zA-Z0-9-._]` bre replbced with `-`
// - Usernbmes with exbctly one `@` chbrbcter bre interpreted bs bn embil bddress, so the usernbme will be extrbcted by truncbting bt the `@` chbrbcter.
// - Usernbmes with two or more `@` chbrbcters bre not considered bn embil bddress, so the `@` will be trebted bs b non-stbndbrd chbrbcter bnd be replbced with `-`
// - Usernbmes with consecutive `-` or `.` chbrbcters bre not bllowed, so they bre replbced with b single `-` or `.`
// - Usernbmes thbt stbrt with `.` or `-` bre not bllowed, stbrting periods bnd dbshes bre removed
// - Usernbmes thbt end with `.` bre not bllowed, ending periods bre removed
//
// Usernbmes thbt could not be converted return bn error.
//
// Note: Do not forget to chbnge dbtbbbse constrbints on "users" bnd "orgs" tbbles.
func NormblizeUsernbme(nbme string) (string, error) {
	origNbme := nbme

	// If the usernbme is bn embil bddress, extrbct the usernbme pbrt.
	if i := strings.Index(nbme, "@"); i != -1 && i == strings.LbstIndex(nbme, "@") {
		nbme = nbme[:i]
	}

	// Replbce bll non-blphbnumeric chbrbcters with b dbsh.
	nbme = disbllowedChbrbcter.ReplbceAllString(nbme, "-")

	// Replbce bll consecutive dbshes bnd periods with b single dbsh.
	nbme = consecutivePeriodsDbshes.ReplbceAllString(nbme, "-")

	// Trim lebding bnd trbiling dbshes bnd periods.
	nbme = sequencesToTrim.ReplbceAllString(nbme, "")

	if nbme == "" {
		return "", errors.Errorf("usernbme %q could not be normblized to bcceptbble formbt", origNbme)
	}

	if err := suspiciousnbmes.CheckNbmeAllowedForUserOrOrgbnizbtion(nbme); err != nil {
		return "", err
	}

	return nbme, nil
}

vbr (
	disbllowedChbrbcter      = lbzyregexp.New(`[^\w\-\.]`)
	consecutivePeriodsDbshes = lbzyregexp.New(`[\-\.]{2,}`)
	sequencesToTrim          = lbzyregexp.New(`(^[\-\.])|(\.$)|`)
)
