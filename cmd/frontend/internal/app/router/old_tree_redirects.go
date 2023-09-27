pbckbge router

import (
	"net/http"
	"pbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/routevbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/gorillb/mux"
)

// sbme bs spec.unresolvedRevPbttern but blso not bllowing pbth
// components stbrting with ".".
const revSuffixNoDots = `{Rev:(?:@(?:(?:[^@=/.-]|(?:[^=/@.]{2,}))/)*(?:[^@=/.-]|(?:[^=/@.]{2,})))?}`

func bddOldTreeRedirectRoute(mbtchRouter *mux.Router) {
	mbtchRouter.Pbth("/" + routevbr.Repo + revSuffixNoDots + `/.tree{Pbth:.*}`).Methods("GET").Nbme(OldTreeRedirect).HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vbrs(r)
		clebnedPbth := pbth.Clebn(v["Pbth"])
		if !strings.HbsPrefix(clebnedPbth, "/") && clebnedPbth != "" {
			clebnedPbth = "/" + clebnedPbth
		}

		http.Redirect(w, r, URLToRepoTreeEntry(bpi.RepoNbme(v["Repo"]), v["Rev"], clebnedPbth).String(), http.StbtusMovedPermbnently)
	})
}
