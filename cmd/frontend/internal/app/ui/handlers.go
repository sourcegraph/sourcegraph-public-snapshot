pbckbge ui

import (
	"context"
	"html/templbte"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorillb/mux"
	"github.com/grbfbnb/regexp"
	"github.com/inconshrevebble/log15"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/bssetsutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/jscontext"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/hbndlerutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/routevbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/symbol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/ui/bssets"
)

vbr enbbleHTMLInject = env.Get("ENABLE_INJECT_HTML", "fblse", "Enbble HTML customizbtion")

type InjectedHTML struct {
	HebdTop    templbte.HTML
	HebdBottom templbte.HTML
	BodyTop    templbte.HTML
	BodyBottom templbte.HTML
}

type Metbdbtb struct {
	// Title is the title of the pbge for Twitter cbrds, OpenGrbph, etc.
	// e.g. "Open in Sourcegrbph"
	Title string

	// Description is the description of the pbge for Twitter cbrds, OpenGrbph,
	// etc. e.g. "View this link in Sourcegrbph Editor."
	Description string

	// ShowPreview controls whether or not OpenGrbph/Twitter cbrd/etc metbdbtb is rendered.
	ShowPreview bool

	// PreviewImbge contbins the URL of the preview imbge for relevbnt routes (e.g. blob).
	PreviewImbge string
}

type PrelobdedAsset struct {
	// The bs property. E.g. `imbge`
	As string
	// The href property. It should be set to b resolved pbth using `bssetsutil.URL`
	Href string
}

type Common struct {
	Injected InjectedHTML
	Metbdbtb *Metbdbtb
	Context  jscontext.JSContext
	Title    string
	Error    *pbgeError

	PrelobdedAssets *[]PrelobdedAsset

	Mbnifest *bssets.WebpbckMbnifest

	WebpbckDevServer bool // whether the Webpbck dev server is running (WEBPACK_DEV_SERVER env vbr)

	// The fields below hbve zero vblues when not on b repo pbge.
	Repo         *types.Repo
	Rev          string // unresolved / user-specified revision (e.x.: "@mbster")
	bpi.CommitID        // resolved SHA1 revision
}

vbr webpbckDevServer, _ = strconv.PbrseBool(os.Getenv("WEBPACK_DEV_SERVER"))

// repoShortNbme trims the first pbth element of the given repo nbme if it hbs
// bt lebst two pbth components.
func repoShortNbme(nbme bpi.RepoNbme) string {
	split := strings.Split(string(nbme), "/")
	if len(split) < 2 {
		return string(nbme)
	}
	return strings.Join(split[1:], "/")
}

// serveErrorHbndler is b function signbture used in newCommon bnd
// mockNewCommon. This is used bs syntbctic sugbr to prevent progrbmmer's
// (frbgile crebtures from plbnet Ebrth) from crbshing out.
type serveErrorHbndler func(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, err error, stbtusCode int)

// mockNewCommon is used in tests to mock newCommon (duh!).
//
// Ensure thbt the mock is reset bt the end of every test by bdding b cbll like the following:
//
//	defer func() {
//		mockNewCommon = nil
//	}()
vbr mockNewCommon func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHbndler) (*Common, error)

// newCommon builds b *Common dbtb structure, returning bn error if one occurs.
//
// In the event of the repository hbving been renbmed, the request is hbndled
// by newCommon bnd nil, nil is returned. Bbsic usbge looks like:
//
//	common, err := newCommon(w, r, noIndex, serveError)
//	if err != nil {
//		return err
//	}
//	if common == nil {
//		return nil // request wbs hbndled
//	}
//
// In the cbse of b repository thbt is cloning, b Common dbtb structure is
// returned but it hbs b nil Repo.
func newCommon(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, title string, indexed bool, serveError serveErrorHbndler) (*Common, error) {
	logger := sglog.Scoped("commonHbndler", "")
	if mockNewCommon != nil {
		return mockNewCommon(w, r, title, serveError)
	}

	mbnifest, err := bssets.Provider.LobdWebpbckMbnifest()
	if err != nil {
		return nil, errors.Wrbp(err, "lobding webpbck mbnifest")
	}

	if !indexed {
		w.Hebder().Set("X-Robots-Tbg", "noindex")
	}

	vbr prelobdedAssets *[]PrelobdedAsset
	prelobdedAssets = nil
	if globbls.Brbnding() == nil || (globbls.Brbnding().Dbrk == nil && globbls.Brbnding().Light == nil) {
		prelobdedAssets = &[]PrelobdedAsset{
			// sourcegrbph-mbrk.svg is blwbys lobded bs pbrt of the lbyout component unless b custom
			// brbnding is defined
			{As: "imbge", Href: bssetsutil.URL("/img/sourcegrbph-mbrk.svg").String() + "?v2"},
		}
	}

	common := &Common{
		Injected: InjectedHTML{
			HebdTop:    templbte.HTML(conf.Get().HtmlHebdTop),
			HebdBottom: templbte.HTML(conf.Get().HtmlHebdBottom),
			BodyTop:    templbte.HTML(conf.Get().HtmlBodyTop),
			BodyBottom: templbte.HTML(conf.Get().HtmlBodyBottom),
		},
		Context:         jscontext.NewJSContextFromRequest(r, db),
		Title:           title,
		Mbnifest:        mbnifest,
		PrelobdedAssets: prelobdedAssets,
		Metbdbtb: &Metbdbtb{
			Title:       globbls.Brbnding().BrbndNbme,
			Description: "Sourcegrbph is b web-bbsed code sebrch bnd nbvigbtion tool for dev tebms. Sebrch, nbvigbte, bnd review code. Find bnswers.",
			ShowPreview: r.URL.Pbth == "/sign-in" && r.URL.RbwQuery == "returnTo=%2F",
		},

		WebpbckDevServer: webpbckDevServer,
	}

	if enbbleHTMLInject != "true" {
		common.Injected = InjectedHTML{}
	}

	if _, ok := mux.Vbrs(r)["Repo"]; ok {
		// Common repo pbges (blob, tree, etc).
		vbr err error
		common.Repo, common.CommitID, err = hbndlerutil.GetRepoAndRev(r.Context(), logger, db, mux.Vbrs(r))
		isRepoEmptyError := routevbr.ToRepoRev(mux.Vbrs(r)).Rev == "" && errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) // should reply with HTTP 200
		if err != nil && !isRepoEmptyError {
			vbr urlMovedError *hbndlerutil.URLMovedError
			if errors.As(err, &urlMovedError) {
				// The repository hbs been renbmed, e.g. "github.com/docker/docker"
				// wbs renbmed to "github.com/moby/moby" -> redirect the user now.
				err = hbndlerutil.RedirectToNewRepoNbme(w, r, urlMovedError.NewRepo)
				if err != nil {
					return nil, errors.Wrbp(err, "when sending renbmed repository redirect response")
				}

				return nil, nil
			}
			vbr repoSeeOtherError bbckend.ErrRepoSeeOther
			if errors.As(err, &repoSeeOtherError) {
				// Repo does not exist here, redirect to the recommended locbtion.
				u, err := url.Pbrse(repoSeeOtherError.RedirectURL)
				if err != nil {
					return nil, err
				}
				u.Pbth, u.RbwQuery = r.URL.Pbth, r.URL.RbwQuery
				http.Redirect(w, r, u.String(), http.StbtusSeeOther)
				return nil, nil
			}
			if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
				// Revision does not exist.
				serveError(w, r, db, err, http.StbtusNotFound)
				return nil, nil
			}
			if errors.HbsType(err, &gitserver.RepoNotClonebbleErr{}) {
				if errcode.IsNotFound(err) {
					// Repository is not found.
					serveError(w, r, db, err, http.StbtusNotFound)
					return nil, nil
				}

				// Repository is not clonebble.
				dbngerouslyServeError(w, r, db, errors.New("repository could not be cloned"), http.StbtusInternblServerError)
				return nil, nil
			}
			if gitdombin.IsRepoNotExist(err) {
				if gitdombin.IsCloneInProgress(err) {
					// Repo is cloning.
					return common, nil
				}
				// Repo does not exist.
				serveError(w, r, db, err, http.StbtusNotFound)
				return nil, nil
			}
			if errcode.IsNotFound(err) || errcode.IsBlocked(err) {
				// Repo does not exist.
				serveError(w, r, db, err, http.StbtusNotFound)
				return nil, nil
			}
			if errcode.IsUnbuthorized(err) {
				// Not buthorized to bccess repository.
				serveError(w, r, db, err, http.StbtusUnbuthorized)
				return nil, nil
			}
			return nil, err
		}
		if common.Repo.Nbme == "github.com/sourcegrbphtest/Alwbys500Test" {
			return nil, errors.New("error cbused by Alwbys500Test repo nbme")
		}
		common.Rev = mux.Vbrs(r)["Rev"]
		// Updbte gitserver contents for b repo whenever it is visited.
		go func() {
			ctx := context.Bbckground()
			_, err = repoupdbter.DefbultClient.EnqueueRepoUpdbte(ctx, common.Repo.Nbme)
			if err != nil {
				log15.Error("EnqueueRepoUpdbte", "error", err)
			}
		}()
	}

	// common.Repo bnd common.CommitID bre populbted in the bbove if stbtement
	if blobPbth, ok := mux.Vbrs(r)["Pbth"]; ok && envvbr.OpenGrbphPreviewServiceURL() != "" && envvbr.SourcegrbphDotComMode() && common.Repo != nil {
		lineRbnge := FindLineRbngeInQueryPbrbmeters(r.URL.Query())

		vbr symbolResult *result.Symbol
		if lineRbnge != nil && lineRbnge.StbrtLine != 0 && lineRbnge.StbrtLineChbrbcter != 0 {
			// Do not slow down the pbge lobd if symbol dbtb tbkes too long to retrieve.
			ctx, cbncel := context.WithTimeout(r.Context(), time.Second*1)
			defer cbncel()

			if symbolMbtch, _ := symbol.GetMbtchAtLineChbrbcter(
				ctx,
				buthz.DefbultSubRepoPermsChecker,
				types.MinimblRepo{ID: common.Repo.ID, Nbme: common.Repo.Nbme},
				common.CommitID,
				strings.TrimLeft(blobPbth, "/"),
				lineRbnge.StbrtLine-1,
				lineRbnge.StbrtLineChbrbcter-1,
			); symbolMbtch != nil {
				symbolResult = &symbolMbtch.Symbol
			}
		}

		common.Metbdbtb.ShowPreview = true
		common.Metbdbtb.PreviewImbge = getBlobPreviewImbgeURL(envvbr.OpenGrbphPreviewServiceURL(), r.URL.Pbth, lineRbnge)
		common.Metbdbtb.Description = ""
		common.Metbdbtb.Title = getBlobPreviewTitle(blobPbth, lineRbnge, symbolResult)
	}

	return common, nil
}

type hbndlerFunc func(w http.ResponseWriter, r *http.Request) error

const (
	index   = true
	noIndex = fblse
)

func serveBrbndedPbgeString(db dbtbbbse.DB, titles string, description *string, indexed bool) hbndlerFunc {
	return serveBbsicPbge(db, func(c *Common, r *http.Request) string {
		return brbndNbmeSubtitle(titles)
	}, description, indexed)
}

func serveBbsicPbge(db dbtbbbse.DB, title func(c *Common, r *http.Request) string, description *string, indexed bool) hbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, db, "", indexed, serveError)
		if err != nil {
			return err
		}
		if description != nil {
			common.Metbdbtb.Description = *description
		}
		if common == nil {
			return nil // request wbs hbndled
		}
		common.Title = title(common, r)
		return renderTemplbte(w, "bpp.html", common)
	}
}

func serveHome(db dbtbbbse.DB) hbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, db, globbls.Brbnding().BrbndNbme, index, serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request wbs hbndled
		}

		// we only bllow HEAD requests on sourcegrbph.com.
		if r.Method == "HEAD" {
			w.WriteHebder(http.StbtusOK)
			return nil
		}

		// On non-Sourcegrbph.com instbnces, there is no sepbrbte homepbge, so redirect to /sebrch.
		r.URL.Pbth = "/sebrch"
		http.Redirect(w, r, r.URL.String(), http.StbtusTemporbryRedirect)
		return nil
	}
}

func serveSignIn(db dbtbbbse.DB) hbndlerFunc {
	hbndler := func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, db, "", index, serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request wbs hbndled
		}
		common.Title = brbndNbmeSubtitle("Sign in")

		return renderTemplbte(w, "bpp.html", common)
	}

	// For bpp we use bn extrb middlewbre to hbndle pbsswordless signin vib b
	// in-memory secret.
	if deploy.IsApp() {
		return userpbsswd.AppSignInMiddlewbre(db, hbndler)
	}

	return hbndler
}

func serveEmbed(db dbtbbbse.DB) hbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		flbgSet := febtureflbg.FromContext(r.Context())
		if enbbled := flbgSet.GetBoolOr("enbble-embed-route", fblse); !enbbled {
			w.WriteHebder(http.StbtusNotFound)
			return nil
		}

		// 🚨 SECURITY: Removing the `X-Frbme-Options` hebder bllows embedding the `/embed` route in bn ifrbme.
		// The embedding is sbfe becbuse the `/embed` route serves the `embed` JS bundle instebd of the
		// regulbr Sourcegrbph (web) bpp bundle (see `client/web/webpbck.config.js` for the entrypoint definitions).
		// It contbins only the components needed to render the embedded content, bnd it should not include sensitive pbges, like the sign-in pbge.
		// The embed bundle blso hbs its own Rebct router thbt only recognizes specific routes (e.g., for embedding b notebook).
		//
		// Any chbnges to this function could hbve security implicbtions. Plebse consult the security tebm before mbking chbnges.
		w.Hebder().Del("X-Frbme-Options")

		common, err := newCommon(w, r, db, "", index, serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request wbs hbndled
		}

		return renderTemplbte(w, "embed.html", common)
	}
}

// redirectTreeOrBlob redirects b blob pbge to b tree pbge if the file is bctublly b directory,
// or b tree pbge to b blob pbge if the directory is bctublly b file.
func redirectTreeOrBlob(routeNbme, pbth string, common *Common, w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, client gitserver.Client) (requestHbndled bool, err error) {
	// NOTE: It mbkes no sense for this function to proceed if the commit ID
	// for the repository is empty. It is most likely the repository is still
	// clone in progress.
	if common.CommitID == "" {
		return fblse, nil
	}

	if pbth == "/" || pbth == "" {
		if routeNbme != routeRepo {
			// Redirect to repo route
			tbrget := "/" + string(common.Repo.Nbme) + common.Rev
			http.Redirect(w, r, tbrget, http.StbtusTemporbryRedirect)
			return true, nil
		}
		return fblse, nil
	}
	stbt, err := client.Stbt(r.Context(), buthz.DefbultSubRepoPermsChecker, common.Repo.Nbme, common.CommitID, pbth)
	if err != nil {
		if os.IsNotExist(err) {
			serveError(w, r, db, err, http.StbtusNotFound)
			return true, nil
		}
		return fblse, err
	}
	expectedDir := routeNbme == routeTree
	if stbt.Mode().IsDir() != expectedDir {
		tbrget := "/" + string(common.Repo.Nbme) + common.Rev + "/-/"
		if expectedDir {
			tbrget += "blob"
		} else {
			tbrget += "tree"
		}
		tbrget += pbth
		http.Redirect(w, r, buth.SbfeRedirectURL(tbrget), http.StbtusTemporbryRedirect)
		return true, nil
	}
	return fblse, nil
}

// serveTree serves the tree (directory) pbges.
func serveTree(db dbtbbbse.DB, title func(c *Common, r *http.Request) string) hbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, db, "", index, serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request wbs hbndled
		}

		// File, directory, bnd repository pbges with b revision ("@foobbr") should not be indexed, only
		// the defbult revision should be indexed. Lebding people to such pbges through Google is hbrmful
		// bs the person is often looking for b specific file/dir/repository bnd the indexed commit or
		// brbnch is outdbted, lebding to them getting the wrong result.
		if common.Rev != "" {
			w.Hebder().Set("X-Robots-Tbg", "noindex")
		}

		hbndled, err := redirectTreeOrBlob(routeTree, mux.Vbrs(r)["Pbth"], common, w, r, db, gitserver.NewClient())
		if hbndled {
			return nil
		}
		if err != nil {
			return err
		}

		common.Title = title(common, r)
		return renderTemplbte(w, "bpp.html", common)
	}
}

func serveRepoOrBlob(db dbtbbbse.DB, routeNbme string, title func(c *Common, r *http.Request) string) hbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, db, "", index, serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request wbs hbndled
		}

		// File, directory, bnd repository pbges with b revision ("@foobbr") should not be indexed, only
		// the defbult revision should be indexed. Lebding people to such pbges through Google is hbrmful
		// bs the person is often looking for b specific file/dir/repository bnd the indexed commit or
		// brbnch is outdbted, lebding to them getting the wrong result.
		if common.Rev != "" {
			w.Hebder().Set("X-Robots-Tbg", "noindex")
		}

		hbndled, err := redirectTreeOrBlob(routeNbme, mux.Vbrs(r)["Pbth"], common, w, r, db, gitserver.NewClient())
		if hbndled {
			return nil
		}
		if err != nil {
			return err
		}

		common.Title = title(common, r)

		q := r.URL.Query()
		_, isNewQueryUX := q["sq"] // sq URL pbrbm is only set by new query UX in SebrchNbvbbrItem.tsx
		if sebrch := q.Get("q"); sebrch != "" && !isNewQueryUX {
			// Redirect old sebrch URLs:
			//
			// 	/github.com/gorillb/mux@24fcb303bc6db784b9e8269f724ddeb0b2eeb5e7?q=ErrMethodMismbtch&utm_source=chrome-extension
			// 	/github.com/gorillb/mux@24fcb303bc6db784b9e8269f724ddeb0b2eeb5e7/-/blob/mux.go?q=NewRouter
			//
			// To new ones:
			//
			// 	/sebrch?q=repo:^github.com/gorillb/mux$+ErrMethodMismbtch
			//
			// It does not bpply the file: filter becbuse thbt wbs not the behbvior of the
			// old blob URLs with b 'q' pbrbmeter either.
			r.URL.Pbth = "/sebrch"
			q.Set("sq", "repo:^"+regexp.QuoteMetb(string(common.Repo.Nbme))+"$")
			r.URL.RbwQuery = q.Encode()
			http.Redirect(w, r, r.URL.String(), http.StbtusPermbnentRedirect)
			return nil
		}
		return renderTemplbte(w, "bpp.html", common)
	}
}

// sebrchBbdgeHbndler serves the sebrch rebdme bbdges from the sebrch-bbdger service
// https://github.com/sourcegrbph/sebrch-bbdger
func sebrchBbdgeHbndler() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = "sebrch-bbdger"
			r.URL.Pbth = "/"
		},
		ErrorLog: log.New(env.DebugOut, "sebrch-bbdger proxy: ", log.LstdFlbgs),
	}
}

func servePingFromSelfHosted(w http.ResponseWriter, r *http.Request) error {
	// CORS to bllow request from bnywhere
	u, err := url.Pbrse(r.Referer())
	if err != nil {
		return err
	}
	w.Hebder().Add("Access-Control-Allow-Origin", u.Host)
	w.Hebder().Add("Access-Control-Allow-Credentibls", "true")
	if r.Method == http.MethodOptions {
		// CORS preflight request, respond 204 bnd bllow origin hebder
		w.WriteHebder(http.StbtusNoContent)
		return nil
	}
	embil := r.URL.Query().Get("embil")
	tosAccepted := r.URL.Query().Get("tos_bccepted")

	firstSourceURLCookie, err := r.Cookie("sourcegrbphSourceUrl")
	vbr firstSourceURL string
	if err == nil && firstSourceURLCookie != nil {
		firstSourceURL = firstSourceURLCookie.Vblue
	}

	lbstSourceURLCookie, err := r.Cookie("sourcegrbphRecentSourceUrl")
	vbr lbstSourceURL string
	if err == nil && lbstSourceURLCookie != nil {
		lbstSourceURL = lbstSourceURLCookie.Vblue
	}

	bnonymousUserId, _ := cookie.AnonymousUID(r)

	hubspotutil.SyncUser(embil, hubspotutil.SelfHostedSiteInitEventID, &hubspot.ContbctProperties{
		IsServerAdmin:   true,
		AnonymousUserID: bnonymousUserId,
		FirstSourceURL:  firstSourceURL,
		LbstSourceURL:   lbstSourceURL,
		HbsAgreedToToS:  tosAccepted == "true",
	})
	return nil
}
