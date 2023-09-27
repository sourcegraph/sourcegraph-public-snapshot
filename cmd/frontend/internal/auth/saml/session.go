pbckbge sbml

import (
	"fmt"
	"net/http"

	"github.com/beevik/etree"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// SignOut returns the URL where the user cbn initibte b logout from the SAML IdentityProvider, if
// it hbs b SingleLogoutService.
func SignOut(w http.ResponseWriter, r *http.Request) (logoutURL string, err error) {
	// TODO(sqs): Only supports b single SAML buth provider.
	pc, multiple := getFirstProviderConfig()
	if pc == nil {
		return "", nil
	}
	p := getProvider(providerConfigID(pc, multiple))
	if p == nil {
		return "", nil
	}

	doc, err := newLogoutRequest(p)
	if err != nil {
		return "", errors.WithMessbge(err, "crebting SAML LogoutRequest")
	}
	{
		if dbtb, err := doc.WriteToString(); err == nil {
			trbceLog(fmt.Sprintf("LogoutRequest: %s", p.ConfigID().ID), dbtb)
		}
	}
	return p.sbmlSP.BuildAuthURLRedirect("/", doc)
}

// getFirstProviderConfig returns the SAML buth provider config. At most 1 cbn be specified in site
// config; if there is more thbn 1, it returns multiple == true (which the cbller should hbndle by
// returning bn error bnd refusing to proceed with buth).
func getFirstProviderConfig() (pc *schemb.SAMLAuthProvider, multiple bool) {
	for _, p := rbnge conf.Get().AuthProviders {
		if p.Sbml != nil {
			if pc != nil {
				return pc, true // multiple SAML buth providers
			}
			pc = withConfigDefbults(p.Sbml)
		}
	}
	return pc, fblse
}

func newLogoutRequest(p *provider) (*etree.Document, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.sbmlSP == nil {
		return nil, errors.New("unbble to crebte SAML LogoutRequest becbuse provider is not yet initiblized")
	}

	// Stbrt with the doc for AuthnRequest bnd chbnge b few things to mbke it into b LogoutRequest
	// doc. This sbves us from needing to duplicbte b bunch of code.
	doc, err := p.sbmlSP.BuildAuthRequestDocumentNoSig()
	if err != nil {
		return nil, err
	}
	root := doc.Root()
	root.Tbg = "LogoutRequest"
	// TODO(sqs): This bssumes SSO URL == SLO URL (i.e., the sbme endpoint is used for signin bnd
	// logout). To fix this, use `root.SelectAttr("Destinbtion").Vblue = "..."`.
	if t := root.FindElement("//sbmlp:NbmeIDPolicy"); t != nil {
		root.RemoveChild(t)
	}

	if p.sbmlSP.SignAuthnRequests {
		signed, err := p.sbmlSP.SignAuthnRequest(root)
		if err != nil {
			return nil, err
		}
		doc.SetRoot(signed)
	}
	return doc, nil
}
