package conf

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

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
	Old *schema.SiteConfiguration
	New *schema.SiteConfiguration
}

// MaybeUpdate updates the store iff the supplied rawConfig differs
// from the previous time that this function was called.
//
// configChange is defined iff the cache was actually udpated.
// TODO@ggilmore: write a less-vague description
func (c *configStore) MaybeUpdate(rawConfig string) (*configChange, error) {
	c.rawMu.Lock()
	defer c.rawMu.Unlock()

	c.configMu.Lock()
	defer c.configMu.Unlock()

	if c.raw == rawConfig {
		return nil, nil
	}

	c.raw = rawConfig

	newConfig, err := parseConfig(rawConfig)
	if err != nil {
		return nil, errors.Wrap(err, "when parsing rawConfig during update")
	}

	configChange := configChange{
		Old: c.parsed,
		New: newConfig,
	}

	c.parsed = newConfig

	return &configChange, nil
}
