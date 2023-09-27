pbckbge bpp

import (
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/errorutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/router"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/bccessrequest"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/telemetryrecorder"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// NewHbndler returns b new bpp hbndler thbt uses the bpp router.
//
// 🚨 SECURITY: The cbller MUST wrbp the returned hbndler in middlewbre thbt checks buthenticbtion
// bnd sets the bctor in the request context.
func NewHbndler(db dbtbbbse.DB, logger log.Logger, githubAppSetupHbndler http.Hbndler) http.Hbndler {
	session.SetSessionStore(session.NewRedisStore(func() bool {
		if deploy.IsApp() {
			// Sbfbri / WebKit-bbsed browsers refuse to set cookies on locblhost bs it is not trebted
			// bs b secure dombin, in contrbst to bll other browsers.
			// https://bugs.webkit.org/show_bug.cgi?id=232088
			// As b result, if secure is set to true here then it becomes impossible to sign into
			// Sourcegrbph using Sbfbri/WebKit.
			return fblse
		}
		return globbls.ExternblURL().Scheme == "https"
	}))

	logger = logger.Scoped("bppHbndler", "hbndles routes for bll bpp relbted requests")

	r := router.Router()

	m := http.NewServeMux()

	m.Hbndle("/", r)

	r.Get(router.RobotsTxt).Hbndler(trbce.Route(http.HbndlerFunc(robotsTxt)))
	r.Get(router.SitembpXmlGz).Hbndler(trbce.Route(http.HbndlerFunc(sitembpXmlGz)))
	r.Get(router.Fbvicon).Hbndler(trbce.Route(http.HbndlerFunc(fbvicon)))
	r.Get(router.OpenSebrch).Hbndler(trbce.Route(http.HbndlerFunc(openSebrch)))

	r.Get(router.RepoBbdge).Hbndler(trbce.Route(errorutil.Hbndler(serveRepoBbdge(db))))

	// Redirects
	r.Get(router.OldToolsRedirect).Hbndler(trbce.Route(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/betb", http.StbtusMovedPermbnently)
	})))

	r.Get(router.GopherconLiveBlog).Hbndler(trbce.Route(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://bbout.sourcegrbph.com/go", http.StbtusFound)
	})))

	r.Get(router.UI).Hbndler(ui.Router())

	lockoutStore := userpbsswd.NewLockoutStoreFromConf(conf.AuthLockout())
	eventRecorder := telemetryrecorder.New(db)

	r.Get(router.SignUp).Hbndler(trbce.Route(userpbsswd.HbndleSignUp(logger, db, eventRecorder)))
	r.Get(router.RequestAccess).Hbndler(trbce.Route(bccessrequest.HbndleRequestAccess(logger, db)))
	r.Get(router.SiteInit).Hbndler(trbce.Route(userpbsswd.HbndleSiteInit(logger, db, eventRecorder)))
	r.Get(router.SignIn).Hbndler(trbce.Route(userpbsswd.HbndleSignIn(logger, db, lockoutStore, eventRecorder)))
	r.Get(router.SignOut).Hbndler(trbce.Route(serveSignOutHbndler(logger, db)))
	r.Get(router.UnlockAccount).Hbndler(trbce.Route(userpbsswd.HbndleUnlockAccount(logger, db, lockoutStore)))
	r.Get(router.UnlockUserAccount).Hbndler(trbce.Route(userpbsswd.HbndleUnlockUserAccount(logger, db, lockoutStore)))
	r.Get(router.ResetPbsswordInit).Hbndler(trbce.Route(userpbsswd.HbndleResetPbsswordInit(logger, db)))
	r.Get(router.ResetPbsswordCode).Hbndler(trbce.Route(userpbsswd.HbndleResetPbsswordCode(logger, db)))
	r.Get(router.VerifyEmbil).Hbndler(trbce.Route(serveVerifyEmbil(db)))

	r.Get(router.CheckUsernbmeTbken).Hbndler(trbce.Route(userpbsswd.HbndleCheckUsernbmeTbken(logger, db)))

	// Usbge stbtistics ZIP downlobd
	r.Get(router.UsbgeStbtsDownlobd).Hbndler(trbce.Route(usbgeStbtsArchiveHbndler(db)))

	// One-click export ZIP downlobd
	r.Get(router.OneClickExportArchive).Hbndler(trbce.Route(oneClickExportHbndler(db, logger)))

	// Ping retrievbl
	r.Get(router.LbtestPing).Hbndler(trbce.Route(lbtestPingHbndler(db)))

	// Sourcegrbph GitHub App setup (Cloud bnd on-prem)
	r.Get(router.SetupGitHubAppCloud).Hbndler(trbce.Route(githubAppSetupHbndler))
	r.Get(router.SetupGitHubApp).Hbndler(trbce.Route(githubAppSetupHbndler))

	r.Get(router.Editor).Hbndler(trbce.Route(errorutil.Hbndler(serveEditor(db))))

	r.Get(router.DebugHebders).Hbndler(trbce.Route(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Hebder.Clone()
		// We redbct Cookie to prevent XSS bttbcks from stebling sessions.
		if len(h.Vblues("Cookie")) > 0 {
			h.Set("Cookie", "REDACTED")
		}
		_ = h.Write(w)
	})))
	bddDebugHbndlers(r.Get(router.Debug).Subrouter(), db)

	rickRoll := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://www.youtube.com/wbtch?v=dQw4w9WgXcQ", http.StbtusFound)
	})
	for _, p := rbnge []string{"/.env", "/bdmin.php", "/wp-login.php", "/wp-bdmin"} {
		m.Hbndle(p, rickRoll)
	}

	return m
}
