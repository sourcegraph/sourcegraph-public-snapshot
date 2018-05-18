package auth

import (
	"fmt"
	"sort"
	"sync"
)

var (
	allProvidersMu sync.Mutex
	allProviders   []*Provider // all configured authentication provider instances (modified by UpdateProviders)
)

// Providers returns a list of all authentication provider instances that are active in the site
// config. The return value is immutable.
func Providers() []*Provider {
	allProvidersMu.Lock()
	defer allProvidersMu.Unlock()
	return allProviders
}

// UpdateProviders updates the set of active authentication provider instances. It adds providers
// whose map value is true and removes those whose map value is false.
//
// It is generally called by site configuration listeners associated with authentication provider
// implementations after any change to the set of configured instances of that type.
func UpdateProviders(updates map[*Provider]bool) {
	allProvidersMu.Lock()
	defer allProvidersMu.Unlock()

	// Copy on write (not copy on read) because this is written rarely and read often.
	oldProviders := allProviders
	allProviders = make([]*Provider, 0, len(oldProviders))
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
		ai := allProviders[i]
		aj := allProviders[j]
		return ai.Public.DisplayName < aj.Public.DisplayName
	})
}
