package openidconnect

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"reflect"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// getProviderConfigForKey returns the OpenID Connect auth provider config with the given key (==
// (providerKey).KeyString()).
func getProviderConfigForKey(key string) *schema.OpenIDConnectAuthProvider {
	for _, p := range conf.AuthProviders() {
		if p.Openidconnect != nil && toProviderKey(p.Openidconnect).KeyString() == key {
			return p.Openidconnect
		}
	}
	return nil
}

func handleGetProviderConfig(w http.ResponseWriter, key string) (provider *provider, handled bool) {
	pc := getProviderConfigForKey(key)
	if pc == nil {
		log15.Error("No OpenID Connect auth provider found with key.", "key", key)
		http.Error(w, "Misconfigured OpenID Connect auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	if pc.Issuer == "" {
		log15.Error("No issuer set for OpenID Connect auth provider (set the openidconnect auth provider's issuer property).", "key", key)
		http.Error(w, "Misconfigured OpenID Connect auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	provider, err := cache.get(pc.Issuer)
	if err != nil {
		log15.Error("Error getting OpenID Connect provider metadata.", "issuer", pc.Issuer, "error", err)
		http.Error(w, "Unexpected error in OpenID Connect authentication provider.", http.StatusInternalServerError)
		return
	}

	// Set config field (copying to avoid race condition).
	tmp := *provider
	tmp.config = *pc
	provider = &tmp

	return provider, false
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

func (k providerKey) KeyString() string {
	b := sha256.Sum256([]byte(strconv.Itoa(len(k.issuerURL)) + ":" + k.issuerURL + ":" + k.clientID))
	return hex.EncodeToString(b[:10])
}

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
