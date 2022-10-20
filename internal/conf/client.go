package conf

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type client struct {
	store       *store
	passthrough ConfigurationSource
	watchersMu  sync.Mutex
	watchers    []chan struct{}

	// sourceUpdates receives events that indicate the configuration source has been
	// updated. It should prompt the client to update the store, and the received channel
	// should be closed when future queries to the client returns the most up to date
	// configuration.
	sourceUpdates <-chan chan struct{}
}

var _ conftypes.UnifiedQuerier = &client{}

var (
	defaultClientOnce sync.Once
	defaultClientVal  *client
)

func DefaultClient() *client {
	defaultClientOnce.Do(func() {
		defaultClientVal = initDefaultClient()
	})
	return defaultClientVal
}

// MockClient returns a client in the same basic configuration as the DefaultClient, but is not limited to a global singleton.
// This is useful to mock configuration in tests without race conditions modifying values when running tests in parallel.
func MockClient() *client {
	return &client{store: newStore()}
}

// Raw returns a copy of the raw configuration.
func Raw() conftypes.RawUnified {
	return DefaultClient().Raw()
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
	return DefaultClient().Get()
}

func SiteConfig() schema.SiteConfiguration {
	return Get().SiteConfiguration
}

func ServiceConnections() conftypes.ServiceConnections {
	return Get().ServiceConnections()
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

func (c *client) SiteConfig() schema.SiteConfiguration {
	return c.Get().SiteConfiguration
}

func (c *client) ServiceConnections() conftypes.ServiceConnections {
	return c.Get().ServiceConnections()
}

// Mock sets up mock data for the site configuration.
//
// Mock is a wrapper around client.Mock.
func Mock(mockery *Unified) {
	DefaultClient().Mock(mockery)
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
	DefaultClient().Watch(f)
}

// Cached will return a wrapper around f which caches the response. The value
// will be recomputed every time the config is updated.
//
// IMPORTANT: The first call to wrapped will block on config initialization.  It will also create a
// long lived goroutine when DefaultClient().Cached is invoked. As a result it's important to NEVER
// call it inside a function to avoid unbounded goroutines that never return.
func Cached[T any](f func() T) (wrapped func() T) {
	g := func() any {
		return f()
	}
	h := DefaultClient().Cached(g)
	return func() T {
		return h().(T)
	}
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

// Cached will return a wrapper around f which caches the response. The value
// will be recomputed every time the config is updated.
//
// The first call to wrapped will block on config initialization.
func (c *client) Cached(f func() any) (wrapped func() any) {
	var once sync.Once
	var val atomic.Value
	return func() any {
		once.Do(func() {
			c.Watch(func() {
				val.Store(f())
			})
		})
		return val.Load()
	}
}

// notifyWatchers runs all the callbacks registered via client.Watch() whenever
// the configuration has changed. It does not block on individual sends.
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

type continuousUpdateOptions struct {
	// delayBeforeUnreachableLog is how long to wait before logging an error upon initial startup
	// due to the frontend being unreachable. It is used to avoid log spam when other services (that
	// contact the frontend for configuration) start up before the frontend.
	delayBeforeUnreachableLog time.Duration

	logger              log.Logger
	sleepBetweenUpdates func() // sleep between updates
}

// continuouslyUpdate runs (*client).fetchAndUpdate in an infinite loop, with error logging and
// random sleep intervals.
//
// The optOnlySetByTests parameter is ONLY customized by tests. All callers in main code should pass
// nil (so that the same defaults are used).
func (c *client) continuouslyUpdate(optOnlySetByTests *continuousUpdateOptions) {
	opts := optOnlySetByTests
	if opts == nil {
		// Apply defaults.
		opts = &continuousUpdateOptions{
			// This needs to be long enough to allow the frontend to fully migrate the PostgreSQL
			// database in most cases, to avoid log spam when running sourcegraph/server for the
			// first time.
			delayBeforeUnreachableLog: 15 * time.Second,
			logger:                    log.Scoped("conf.client", "configuration client"),
			sleepBetweenUpdates: func() {
				jitter := time.Duration(rand.Int63n(5 * int64(time.Second)))
				time.Sleep(jitter)
			},
		}
	}

	isFrontendUnreachableError := func(err error) bool {
		var e *net.OpError
		return errors.As(err, &e) && e.Op == "dial"
	}

	waitForSleep := func() <-chan struct{} {
		c := make(chan struct{}, 1)
		go func() {
			opts.sleepBetweenUpdates()
			close(c)
		}()
		return c
	}

	// Make an initial fetch an update - this is likely to error, so just discard the
	// error on this initial attempt.
	_ = c.fetchAndUpdate(opts.logger)

	start := time.Now()
	for {
		logger := opts.logger

		// signalDoneReading, if set, indicates that we were prompted to update because
		// the source has been updated.
		var signalDoneReading chan struct{}
		select {
		case signalDoneReading = <-c.sourceUpdates:
			// Config was changed at source, so let's check now
			logger = logger.With(log.String("triggered_by", "sourceUpdates"))
		case <-waitForSleep():
			// File possibly changed at source, so check now.
			logger = logger.With(log.String("triggered_by", "waitForSleep"))
		}

		logger.Debug("checking for updates")
		err := c.fetchAndUpdate(logger)
		if err != nil {
			// Suppress log messages for errors caused by the frontend being unreachable until we've
			// given the frontend enough time to initialize (in case other services start up before
			// the frontend), to reduce log spam.
			if time.Since(start) > opts.delayBeforeUnreachableLog || !isFrontendUnreachableError(err) {
				logger.Error("received error during background config update", log.Error(err))
			}
		} else {
			// We successfully fetched the config, we reset the timer to give
			// frontend time if it needs to restart
			start = time.Now()
		}

		// Indicate that we are done reading, if we were prompted to update by the updates
		// channel
		if signalDoneReading != nil {
			close(signalDoneReading)
		}
	}
}

func (c *client) fetchAndUpdate(logger log.Logger) error {
	var (
		ctx       = context.Background()
		newConfig conftypes.RawUnified
		err       error
	)
	if c.passthrough != nil {
		newConfig, err = c.passthrough.Read(ctx)
	} else {
		newConfig, err = internalapi.Client.Configuration(ctx)
	}
	if err != nil {
		return errors.Wrap(err, "unable to fetch new configuration")
	}

	configChange, err := c.store.MaybeUpdate(newConfig)
	if err != nil {
		return errors.Wrap(err, "unable to update new configuration")
	}

	if configChange.Changed {
		logger.Info("config changed, notifying watchers",
			log.Int("watchers", len(c.watchers)))
		c.notifyWatchers()
	} else {
		logger.Debug("no config changes detected")
	}

	return nil
}
