package conf

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf/parse"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Store manages the in-memory storage, access,
// and updating of the site configuration in a threadsafe manner.
type Store struct {
	configMu  sync.RWMutex
	lastValid *schema.SiteConfiguration
	mock      *schema.SiteConfiguration

	rawMu sync.RWMutex
	raw   string

	ready chan struct{}
	once  sync.Once
}

// NewStore returns a new configuration store.
func NewStore() *Store {
	return &Store{
		ready: make(chan struct{}),
	}
}

// LastValid returns the last valid site configuration that this
// store was updated with.
func (s *Store) LastValid() *schema.SiteConfiguration {
	s.WaitUntilInitialized()

	s.configMu.RLock()
	defer s.configMu.RUnlock()

	if s.mock != nil {
		return s.mock
	}

	return s.lastValid
}

// Raw returns the last raw JSON string that this store was updated with.
func (s *Store) Raw() string {
	s.WaitUntilInitialized()

	s.rawMu.RLock()
	defer s.rawMu.RUnlock()
	return s.raw
}

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
func (s *Store) Mock(mockery *schema.SiteConfiguration) {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	s.mock = mockery
	s.initialize()
}

type UpdateResult struct {
	Changed bool
	Old     *schema.SiteConfiguration
	New     *schema.SiteConfiguration
}

// MaybeUpdate attempts to update the store with the supplied rawConfig.
//
// If the rawConfig isn't syntactically valid JSON, the store's LastValid field.
// won't be updating and a parsing error will be returned
// from the previous time that this function was called.
//
// configChange is defined iff the cache was actually udpated.
// TODO@ggilmore: write a less-vague description
func (s *Store) MaybeUpdate(rawConfig string) (UpdateResult, error) {
	s.rawMu.Lock()
	defer s.rawMu.Unlock()

	s.configMu.Lock()
	defer s.configMu.Unlock()

	result := UpdateResult{
		Changed: false,
		Old:     s.lastValid,
		New:     s.lastValid,
	}

	if s.raw == rawConfig {
		return result, nil
	}

	s.raw = rawConfig

	newConfig, err := parse.ParseConfigEnvironment(rawConfig)
	if err != nil {
		return result, errors.Wrap(err, "when parsing rawConfig during update")
	}

	result.Changed = true
	result.New = newConfig
	s.lastValid = newConfig

	s.initialize()

	return result, nil
}

// WaitUntilInitialized blocks and only returns to the caller once the store
// has initialized with a syntactically valid configuration file (via MaybeUpdate() or Mock()).
func (s *Store) WaitUntilInitialized() {
	mode := getMode()
	if mode == modeServer {
		select {
		// Frontend has initialized its configuration server.
		case <-configurationServerFrontendOnlyInitialized:
		// We assume that we're in an unrecoverable deadlock if frontend hasn't
		// started its configuration server after 30 seconds.
		case <-time.After(30 * time.Second):
			panic("deadlock detected: you have called conf.Get or conf.Watch before the frontend has been initialized (you may need to use a goroutine)")
		}
	}

	<-s.ready
}

func (s *Store) initialize() {
	s.once.Do(func() {
		close(s.ready)
	})
}
