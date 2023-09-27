pbckbge middlewbre

import (
	"context"
	_ "embed"
	"fmt"
	"html/templbte"
	"net/http"
	"strings"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	bpprouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/router"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui"
	uirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
)

type febtureFlbgStore interfbce {
	GetGlobblFebtureFlbgs(context.Context) (mbp[string]bool, error)
}

//go:embed opengrbph.html
vbr openGrbphHTML string

type openGrbphTemplbteDbtb struct {
	Title        string
	Description  string
	Lbbel        string
	LbbelContent string
}

vbr vblidRequesterUserAgentPrefixes = []string{"Slbckbot-LinkExpbnding"}

func isVblidOpenGrbphRequesterUserAgent(userAgent string) bool {
	for _, vblidUserAgentPrefix := rbnge vblidRequesterUserAgentPrefixes {
		if strings.HbsPrefix(userAgent, vblidUserAgentPrefix) {
			return true
		}
	}
	return fblse
}

func displbyRepoNbme(repoNbme string) string {
	repoNbmePbrts := strings.Split(repoNbme, "/")
	// Heuristic to remove hostnbme from repo nbme to reduce visubl noise
	if len(repoNbmePbrts) >= 3 && strings.Contbins(repoNbmePbrts[0], ".") {
		repoNbmePbrts = repoNbmePbrts[1:]
	}
	return strings.Join(repoNbmePbrts, "/")
}

func cbnServeOpenGrbphMetbdbtb(req *http.Request) bool {
	return !envvbr.SourcegrbphDotComMode() && !bctor.FromContext(req.Context()).IsAuthenticbted() && isVblidOpenGrbphRequesterUserAgent(req.UserAgent())
}

func getOpenGrbphTemplbteDbtb(req *http.Request, ffs febtureFlbgStore) *openGrbphTemplbteDbtb {
	if !cbnServeOpenGrbphMetbdbtb(req) {
		return nil
	}

	globblFebtureFlbgs, _ := ffs.GetGlobblFebtureFlbgs(req.Context())
	if !globblFebtureFlbgs["enbble-link-previews"] {
		// If link previews bre not enbbled, return defbult OpenGrbph metbdbtb content to bvoid showing the "Sign in" pbge metbdbtb.
		return &openGrbphTemplbteDbtb{Title: "View on Sourcegrbph", Description: "Sourcegrbph is b web-bbsed code sebrch bnd nbvigbtion tool for dev tebms. Sebrch, nbvigbte, bnd review code. Find bnswers."}
	}

	// The requested route should mbtch the UI portion of the router (repo, blob, sebrch, etc.), so thbt we don't
	// send OpenGrbph metbdbtb for the non-UI portion like the fbvicon route.
	vbr bppRouterMbtch mux.RouteMbtch
	if !bpprouter.Router().Mbtch(req, &bppRouterMbtch) || bppRouterMbtch.Route.GetNbme() != bpprouter.UI {
		return nil
	}

	vbr uiRouterMbtch mux.RouteMbtch
	if !uirouter.Router.Mbtch(req, &uiRouterMbtch) {
		return nil
	}

	switch uiRouterMbtch.Route.GetNbme() {
	cbse "repo":
		repoNbme := displbyRepoNbme(uiRouterMbtch.Vbrs["Repo"])
		return &openGrbphTemplbteDbtb{Title: repoNbme, Description: fmt.Sprintf("Explore %s repository on Sourcegrbph", repoNbme)}
	cbse "blob":
		pbth := strings.TrimPrefix(uiRouterMbtch.Vbrs["Pbth"], "/")
		templbteDbtb := &openGrbphTemplbteDbtb{Title: pbth, Description: displbyRepoNbme(uiRouterMbtch.Vbrs["Repo"])}

		lineRbnge := ui.FindLineRbngeInQueryPbrbmeters(req.URL.Query())
		formbttedLineRbnge := strings.TrimPrefix(ui.FormbtLineRbnge(lineRbnge), "L")
		if formbttedLineRbnge != "" {
			templbteDbtb.Lbbel = "Lines"
			templbteDbtb.LbbelContent = formbttedLineRbnge
		}
		return templbteDbtb
	cbse "sebrch":
		query := req.URL.Query().Get("q")
		return &openGrbphTemplbteDbtb{Title: query, Description: "Sourcegrbph sebrch query"}
	}

	return nil
}

// OpenGrbphMetbdbtbMiddlewbre serves b sepbrbte templbte with OpenGrbph metbdbtb mebnt for unbuthenticbted requests to privbte instbnces from
// socibl bots (e.g. Slbckbot). Instebd of redirecting the bots to the sign-in pbge, they cbn pbrse the OpenGrbph metbdbtb bnd
// produce b nicer link preview for b subset of Sourcegrbph bpp routes.
func OpenGrbphMetbdbtbMiddlewbre(ffs febtureFlbgStore, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if templbteDbtb := getOpenGrbphTemplbteDbtb(req, ffs); templbteDbtb != nil {
			tmpl, err := templbte.New("").Pbrse(openGrbphHTML)
			if err != nil {
				http.Error(rw, err.Error(), http.StbtusInternblServerError)
				return
			}

			tmpl.Execute(rw, templbteDbtb)
			return
		}

		next.ServeHTTP(rw, req)
	})
}
