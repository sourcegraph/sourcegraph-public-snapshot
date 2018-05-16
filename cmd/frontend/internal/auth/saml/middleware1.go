package saml

import (
	"net/http"
	"strings"

	"github.com/crewjam/saml/samlsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// authHandler1 is the old SAML HTTP auth handler. It is still in use (if the enhancedSAML
// experiment is disabled, which is the default).
//
// It uses github.com/crewjam/saml and does not let us use our own session cookie (it exposes only a
// middleware that handles its own cookie authentication, using JWT). This is nice for simple
// applications of SAML, but it makes it difficult to support multiple auth providers with SAML and
// expose other SAML functionality (such as token revocation).
func authHandler1(w http.ResponseWriter, r *http.Request, next http.Handler, isAPIRequest bool) {
	pc, handled := handleGetFirstProviderConfig(w)
	if handled {
		return
	}
	if handled := authHandlerCommon(w, r, next, pc); handled {
		return
	}

	// Otherwise, we need to use SAML.
	samlSP, err := cache1.get(*pc)
	if err != nil {
		log15.Error("Error getting SAML service provider.", "error", err)
		http.Error(w, "Unexpected error in SAML authentication provider.", http.StatusInternalServerError)
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
			http.Error(w, "Authentication required.", http.StatusUnauthorized)
			return
		}
	}

	// Require SAML auth to proceed (redirect to SAML login flow).
	samlSP.RequireAccount(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpID := samlSP.ServiceProvider.IDPMetadata.EntityID

		token := samlsp.Token(r.Context())
		if token == nil {
			log15.Error("No SAML token in request context.")
			http.Error(w, "Error getting SAML token.", http.StatusInternalServerError)
			return
		}

		samlActor, safeErrMsg, err := getActorFromSAML(r.Context(), token.Subject, idpID, token.Attributes)
		if err != nil {
			log15.Error("Error looking up SAML-authenticated user.", "error", err, "userErr", safeErrMsg)
			http.Error(w, safeErrMsg, http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(actor.WithActor(r.Context(), samlActor)))
	})).ServeHTTP(w, r)
}
