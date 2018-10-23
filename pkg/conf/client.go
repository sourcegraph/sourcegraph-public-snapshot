package conf

import (
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type client struct {
	configMu sync.RWMutex
	config   *schema.SiteConfiguration

	rawMu sync.RWMutex
	raw   string

	watchersMu sync.Mutex
	watchers   []chan struct{}
}

var DefaultClient = Client()

func Client() *client {
	return &client{}
}

// Get returns a copy of the configuration. The returned value should NEVER be
// modified.
//
// Important: The configuration can change while the process is running! Code
// should only call this in response to conf.Watch OR it should invoke it
// periodically or in direct response to a user action (e.g. inside an HTTP
// handler) to ensure it responds to configuration changes while the process
// is running.
//
// There are a select few configuration options that do restart the server (for
// example, TLS or which port the frontend listens on) but these are the
// exception rather than the rule. In general, ANY use of configuration should
// be done in such a way that it responds to config changes while the process
// is running.
//
// Get is a wrapper around client.Get().
func Get() *schema.SiteConfiguration {
	return DefaultClient.Get()
}

// Get returns a copy of the configuration. The returned value should NEVER be
// modified.
//
// Important: The configuration can change while the process is running! Code
// should only call this in response to conf.Watch OR it should invoke it
// periodically or in direct response to a user action (e.g. inside an HTTP
// handler) to ensure it responds to configuration changes while the process
// is running.
//
// There are a select few configuration options that do restart the server (for
// example, TLS or which port the frontend listens on) but these are the
// exception rather than the rule. In general, ANY use of configuration should
// be done in such a way that it responds to config changes while the process
// is running.
func (c *client) Get() *schema.SiteConfiguration {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	if mockGetData != nil {
		return mockGetData
	}
	return c.config
}

var mockGetData *schema.SiteConfiguration

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
func Mock(mockery *schema.SiteConfiguration) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	mockGetData = mockery
}

// Watch calls the given function in a separate goroutine whenever the
// configuration has changed. The new configuration can be received by calling
// conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
//
// Get is a wrapper around client.Get().
func Watch(f func()) {
	DefaultClient.Watch(f)
}

// Watch calls the given function in a separate goroutine whenever the
// configuration has changed. The new configuration can be received by calling
// conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
func (c *client) Watch(f func()) {
	// Add the watcher channel now, rather than after invoking f below, in case
	// an update were to happen while we were invoking f.
	notify := make(chan struct{}, 1)
	c.watchersMu.Lock()
	c.watchers = append(watchers, notify)
	c.watchersMu.Unlock()

	// Call the function now, to use the current configuration.
	f()

	go func() {
		// Invoke f when the configuration has changed.
		for {
			<-notify
			f()
		}
	}()
}

func (c *client) continouslyUpdate() {
	go func() {
		for {
			time.Sleep(5 * time.Second)

			newConfig, err := c.fetchConfig()
			if err != nil {
				log.Printf("unable to fetch new configuration, err: %s", err)
				continue
			}

			if !c.shouldUpdate(newConfig) {
				continue
			}

			err = c.update(newConfig)
			if err != nil {
				log.Printf("unable to update new configuration, err: %s", err)
				continue
			}

			c.watchersMu.Lock()
			defer c.watchersMu.Unlock()

			for _, watcher := range c.watchers {
				// Perform a non-blocking send.
				//
				// Since the watcher channels that we are sending on have a
				// buffer of 1, it is guaranteed the watcher will
				// reconsider the config at some point in the future even
				// if this send fails.
				select {
				case watcher <- struct{}{}:
				default:
				}
			}
		}
	}()
}

func (c *client) update(rawConfig string) error {
	rawMu.Lock()
	c.raw = rawConfig
	rawMu.Unlock()

	tmpConfig, err := parseConfig(rawConfig)
	if err != nil {
		return errors.Wrap(err, "when parsing rawConfig during update")
	}

	c.configMu.Lock()
	defer c.configMu.Unlock()
	c.config = tmpConfig

	return nil
}

func (c *client) fetchConfig() (string, error) {
	return "TEST", nil
}

func (c *client) shouldUpdate(newRawConfig string) bool {
	c.rawMu.Lock()
	oldRawConfig := c.raw
	c.rawMu.Unlock()
	return oldRawConfig != newRawConfig
}
