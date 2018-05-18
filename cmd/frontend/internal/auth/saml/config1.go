package saml

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// getFirstProviderConfig returns the SAML auth provider config. At most 1 can be specified in site
// config; if there is more than 1, it returns multiple == true (which the caller should handle by
// returning an error and refusing to proceed with auth).
func getFirstProviderConfig() (pc *schema.SAMLAuthProvider, multiple bool) {
	for _, p := range conf.AuthProviders() {
		if p.Saml != nil {
			if pc != nil {
				return pc, true // multiple SAML auth providers
			}
			pc = p.Saml
		}
	}
	return pc, false
}

func handleGetFirstProviderConfig(w http.ResponseWriter) (pc *schema.SAMLAuthProvider, handled bool) {
	pc, multiple := getFirstProviderConfig()
	if multiple {
		log15.Error("At most 1 SAML auth provider may be set in site config.")
		http.Error(w, "Misconfigured SAML auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	return pc, false
}
