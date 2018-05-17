package saml

import (
	"net/http"
	"reflect"

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

func providersOfType(ps []schema.AuthProviders) []*schema.SAMLAuthProvider {
	var pcs []*schema.SAMLAuthProvider
	for _, p := range ps {
		if p.Saml != nil {
			pcs = append(pcs, p.Saml)
		}
	}
	return pcs
}

type providerKey struct{ idpMetadata, idpMetadataURL, spCertificate string }

func toProviderKey(pc *schema.SAMLAuthProvider) providerKey {
	return providerKey{
		idpMetadata:    pc.IdentityProviderMetadata,
		idpMetadataURL: pc.IdentityProviderMetadataURL,
		spCertificate:  pc.ServiceProviderCertificate,
	}
}

func toKeyMap(pcs []*schema.SAMLAuthProvider) map[providerKey]*schema.SAMLAuthProvider {
	m := make(map[providerKey]*schema.SAMLAuthProvider, len(pcs))
	for _, pc := range pcs {
		m[toProviderKey(pc)] = pc
	}
	return m
}

type configOp int

const (
	opAdded configOp = iota
	opChanged
	opRemoved
)

func diffProviderConfig(old, new []*schema.SAMLAuthProvider) map[schema.SAMLAuthProvider]configOp {
	oldMap := toKeyMap(old)
	diff := map[schema.SAMLAuthProvider]configOp{}
	for _, newPC := range new {
		newKey := toProviderKey(newPC)
		if oldPC, ok := oldMap[newKey]; ok {
			if !reflect.DeepEqual(oldPC, newPC) {
				diff[*newPC] = opChanged
			}
			delete(oldMap, newKey)
		} else {
			diff[*newPC] = opAdded
		}
	}
	for _, oldPC := range oldMap {
		diff[*oldPC] = opRemoved
	}
	return diff
}
