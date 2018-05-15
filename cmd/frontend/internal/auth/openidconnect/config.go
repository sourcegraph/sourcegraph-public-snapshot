package openidconnect

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

type providerKey struct{ issuerURL, clientID string }

func toProviderKey(pc *schema.OpenIDConnectAuthProvider) providerKey {
	return providerKey{issuerURL: pc.Issuer, clientID: pc.ClientID}
}

func getProviderConfig(pk providerKey) *schema.OpenIDConnectAuthProvider {
	pc := conf.AuthProvider().Openidconnect
	if pc != nil && toProviderKey(pc) == pk {
		return pc
	}
	return nil
}
