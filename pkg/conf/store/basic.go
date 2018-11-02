package store

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

// BasicStore manages the in-memory storage, access,
// and updating of the basic site configuration in a threadsafe manner.
type BasicStore struct {
	configMu  sync.RWMutex
	lastValid *schema.BasicSiteConfiguration
	mock      *schema.BasicSiteConfiguration

	rawMu sync.RWMutex
	raw   string

	ready chan struct{}
	once  sync.Once
}

// NewBasicStore returns a new configuration store.
func NewBasicStore() *BasicStore {
	return &BasicStore{
		ready: make(chan struct{}),
	}
}

// LastValid returns the last valid site configuration that this
// store was updated with.
func (b *BasicStore) LastValid() *schema.BasicSiteConfiguration {
	b.WaitUntilInitialized()

	b.configMu.RLock()
	defer b.configMu.RUnlock()

	if b.mock != nil {
		return b.mock
	}

	return b.lastValid
}

// Raw returns the last raw JSON string that this store was updated with.
func (b *BasicStore) Raw() string {
	b.WaitUntilInitialized()

	b.rawMu.RLock()
	defer b.rawMu.RUnlock()
	return b.raw
}

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
func (b *BasicStore) Mock(mockery *schema.BasicSiteConfiguration) {
	b.configMu.Lock()
	defer b.configMu.Unlock()

	b.mock = mockery
	b.initialize()
}

type BasicUpdateResult struct {
	Changed bool
	Old     *schema.BasicSiteConfiguration
	New     *schema.BasicSiteConfiguration
}

// MaybeUpdate attempts to update the store with the supplied rawConfig.
//
// If the rawConfig isn't syntactically valid JSON, the store's LastValid field.
// won't be updating and a parsing error will be returned
// from the previous time that this function was called.
//
// configChange is defined iff the cache was actually udpated.
// TODO@ggilmore: write a less-vague description
func (b *BasicStore) MaybeUpdate(rawConfig string) (BasicUpdateResult, error) {
	b.rawMu.Lock()
	defer b.rawMu.Unlock()

	b.configMu.Lock()
	defer b.configMu.Unlock()

	result := BasicUpdateResult{
		Changed: false,
		Old:     b.lastValid,
		New:     b.lastValid,
	}

	if b.raw == rawConfig {
		return result, nil
	}

	b.raw = rawConfig

	newConfig, err := conftypes.ParseBasic(rawConfig)
	if err != nil {
		return result, errors.Wrap(err, "when parsing rawConfig during update")
	}

	result.Changed = true
	result.New = newConfig
	b.lastValid = newConfig

	b.initialize()

	return result, nil
}

// WaitUntilInitialized blocks and only returns to the caller once the store
// has initialized with a syntactically valid configuration file (via MaybeUpdate() or Mock()).
func (b *BasicStore) WaitUntilInitialized() {
	<-b.ready
}

func (b *BasicStore) initialize() {
	b.once.Do(func() {
		close(b.ready)
	})
}
