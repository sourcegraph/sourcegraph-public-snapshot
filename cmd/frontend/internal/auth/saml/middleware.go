package saml

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Middleware is middleware for SAML authentication, adding endpoints under the auth path prefix to
// enable the login flow an requiring login for all other endpoints.
//
// 🚨 SECURITY
var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			getAuthHandler()(w, r, next, true)
		})
	},
	App: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			getAuthHandler()(w, r, next, false)
		})
	},
}

func authHandlerCommon(w http.ResponseWriter, r *http.Request, next http.Handler, pc *schema.SAMLAuthProvider) (handled bool) {
	// Check the SAML auth provider configuration.
	if pc != nil && (pc.ServiceProviderCertificate == "" || pc.ServiceProviderPrivateKey == "") {
		log15.Error("No certificate and/or private key set for SAML auth provider in site configuration.")
		http.Error(w, "misconfigured SAML auth provider", http.StatusInternalServerError)
		return true
	}

	// If SAML isn't enabled, skip SAML auth.
	if pc == nil {
		next.ServeHTTP(w, r)
		return true
	}

	return false
}

// getAuthHandler returns the auth HTTP handler to use, depending on whether the enhancedSAML
// experiment is enabled.
func getAuthHandler() func(http.ResponseWriter, *http.Request, http.Handler, bool) {
	if conf.EnhancedSAMLEnabled() {
		return authHandler2
	}
	return authHandler1
}

// getActorFromSAML translates the SAML token's claims (set in request context by the SAML
// middleware) into an Actor.
func getActorFromSAML(ctx context.Context, subjectNameID, idpID string, attr interface {
	Get(string) string
}) (_ *actor.Actor, safeErrMsg string, err error) {
	email := attr.Get("email")
	if email == "" && mightBeEmail(subjectNameID) {
		email = subjectNameID
	}
	login := attr.Get("login")
	if login == "" {
		login = attr.Get("uid")
	}
	displayName := attr.Get("displayName")
	if displayName == "" {
		displayName = attr.Get("givenName")
	}
	if displayName == "" {
		displayName = login
	}
	if displayName == "" {
		displayName = email
	}
	if displayName == "" {
		displayName = subjectNameID
	}
	if login == "" {
		login = email
	}
	if login == "" {
		return nil, "The SAML authentication provider did not contain an email attribute.", errors.New("SAML response did not contain email")
	}
	login, err = auth.NormalizeUsername(login)
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://about.sourcegraph.com/docs/config/authentication#username-normalization.", login), err
	}

	userID, safeErrMsg, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		Username:        login,
		Email:           email,
		EmailIsVerified: email != "", // TODO(sqs): https://github.com/sourcegraph/sourcegraph/issues/10118
		DisplayName:     displayName,
		// SAML has no standard way of providing an avatar URL.
	},
		db.ExternalAccountSpec{ServiceType: "saml", ServiceID: idpID, AccountID: subjectNameID},
	)
	if err != nil {
		return nil, safeErrMsg, err
	}
	return actor.FromUser(userID), "", nil
}

func mightBeEmail(s string) bool {
	return strings.Count(s, "@") == 1
}
