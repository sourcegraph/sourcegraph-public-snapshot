package conf

import (
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	go defaultClient.run()
}

type client struct {
	store *configStore

	watchersMu sync.Mutex
	watchers   []chan struct{}

	// barrier to block handling requests until the
	ready chan struct{}

	mockMu      sync.RWMutex
	mockGetData *schema.SiteConfiguration
}

var defaultClient = &client{}

func (c *client) run() {
	for {
		err := c.fetchAndUpdate()
		if err == nil {
			break
		}

		log.Printf("received error during initial configuration update, err: %s", err)
		time.Sleep(1 * time.Second)
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
	return defaultClient.Get()
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

	c.mockMu.RLock()
	defer c.mockMu.RUnlock()

	if c.mockGetData != nil {
		return c.mockGetData
	}

	return c.store.Parsed()
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
//
// GetTODO is a wrapper around client.GetTODO.
func GetTODO() *schema.SiteConfiguration {
	return defaultClient.GetTODO()
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
func (c *client) GetTODO() *schema.SiteConfiguration {
	return c.Get()
}

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
//
// Mock is a wrapper around client.Mock.
func Mock(mockery *schema.SiteConfiguration) {
	defaultClient.Mock(mockery)
}

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
func (c *client) Mock(mockery *schema.SiteConfiguration) {
	c.mockMu.Lock()
	c.mockGetData = mockery
	c.mockMu.Unlock()
}

// Watch calls the given function in a separate goroutine whenever the
// configuration has changed. The new configuration can be received by calling
// conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
//
// Watch is a wrapper around client.Watch.
func Watch(f func()) {
	defaultClient.Watch(f)
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
	c.watchers = append(c.watchers, notify)
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

	configChange, err := c.store.MaybeUpdate(newRawConfig)
	if err != nil {
		return errors.Wrap(err, "unable to update new configuration")
	}

	if configChange.Changed {
		c.notifyWatchers()
	}

	return nil
}

func (c *client) fetchConfig() (string, error) {
	return "TEST", nil
}
