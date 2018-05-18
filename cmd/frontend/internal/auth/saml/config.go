package saml

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

// getProviderConfigForKey returns the SAML auth provider config with the given key (==
// (providerKey).KeyString()).
func getProviderConfigForKey(key string) *schema.SAMLAuthProvider {
	for _, p := range conf.AuthProviders() {
		if p.Saml != nil && toProviderKey(p.Saml).KeyString() == key {
			return p.Saml
		}
	}
	return nil
}

func handleGetProviderConfig(w http.ResponseWriter, key string) (provider *provider, handled bool) {
	pc := getProviderConfigForKey(key)
	if pc == nil {
		log15.Error("No SAML auth provider found with key.", "key", key)
		http.Error(w, "Misconfigured SAML auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	provider, err := cache2.get(*pc)
	if err != nil {
		log15.Error("Error getting SAML provider metadata.", "error", err)
		http.Error(w, "Unexpected error in SAML authentication provider.", http.StatusInternalServerError)
		return
	}
	return provider, false
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

func (k providerKey) KeyString() string {
	b := sha256.Sum256([]byte(strconv.Itoa(len(k.idpMetadata)) + ":" + strconv.Itoa(len(k.idpMetadataURL)) + ":" + k.idpMetadata + ":" + k.idpMetadataURL + ":" + k.spCertificate))
	return hex.EncodeToString(b[:10])
}

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
