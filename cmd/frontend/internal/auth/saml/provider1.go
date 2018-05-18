package saml

import (
	"context"
	"crypto/rsa"
	"encoding/xml"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

var mockGetServiceProvider1 func(*schema.SAMLAuthProvider) (*samlsp.Middleware, error)

func getServiceProvider1(ctx context.Context, pc *schema.SAMLAuthProvider) (*samlsp.Middleware, error) {
	if mockGetServiceProvider1 != nil {
		return mockGetServiceProvider1(pc)
	}

	c, err := readProviderConfig(pc, conf.Get().AppURL)
	if err != nil {
		return nil, err
	}

	opt := samlsp.Options{
		URL:          *c.entityID,
		Key:          c.keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:  c.keyPair.Leaf,
		CookieMaxAge: session.DefaultExpiryPeriod,
		CookieSecure: c.entityID.Scheme == "https",
	}

	// Do our own HTTP request instead of relying on the github.com/crewjam/saml/samlsp package, so
	// we have more control over its ctx and can share more code with the new SAML impl.
	metadataXML, err := readIdentityProviderMetadata(ctx, c)
	if err != nil {
		return nil, err
	}
	opt.IDPMetadata, err = unmarshalEntityDescriptor1(metadataXML)
	if err != nil {
		return nil, errors.Wrap(err, "parsing SAML Identity Provider metadata XML")
	}

	samlSP, err := samlsp.New(opt)
	if err != nil {
		return nil, err
	}
	samlSP.ClientToken.(*samlsp.ClientCookies).Name = "sg-session"
	samlSP.ServiceProvider.AuthnNameIDFormat = saml.NameIDFormat(getNameIDFormat(pc))

	// Cookie domains can't contain port numbers. Work around a bug in github.com/crewjam/saml where
	// it uses appURL.Host (which includes the appURL's port number, if any, thereby causing warning
	// messages like `net/http: invalid Cookie.Domain "localhost:3080"; dropping domain attribute`.
	samlSP.ClientToken.(*samlsp.ClientCookies).Domain = c.entityID.Hostname()

	return samlSP, nil
}

// unmarshalEntityDescriptor1 unmarshals from an XML root <EntityDescriptor> or <EntitiesDescriptor>
// element. If the latter, it returns the first <EntityDescriptor> child that has an
// IDPSSODescriptor.
//
// Taken from github.com/crewjam/saml.
func unmarshalEntityDescriptor1(data []byte) (*saml.EntityDescriptor, error) {
	var entity *saml.EntityDescriptor
	if err := xml.Unmarshal(data, &entity); err != nil {
		// This comparison is ugly, but it is how the error is generated in encoding/xml.
		if err.Error() != "expected element type <EntityDescriptor> but have <EntitiesDescriptor>" {
			return nil, err
		}
		var entities *saml.EntitiesDescriptor
		if err := xml.Unmarshal(data, &entities); err != nil {
			return nil, err
		}
		for i, e := range entities.EntityDescriptors {
			if len(e.IDPSSODescriptors) > 0 {
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
