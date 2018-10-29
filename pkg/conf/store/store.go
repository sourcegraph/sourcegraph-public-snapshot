package store

import (
	"sync"

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

func New() *Store {
	return &Store{
		ready: make(chan struct{}),
	}
}

// WaitUntilInitialized blocks and only returns to the caller once the store
// has initialized with a syntactically valid configuration file (via MaybeUpdate() or Mock()).
func (s *Store) WaitUntilInitialized() {
	<-s.ready
}

func (s *Store) initialize() {
	s.once.Do(func() {
		close(s.ready)
	})
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

// Raw returns the raw JSON string that this store was updated with.
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

type configChange struct {
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
func (s *Store) MaybeUpdate(rawConfig string) (configChange, error) {
	s.rawMu.Lock()
	defer s.rawMu.Unlock()

	s.configMu.Lock()
	defer s.configMu.Unlock()

	if s.raw == rawConfig {
		return configChange{
			Changed: false,
		}, nil
	}

	s.raw = rawConfig

	newConfig, err := parse.ParseConfigEnvironment_DEPRECATED(rawConfig)
	if err != nil {
		return configChange{}, errors.Wrap(err, "when parsing rawConfig during update")
	}

	s.initialize()

	s.lastValid = newConfig

	return configChange{
		Changed: true,
		Old:     s.lastValid,
		New:     newConfig,
	}, nil
}
