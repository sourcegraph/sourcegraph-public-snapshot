package auth

import (
	"encoding/json"
	"sort"
	"sync"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	// curProviders is a map (label -> (config string -> Provider)). The first key is the label
	// under which the provider was registered. The second key is the normalized JSON serialization
	// of Provider.Config().
	curProviders   = map[string]map[string]Provider{}
	curProvidersMu sync.RWMutex

	MockProviders []Provider
)

// UpdateProviders updates the set of active authentication provider instances. It replaces the
// current set of Providers under the specified label with the new set.
func UpdateProviders(label string, providers []Provider) {
	curProvidersMu.Lock()
	defer curProvidersMu.Unlock()

	if providers == nil {
		delete(curProviders, label)
		return
	}

	newLabelProviders := map[string]Provider{}
	for _, p := range providers {
		k, err := json.Marshal(p.Config())
		if err != nil {
			log15.Error("Omitting auth provider, because could not JSON-marshal its config", "error", err, "configID", p.ConfigID())
			continue
		}
		newLabelProviders[string(k)] = p
	}
	curProviders[label] = newLabelProviders
}

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
	for _, lp := range curProviders {
		ct += len(lp)
	}
	providers := make([]Provider, 0, ct)
	for _, lp := range curProviders {
		for _, p := range lp {
			providers = append(providers, p)
		}
	}

	sort.Sort(sortProviders(providers))

	return providers
}

type sortProviders []Provider

func (p sortProviders) Len() int {
	return len(p)
}
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

	for _, lp := range curProviders {
		for _, p := range lp {
			if p.ConfigID() == id {
				return p
			}
		}
	}
	return nil
}
