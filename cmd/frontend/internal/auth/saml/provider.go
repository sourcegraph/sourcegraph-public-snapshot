package saml

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/beevik/etree"
	"github.com/pkg/errors"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/net/context/ctxhttp"
)

const providerType = "saml"

type provider struct {
	config schema.SAMLAuthProvider

	mu         sync.Mutex
	samlSP     *saml2.SAMLServiceProvider
	refreshErr error
}

// ID implements auth.Provider.
func (p *provider) ID() auth.ProviderID {
	return auth.ProviderID{
		Type: providerType,
		ID:   toProviderID(&p.config).KeyString(),
	}
}

// Config implements auth.Provider.
func (p *provider) Config() schema.AuthProviders {
	return schema.AuthProviders{Saml: &p.config}
}

// Refresh implements auth.Provider.
func (p *provider) Refresh(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.samlSP, p.refreshErr = getServiceProvider(ctx, &p.config)
	return p.refreshErr
}

// CachedInfo implements auth.Provider.
func (p *provider) CachedInfo() *auth.ProviderInfo {
	info := auth.ProviderInfo{
		DisplayName: p.config.DisplayName,
		AuthenticationURL: (&url.URL{
			Path:     path.Join(auth.AuthURLPrefix, "saml", "login"),
			RawQuery: (url.Values{"p": []string{toProviderID(&p.config).KeyString()}}).Encode(),
		}).String(),
	}
	if info.DisplayName == "" {
		info.DisplayName = "SAML"
	}
	return &info
}

func getServiceProvider(ctx context.Context, pc *schema.SAMLAuthProvider) (*saml2.SAMLServiceProvider, error) {
	c, err := readProviderConfig(pc, conf.Get().AppURL)
	if err != nil {
		return nil, err
	}

	idpMetadata, err := readIdentityProviderMetadata(ctx, c)
	if err != nil {
		return nil, err
	}

	metadata, err := unmarshalEntityDescriptor(idpMetadata)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing SAML Identity Provider metadata")
	}

	idpCertStore := dsig.MemoryX509CertificateStore{Roots: []*x509.Certificate{}}
	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		for i, xcert := range kd.KeyInfo.X509Data.X509Certificates {
			if xcert.Data == "" {
				return nil, fmt.Errorf("SAML Identity Provider metadata certificate %d is empty", i)
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("decoding SAML Identity Provider metadata certificate %d", i))
			}
			idpCert, err := x509.ParseCertificate(certData)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("parsing SAML Identity Provider metadata certificate %d X.509 data", i))
			}
			idpCertStore.Roots = append(idpCertStore.Roots, idpCert)
		}
	}

	// The SP's signing and encryption keys. Some SAML IdPs
	var spKeyStore dsig.X509KeyStore
	var signRequests bool
	if c.keyPair != nil {
		spKeyStore = dsig.TLSCertKeyStore(*c.keyPair)
		signRequests = pc.SignRequests == nil || *pc.SignRequests
	} else {
		// If the SP private key isn't specified, then the IdP must not care to validate.
		spKeyStore = dsig.RandomKeyStoreForTest()
		if pc.SignRequests != nil && *pc.SignRequests {
			return nil, errors.New("signRequests is true for SAML Service Provider but no private key and cert are given")
		}
	}

	issuerURL := c.entityID.ResolveReference(&url.URL{Path: path.Join(c.entityID.Path, "/saml/metadata")}).String()
	return &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
		IdentityProviderIssuer:      metadata.EntityID,
		ServiceProviderIssuer:       issuerURL,
		AssertionConsumerServiceURL: c.entityID.ResolveReference(&url.URL{Path: path.Join(c.entityID.Path, "/saml/acs")}).String(),
		SignAuthnRequests:           signRequests,
		AudienceURI:                 issuerURL,
		IDPCertificateStore:         &idpCertStore,
		SPKeyStore:                  spKeyStore,
		NameIdFormat:                getNameIDFormat(pc),
		SkipSignatureValidation:     pc.InsecureSkipAssertionSignatureValidation,
		ValidateEncryptionCert:      true,
	}, nil
}

// entitiesDescriptor represents the SAML EntitiesDescriptor object.
type entitiesDescriptor struct {
	XMLName             xml.Name       `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntitiesDescriptor"`
	ID                  *string        `xml:",attr,omitempty"`
	ValidUntil          *time.Time     `xml:"validUntil,attr,omitempty"`
	CacheDuration       *time.Duration `xml:"cacheDuration,attr,omitempty"`
	Name                *string        `xml:",attr,omitempty"`
	Signature           *etree.Element
	EntitiesDescriptors []entitiesDescriptor     `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntitiesDescriptor"`
	EntityDescriptors   []types.EntityDescriptor `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
}

// unmarshalEntityDescriptor unmarshals from an XML root <EntityDescriptor> or <EntitiesDescriptor>
// element. If the latter, it returns the first <EntityDescriptor> child that has an
// IDPSSODescriptor.
//
// Taken from github.com/crewjam/saml.
func unmarshalEntityDescriptor(data []byte) (*types.EntityDescriptor, error) {
	var entity *types.EntityDescriptor
	if err := xml.Unmarshal(data, &entity); err != nil {
		// This comparison is ugly, but it is how the error is generated in encoding/xml.
		if err.Error() != "expected element type <EntityDescriptor> but have <EntitiesDescriptor>" {
			return nil, err
		}
		var entities *entitiesDescriptor
		if err := xml.Unmarshal(data, &entities); err != nil {
			return nil, err
		}
		for i, e := range entities.EntityDescriptors {
			if e.IDPSSODescriptor != nil {
				entity = &entities.EntityDescriptors[i]
				break
			}
		}
		if entity == nil {
			return nil, errors.New("no entity found with IDPSSODescriptor")
		}
	}
	return entity, nil
}

type providerConfig struct {
	entityID        *url.URL
	keyPair         *tls.Certificate
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
	if pc.ServiceProviderCertificate != "" && pc.ServiceProviderPrivateKey != "" {
		keyPair, err := tls.X509KeyPair([]byte(pc.ServiceProviderCertificate), []byte(pc.ServiceProviderPrivateKey))
		if err != nil {
			return nil, err
		}
		keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			return nil, err
		}
		c.keyPair = &keyPair
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
