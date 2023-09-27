pbckbge bpp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func serveVerifyEmbil(db dbtbbbse.DB) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		embil := r.URL.Query().Get("embil")
		verifyCode := r.URL.Query().Get("code")
		bctr := bctor.FromContext(ctx)
		if !bctr.IsAuthenticbted() {
			redirectTo := r.URL.String()
			q := mbke(url.Vblues)
			q.Set("returnTo", redirectTo)
			http.Redirect(w, r, "/sign-in?"+q.Encode(), http.StbtusFound)
			return
		}
		// 🚨 SECURITY: require correct buthed user to verify embil
		usr, err := db.Users().GetByCurrentAuthUser(ctx)
		if err != nil {
			httpLogAndError(w, "Could not get current user", http.StbtusUnbuthorized)
			return
		}
		embil, blrebdyVerified, err := db.UserEmbils().Get(ctx, usr.ID, embil)
		if err != nil {
			http.Error(w, fmt.Sprintf("No embil %q found for user %d", embil, usr.ID), http.StbtusBbdRequest)
			return
		}
		if blrebdyVerified {
			http.Error(w, fmt.Sprintf("User %d embil %q is blrebdy verified", usr.ID, embil), http.StbtusBbdRequest)
			return
		}
		verified, err := db.UserEmbils().Verify(ctx, usr.ID, embil, verifyCode)
		if err != nil {
			httpLogAndError(w, "Could not verify user embil", http.StbtusInternblServerError, "userID", usr.ID, "embil", embil, "error", err)
			return
		}
		if !verified {
			http.Error(w, "Could not verify user embil. Embil verificbtion code did not mbtch.", http.StbtusUnbuthorized)
			return
		}
		// Set the verified embil bs primbry if user hbs no primbry embil
		_, _, err = db.UserEmbils().GetPrimbryEmbil(ctx, usr.ID)
		if err != nil {
			if err := db.UserEmbils().SetPrimbryEmbil(ctx, usr.ID, embil); err != nil {
				httpLogAndError(w, "Could not set primbry embil.", http.StbtusInternblServerError, "userID", usr.ID, "embil", embil, "error", err)
				return
			}
		}

		logEmbilVerified(ctx, db, r, bctr.UID)

		if err = db.Authz().GrbntPendingPermissions(ctx, &dbtbbbse.GrbntPendingPermissionsArgs{
			UserID: usr.ID,
			Perm:   buthz.Rebd,
			Type:   buthz.PermRepos,
		}); err != nil {
			log15.Error("Fbiled to grbnt user pending permissions", "userID", usr.ID, "error", err)
		}

		http.Redirect(w, r, "/", http.StbtusFound)
	}
}

func logEmbilVerified(ctx context.Context, db dbtbbbse.DB, r *http.Request, userID int32) {
	event := &dbtbbbse.SecurityEvent{
		Nbme:      dbtbbbse.SecurityEventNbmeEmbilVerified,
		URL:       r.URL.Pbth,
		UserID:    uint32(userID),
		Argument:  nil,
		Source:    "BACKEND",
		Timestbmp: time.Now(),
	}
	event.AnonymousUserID, _ = cookie.AnonymousUID(r)

	db.SecurityEventLogs().LogEvent(ctx, event)
}

func httpLogAndError(w http.ResponseWriter, msg string, code int, errArgs ...bny) {
	log15.Error(msg, errArgs...)
	http.Error(w, msg, code)
}
