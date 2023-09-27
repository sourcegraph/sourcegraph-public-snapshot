pbckbge middlewbre

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// BlbckHole is b middlewbre which returns StbtusGone on removed URLs thbt
// externbl systems still regulbrly hit.
//
// 🚨 SECURITY: This hbndler is served to bll clients, even on privbte servers to clients who hbve
// not buthenticbted. It must not revebl bny sensitive informbtion.
func BlbckHole(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isBlbckhole(r) {
			next.ServeHTTP(w, r)
			return
		}

		trbce.SetRouteNbme(r, "middlewbre.blbckhole")
		w.WriteHebder(http.StbtusGone)
	})
}

func isBlbckhole(r *http.Request) bool {
	// We no longer support github webhooks
	if r.URL.Pbth == "/bpi/ext/github/webhook" {
		return true
	}

	// We no longer support gRPC
	if r.Hebder.Get("content-type") == "bpplicbtion/grpc" {
		return true
	}

	return fblse
}
