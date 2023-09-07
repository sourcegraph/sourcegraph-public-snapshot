// Package auth is imported for side-effects to enable enterprise-only SSO.
package auth

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/app"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/azureoauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/bitbucketcloudoauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/confauth"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/gerrit"
	githubapp "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/githubappauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/githuboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/gitlaboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/httpheader"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/saml"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/sourcegraphoperator"
	internalauth "github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Init must be called by the frontend to initialize the auth middlewares.
func Init(logger log.Logger, db database.DB) {
	logger = logger.Scoped("auth", "provides enterprise authentication middleware")
	azureoauth.Init(logger, db)
	bitbucketcloudoauth.Init(logger, db)
	gerrit.Init()
	githuboauth.Init(logger, db)
	gitlaboauth.Init(logger, db)
	httpheader.Init()
	openidconnect.Init()
	saml.Init()
	sourcegraphoperator.Init()

	// Register enterprise auth middleware
	auth.RegisterMiddlewares(
		openidconnect.Middleware(db),
		sourcegraphoperator.Middleware(db),
		saml.Middleware(db),
		httpheader.Middleware(db),
		githuboauth.Middleware(db),
		gitlaboauth.Middleware(db),
		bitbucketcloudoauth.Middleware(db),
		azureoauth.Middleware(db),
		githubapp.Middleware(db),
		confauth.Middleware(),
	)
	// Register app-level sign-out handler
	app.RegisterSSOSignOutHandler(ssoSignOutHandler)
}

func ssoSignOutHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.Scoped("ssoSignOutHandler", "Signing out from SSO providers")
	for _, p := range conf.Get().AuthProviders {
		var err error
		switch {
		case p.Openidconnect != nil:
			_, err = openidconnect.SignOut(w, r, openidconnect.SessionKey, openidconnect.GetProvider)
		case p.Saml != nil:
			_, err = saml.SignOut(w, r)
		}
		if err != nil {
			logger.Error("failed to clear auth provider session data", log.Error(err))
		}
	}

	if p := sourcegraphoperator.GetOIDCProvider(internalauth.SourcegraphOperatorProviderType); p != nil {
		_, err := openidconnect.SignOut(
			w,
			r,
			sourcegraphoperator.SessionKey,
			func(string) *openidconnect.Provider {
				return p
			},
		)
		if err != nil {
			logger.Error("failed to clear auth provider session data", log.Error(err))
		}
	}
}
