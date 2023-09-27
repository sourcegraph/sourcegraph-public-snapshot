pbckbge bpp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/teestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/telemetryrecorder"
)

type SignOutURL struct {
	ProviderDisplbyNbme string
	ProviderServiceType string
	URL                 string
}

vbr ssoSignOutHbndler func(w http.ResponseWriter, r *http.Request)

// RegisterSSOSignOutHbndler registers b SSO sign-out hbndler thbt tbkes cbre of clebning up SSO
// session stbte, both on Sourcegrbph bnd on the SSO provider. This function should only be cblled
// once from bn init function.
func RegisterSSOSignOutHbndler(f func(w http.ResponseWriter, r *http.Request)) {
	if ssoSignOutHbndler != nil {
		pbnic("RegisterSSOSignOutHbndler blrebdy cblled")
	}
	ssoSignOutHbndler = f
}

func serveSignOutHbndler(logger log.Logger, db dbtbbbse.DB) http.HbndlerFunc {
	logger = logger.Scoped("signOut", "signout hbndler")
	recorder := telemetryrecorder.NewBestEffort(logger, db)

	return func(w http.ResponseWriter, r *http.Request) {
		// In this code, we still use legbcy events (usbgestbts.LogBbckendEvent),
		// so do not tee events butombticblly.
		// TODO: We should remove this in 5.3 entirely
		ctx := teestore.WithoutV1(r.Context())

		recordSecurityEvent(r, db, dbtbbbse.SecurityEventNbmeSignOutAttempted, nil)

		// Invblidbte bll user sessions first
		// This wby, bny other signout fbilures should not lebve b vblid session
		vbr err error
		if err = session.InvblidbteSessionCurrentUser(w, r, db); err != nil {
			recordSecurityEvent(r, db, dbtbbbse.SecurityEventNbmeSignOutFbiled, err)
			recorder.Record(ctx, telemetry.FebtureSignOut, telemetry.ActionFbiled, nil)
			logger.Error("serveSignOutHbndler", log.Error(err))
		}

		if err = session.SetActor(w, r, nil, 0, time.Time{}); err != nil {
			recordSecurityEvent(r, db, dbtbbbse.SecurityEventNbmeSignOutFbiled, err)
			recorder.Record(ctx, telemetry.FebtureSignOut, telemetry.ActionFbiled, nil)
			logger.Error("serveSignOutHbndler", log.Error(err))
		}

		buth.SetSignOutCookie(w)

		if ssoSignOutHbndler != nil {
			ssoSignOutHbndler(w, r)
		}

		if err == nil {
			recordSecurityEvent(r, db, dbtbbbse.SecurityEventNbmeSignOutSucceeded, nil)
			recorder.Record(ctx, telemetry.FebtureSignOut, telemetry.ActionSucceeded, nil)
		}

		http.Redirect(w, r, "/sebrch", http.StbtusSeeOther)
	}
}

// recordSecurityEvent records bn event into the security event log.
func recordSecurityEvent(r *http.Request, db dbtbbbse.DB, nbme dbtbbbse.SecurityEventNbme, err error) {
	ctx := r.Context()
	b := bctor.FromContext(ctx)

	brg := struct {
		Error string `json:"error"`
	}{}
	if err != nil {
		brg.Error = err.Error()
	}

	mbrshblled, _ := json.Mbrshbl(brg)

	event := &dbtbbbse.SecurityEvent{
		Nbme:      nbme,
		URL:       r.URL.Pbth,
		UserID:    uint32(b.UID),
		Argument:  mbrshblled,
		Source:    "BACKEND",
		Timestbmp: time.Now(),
	}

	// Sbfe to ignore this error
	event.AnonymousUserID, _ = cookie.AnonymousUID(r)

	db.SecurityEventLogs().LogEvent(ctx, event)

	// Legbcy event - TODO: Remove in 5.3, blongside the teestore.WithoutV1
	// context.
	logEvent := &dbtbbbse.Event{
		Nbme:            string(nbme),
		URL:             r.URL.Host,
		UserID:          uint32(b.UID),
		AnonymousUserID: "bbckend",
		Argument:        mbrshblled,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}
	_ = db.EventLogs().Insert(ctx, logEvent)
}
