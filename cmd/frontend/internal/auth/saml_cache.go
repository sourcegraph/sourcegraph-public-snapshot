package auth

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/crewjam/saml/samlsp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Start trying to populate the cache of SAML IdP metadata immediately upon server startup and site
// config changes so users don't incur the wait on the first auth flow request.
func init() {
	var (
		first = true

		samlMu sync.Mutex
		pc     *schema.SAMLAuthProvider
	)
	conf.Watch(func() {
		samlMu.Lock()
		defer samlMu.Unlock()

		// Only react when the config changes.
		newPC := conf.AuthProvider().Saml
		if reflect.DeepEqual(newPC, pc) {
			return
		}

		if first {
			log15.Info("Reloading changed SAML authentication provider configuration.")
			first = false
		}
		pc = newPC
		if pc != nil {
			go func(pc schema.SAMLAuthProvider) {
				if _, err := samlCache.get(pc); err != nil {
					log15.Error("Error prefetching SAML service provider metadata.", "error", err)
				}
			}(*pc)
		}
	})
}

// samlCache is the singleton SAML service provider metadata cache.
var samlCache samlProviderCache

// samlProviderCache caches SAML service provider metadata (which is retrieved using the site config
// for the provider).
type samlProviderCache struct {
	mu   sync.Mutex
	data map[schema.SAMLAuthProvider]*samlProviderCacheEntry // auth provider config -> entry
}

type samlProviderCacheEntry struct {
	once    sync.Once
	val     *samlsp.Middleware
	err     error
	expires time.Time
}

// get gets the SAML service provider with the specified config. If the service provider is cached,
// it returns it from the cache; otherwise it performs a network request to look up the provider. At
// most one network request will be in flight for a given provider config; later requests block on
// the original request.
func (c *samlProviderCache) get(pc schema.SAMLAuthProvider) (*samlsp.Middleware, error) {
	c.mu.Lock()
	if c.data == nil {
		c.data = map[schema.SAMLAuthProvider]*samlProviderCacheEntry{}
	}
	e, ok := c.data[pc]
	if !ok || time.Now().After(e.expires) {
		e = &samlProviderCacheEntry{}
		c.data[pc] = e
	}
	c.mu.Unlock()

	fetched := false // whether it was fetched in *this* func call
	e.once.Do(func() {
		e.val, e.err = getSAMLServiceProvider(&pc)
		e.err = errors.WithMessage(e.err, "retrieving SAML SP metadata from issuer")
		fetched = true

		var ttl time.Duration
		if e.err == nil {
			ttl = 5 * time.Minute
		} else {
			ttl = 5 * time.Second
		}
		e.expires = time.Now().Add(ttl)
	})

	err := e.err
	if !fetched {
		err = errors.WithMessage(err, "(cached error)") // make debugging easier
	}

	return e.val, err
}

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
