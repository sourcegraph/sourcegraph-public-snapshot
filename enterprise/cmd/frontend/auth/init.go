// Package auth is imported for side-effects to enable enterprise-only SSO.
package auth

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/githuboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/gitlaboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/httpheader"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/saml"
)

func init() {
	// Register enterprise auth middleware
	auth.RegisterMiddlewares(
		openidconnect.Middleware,
		saml.Middleware,
		httpheader.Middleware,
		githuboauth.Middleware,
		gitlaboauth.Middleware,
	)
}
