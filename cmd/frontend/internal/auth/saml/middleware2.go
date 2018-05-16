package saml

import (
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// authHandler2 is the new SAML HTTP auth handler. It is used when the enhancedSAML experiment is
// enabled.
//
// It uses github.com/russelhaering/gosaml2 and (unlike authHandler1) makes it possible to support
// multiple auth providers with SAML and expose more SAML functionality.
func authHandler2(w http.ResponseWriter, r *http.Request, next http.Handler, isAPIRequest bool) {
	pc, handled := handleGetFirstProviderConfig(w)
	if handled {
		return
	}
	if handled := authHandlerCommon(w, r, next, pc); handled {
		return
	}

	// Otherwise, we need to use SAML.
	sp, err := cache2.get(*pc)
	if err != nil {
		log15.Error("Error getting SAML service provider.", "error", err)
		http.Error(w, "Unexpected error in SAML authentication provider.", http.StatusInternalServerError)
		return
	}

	// Delegate to SAML ACS and metadata endpoint handlers.
	if !isAPIRequest && strings.HasPrefix(r.URL.Path, auth.AuthURLPrefix+"/saml/") {
		samlSPHandler(w, r, sp)
		return
	}

	// Respond to unauthenticated request.
	if isAPIRequest {
		http.Error(w, "Authentication required.", http.StatusUnauthorized)
		return
	}
	// Redirect the user to where they can sign in.
	redirectToAuthURL(w, r, sp, r.URL.String())
}

func samlSPHandler(w http.ResponseWriter, r *http.Request, sp *saml2.SAMLServiceProvider) {
	requestPath := strings.TrimPrefix(r.URL.Path, auth.AuthURLPrefix)

	// Handle GET endpoints.
	if r.Method == "GET" {
		switch requestPath {
		case "/saml/metadata":
			metadata, err := sp.Metadata()
			if err != nil {
				log15.Error("Error generating SAML service provider metadata.", "err", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			buf, err := xml.MarshalIndent(metadata, "", "  ")
			if err != nil {
				log15.Error("Error encoding SAML service provider metadata.", "err", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/samlmetadata+xml; charset=utf-8")
			w.Write(buf)
			return
		}
	}

	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// Handle POST endpoints.
	switch requestPath {
	case "/saml/acs":
		info, err := sp.RetrieveAssertionInfo(r.FormValue("SAMLResponse"))
		if err != nil {
			log15.Error("Error validating SAML assertions.", "err", err)
			http.Error(w, "Error validating SAML assertions.", http.StatusForbidden)
			return
		}
		if wi := info.WarningInfo; wi.InvalidTime || wi.NotInAudience {
			log15.Error("Error validating SAML assertions", "warningInfo", wi)
			http.Error(w, "Error validating SAML assertions.", http.StatusForbidden)
			return
		}

		actor, safeErrMsg, err := getActorFromSAML(r.Context(), info.NameID, sp.IdentityProviderIssuer, samlAssertionValues(info.Values))
		if err != nil {
			log15.Error("Error looking up SAML-authenticated user.", "err", err, "userErr", safeErrMsg)
			http.Error(w, safeErrMsg, http.StatusInternalServerError)
			return
		}
		var exp time.Duration
		if info.SessionNotOnOrAfter != nil {
			exp = time.Until(*info.SessionNotOnOrAfter)
		}
		if err := session.SetActor(w, r, actor, exp); err != nil {
			log15.Error("Error setting SAML-authenticated actor in session.", "err", err)
			http.Error(w, "Error starting SAML-authenticated session.", http.StatusInternalServerError)
			return
		}

		// 🚨 SECURITY: Call auth.SafeRedirectURL to avoid an open-redirect vuln.
		http.Redirect(w, r, auth.SafeRedirectURL(r.FormValue("RelayState")), http.StatusFound)

	case "/saml/logout":
		// TODO(sqs): Fully validate the LogoutResponse here (i.e., also validate that the document
		// is a valid LogoutResponse). It is possible that this request is being spoofed, but it
		// doesn't let an attacker do very much (just log a user out and redirect).
		//
		// 🚨 SECURITY: If this logout handler starts to do anything more advanced, it probably must
		// validate the LogoutResponse to avoid being vulnerable to spoofing.
		_, err := sp.ValidateEncodedResponse(r.FormValue("SAMLResponse"))
		if err != nil && !strings.HasPrefix(err.Error(), "unable to unmarshal response:") {
			log15.Error("Error validating SAML logout response.", "err", err)
			http.Error(w, "Error validating SAML logout response.", http.StatusForbidden)
			return
		}

		// If this is an SP-initiated logout, then the actor has already been cleared from the
		// session (but there's no harm in clearing it again). If it's an IdP-initiated logout,
		// then it hasn't, and we must clear it here.
		if err := session.SetActor(w, r, nil, 0); err != nil {
			log15.Error("Error clearing actor from session in SAML logout handler.", "err", err)
			http.Error(w, "Error signing out of SAML-authenticated session.", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)

	default:
		http.Error(w, "", http.StatusNotFound)
	}
}

type samlAssertionValues saml2.Values

func (v samlAssertionValues) Get(key string) string {
	for _, a := range v {
		if a.Name == key || a.FriendlyName == key {
			return a.Values[0].Value
		}
	}
	return ""
}

func redirectToAuthURL(w http.ResponseWriter, r *http.Request, sp *saml2.SAMLServiceProvider, returnToURL string) {
	authURL, err := buildAuthURLRedirect(sp, auth.SafeRedirectURL(returnToURL))
	if err != nil {
		log15.Error("Failed to build SAML auth URL.", "err", err)
		http.Error(w, "Unexpected error in SAML authentication provider.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func buildAuthURLRedirect(sp *saml2.SAMLServiceProvider, relayState string) (string, error) {
	doc, err := sp.BuildAuthRequestDocument()
	if err != nil {
		return "", err
	}
	return sp.BuildAuthURLRedirect(relayState, doc)
}
