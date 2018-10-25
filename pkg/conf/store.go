package conf

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// configStore is a threadsafe
type configStore struct {
	configMu sync.RWMutex
	parsed   *schema.SiteConfiguration

	rawMu sync.RWMutex
	raw   string
}

func (c *configStore) Parsed() *schema.SiteConfiguration {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	return c.parsed
}

func (c *configStore) Raw() string {
	c.rawMu.RLock()
	defer c.rawMu.RUnlock()
	return c.raw
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

	c.parsed = newConfig

	return configChange{
		Changed: true,
		Old:     c.parsed,
		New:     newConfig,
	}, nil
}
