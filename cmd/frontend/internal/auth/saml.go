package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/crewjam/saml/samlsp"
	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

// SAML App creation vars
var samlProvider = conf.AuthSAML()

// newSAMLAuthHandler wraps the passed in handler with SAML authentication, adding endpoints under the auth
// path prefix to enable the login flow an requiring login for all other endpoints.
//
// 🚨 SECURITY
func newSAMLAuthHandler(createCtx context.Context, handler http.Handler, appURL string) (http.Handler, error) {
	if samlProvider == nil {
		return nil, errors.New("No SAML ID Provider specified")
	}
	if samlProvider.ServiceProviderCertificate == "" {
		return nil, errors.New("No SAML Service Provider certificate")
	}
	if samlProvider.ServiceProviderPrivateKey == "" {
		return nil, errors.New("No SAML Service Provider private key")
	}

	entityIDURL, err := url.Parse(appURL + authURLPrefix)
	if err != nil {
		return nil, err
	}
	keyPair, err := tls.X509KeyPair([]byte(samlProvider.ServiceProviderCertificate), []byte(samlProvider.ServiceProviderPrivateKey))
	if err != nil {
		return nil, err
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, err
	}

	opt := samlsp.Options{
		URL:          *entityIDURL,
		Key:          keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:  keyPair.Leaf,
		CookieMaxAge: session.DefaultExpiryPeriod,
		CookieSecure: entityIDURL.Scheme == "https",
	}

	// Allow specifying either URL to SAML Identity Provider metadata XML file, or the XML
	// file contents directly.
	switch {
	case samlProvider.IdentityProviderMetadataURL != "" && samlProvider.IdentityProviderMetadata != "":
		return nil, errors.New("invalid SAML configuration: set either identityProviderMetadataURL or identityProviderMetadata, not both")
	case samlProvider.IdentityProviderMetadataURL != "":
		opt.IDPMetadataURL, err = url.Parse(samlProvider.IdentityProviderMetadataURL)
		if err != nil {
			return nil, errors.Wrap(err, "parsing SAML Identity Provider metadata URL")
		}
	case samlProvider.IdentityProviderMetadata != "":
		if err := xml.Unmarshal([]byte(samlProvider.IdentityProviderMetadata), &opt.IDPMetadata); err != nil {
			return nil, errors.Wrap(err, "parsing SAML Identity Provider metadata XML (note: a root element of <EntityDescriptor> is expected)")
		}
	default:
		return nil, errors.New("invalid SAML configuration: must provide the SAML metadata, using either identityProviderMetadataURL (URL where XML file is available) or identityProviderMetadata (XML file contents)")
	}

	samlSP, err := samlsp.New(opt)
	if err != nil {
		return nil, err
	}
	samlSP.CookieName = "sg-session"

	idpID := samlSP.ServiceProvider.IDPMetadata.EntityID
	authedHandler := session.SessionHeaderToCookieMiddleware(samlSP.RequireAccount(samlToActorMiddleware(handler, idpID)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle SAML ACS and metadata endpoints
		if strings.HasPrefix(r.URL.Path, authURLPrefix+"/saml/") {
			samlSP.ServeHTTP(w, r)
			return
		}
		// Handle all other endpoints
		authedHandler.ServeHTTP(w, r)
	}), nil
}

// samlToActorMiddleware translates the SAML session into an Actor and sets it in the request context
// before delegating to its child handler.
func samlToActorMiddleware(h http.Handler, idpID string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actr, err := getActorFromSAML(r, idpID)
		if err != nil {
			log15.Error("could not map SAML assertion to user", "error", err)
			http.Error(w, "could not map SAML assertion to user", http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(w, r.WithContext(actor.WithActor(r.Context(), actr)))
	})
}

// getActorFromSAML translates the SAML session into an Actor.
func getActorFromSAML(r *http.Request, idpID string) (*actor.Actor, error) {
	ctx := r.Context()
	subject := r.Header.Get("X-Saml-Subject") // this header is set by the SAML library after extracting the value from the JWT cookie
	externalID := samlToExternalID(idpID, subject)

	email := r.Header.Get("X-Saml-Email")
	if email == "" && mightBeEmail(subject) {
		email = subject
	}
	login := r.Header.Get("X-Saml-Login")
	if login == "" {
		login = r.Header.Get("X-Saml-Uid")
	}
	displayName := r.Header.Get("X-Saml-DisplayName")
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
	login, err := NormalizeUsername(login)
	if err != nil {
		return nil, err
	}

	userID, err := createOrUpdateUser(ctx, db.NewUser{
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
