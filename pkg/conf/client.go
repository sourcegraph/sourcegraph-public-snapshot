package conf

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

type client struct {
	store       *Store
	passthrough ConfigurationSource
	watchersMu  sync.Mutex
	watchers    []chan struct{}
}

var defaultClient *client

// Raw returns a copy of the raw configuration.
func Raw() conftypes.RawUnified {
	return defaultClient.Raw()
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
// There are a select few configuration options that do restart the server, but these are the
// exception rather than the rule. In general, ANY use of configuration should
// be done in such a way that it responds to config changes while the process
// is running.
//
// Get is a wrapper around client.Get.
func Get() *Unified {
	return defaultClient.Get()
}

// Raw returns a copy of the raw configuration.
func (c *client) Raw() conftypes.RawUnified {
	return c.store.Raw()
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
// There are a select few configuration options that do restart the server but these are the
// exception rather than the rule. In general, ANY use of configuration should
// be done in such a way that it responds to config changes while the process
// is running.
func (c *client) Get() *Unified {
	return c.store.LastValid()
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
//
// GetTODO is a wrapper around client.GetTODO.
func GetTODO() *Unified {
	return defaultClient.GetTODO()
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
func (c *client) GetTODO() *Unified {
	return c.Get()
}

// Mock sets up mock data for the site configuration.
//
// Mock is a wrapper around client.Mock.
func Mock(mockery *Unified) {
	defaultClient.Mock(mockery)
}

// Mock sets up mock data for the site configuration.
func (c *client) Mock(mockery *Unified) {
	c.store.Mock(mockery)
}

// Watch calls the given function whenever the configuration has changed. The new configuration is
// accessed by calling conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
//
// Watch is a wrapper around client.Watch.
//
// IMPORTANT: Watch will block on config initialization. It therefore should *never* be called
// synchronously in `init` functions.
func Watch(f func()) {
	defaultClient.Watch(f)
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
	c.watchers = append(c.watchers, notify)
	c.watchersMu.Unlock()

	// Call the function now, to use the current configuration.
	c.store.WaitUntilInitialized()
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

func (c *client) continuouslyUpdate() {
	for {
		err := c.fetchAndUpdate()
		if err != nil {
			log.Printf("received error during background config update, err: %s", err)
		}

		jitter := time.Duration(rand.Int63n(5 * int64(time.Second)))
		time.Sleep(jitter)
	}
}

func (c *client) fetchAndUpdate() error {
	ctx := context.Background()
	var (
		newConfig conftypes.RawUnified
		err       error
	)
	if c.passthrough != nil {
		newConfig, err = c.passthrough.Read(ctx)
	} else {
		newConfig, err = api.InternalClient.Configuration(ctx)
	}
	if err != nil {
		return errors.Wrap(err, "unable to fetch new configuration")
	}

	configChange, err := c.store.MaybeUpdate(newConfig)
	if err != nil {
		return errors.Wrap(err, "unable to update new configuration")
	}

	if configChange.Changed {
		c.notifyWatchers()
	}

	return nil
}
