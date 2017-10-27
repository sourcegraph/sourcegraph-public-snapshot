package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/crewjam/saml/samlsp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	// SAML App creation vars
	samlSPCert         = env.Get("SAML_CERT", "", "SAML Service Provider certificate")
	samlSPKey          = env.Get("SAML_KEY", "", "SAML Service Provider private key")
	samlIDPMetadataURL = env.Get("SAML_ID_PROVIDER_METADATA_URL", "", "SAML Identity Provider metadata URL")

	idpMetadataURL *url.URL
)

func init() {
	var err error
	idpMetadataURL, err = url.Parse(samlIDPMetadataURL)
	if err != nil {
		log.Fatalf("Could not parse the Identity Provider metadata URL: %s", err)
	}
}

// newSAMLAuthHandler wraps the passed in handler with SAML authentication, adding endpoints under the auth
// path prefix to enable the login flow an requiring login for all other endpoints.
//
// 🚨 SECURITY
func newSAMLAuthHandler(createCtx context.Context, handler http.Handler, appURL string) (http.Handler, error) {
	if samlIDPMetadataURL == "" {
		return nil, errors.New("No SAML ID Provider specified")
	}
	if samlSPCert == "" {
		return nil, errors.New("No SAML Service Provider certificate")
	}
	if samlSPKey == "" {
		return nil, errors.New("No SAML Service Provider private key")
	}

	entityIDURL, err := url.Parse(appURL + authURLPrefix)
	if err != nil {
		return nil, err
	}
	keyPair, err := tls.X509KeyPair([]byte(samlSPCert), []byte(samlSPKey))
	if err != nil {
		return nil, err
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, err
	}

	samlSP, err := samlsp.New(samlsp.Options{
		URL:            *entityIDURL,
		Key:            keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:    keyPair.Leaf,
		IDPMetadataURL: idpMetadataURL,
	})
	if err != nil {
		return nil, err
	}

	authedHandler := samlSP.RequireAccount(handler)
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
