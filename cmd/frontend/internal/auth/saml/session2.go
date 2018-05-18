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
	pc, _ := getFirstProviderConfig()
	if pc == nil || !conf.EnhancedSAMLEnabled() {
		return "", nil
	}

	p, err := cache2.get(*pc)
	if err != nil {
		return "", errors.WithMessage(err, "looking up SAML provider metadata")
	}

	doc, err := newLogoutRequest(p)
	if err != nil {
		return "", errors.WithMessage(err, "creating SAML LogoutRequest")
	}
	return p.BuildAuthURLRedirect("/", doc)
}

func newLogoutRequest(p *provider) (*etree.Document, error) {
	// Start with the doc for AuthnRequest and change a few things to make it into a LogoutRequest
	// doc. This saves us from needing to duplicate a bunch of code.
	doc, err := p.BuildAuthRequestDocumentNoSig()
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

	if p.SignAuthnRequests {
		signed, err := p.SignAuthnRequest(root)
		if err != nil {
			return nil, err
		}
		doc.SetRoot(signed)
	}
	return doc, nil
}
