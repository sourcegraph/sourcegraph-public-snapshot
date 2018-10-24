package conf

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type configCache struct {
	configMu sync.RWMutex
	parsed   *schema.SiteConfiguration

	rawMu sync.RWMutex
	raw   string
}

func (c *configCache) Parsed() *schema.SiteConfiguration {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	return c.parsed
}

func (c *configCache) Raw() string {
	c.rawMu.RLock()
	defer c.rawMu.RUnlock()
	return c.raw
}

func (c *configCache) Update(rawConfig string) error {
	c.rawMu.Lock()
	defer c.rawMu.Unlock()

	c.configMu.Lock()
	defer c.configMu.Unlock()

	c.raw = rawConfig

	tmpConfig, err := parseConfig(rawConfig)
	if err != nil {
		return errors.Wrap(err, "when parsing rawConfig during update")
	}

	c.parsed = tmpConfig
	return nil
}

// IsDirty reports whether the config has been changed since this process started.
// This can occur when config is read from a file and the file has changed on disk.
func (c *configCache) IsDirty(newRawConfig string) bool {
	c.rawMu.RLock()
	oldRawConfig := c.raw
	c.rawMu.RUnlock()

	return oldRawConfig != newRawConfig
}
