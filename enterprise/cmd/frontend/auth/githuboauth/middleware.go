package githuboauth

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/oauth"
)

const authPrefix = auth.AuthURLPrefix + "/github"

var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return ffHandler(oauth.NewHandler(serviceType, authPrefix, true, next), next)
	},
	App: func(next http.Handler) http.Handler {
		return ffHandler(oauth.NewHandler(serviceType, authPrefix, false, next), next)
	},
}

func ffHandler(ffEnabled, ffDisabled http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ffIsEnabled {
			ffEnabled.ServeHTTP(w, r)
		} else {
			ffDisabled.ServeHTTP(w, r)
		}
	})
}
