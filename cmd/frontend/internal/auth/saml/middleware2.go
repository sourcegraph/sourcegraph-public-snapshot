package saml

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// All SAML new implementation (2 not 1) endpoints are under this path prefix.
const authPrefix = auth.AuthURLPrefix + "/saml"

// authHandler2 is the new SAML HTTP auth handler. It is used when the enhancedSAML experiment is
// enabled.
//
// It uses github.com/russelhaering/gosaml2 and (unlike authHandler1) makes it possible to support
// multiple auth providers with SAML and expose more SAML functionality.
func authHandler2(w http.ResponseWriter, r *http.Request, next http.Handler, isAPIRequest bool) {
	// Delegate to SAML ACS and metadata endpoint handlers.
	if !isAPIRequest && strings.HasPrefix(r.URL.Path, auth.AuthURLPrefix+"/saml/") {
		samlSPHandler(w, r)
		return
	}

	// If the actor is authenticated and not performing a SAML operation, then proceed to next.
	if actor.FromContext(r.Context()).IsAuthenticated() {
		next.ServeHTTP(w, r)
		return
	}

	// If there is only one auth provider configured, the single auth provider is SAML, and it's an
	// app request, redirect to signin immediately. The user wouldn't be able to do anything else
	// anyway; there's no point in showing them a signin screen with just a single signin option.
	if ps := conf.AuthProviders(); len(ps) == 1 && ps[0].Saml != nil && !isAPIRequest {
		pc := ps[0].Saml
		sp, err := cache2.get(*pc)
		if err != nil {
			log15.Error("Error getting SAML service provider.", "error", err)
			http.Error(w, "Unexpected error in SAML authentication provider.", http.StatusInternalServerError)
			return
		}
		key := toProviderKey(pc).KeyString()
		redirectToAuthURL(w, r, key, sp, auth.SafeRedirectURL(r.URL.String()))
		return
	}

	next.ServeHTTP(w, r)
}

func samlSPHandler(w http.ResponseWriter, r *http.Request) {
	requestPath := strings.TrimPrefix(r.URL.Path, authPrefix)

	// Handle GET endpoints.
	if r.Method == "GET" {
		// All of these endpoints expect the provider key in the URL query.
		key := r.URL.Query().Get("p")
		sp, handled := handleGetProviderConfig(w, key)
		if handled {
			return
		}

		switch requestPath {
		case "/metadata":
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

		case "/login":
			// It is safe to use r.Referer() because the redirect-to URL will be checked later,
			// before the client is actually instructed to navigate there.
			redirectToAuthURL(w, r, key, sp, r.Referer())
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

	// The remaining endpoints all expect the provider key in the POST data's RelayState.
	var relayState relayState
	if err := relayState.decode(r.FormValue("RelayState")); err != nil {
		log15.Error("Error decoding SAML relay state.", "err", err)
		http.Error(w, "Error decoding SAML relay state.", http.StatusForbidden)
		return
	}
	sp, handled := handleGetProviderConfig(w, relayState.ProviderKey)
	if handled {
		return
	}

	// Handle POST endpoints.
	switch requestPath {
	case "/acs":
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

		serviceID := sp.IdentityProviderIssuer + ":" + sp.config.certFingerprint
		actor, safeErrMsg, err := getActorFromSAML(r.Context(), info.NameID, serviceID, samlAssertionValues(info.Values))
		if err != nil {
			log15.Error("Error looking up SAML-authenticated user.", "err", err, "userErr", safeErrMsg)
			http.Error(w, safeErrMsg, http.StatusInternalServerError)
			return
		}
		var exp time.Duration
		if info.SessionNotOnOrAfter != nil {
			// 🚨 SECURITY: TODO(sqs): We *should* uncomment the line below to make our own sessions
			// only last for as long as the IdP said the authn grant is active for. Unfortunately,
			// until we support refreshing SAML authn in the background
			// (https://github.com/sourcegraph/sourcegraph/issues/11340), this provides a bad user
			// experience because users need to re-authenticate via SAML every minute or so
			// (assuming their SAML IdP, like many, has a 1-minute access token validity period).
			exp = time.Until(*info.SessionNotOnOrAfter)
		}
		if err := session.SetActor(w, r, actor, exp); err != nil {
			log15.Error("Error setting SAML-authenticated actor in session.", "err", err)
			http.Error(w, "Error starting SAML-authenticated session.", http.StatusInternalServerError)
			return
		}

		// 🚨 SECURITY: Call auth.SafeRedirectURL to avoid an open-redirect vuln.
		http.Redirect(w, r, auth.SafeRedirectURL(relayState.ReturnToURL), http.StatusFound)

	case "/logout":
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

func redirectToAuthURL(w http.ResponseWriter, r *http.Request, providerKey string, sp *provider, returnToURL string) {
	authURL, err := buildAuthURLRedirect(sp, relayState{
		ProviderKey: providerKey,
		ReturnToURL: auth.SafeRedirectURL(returnToURL),
	})
	if err != nil {
		log15.Error("Failed to build SAML auth URL.", "err", err)
		http.Error(w, "Unexpected error in SAML authentication provider.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func buildAuthURLRedirect(sp *provider, relayState relayState) (string, error) {
	doc, err := sp.BuildAuthRequestDocument()
	if err != nil {
		return "", err
	}
	return sp.BuildAuthURLRedirect(relayState.encode(), doc)
}

type relayState struct {
	ProviderKey string `json:"k"`
	ReturnToURL string `json:"r"`
}

// encode returns the base64-encoded JSON representation of the relay state.
func (s *relayState) encode() string {
	b, _ := json.Marshal(s)
	return base64.StdEncoding.EncodeToString(b)
}

// Decode decodes the base64-encoded JSON representation of the relay state into the receiver.
func (s *relayState) decode(encoded string) error {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, s)
}
