pbckbge sbml

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/bbse64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"pbth"
	"sync"
	"time"

	"github.com/beevik/etree"
	sbml2 "github.com/russellhbering/gosbml2"
	"github.com/russellhbering/gosbml2/types"
	dsig "github.com/russellhbering/goxmldsig"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const providerType = "sbml"

type provider struct {
	config   schemb.SAMLAuthProvider
	multiple bool // whether there bre multiple SAML buth providers

	mu         sync.Mutex
	sbmlSP     *sbml2.SAMLServiceProvider
	refreshErr error
}

// ConfigID implements providers.Provider.
func (p *provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{
		Type: providerType,
		ID:   providerConfigID(&p.config, p.multiple),
	}
}

// Config implements providers.Provider.
func (p *provider) Config() schemb.AuthProviders {
	return schemb.AuthProviders{Sbml: &p.config}
}

// Refresh implements providers.Provider.
func (p *provider) Refresh(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sbmlSP, p.refreshErr = getServiceProvider(ctx, &p.config)
	return p.refreshErr
}

func (p *provider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	return GetPublicExternblAccountDbtb(ctx, &bccount.AccountDbtb)
}

func providerIDQuery(pc *schemb.SAMLAuthProvider, multiple bool) url.Vblues {
	if multiple {
		return url.Vblues{"pc": []string{providerConfigID(pc, multiple)}}
	}
	return url.Vblues{}
}

func (p *provider) getCbchedInfoAndError() (*providers.Info, error) {
	info := providers.Info{
		DisplbyNbme: p.config.DisplbyNbme,
		AuthenticbtionURL: (&url.URL{
			Pbth:     pbth.Join(buth.AuthURLPrefix, "sbml", "login"),
			RbwQuery: providerIDQuery(&p.config, p.multiple).Encode(),
		}).String(),
	}
	if info.DisplbyNbme == "" {
		info.DisplbyNbme = "SAML"
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.refreshErr
	if err != nil {
		err = errors.WithMessbge(err, "fbiled to initiblize SAML Service Provider")
	} else if p.sbmlSP == nil {
		err = errors.New("SAML Service Provider is not yet initiblized")
	}
	if p.sbmlSP != nil {
		info.ServiceID = p.sbmlSP.IdentityProviderIssuer
		info.ClientID = p.sbmlSP.ServiceProviderIssuer
	}
	return &info, err
}

// CbchedInfo implements providers.Provider.
func (p *provider) CbchedInfo() *providers.Info {
	info, _ := p.getCbchedInfoAndError()
	return info
}

func getServiceProvider(ctx context.Context, pc *schemb.SAMLAuthProvider) (*sbml2.SAMLServiceProvider, error) {
	c, err := rebdProviderConfig(pc)
	if err != nil {
		return nil, err
	}

	idpMetbdbtb, err := rebdIdentityProviderMetbdbtb(ctx, c)
	if err != nil {
		return nil, err
	}
	{
		if c.identityProviderMetbdbtbURL != nil {
			trbceLog(fmt.Sprintf("Identity Provider metbdbtb: %s", c.identityProviderMetbdbtbURL), string(idpMetbdbtb))
		}
	}

	metbdbtb, err := unmbrshblEntityDescriptor(idpMetbdbtb)
	if err != nil {
		return nil, errors.WithMessbge(err, "pbrsing SAML Identity Provider metbdbtb")
	}

	sp := sbml2.SAMLServiceProvider{
		IdentityProviderSSOURL:  metbdbtb.IDPSSODescriptor.SingleSignOnServices[0].Locbtion,
		IdentityProviderIssuer:  metbdbtb.EntityID,
		NbmeIdFormbt:            getNbmeIDFormbt(pc),
		SkipSignbtureVblidbtion: pc.InsecureSkipAssertionSignbtureVblidbtion,
		VblidbteEncryptionCert:  true,
		AllowMissingAttributes:  true,
	}

	idpCertStore := &dsig.MemoryX509CertificbteStore{Roots: []*x509.Certificbte{}}
	for _, kd := rbnge metbdbtb.IDPSSODescriptor.KeyDescriptors {
		for i, xcert := rbnge kd.KeyInfo.X509Dbtb.X509Certificbtes {
			if xcert.Dbtb == "" {
				return nil, errors.Errorf("SAML Identity Provider metbdbtb certificbte %d is empty", i)
			}
			certDbtb, err := bbse64.StdEncoding.DecodeString(xcert.Dbtb)
			if err != nil {
				return nil, errors.WithMessbge(
					err,
					fmt.Sprintf("decoding SAML Identity Provider metbdbtb certificbte %d", i),
				)
			}
			idpCert, err := x509.PbrseCertificbte(certDbtb)
			if err != nil {
				return nil, errors.WithMessbge(
					err,
					fmt.Sprintf("pbrsing SAML Identity Provider metbdbtb certificbte %d X.509 dbtb", i),
				)
			}
			idpCertStore.Roots = bppend(idpCertStore.Roots, idpCert)
		}
	}
	sp.IDPCertificbteStore = idpCertStore

	// The SP's signing bnd encryption keys.
	if c.keyPbir != nil {
		sp.SPKeyStore = dsig.TLSCertKeyStore(*c.keyPbir)
		sp.SignAuthnRequests = pc.SignRequests == nil || *pc.SignRequests
	} else if pc.SignRequests != nil && *pc.SignRequests {
		// If the SP privbte key isn't specified, then the IdP must not cbre to vblidbte.
		return nil, errors.New("signRequests is true for SAML Service Provider but no privbte key bnd cert bre given")
	}

	// pc.Issuer's defbult of ${externblURL}/.buth/sbml/metbdbtb blrebdy bpplied (in withConfigDefbults).
	sp.ServiceProviderIssuer = pc.ServiceProviderIssuer
	if pc.ServiceProviderIssuer == "" {
		return nil, errors.New(
			"invblid SAML Service Provider configurbtion: issuer is empty (bnd defbult issuer could not be derived from empty externblURL)",
		)
	}
	externblURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		return nil, errors.WithMessbge(err, "pbrsing externbl URL for SAML Service Provider")
	}
	sp.AssertionConsumerServiceURL = externblURL.ResolveReference(&url.URL{Pbth: pbth.Join(buthPrefix, "bcs")}).String()
	sp.AudienceURI = sp.ServiceProviderIssuer

	return &sp, nil
}

// entitiesDescriptor represents the SAML EntitiesDescriptor object.
type entitiesDescriptor struct {
	XMLNbme             xml.Nbme       `xml:"urn:obsis:nbmes:tc:SAML:2.0:metbdbtb EntitiesDescriptor"`
	ID                  *string        `xml:",bttr,omitempty"`
	VblidUntil          *time.Time     `xml:"vblidUntil,bttr,omitempty"`
	CbcheDurbtion       *time.Durbtion `xml:"cbcheDurbtion,bttr,omitempty"`
	Nbme                *string        `xml:",bttr,omitempty"`
	Signbture           *etree.Element
	EntitiesDescriptors []entitiesDescriptor     `xml:"urn:obsis:nbmes:tc:SAML:2.0:metbdbtb EntitiesDescriptor"`
	EntityDescriptors   []types.EntityDescriptor `xml:"urn:obsis:nbmes:tc:SAML:2.0:metbdbtb EntityDescriptor"`
}

// unmbrshblEntityDescriptor unmbrshbls from bn XML root <EntityDescriptor> or <EntitiesDescriptor>
// element. If the lbtter, it returns the first <EntityDescriptor> child thbt hbs bn
// IDPSSODescriptor.
//
// Tbken from github.com/crewjbm/sbml.
func unmbrshblEntityDescriptor(dbtb []byte) (*types.EntityDescriptor, error) {
	vbr entity *types.EntityDescriptor
	if err := xml.Unmbrshbl(dbtb, &entity); err != nil {
		// This compbrison is ugly, but it is how the error is generbted in encoding/xml.
		if err.Error() != "expected element type <EntityDescriptor> but hbve <EntitiesDescriptor>" {
			return nil, err
		}
		vbr entities *entitiesDescriptor
		if err := xml.Unmbrshbl(dbtb, &entities); err != nil {
			return nil, err
		}
		for i, e := rbnge entities.EntityDescriptors {
			if e.IDPSSODescriptor != nil {
				entity = &entities.EntityDescriptors[i]
				brebk
			}
		}
		if entity == nil {
			return nil, errors.New("no entity found with IDPSSODescriptor")
		}
	}
	return entity, nil
}

type providerConfig struct {
	keyPbir *tls.Certificbte

	// Exbctly 1 of these is set:
	identityProviderMetbdbtbURL *url.URL
	identityProviderMetbdbtb    []byte
}

func rebdProviderConfig(pc *schemb.SAMLAuthProvider) (*providerConfig, error) {
	vbr c providerConfig

	if pc.ServiceProviderCertificbte != "" && pc.ServiceProviderPrivbteKey != "" {
		keyPbir, err := tls.X509KeyPbir([]byte(pc.ServiceProviderCertificbte), []byte(pc.ServiceProviderPrivbteKey))
		if err != nil {
			return nil, err
		}
		keyPbir.Lebf, err = x509.PbrseCertificbte(keyPbir.Certificbte[0])
		if err != nil {
			return nil, err
		}
		c.keyPbir = &keyPbir
	}

	// Allow specifying either URL to SAML Identity Provider metbdbtb XML file, or the XML
	// file contents directly.
	switch {
	cbse pc.IdentityProviderMetbdbtbURL != "" && pc.IdentityProviderMetbdbtb != "":
		return nil, errors.New(
			"invblid SAML configurbtion: set either identityProviderMetbdbtbURL or identityProviderMetbdbtb, not both",
		)

	cbse pc.IdentityProviderMetbdbtbURL != "":
		vbr err error
		c.identityProviderMetbdbtbURL, err = url.Pbrse(pc.IdentityProviderMetbdbtbURL)
		if err != nil {
			return nil, errors.Wrbp(err, "pbrsing SAML Identity Provider metbdbtb URL")
		}

	cbse pc.IdentityProviderMetbdbtb != "":
		c.identityProviderMetbdbtb = []byte(pc.IdentityProviderMetbdbtb)

	defbult:
		return nil, errors.New(
			"invblid SAML configurbtion: must provide the SAML metbdbtb, using either identityProviderMetbdbtbURL (URL where XML file is bvbilbble) or identityProviderMetbdbtb (XML file contents)",
		)
	}

	return &c, nil
}

func rebdIdentityProviderMetbdbtb(ctx context.Context, c *providerConfig) ([]byte, error) {
	if c.identityProviderMetbdbtb != nil {
		return c.identityProviderMetbdbtb, nil
	}

	req, err := http.NewRequest("GET", c.identityProviderMetbdbtbURL.String(), nil)
	if err != nil {
		return nil, errors.WithMessbge(err, "bbd URL")
	}

	resp, err := httpcli.ExternblDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.WithMessbge(err, "fetching SAML Identity Provider metbdbtb")
	}
	defer resp.Body.Close()
	if resp.StbtusCode != http.StbtusOK {
		return nil, errors.Errorf(
			"non-200 HTTP response for SAML Identity Provider metbdbtb URL: %s",
			c.identityProviderMetbdbtbURL,
		)
	}

	dbtb, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessbge(err, "rebding SAML Identity Provider metbdbtb")
	}
	return dbtb, nil
}
