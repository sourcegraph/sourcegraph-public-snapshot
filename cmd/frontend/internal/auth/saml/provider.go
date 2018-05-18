package saml

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/net/context/ctxhttp"
)

const providerType = "saml"

type providerConfig struct {
	entityID        *url.URL
	keyPair         tls.Certificate
	certFingerprint string

	// Exactly 1 of these is set:
	identityProviderMetadataURL *url.URL
	identityProviderMetadata    []byte
}

func readProviderConfig(pc *schema.SAMLAuthProvider, appURLStr string) (*providerConfig, error) {
	appURL, err := url.Parse(appURLStr)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing app URL for SAML service provider client")
	}

	var c providerConfig
	c.entityID, err = url.Parse(appURL.ResolveReference(&url.URL{Path: auth.AuthURLPrefix}).String())
	if err != nil {
		return nil, err
	}
	c.keyPair, err = tls.X509KeyPair([]byte(pc.ServiceProviderCertificate), []byte(pc.ServiceProviderPrivateKey))
	if err != nil {
		return nil, err
	}
	c.keyPair.Leaf, err = x509.ParseCertificate(c.keyPair.Certificate[0])
	if err != nil {
		return nil, err
	}
	c.certFingerprint = certFingerprint(c.keyPair.Leaf)

	// Allow specifying either URL to SAML Identity Provider metadata XML file, or the XML
	// file contents directly.
	switch {
	case pc.IdentityProviderMetadataURL != "" && pc.IdentityProviderMetadata != "":
		return nil, errors.New("invalid SAML configuration: set either identityProviderMetadataURL or identityProviderMetadata, not both")

	case pc.IdentityProviderMetadataURL != "":
		c.identityProviderMetadataURL, err = url.Parse(pc.IdentityProviderMetadataURL)
		if err != nil {
			return nil, errors.Wrap(err, "parsing SAML Identity Provider metadata URL")
		}

	case pc.IdentityProviderMetadata != "":
		c.identityProviderMetadata = []byte(pc.IdentityProviderMetadata)

	default:
		return nil, errors.New("invalid SAML configuration: must provide the SAML metadata, using either identityProviderMetadataURL (URL where XML file is available) or identityProviderMetadata (XML file contents)")
	}

	return &c, nil
}

func certFingerprint(cert *x509.Certificate) string {
	d := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return base64.RawStdEncoding.EncodeToString(d[:])
}

func readIdentityProviderMetadata(ctx context.Context, c *providerConfig) ([]byte, error) {
	if c.identityProviderMetadata != nil {
		return []byte(c.identityProviderMetadata), nil
	}

	resp, err := ctxhttp.Get(ctx, nil, c.identityProviderMetadataURL.String())
	if err != nil {
		return nil, errors.WithMessage(err, "fetching SAML Identity Provider metadata")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 HTTP response for SAML Identity Provider metadata URL: %s", c.identityProviderMetadataURL)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "reading SAML Identity Provider metadata")
	}
	return data, nil
}
