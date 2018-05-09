package auth

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"net/url"

	"github.com/crewjam/saml/samlsp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/schema"
)

func getSAMLServiceProvider(pc *schema.SAMLAuthProvider) (*samlsp.Middleware, error) {
	entityIDURL, err := url.Parse(globals.AppURL.String() + authURLPrefix)
	if err != nil {
		return nil, err
	}
	keyPair, err := tls.X509KeyPair([]byte(pc.ServiceProviderCertificate), []byte(pc.ServiceProviderPrivateKey))
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
	case pc.IdentityProviderMetadataURL != "" && pc.IdentityProviderMetadata != "":
		return nil, errors.New("invalid SAML configuration: set either identityProviderMetadataURL or identityProviderMetadata, not both")
	case pc.IdentityProviderMetadataURL != "":
		opt.IDPMetadataURL, err = url.Parse(pc.IdentityProviderMetadataURL)
		if err != nil {
			return nil, errors.Wrap(err, "parsing SAML Identity Provider metadata URL")
		}
	case pc.IdentityProviderMetadata != "":
		if err := xml.Unmarshal([]byte(pc.IdentityProviderMetadata), &opt.IDPMetadata); err != nil {
			return nil, errors.Wrap(err, "parsing SAML Identity Provider metadata XML (note: a root element of <EntityDescriptor> is expected)")
		}
	default:
		return nil, errors.New("invalid SAML configuration: must provide the SAML metadata, using either identityProviderMetadataURL (URL where XML file is available) or identityProviderMetadata (XML file contents)")
	}

	samlSP, err := samlsp.New(opt)
	if err != nil {
		return nil, err
	}
	samlSP.ClientToken.(*samlsp.ClientCookies).Name = "sg-session"

	return samlSP, nil
}
