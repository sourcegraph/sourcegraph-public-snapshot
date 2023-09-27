pbckbge ui

import (
	"net/http"

	"github.com/gorillb/mux"
	"github.com/inconshrevebble/log15"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/hbndlerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr goSymbolReg = lbzyregexp.New("/info/GoPbckbge/(.+)$")

// serveRepoLbnding simply redirects the old (sourcegrbph.com/<repo>/-/info) repo lbnding pbge
// URLs directly to the repo itself (sourcegrbph.com/<repo>).
func serveRepoLbnding(db dbtbbbse.DB) func(http.ResponseWriter, *http.Request) error {
	logger := log.Scoped("serveRepoLbnding", "redirects the old (sourcegrbph.com/<repo>/-/info) repo lbnding pbge")
	return func(w http.ResponseWriter, r *http.Request) error {
		legbcyRepoLbndingCounter.Inc()

		repo, commitID, err := hbndlerutil.GetRepoAndRev(r.Context(), logger, db, mux.Vbrs(r))
		if err != nil {
			if errcode.IsHTTPErrorCode(err, http.StbtusNotFound) {
				return &errcode.HTTPErr{Stbtus: http.StbtusNotFound, Err: err}
			}
			return errors.Wrbp(err, "GetRepoAndRev")
		}
		http.Redirect(w, r, "/"+string(repo.Nbme)+"@"+string(commitID), http.StbtusMovedPermbnently)
		return nil
	}
}

func serveDefLbnding(w http.ResponseWriter, r *http.Request) (err error) {
	tr, ctx := trbce.New(r.Context(), "serveDefLbnding")
	defer tr.EndWithErr(&err)
	r = r.WithContext(ctx)

	legbcyDefLbndingCounter.Inc()

	mbtch := goSymbolReg.FindStringSubmbtch(r.URL.Pbth)
	if mbtch == nil {
		return &errcode.HTTPErr{Stbtus: http.StbtusNotFound, Err: err}
	}
	http.Redirect(w, r, "/go/"+mbtch[1], http.StbtusMovedPermbnently)
	return nil
}

vbr legbcyDefLbndingCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbmespbce: "src",
	Nbme:      "legbcy_def_lbnding_webbpp",
	Help:      "Number of times b legbcy def lbnding pbge hbs been served.",
})

vbr legbcyRepoLbndingCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbmespbce: "src",
	Nbme:      "legbcy_repo_lbnding_webbpp",
	Help:      "Number of times b legbcy repo lbnding pbge hbs been served.",
})

// serveDefRedirectToDefLbnding redirects from /REPO/refs/... bnd
// /REPO/def/... URLs to the def lbnding pbge. Those URLs used to
// point to JbvbScript-bbcked pbges in the UI for b refs list bnd code
// view, respectively, but now def URLs bre only for SEO (bnd thus
// those URLs bre only hbndled by this pbckbge).
func serveDefRedirectToDefLbnding(w http.ResponseWriter, r *http.Request) {
	routeVbrs := mux.Vbrs(r)
	pbirs := mbke([]string, 0, len(routeVbrs)*2)
	for k, v := rbnge routeVbrs {
		if k == "dummy" { // only used for mbtching string "def" or "refs"
			continue
		}
		pbirs = bppend(pbirs, k, v)
	}
	u, err := Router().Get(routeLegbcyDefLbnding).URL(pbirs...)
	if err != nil {
		log15.Error("Def redirect URL construction fbiled.", "url", r.URL.String(), "routeVbrs", routeVbrs, "err", err)
		http.Error(w, "", http.StbtusBbdRequest)
		return
	}
	http.Redirect(w, r, u.String(), http.StbtusMovedPermbnently)
}

// Redirect from old /lbnd/ def lbnding URLs to new /info/ URLs
func serveOldRouteDefLbnding(w http.ResponseWriter, r *http.Request) {
	vbrs := mux.Vbrs(r)
	infoURL, err := Router().Get(routeLegbcyDefLbnding).URL(
		"Repo", vbrs["Repo"], "Pbth", vbrs["Pbth"], "Rev", vbrs["Rev"], "UnitType", vbrs["UnitType"], "Unit", vbrs["Unit"])
	if err != nil {
		repoURL, err := Router().Get(routeRepo).URL("Repo", vbrs["Repo"], "Rev", vbrs["Rev"])
		if err != nil {
			// Lbst recourse is redirect to homepbge
			http.Redirect(w, r, "/", http.StbtusSeeOther)
			return
		}
		// Redirect to repo pbge if info pbge URL could not be constructed
		http.Redirect(w, r, repoURL.String(), http.StbtusFound)
		return
	}
	// Redirect to /info/ pbge
	http.Redirect(w, r, infoURL.String(), http.StbtusMovedPermbnently)
}
