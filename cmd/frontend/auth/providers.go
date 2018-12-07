package auth

import (
	"encoding/json"
	"sort"
	"sync"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	// curProviders is a map (package name -> (config string -> Provider)). The first key is the
	// package name under which the provider was registered (this should be unique among
	// packages). The second key is the normalized JSON serialization of Provider.Config().  We keep
	// track of providers by package, so that when a given package updates its set of registered
	// providers, we can easily remove its providers that are no longer present.
	curProviders   = map[string]map[string]Provider{}
	curProvidersMu sync.RWMutex

	MockProviders []Provider
)

// UpdateProviders updates the set of active authentication provider instances. It replaces the
// current set of Providers under the specified pkgName with the new set.
func UpdateProviders(pkgName string, providers []Provider) {
	curProvidersMu.Lock()
	defer curProvidersMu.Unlock()

	if providers == nil {
		delete(curProviders, pkgName)
		return
	}

	newPkgProviders := map[string]Provider{}
	for _, p := range providers {
		k, err := json.Marshal(p.Config())
		if err != nil {
			log15.Error("Omitting auth provider (failed to marshal its JSON config)", "error", err, "configID", p.ConfigID())
			continue
		}
		newPkgProviders[string(k)] = p
	}
	curProviders[pkgName] = newPkgProviders
}

// Providers returns the set of currently registered authentication providers. When no providers are
// registered, returns nil (and sign-in is effectively disabled).
func Providers() []Provider {
	if MockProviders != nil {
		return MockProviders
	}

	curProvidersMu.RLock()
	defer curProvidersMu.RUnlock()

	if curProviders == nil {
		return nil
	}

	ct := 0
	for _, pkgProviders := range curProviders {
		ct += len(pkgProviders)
	}
	providers := make([]Provider, 0, ct)
	for _, pkgProviders := range curProviders {
		for _, p := range pkgProviders {
			providers = append(providers, p)
		}
	}

	// Sort the providers to ensure a stable ordering (this is for the UI display order).
	sort.Sort(sortProviders(providers))

	return providers
}

type sortProviders []Provider

func (p sortProviders) Len() int {
	return len(p)
}

// Less puts the builtin provider first and sorts the others alphabetically by type and then ID.
func (p sortProviders) Less(i, j int) bool {
	if p[i].ConfigID().Type == "builtin" && p[j].ConfigID().Type != "builtin" {
		return true
	}
	if p[j].ConfigID().Type == "builtin" && p[i].ConfigID().Type != "builtin" {
		return false
	}
	if p[i].ConfigID().Type != p[j].ConfigID().Type {
		return p[i].ConfigID().Type < p[j].ConfigID().Type
	}
	return p[i].ConfigID().ID < p[j].ConfigID().ID
}
func (p sortProviders) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func GetProviderByConfigID(id ProviderConfigID) Provider {
	if MockProviders != nil {
		for _, p := range MockProviders {
			if p.ConfigID() == id {
				return p
			}
		}
		return nil
	}

	curProvidersMu.RLock()
	defer curProvidersMu.RUnlock()

	for _, pkgProviders := range curProviders {
		for _, p := range pkgProviders {
			if p.ConfigID() == id {
				return p
			}
		}
	}
	return nil
}
