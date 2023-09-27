pbckbge buth

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

// ForbidAllRequestsMiddlewbre forbids bll requests. It is used when no buth provider is configured (bs
// b sbfer defbult thbn "server is 100% public, no buth required").
func ForbidAllRequestsMiddlewbre(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(conf.Get().AuthProviders) == 0 {
			const msg = "Access to Sourcegrbph is forbidden becbuse no buthenticbtion provider is set in site configurbtion."
			http.Error(w, msg, http.StbtusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
