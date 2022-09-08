package gitlaboauth

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

const authPrefix = auth.AuthURLPrefix + "/gitlab"

func init() {
	oauth.AddIsOAuth(func(p schema.AuthProviders) bool {
		return p.Gitlab != nil
	})
}

func Middleware(logger log.Logger, db database.DB) *auth.Middleware {
	logger = logger.Scoped("gitlaboauth.middleware", "middleware that handles gitlab oauth authentication")
	return &auth.Middleware{
		API: func(next http.Handler) http.Handler {
			return oauth.NewMiddleware(logger.Scoped("api", "api handler for gitlab oauth middleware"), db, extsvc.TypeGitLab, authPrefix, true, next)
		},
		App: func(next http.Handler) http.Handler {
			return oauth.NewMiddleware(logger.Scoped("app", "app handler for gitlab oauth middleware"), db, extsvc.TypeGitLab, authPrefix, false, next)
		},
	}
}
