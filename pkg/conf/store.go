package conf

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// configStore manages the in-memory storage, access,
// and updating of the site configuration in a threadsafe manner.
type configStore struct {
	configMu sync.RWMutex
	parsed   *schema.SiteConfiguration
	mock     *schema.SiteConfiguration

	rawMu sync.RWMutex
	raw   string

	ready chan struct{}
	once  sync.Once
}

func Store() *configStore {
	return &configStore{
		ready: make(chan struct{}),
	}
}

// WaitUntilInitialized blocks and only returns to the caller once the store
// has initialized with a syntactically valid configuration file (via MaybeUpdate() or Mock()).
func (c *configStore) WaitUntilInitialized() {
	<-c.ready
}

func (c *configStore) initialize() {
	c.once.Do(func() {
		close(c.ready)
	})
}

// Parsed returns the last valid site configuration that this
// store was updated with.
func (c *configStore) Parsed() *schema.SiteConfiguration {
	c.WaitUntilInitialized()

	c.configMu.RLock()
	defer c.configMu.RUnlock()

	if c.mock != nil {
		return c.mock
	}

	return c.parsed
}

// Raw returns the raw JSON string that this store was updated with.
func (c *configStore) Raw() string {
	c.WaitUntilInitialized()

	c.rawMu.RLock()
	defer c.rawMu.RUnlock()
	return c.raw
}

func (c *configStore) Mock(mockery *schema.SiteConfiguration) {
	c.configMu.Lock()
	defer c.configMu.Unlock()

	c.mock = mockery
	c.initialize()
}

type configChange struct {
	Changed bool
	Old     *schema.SiteConfiguration
	New     *schema.SiteConfiguration
}

// MaybeUpdate updates the store iff the supplied rawConfig differs
// from the previous time that this function was called.
//
// configChange is defined iff the cache was actually udpated.
// TODO@ggilmore: write a less-vague description
func (c *configStore) MaybeUpdate(rawConfig string) (configChange, error) {
	c.rawMu.Lock()
	defer c.rawMu.Unlock()

	c.configMu.Lock()
	defer c.configMu.Unlock()

	if c.raw == rawConfig {
		return configChange{
			Changed: false,
		}, nil
	}

	c.raw = rawConfig

	newConfig, err := parseConfig(rawConfig)
	if err != nil {
		return configChange{}, errors.Wrap(err, "when parsing rawConfig during update")
	}

	c.initialize()

	c.parsed = newConfig

	return configChange{
		Changed: true,
		Old:     c.parsed,
		New:     newConfig,
	}, nil
}
