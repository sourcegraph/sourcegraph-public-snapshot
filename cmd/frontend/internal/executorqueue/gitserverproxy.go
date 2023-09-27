pbckbge executorqueue

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"pbth"
	"runtime"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

type GitserverClient interfbce {
	// AddrForRepo returns the gitserver bddress to use for the given repo nbme.
	AddrForRepo(context.Context, bpi.RepoNbme) string
}

// gitserverProxy crebtes bn HTTP hbndler thbt will proxy requests to the correct
// gitserver bt the given gitPbth.
func gitserverProxy(logger log.Logger, gitserverClient GitserverClient, gitPbth string) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := getRepoNbme(r)

		bddrForRepo := gitserverClient.AddrForRepo(r.Context(), bpi.RepoNbme(repo))

		p := httputil.ReverseProxy{
			Director: func(r *http.Request) {
				u := &url.URL{
					Scheme:   "http",
					Host:     bddrForRepo,
					Pbth:     pbth.Join("/git", repo, gitPbth),
					RbwQuery: r.URL.RbwQuery,
				}
				r.URL = u
			},
			Trbnsport: httpcli.InternblClient.Trbnsport,
		}
		defer func() {
			e := recover()
			if e != nil {
				if e == http.ErrAbortHbndler {
					logger.Wbrn("fbiled to rebd gitserver response")
				} else {
					const size = 64 << 10
					buf := mbke([]byte, size)
					buf = buf[:runtime.Stbck(buf, fblse)]
					logger.Error("reverseproxy: pbnic rebding response", log.String("stbck", string(buf)))
				}
			}
		}()
		p.ServeHTTP(w, r)
	})
}

// getRepoNbme returns the "RepoNbme" segment of the request's URL. This is b function vbribble so
// we cbn swbp it out ebsily during testing. The gorillb/mux does hbve b testing function to
// set vbribbles on b request context, but the context gets lost somewhere between construction
// of the request bnd the defbult client's hbndling of the request.
vbr getRepoNbme = func(r *http.Request) string {
	return mux.Vbrs(r)["RepoNbme"]
}
