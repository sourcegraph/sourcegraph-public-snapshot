package store

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

// CoreStore manages the in-memory storage, access,
// and updating of the core site configuration in a threadsafe manner.
type CoreStore struct {
	configMu  sync.RWMutex
	lastValid *schema.CoreSiteConfiguration
	mock      *schema.CoreSiteConfiguration

	rawMu sync.RWMutex
	raw   string

	ready chan struct{}
	once  sync.Once
}

// NewCoreStore returns a new configuration store.
func NewCoreStore() *CoreStore {
	return &CoreStore{
		ready: make(chan struct{}),
	}
}

// LastValid returns the last valid site configuration that this
// store was updated with.
func (c *CoreStore) LastValid() *schema.CoreSiteConfiguration {
	c.WaitUntilInitialized()

	c.configMu.RLock()
	defer c.configMu.RUnlock()

	if c.mock != nil {
		return c.mock
	}

	return c.lastValid
}

// Raw returns the last raw JSON string that this store was updated with.
func (c *CoreStore) Raw() string {
	c.WaitUntilInitialized()

	c.rawMu.RLock()
	defer c.rawMu.RUnlock()
	return c.raw
}

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watcherc.
func (c *CoreStore) Mock(mockery *schema.CoreSiteConfiguration) {
	c.configMu.Lock()
	defer c.configMu.Unlock()

	c.mock = mockery
	c.initialize()
}

type CoreUpdateResult struct {
	Changed bool
	Old     *schema.CoreSiteConfiguration
	New     *schema.CoreSiteConfiguration
}

// MaybeUpdate attempts to update the store with the supplied rawConfig.
//
// If the rawConfig isn't syntactically valid JSON, the store's LastValid field.
// won't be updating and a parsing error will be returned
// from the previous time that this function was called.
//
// configChange is defined iff the cache was actually udpated.
// TODO@ggilmore: write a less-vague description
func (c *CoreStore) MaybeUpdate(rawConfig string) (CoreUpdateResult, error) {
	c.rawMu.Lock()
	defer c.rawMu.Unlock()

	c.configMu.Lock()
	defer c.configMu.Unlock()

	result := CoreUpdateResult{
		Changed: false,
		Old:     c.lastValid,
		New:     c.lastValid,
	}

	if c.raw == rawConfig {
		return result, nil
	}

	c.raw = rawConfig

	newConfig, err := conftypes.ParseCore(rawConfig)
	if err != nil {
		return result, errors.Wrap(err, "when parsing rawConfig during update")
	}

	result.Changed = true
	result.New = newConfig
	c.lastValid = newConfig

	c.initialize()

	return result, nil
}

// WaitUntilInitialized blocks and only returns to the caller once the store
// has initialized with a syntactically valid configuration file (via MaybeUpdate() or Mock()).
func (c *CoreStore) WaitUntilInitialized() {
	<-c.ready
}

func (c *CoreStore) initialize() {
	c.once.Do(func() {
		close(c.ready)
	})
}
