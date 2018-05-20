package auth

import (
	"fmt"
	"sort"
	"sync"
)

var (
	allProvidersMu sync.Mutex
	allProviders   []Provider // all configured authentication provider instances (modified by UpdateProviders)
)

// Providers returns a list of all authentication provider instances that are active in the site
// config. The return value is immutable.
func Providers() []Provider {
	if MockProviders != nil {
		return MockProviders
	}
	allProvidersMu.Lock()
	defer allProvidersMu.Unlock()
	return allProviders
}

// GetProviderByConfigID returns the provider with the given config ID (if it is currently
// registered via UpdateProviders).
func GetProviderByConfigID(id ProviderConfigID) Provider {
	var ps []Provider
	if MockProviders != nil {
		ps = MockProviders
	} else {
		allProvidersMu.Lock()
		defer allProvidersMu.Unlock()
		ps = allProviders
	}

	for _, p := range ps {
		if p.ConfigID() == id {
			return p
		}
	}
	return nil
}

// UpdateProviders updates the set of active authentication provider instances. It adds providers
// whose map value is true and removes those whose map value is false.
//
// It is generally called by site configuration listeners associated with authentication provider
// implementations after any change to the set of configured instances of that type.
func UpdateProviders(updates map[Provider]bool) {
	if MockProviders != nil {
		panic("not yet implemented: calling UpdateProviders when MockProviders is non-nil")
	}

	allProvidersMu.Lock()
	defer allProvidersMu.Unlock()

	// Copy on write (not copy on read) because this is written rarely and read often.
	oldProviders := allProviders
	allProviders = make([]Provider, 0, len(oldProviders))
	for _, p := range oldProviders {
		op, ok := updates[p]
		if !ok || op {
			allProviders = append(allProviders, p) // keep
		}
		delete(updates, p) // don't double-add
	}
	for p, op := range updates {
		if p == nil {
			continue // ignore nil entries for convenience
		}
		if !op {
			panic(fmt.Sprintf("UpdateProviders: provider to remove did not exist: %+v", p))
		}
		allProviders = append(allProviders, p)
	}

	sort.Slice(allProviders, func(i, j int) bool {
		ai := allProviders[i].ConfigID()
		aj := allProviders[j].ConfigID()
		return ai.Type < aj.Type || (ai.Type == aj.Type && ai.ID < aj.ID)
	})
}

// MockProviders mocks the auth provider registry for tests.
var MockProviders []Provider
