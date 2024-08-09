package modelconfig

import (
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Service is the system-wide component for obtaining the set of
// LLM models the current Sourcegraph instance is configured to use.
//
// You can obtain the global, package-level instance of this interface by
// calling the Get() method. It is safe for concurrent reads.
//
// Updating the system's model configuration is done by updating the Site
// Configuration. The implementation of the Service will listen
// for configuration changes and update the ModelConfiguration object in-
// memory as appropriate.
type Service interface {
	// Get returns the current model configuration for this Sourcegraph instance.
	// Callers should not modify the returned data, and treat it as if it were
	// immutable.
	Get() (*types.ModelConfiguration, error)
}

// Global instance of the model config service. We don't initialize via
// a sync.Once because we ensure Init cannot be called twice, and assume it
// will not be called concurrently.
var singletonConfigService *service

// Get returns the singleton ModelConfigService.
//
// This requires that the Init function has been called before hand, which
// is typically done on application startup.
func Get() Service {
	if singletonConfigService == nil {
		// Return a valid Service, but it'll just return an error if you
		// ever try to use it.
		return &failingSvc{
			message: "ModelConfigService not initialized. Init not called.",
		}
	}
	return singletonConfigService
}

// InitMock initializes the global modelconfig service for use in unit tests.
// Will fail if Init() has already been called.
func InitMock() error {
	// If the singletons ervice has already been initialized, fail if it wasn't
	// already done by another unit test. We don't want to mix tests and "real"
	// behaviors.
	if singletonConfigService != nil && !singletonConfigService.inTestMode {
		return errors.New("service already initialized via Init")
	}
	if singletonConfigService == nil {
		singletonConfigService = &service{
			inTestMode: true,
		}
	}
	return nil
}

// ResetMock will reset the mock modelconfig service to pick up any recent
// site config changes.
func ResetMock() error {
	// We intentionally do not provide static or Cody Gateway data so tests
	// can be deterministic.
	return ResetMockWithStaticData(nil)
}

func ResetMockWithStaticData(staticConfig *types.ModelConfiguration) error {
	if singletonConfigService == nil {
		return errors.New("service not configured, call InitMock")
	}
	if !singletonConfigService.inTestMode {
		return errors.New("service not in test mode, refusing to reset mock")
	}

	// Load the latest site config.
	logger := log.Scoped("modelconfigResetMock")
	siteConfig := conf.Get().SiteConfiguration
	siteModelConfig, err := maybeGetSiteModelConfiguration(logger, siteConfig)
	if err != nil {
		return errors.Wrap(err, "converting completion config")
	}

	// Rebuild the modelconfig data.
	b := builder{
		staticData:      staticConfig,
		codyGatewayData: nil,
		siteConfigData:  siteModelConfig,
	}
	newConfig, err := b.build()
	if err != nil {
		return errors.Wrap(err, "building modelconfig")
	}
	singletonConfigService.set(newConfig)
	return nil
}

// service implements the Service interface, and exposes a thread-safe `set` method
// for updating the current configuration.
type service struct {
	// currentConfig is the "source of truth" for this Sg instance's model configuration.
	currentConfig   *types.ModelConfiguration
	currentConfigMu sync.RWMutex

	// inTestMode is set IFF the singleton config service was initialized via a call
	// to InitMock, and allows certain operations to help testing.
	inTestMode bool
}

func (svc *service) Get() (*types.ModelConfiguration, error) {
	svc.currentConfigMu.RLock()
	defer svc.currentConfigMu.RUnlock()

	// Create a deep copy of the current configuration, so callers can operate on
	// older versions without worrying about data races or other types of errors.
	cfgCopy, err := deepCopy(svc.currentConfig)
	if err != nil {
		return nil, err
	}
	return cfgCopy, nil
}

// set updates the set.currentConfig to the supplied value. It is assumped that
// the Service will "own" the pointer, and the caller will no longer modify it.
func (svc *service) set(newConfig *types.ModelConfiguration) {
	// Block until the lock is available.
	svc.currentConfigMu.Lock()
	defer svc.currentConfigMu.Unlock()

	svc.currentConfig = newConfig
}

// failingSvc implements the Service interface, but only ever returns an error.
type failingSvc struct {
	message string
}

func (fs *failingSvc) Get() (*types.ModelConfiguration, error) {
	return nil, errors.New(fs.message)
}
