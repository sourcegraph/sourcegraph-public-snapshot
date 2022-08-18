package authz

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// authzProvidersReady and authzProvidersReadyOnce together indicate when
	// GetProviders should no longer block. It should block until SetProviders
	// is called at least once.
	authzProvidersReadyOnce sync.Once
	authzProvidersReady     = make(chan struct{})

	// authzProviders is the currently registered list of authorization providers.
	authzProviders []Provider

	// authzMu protects access to both allowAccessByDefault and authzProviders
	authzMu sync.RWMutex
)

// SetProviders sets the current authz parameters. It is concurrency-safe.
// TODO
func SetProviders(_ bool, z []Provider) {
	authzMu.Lock()
	defer authzMu.Unlock()

	authzProviders = z

	authzProvidersReadyOnce.Do(func() {
		close(authzProvidersReady)
	})
}

// GetProviders returns the current authz parameters. It is concurrency-safe.
//
// It blocks until SetProviders has been called at least once.
func GetProviders() (providers []Provider) {
	if !isTest {
		<-authzProvidersReady
	}
	authzMu.Lock()
	defer authzMu.Unlock()

	if authzProviders == nil {
		return nil
	}
	providers = make([]Provider, len(authzProviders))
	copy(providers, authzProviders)
	return providers
}

var isTest = (func() bool {
	path, _ := os.Executable()
	return filepath.Ext(path) == ".test" ||
		strings.Contains(path, "/T/___") || // Test path used by GoLand
		filepath.Base(path) == "__debug_bin" // Debug binary used by VSCode
})()
