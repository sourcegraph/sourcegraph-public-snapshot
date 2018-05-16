package openidconnect

import (
	"net/http"
	"reflect"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// getFirstProviderConfig returns the OpenID Connect auth provider config. At most 1 can be
// specified in site config; if there is more than 1, it returns multiple == true (which the caller
// should handle by returning an error and refusing to proceed with auth).
func getFirstProviderConfig() (pc *schema.OpenIDConnectAuthProvider, multiple bool) {
	for _, p := range conf.AuthProviders() {
		if p.Openidconnect != nil {
			if pc != nil {
				return pc, true // multiple OpenID Connect auth providers
			}
			pc = p.Openidconnect
		}
	}
	return pc, false
}

func handleGetFirstProviderConfig(w http.ResponseWriter) (pc *schema.OpenIDConnectAuthProvider, handled bool) {
	pc, multiple := getFirstProviderConfig()
	if multiple {
		log15.Error("At most 1 OpenID Connect auth provider may be set in site config.")
		http.Error(w, "Misconfigured OpenID Connect auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	return pc, false
}

func providersOfType(ps []schema.AuthProviders) []*schema.OpenIDConnectAuthProvider {
	var pcs []*schema.OpenIDConnectAuthProvider
	for _, p := range ps {
		if p.Openidconnect != nil {
			pcs = append(pcs, p.Openidconnect)
		}
	}
	return pcs
}

type providerKey struct{ issuerURL, clientID string }

func toProviderKey(pc *schema.OpenIDConnectAuthProvider) providerKey {
	return providerKey{issuerURL: pc.Issuer, clientID: pc.ClientID}
}

func getProviderConfig(pk providerKey) *schema.OpenIDConnectAuthProvider {
	// Assumes at most 1 auth provider can match this because pkg/conf enforces that each auth
	// provider type may appear at most once.
	for _, p := range conf.AuthProviders() {
		if pc := p.Openidconnect; pc != nil && toProviderKey(pc) == pk {
			return pc
		}
	}
	return nil
}

func toKeyMap(pcs []*schema.OpenIDConnectAuthProvider) map[providerKey]*schema.OpenIDConnectAuthProvider {
	m := make(map[providerKey]*schema.OpenIDConnectAuthProvider, len(pcs))
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

func diffProviderConfig(old, new []*schema.OpenIDConnectAuthProvider) map[schema.OpenIDConnectAuthProvider]configOp {
	oldMap := toKeyMap(old)
	diff := map[schema.OpenIDConnectAuthProvider]configOp{}
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
