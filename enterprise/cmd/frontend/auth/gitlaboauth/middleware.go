package gitlaboauth

import (
	"net/http"

	"sourcegraph.com/cmd/frontend/auth"
	"sourcegraph.com/enterprise/cmd/frontend/auth/oauth"
	"sourcegraph.com/pkg/extsvc/gitlab"
	"sourcegraph.com/schema"
)

const authPrefix = auth.AuthURLPrefix + "/gitlab"

func init() {
	oauth.AddIsOAuth(func(p schema.AuthProviders) bool {
		return p.Gitlab != nil
	})
}

var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return oauth.NewHandler(gitlab.ServiceType, authPrefix, true, next)
	},
	App: func(next http.Handler) http.Handler {
		return oauth.NewHandler(gitlab.ServiceType, authPrefix, false, next)
	},
}
