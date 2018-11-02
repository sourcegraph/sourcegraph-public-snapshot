package conf

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

type client struct {
	basicStore   *Store
	basicFetcher basicFetcher

	coreStore   *Store
	coreFetcher coreFetcher

	watchersMu sync.Mutex
	watchers   []chan struct{}
}

var defaultClient *client

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
func Get() *conftypes.SiteConfiguration {
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
func (c *client) Get() *conftypes.SiteConfiguration {
	// TODO@ggilmore: Figure out whether or not dealing with nulls in this way is okay
	// - this only to deal with accessing sub fields w/ struct embedding and pointers
	basic := c.basicStore.LastValid()
	if basic == nil {
		basic = &schema.BasicSiteConfiguration{}
	}

	core := c.coreStore.LastValid()
	if core == nil {
		core = &schema.CoreSiteConfiguration{}
	}
	return &conftypes.SiteConfiguration{
		BasicSiteConfiguration: *basic.(*schema.BasicSiteConfiguration),
		CoreSiteConfiguration:  *core.(*schema.CoreSiteConfiguration),
	}
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
//
// GetTODO is a wrapper around client.GetTODO.
func GetTODO() *conftypes.SiteConfiguration {
	return defaultClient.GetTODO()
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
func (c *client) GetTODO() *conftypes.SiteConfiguration {
	return c.Get()
}

// Mock sets up mock data for the site configuration.
//
// Mock is a wrapper around client.Mock.
func Mock(mockery *conftypes.SiteConfiguration) {
	defaultClient.Mock(mockery)
}

// Mock sets up mock data for the site configuration.
func (c *client) Mock(mockery *conftypes.SiteConfiguration) {
	// TODO@ggilmore: Is a nil guard here necessary?
	if mockery == nil {
		mockery = &conftypes.SiteConfiguration{}
	}

	c.basicStore.Mock(&mockery.BasicSiteConfiguration)
	c.coreStore.Mock(&mockery.CoreSiteConfiguration)
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
	// Add the watcher channel now, rather than after invoking f below, in case
	// an update were to happen while we were invoking f.
	notify := make(chan struct{}, 1)
	c.watchersMu.Lock()
	c.watchers = append(c.watchers, notify)
	c.watchersMu.Unlock()

	// Call the function now, to use the current configuration.
	c.basicStore.WaitUntilInitialized()
	c.coreStore.WaitUntilInitialized()
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
		var errs *multierror.Error

		errs = multierror.Append(errs, c.fetchAndUpdateBasic())
		errs = multierror.Append(errs, c.fetchAndUpdateCore())

		if errs.ErrorOrNil() != nil {
			log.Printf("received errors during background config updates, errs: %s", errs.ErrorOrNil())
		}

		jitter := time.Duration(rand.Int63n(5 * int64(time.Second)))
		time.Sleep(jitter)
	}
}

func (c *client) fetchAndUpdateBasic() error {
	newRawConfig, err := c.basicFetcher.FetchBasicConfig()
	if err != nil {
		return errors.Wrap(err, "unable to fetch new basic configuration")
	}

	configChange, err := c.basicStore.MaybeUpdate(newRawConfig, conftypes.ParseBasic)
	if err != nil {
		return errors.Wrap(err, "unable to update new basic configuration")
	}

	if configChange.Changed {
		c.notifyWatchers()
	}
	return nil
}

func (c *client) fetchAndUpdateCore() error {
	newRawConfig, err := c.coreFetcher.FetchCoreConfig()
	if err != nil {
		return errors.Wrap(err, "unable to fetch new core configuration")
	}

	configChange, err := c.coreStore.MaybeUpdate(newRawConfig, conftypes.ParseCore)
	if err != nil {
		return errors.Wrap(err, "unable to update new core configuration")
	}

	if configChange.Changed {
		c.notifyWatchers()
	}
	return nil
}

type basicFetcher interface {
	FetchBasicConfig() (rawJSON string, err error)
}

// Fetch the raw configuration JSON via our internal API.
type httpBasicFetcher struct{}

func (h httpBasicFetcher) FetchBasicConfig() (string, error) {
	rawJSON, err := api.InternalClient.ConfigurationBasicRawJSON(context.Background())
	return rawJSON, err
}

// Fetch the raw configuration directly via conf.DefaultServerFrontendOnly.
// This is needed by frontend, otherwise we'll run into a deadlock issue since
// frontend needs to read the site configuration before it can start serving
// the internal api.
//
// WARNING: Only frontend should use this fetcher! Other services
// that attempt to use it will panic.
type passthroughBasicFetcherFrontendOnly struct{}

func (p passthroughBasicFetcherFrontendOnly) FetchBasicConfig() (string, error) {
	return configurationServerFrontendOnly.RawBasic(), nil
}

type coreFetcher interface {
	FetchCoreConfig() (rawJSON string, err error)
}

// Fetch the raw configuration JSON via our internal API.
type httpCoreFetcher struct{}

func (h httpCoreFetcher) FetchCoreConfig() (string, error) {
	rawJSON, err := api.InternalClient.ConfigurationCoreRawJSON(context.Background())
	return rawJSON, err
}

// Fetch the raw configuration directly via conf.DefaultServerFrontendOnly.
// This is needed by frontend, otherwise we'll run into a deadlock issue since
// frontend needs to read the site configuration before it can start serving
// the internal api.
//
// WARNING: Only frontend should use this fetcher! Other services
// that attempt to use it will panic.
type passthroughCoreFetcherFrontendOnly struct{}

func (p passthroughCoreFetcherFrontendOnly) FetchCoreConfig() (string, error) {
	return configurationServerFrontendOnly.RawCore(), nil
}
