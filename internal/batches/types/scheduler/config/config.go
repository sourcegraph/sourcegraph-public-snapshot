package config

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/batches/types/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/schema"
)

// This is a singleton because, well, the entire site configuration system
// essentially is.
var (
	config *configuration
	mu     sync.Mutex
)

// ActiveWindow returns the window configuration in effect at the present time.
// This is not a live object, and may become outdated if held for long periods.
func ActiveWindow() *window.Configuration {
	return ensureConfig().Active()
}

// Subscribe returns a channel that will receive a message with the new
// configuration each time it is updated.
func Subscribe() chan *window.Configuration {
	return ensureConfig().Subscribe()
}

// Unsubscribe removes a channel returned from Subscribe() from the notification
// list.
func Unsubscribe(ch chan *window.Configuration) {
	ensureConfig().Unsubscribe(ch)
}

// Reset destroys the existing singleton and forces it to be reinitialised the
// next time Active() is called. This should never be used in non-testing code.
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	config = nil
}

// ensureConfig grabs the current configuration, lazily constructing it if
// necessary. It momentarily locks the singleton mutex, but releases it when it
// returns the config. This protects us against race conditions when overwriting
// the config, since Go doesn't guarantee even pointer writes are atomic, but
// doesn't provide any safety to the user. As a result, this shouldn't be used
// for anything that involves writing to the config.
func ensureConfig() *configuration {
	mu.Lock()
	defer mu.Unlock()

	if config == nil {
		config = newConfiguration()
	}
	return config
}

// configuration wraps window.Configuration in a thread-safe manner, while
// allowing consuming code to subscribe to configuration updates.
type configuration struct {
	mu          sync.RWMutex
	active      *window.Configuration
	raw         *[]*schema.BatchChangeRolloutWindow
	subscribers map[chan *window.Configuration]struct{}
}

func newConfiguration() *configuration {
	c := &configuration{subscribers: map[chan *window.Configuration]struct{}{}}

	first := true
	conf.Watch(func() {
		// Technically, if RWMutex instances could be up- and downgraded through
		// their life, we only really need a write lock briefly below when we
		// write to c.active and c.raw. However, Go's sync.RWMutex doesn't
		// provide that, so we'll just write-lock the whole time. Given there
		// shouldn't be a lot of contention around this type, that should be
		// fine.
		c.mu.Lock()
		defer c.mu.Unlock()

		incoming := conf.Get().BatchChangesRolloutWindows

		// If this isn't the first time the watcher has been called and the raw
		// configuration hasn't changed, we don't need to do anything here.
		if !first && sameConfiguration(c.raw, incoming) {
			return
		}

		cfg, err := window.NewConfiguration(incoming)
		if err != nil {
			if c.active == nil {
				log15.Warn("invalid batch changes rollout configuration detected, using the default")
			} else {
				log15.Warn("invalid batch changes rollout configuration detected, using the previous configuration")
			}
			return
		}

		// Set up the current state.
		c.active = cfg
		c.raw = incoming
		first = false

		// Notify subscribers.
		c.notify()
	})

	return c
}

func (c *configuration) Active() *window.Configuration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.active
}

func (c *configuration) Subscribe() chan *window.Configuration {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan *window.Configuration)
	config.subscribers[ch] = struct{}{}

	return ch
}

func (c *configuration) Unsubscribe(ch chan *window.Configuration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(config.subscribers, ch)
}

func (c *configuration) notify() {
	// This should only be called from functions that have already locked the
	// configuration mutex for at least read access.
	for subscriber := range c.subscribers {
		// We don't need to block on this, and we don't want any accidentally
		// closed channels to cause a panic, so we'll wrap this in
		// goroutine.Go() to fire and forget the updates.
		func(ch chan *window.Configuration, active *window.Configuration) {
			goroutine.Go(func() { ch <- active })
		}(subscriber, c.active)
	}
}

func sameConfiguration(prev, next *[]*schema.BatchChangeRolloutWindow) bool {
	// We only want to update if the actual rollout window configuration
	// changed. This is an inefficient, but effective way of figuring that out;
	// since site configurations shouldn't be changing _that_ often, the cost is
	// acceptable here.
	oldJson, err := json.Marshal(prev)
	if err != nil {
		log15.Warn("unable to marshal old batch changes rollout configuration to JSON", "err", err)
	}

	newJson, err := json.Marshal(next)
	if err != nil {
		log15.Warn("unable to marshal new batch changes rollout configuration to JSON", "err", err)
	}

	return bytes.Equal(oldJson, newJson)
}
