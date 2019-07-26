package githuboauth

import (
	"net/http"

	"sourcegraph.com/cmd/frontend/auth"
	"sourcegraph.com/enterprise/cmd/frontend/auth/oauth"
	"sourcegraph.com/pkg/extsvc/github"
	"sourcegraph.com/schema"
)

const authPrefix = auth.AuthURLPrefix + "/github"

func init() {
	oauth.AddIsOAuth(func(p schema.AuthProviders) bool {
		return p.Github != nil
	})
}

var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return oauth.NewHandler(github.ServiceType, authPrefix, true, next)
	},
	App: func(next http.Handler) http.Handler {
		return oauth.NewHandler(github.ServiceType, authPrefix, false, next)
	},
}
