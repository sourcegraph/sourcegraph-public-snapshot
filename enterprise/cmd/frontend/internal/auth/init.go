// Package auth is imported for side-effects to enable enterprise-only SSO.
package auth

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/app"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/githuboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/gitlaboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/httpheader"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/saml"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Init must be called by the frontend to initialize the auth middlewares.
func Init(db database.DB) {
	openidconnect.Init()
	saml.Init()
	httpheader.Init()
	githuboauth.Init(db)
	gitlaboauth.Init(db)

	// Register enterprise auth middleware
	auth.RegisterMiddlewares(
		openidconnect.Middleware(db),
		saml.Middleware(db),
		httpheader.Middleware(db),
		githuboauth.Middleware(db),
		gitlaboauth.Middleware(db),
	)
	// Register app-level sign-out handler
	app.RegisterSSOSignOutHandler(ssoSignOutHandler)

	// Warn about usage of auth providers that are not enabled by the license.
	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		// Only site admins can act on this alert, so only show it to site admins.
		if !args.IsSiteAdmin {
			return nil
		}

		if licensing.IsFeatureEnabledLenient(licensing.FeatureSSO) {
			return nil
		}

		collected := make(map[string]struct{})
		var names []string
		for _, p := range conf.Get().AuthProviders {
			// Only built-in authentication provider is allowed by default.
			if p.Builtin != nil {
				continue
			}

			var name string
			switch {
			case p.Github != nil:
				name = "GitHub OAuth"
			case p.Gitlab != nil:
				name = "GitLab OAuth"
			case p.HttpHeader != nil:
				name = "HTTP header"
			case p.Openidconnect != nil:
				name = "OpenID Connect"
			case p.Saml != nil:
				name = "SAML"
			default:
				name = "Other"
			}

			if _, ok := collected[name]; !ok {
				collected[name] = struct{}{}
				names = append(names, name)
			}
		}
		if len(names) == 0 {
			return nil
		}

		sort.Strings(names)
		return []*graphqlbackend.Alert{{
			TypeValue:    graphqlbackend.AlertTypeError,
			MessageValue: fmt.Sprintf("A Sourcegraph license is required to enable following authentication providers: %s. [**Get a license.**](/site-admin/license)", strings.Join(names, ", ")),
		}}
	})
}

func ssoSignOutHandler(w http.ResponseWriter, r *http.Request) {
	for _, p := range conf.Get().AuthProviders {
		var err error
		switch {
		case p.Openidconnect != nil:
			_, err = openidconnect.SignOut(w, r)
		case p.Saml != nil:
			_, err = saml.SignOut(w, r)
		}
		if err != nil {
			log15.Error("Error clearing auth provider session data.", "err", err)
		}
	}
}
