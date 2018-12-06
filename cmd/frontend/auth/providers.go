package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

var (
	// allProvidersRegistered is closed once all providers are initially
	// registered.
	allProvidersRegisteredOnce sync.Once
	allProvidersRegistered     = make(chan struct{})

	// allProviders should not be accessed directly, use Providers() instead.
	allProvidersMu sync.RWMutex
	allProviders   []Provider
)

// Providers returns a list of all authentication provider instances that are active in the site
// config. The return value is immutable.
func Providers() []Provider {
	if MockProviders != nil {
		return MockProviders
	}

	<-allProvidersRegistered
	allProvidersMu.RLock()
	defer allProvidersMu.RUnlock()
	return allProviders
}

// GetProviderByConfigID returns the provider with the given config ID (if it is currently
// registered via UpdateProviders).
func GetProviderByConfigID(id ProviderConfigID) Provider {
	var ps []Provider
	if MockProviders != nil {
		ps = MockProviders
	} else {
		<-allProvidersRegistered
		allProvidersMu.RLock()
		defer allProvidersMu.RUnlock()
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

var isTest = (func() bool {
	path, _ := os.Executable()
	return filepath.Ext(path) == ".test"
})()

func init() {
	if isTest {
		// Tests do not usually call ConfWatch or if they do, conf.Watch is not
		// fired due to no config mocking in the tests, so we close the channel
		// now.
		allProvidersRegisteredOnce.Do(func() {
			close(allProvidersRegistered)
		})
	}
}

var needRegisteredProviders int32

// ConfWatch should be called strictly from init functions directly, not
// asynchronously.
//
// It is guaranteed that any watcher added via this method during init will run
// at least once before Providers or GetProviderByConfigID can return. This
// inherently guarantees that any auth providers registered via this function
// will be used and there is no time period between registration and the time
// when conf.Watch fires where requests could go by without authentication.
func ConfWatch(f func()) {
	atomic.AddInt32(&needRegisteredProviders, 1)

	go func() {
		init := true
		conf.Watch(func() {
			f()

			if !init {
				return
			}
			init = false
			if atomic.AddInt32(&needRegisteredProviders, -1) <= 0 {
				// Done registering providers.
				allProvidersRegisteredOnce.Do(func() {
					close(allProvidersRegistered)
				})
			}
		})
	}()
}
