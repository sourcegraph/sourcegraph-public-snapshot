package saml

import (
	"net/http"

	"github.com/beevik/etree"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// SignOut returns the URL where the user can initiate a logout from the SAML IdentityProvider, if
// it has a SingleLogoutService.
func SignOut(w http.ResponseWriter, r *http.Request) (logoutURL string, err error) {
	// TODO!(sqs): Only supports a single SAML auth provider.
	pc, _ := getFirstProviderConfig()
	if pc == nil || !conf.EnhancedSAMLEnabled() {
		return "", nil
	}
	p := getProvider(toProviderID(pc).KeyString())
	if p == nil {
		return "", nil
	}

	doc, err := newLogoutRequest(p)
	if err != nil {
		return "", errors.WithMessage(err, "creating SAML LogoutRequest")
	}
	return p.samlSP.BuildAuthURLRedirect("/", doc)
}

func newLogoutRequest(p *provider) (*etree.Document, error) {
	// Start with the doc for AuthnRequest and change a few things to make it into a LogoutRequest
	// doc. This saves us from needing to duplicate a bunch of code.
	doc, err := p.samlSP.BuildAuthRequestDocumentNoSig()
	if err != nil {
		return nil, err
	}
	root := doc.Root()
	root.Tag = "LogoutRequest"
	// TODO(sqs): This assumes SSO URL == SLO URL (i.e., the same endpoint is used for signin and
	// logout). To fix this, use `root.SelectAttr("Destination").Value = "..."`.
	if t := root.FindElement("//samlp:NameIDPolicy"); t != nil {
		root.RemoveChild(t)
	}

	if p.samlSP.SignAuthnRequests {
		signed, err := p.samlSP.SignAuthnRequest(root)
		if err != nil {
			return nil, err
		}
		doc.SetRoot(signed)
	}
	return doc, nil
}
