package saml

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/net/context/ctxhttp"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Start trying to populate the cache of SAML IdP metadata immediately upon server startup and site
// config changes so users don't incur the wait on the first auth flow request.
func init() {
	var (
		init = true

		mu sync.Mutex
		pc *schema.SAMLAuthProvider
	)
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// Only react when the config changes.
		newPC := conf.AuthProvider().Saml
		if reflect.DeepEqual(newPC, pc) {
			return
		}

		if !init {
			log15.Info("Reloading changed SAML authentication provider configuration.")
		}
		pc = newPC
		if pc != nil {
			go func(pc schema.SAMLAuthProvider) {
				var err error
				if conf.EnhancedSAMLEnabled() {
					_, err = cache2.get(pc)
				} else {
					_, err = cache1.get(pc)
				}
				if err != nil {
					log15.Error("Error prefetching SAML service provider metadata.", "error", err)
				}
			}(*pc)
		}
	})
	init = false
}

type providerConfig struct {
	entityID *url.URL
	keyPair  tls.Certificate

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
