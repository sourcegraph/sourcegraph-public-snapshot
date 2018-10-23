package conf

import (
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	go Client.run()
}

type client struct {
	configMu sync.RWMutex
	config   *schema.SiteConfiguration

	rawMu sync.RWMutex
	raw   string

	ready chan struct{}

	watchersMu sync.Mutex
	watchers   []chan struct{}
}

var Client = &client{}

func (c *client) run() {
	err := c.fetchAndUpdate()
	if err != nil {
		return log.Fatalf("received error during initial configuration update, err: %s", err)
	}

	go func() { c.continouslyUpdate(5 * time.Second) }()
	
	close(c.ready)
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
// Get is a wrapper around client.Get.
func Get() *schema.SiteConfiguration {
	return Client.Get()
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
	<-c.ready

	c.configMu.RLock()
	defer c.configMu.RUnlock()
	if mockGetData != nil {
		return mockGetData
	}
	return c.config
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
//
// GetTODO is a wrapper around client.GetTODO.
func GetTODO() *schema.SiteConfiguration {
	return DefaultClient.GetTODO()
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
func (c *client) GetTODO() *schema.SiteConfiguration {
	return c.Get()
}

var mockGetData *schema.SiteConfiguration

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
//
// Mock is a wrapper around client.Mock.
func Mock(mockery *schema.SiteConfiguration) {
	Client.Mock(mockery)
}

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
func (c *client) Mock(mockery *schema.SiteConfiguration) {
	c.configMu.Lock()
	mockGetData = mockery
	c.configMu.Unlock()
}

// Watch calls the given function in a separate goroutine whenever the
// configuration has changed. The new configuration can be received by calling
// conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
//
// Watch is a wrapper around client.Watch.
func Watch(f func()) {
	Client.Watch(f)
}

// Watch calls the given function in a separate goroutine whenever the
// configuration has changed. The new configuration can be received by calling
// conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
func (c *client) Watch(f func()) {
	<-c.ready

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

// notifyWatchers runs all the callbacks registered via client.Watch() whenever
// the configuration has changed.
func (c *client) notifyWatchers() {
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

func (c *client) continouslyUpdate(interval time.Duration) {
	for {
		err := c.fetchAndUpdate()
		if err != nil {
			log.Printf("received error during background config update, err: %s", err)
		}

		time.Sleep(interval)
	}
}

func (c *client) fetchAndUpdate() error {
	newRawConfig, err := c.fetchConfig()
	if err != nil {
		return errors.Wrap(err, "unable to fetch new configuration")
	}

	if !c.shouldUpdate(newRawConfig) {
		return nil
	}

	err = c.updateCache(newRawConfig)
	if err != nil {
		return errors.Wrap(err, "unable to update new configuration")
	}

	c.notifyWatchers()
	return nil
}

func (c *client) fetchConfig() (string, error) {
	return "TEST", nil
}

func (c *client) updateCache(rawConfig string) error {
	c.rawMu.Lock()
	c.raw = rawConfig
	c.rawMu.Unlock()

	tmpConfig, err := parseConfig(rawConfig)
	if err != nil {
		return errors.Wrap(err, "when parsing rawConfig during update")
	}

	c.configMu.Lock()
	c.config = tmpConfig
	c.configMu.Unlock()

	return nil
}

func (c *client) shouldUpdate(newRawConfig string) bool {
	c.rawMu.RLock()
	oldRawConfig := c.raw
	c.rawMu.RUnlock()

	return oldRawConfig != newRawConfig
}
