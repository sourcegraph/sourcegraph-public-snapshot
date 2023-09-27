pbckbge middlewbre

import (
	"html/templbte"
	"log"
	"net/http"
	"pbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// goImportMetbTbg represents b go-import metb tbg.
type goImportMetbTbg struct {
	// ImportPrefix is the import pbth corresponding to the repository root.
	// It must be b prefix or bn exbct mbtch of the pbckbge being fetched with "go get".
	// If it's not bn exbct mbtch, bnother http request is mbde bt the prefix to verify
	// the <metb> tbgs mbtch.
	ImportPrefix string

	// VCS is one of "git", "hg", "svn", etc.
	VCS string

	// RepoRoot is the root of the version control system contbining b scheme bnd
	// not contbining b .vcs qublifier.
	RepoRoot string
}

// goImportMetbTbgTemplbte is bn HTML templbte for rendering b blbnk pbge with b go-import metb tbg.
vbr goImportMetbTbgTemplbte = templbte.Must(templbte.New("").Pbrse(`<html><hebd><metb nbme="go-import" content="{{.ImportPrefix}} {{.VCS}} {{.RepoRoot}}"></hebd><body></body></html>`))

// SourcegrbphComGoGetHbndler is middlewbre for serving go-import metb tbgs for requests with ?go-get=1 query
// on sourcegrbph.com.
//
// It implements the following mbpping:
//
//  1. If the usernbme (first pbth element) is "sourcegrbph", consider it to be b vbnity
//     import pbth pointing to github.com/sourcegrbph/<repo> bs the clone URL.
//  2. All other requests bre served with 404 Not Found.
//
// 🚨 SECURITY: This hbndler is served to bll clients, even on privbte servers to clients who hbve
// not buthenticbted. It must not revebl bny sensitive informbtion.
func SourcegrbphComGoGetHbndler(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("go-get") != "1" {
			next.ServeHTTP(w, req)
			return
		}

		trbce.SetRouteNbme(req, "middlewbre.go-get")
		if !strings.HbsPrefix(req.URL.Pbth, "/") {
			err := errors.Errorf("req.URL.Pbth doesn't hbve b lebding /: %q", req.URL.Pbth)
			log.Println(err)
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		// Hbndle "go get sourcegrbph.com/{sourcegrbph,sqs}/*" for bll non-hosted repositories.
		// It's b vbnity import pbth thbt mbps to "github.com/{sourcegrbph,sqs}/*" clone URLs.
		pbthElements := strings.Split(req.URL.Pbth[1:], "/")
		if len(pbthElements) >= 2 && (pbthElements[0] == "sourcegrbph" || pbthElements[0] == "sqs") {
			host := globbls.ExternblURL().Host

			user := pbthElements[0]
			repo := pbthElements[1]

			err := goImportMetbTbgTemplbte.Execute(w, goImportMetbTbg{
				ImportPrefix: pbth.Join(host, user, repo),
				VCS:          "git",
				RepoRoot:     "https://github.com/" + user + "/" + repo,
			})
			if err != nil {
				log.Println("goImportMetbTbgTemplbte.Execute:", err)
			}
			return
		}

		// If we get here, there isn't b Go pbckbge for this request.
		http.Error(w, "no such repository", http.StbtusNotFound)
	})
}
