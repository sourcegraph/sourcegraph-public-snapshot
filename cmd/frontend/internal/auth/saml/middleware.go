package saml

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/crewjam/saml/samlsp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Middleware is middleware for SAML authentication, adding endpoints under the auth path prefix to
// enable the login flow an requiring login for all other endpoints.
//
// 🚨 SECURITY
var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHandler(w, r, next, true)
		})
	},
	App: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHandler(w, r, next, false)
		})
	},
}

func authHandler(w http.ResponseWriter, r *http.Request, next http.Handler, isAPIRequest bool) {
	// Check the SAML auth provider configuration.
	pc := conf.AuthProvider().Saml
	if pc != nil && (pc.ServiceProviderCertificate == "" || pc.ServiceProviderPrivateKey == "") {
		log15.Error("No certificate and/or private key set for SAML auth provider in site configuration.")
		http.Error(w, "misconfigured SAML auth provider", http.StatusInternalServerError)
		return
	}

	// If SAML isn't enabled, or actor is already authenticated (e.g., via access token), skip SAML auth.
	if pc == nil || actor.FromContext(r.Context()).IsAuthenticated() {
		next.ServeHTTP(w, r)
		return
	}

	// Otherwise, we need to use SAML.
	samlSP, err := cache.get(*pc)
	if err != nil {
		log15.Error("Error getting SAML service provider.", "error", err)
		http.Error(w, "unexpected error in SAML authentication provider", http.StatusInternalServerError)
		return
	}

	if !isAPIRequest {
		// Delegate to SAML ACS and metadata endpoint handlers.
		if strings.HasPrefix(r.URL.Path, auth.AuthURLPrefix+"/saml/") {
			samlSP.ServeHTTP(w, r)
			return
		}
	}

	// HACK: Return HTTP 401 for API requests without a cookie (unlike for app requests, where we
	// need to HTTP 302 redirect the user to the SAML login flow). This behavior provides no
	// additional security but is nicer for API requests.
	if isAPIRequest {
		if c, _ := r.Cookie(samlSP.ClientToken.(*samlsp.ClientCookies).Name); c == nil {
			http.Error(w, "requires authentication", http.StatusUnauthorized)
			return
		}
	}

	// Require SAML auth to proceed (redirect to SAML login flow).
	samlSP.RequireAccount(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpID := samlSP.ServiceProvider.IDPMetadata.EntityID
		samlActor, err := getActorFromSAML(r.Context(), idpID)
		if err != nil {
			log15.Error("Error looking up SAML-authenticated user.", "error", err)
			http.Error(w, "Error looking up SAML-authenticated user. "+auth.CouldNotGetUserDescription, http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(actor.WithActor(r.Context(), samlActor)))
	})).ServeHTTP(w, r)
}

// getActorFromSAML translates the SAML token's claims (set in request context by the SAML
// middleware) into an Actor.
func getActorFromSAML(ctx context.Context, idpID string) (*actor.Actor, error) {
	token := samlsp.Token(ctx)
	if token == nil {
		return nil, errors.New("no SAML token in request context")
	}

	subject := token.Subject
	externalID := samlToExternalID(idpID, subject)

	email := token.Attributes.Get("email")
	if email == "" && mightBeEmail(subject) {
		email = subject
	}
	login := token.Attributes.Get("login")
	if login == "" {
		login = token.Attributes.Get("uid")
	}
	displayName := token.Attributes.Get("displayName")
	if displayName == "" {
		displayName = token.Attributes.Get("givenName")
	}
	if displayName == "" {
		displayName = login
	}
	if displayName == "" {
		displayName = email
	}
	if displayName == "" {
		displayName = subject
	}
	if login == "" {
		login = email
	}
	if login == "" {
		return nil, fmt.Errorf("could not create user, because SAML assertion did not contain email attribute statement")
	}
	login, err := auth.NormalizeUsername(login)
	if err != nil {
		return nil, err
	}

	userID, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		ExternalProvider: idpID,
		ExternalID:       externalID,
		Username:         login,
		Email:            email,
		DisplayName:      displayName,
		// SAML has no standard way of providing an avatar URL.
	})
	if err != nil {
		return nil, err
	}
	return actor.FromUser(userID), nil
}

func samlToExternalID(idpID, subject string) string {
	return fmt.Sprintf("%s:%s", idpID, subject)
}

func mightBeEmail(s string) bool {
	return strings.Count(s, "@") == 1
}
