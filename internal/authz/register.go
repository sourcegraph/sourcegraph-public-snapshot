package authz

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var (
	// allowAccessByDefault, if set to true, grants all users access to repositories that are
	// not matched by any authz provider. The default value is true. It is only set to false in
	// error modes (when the configuration is in a state where interpreting it literally could lead
	// to leakage of private repositories).
	allowAccessByDefault = true

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
func SetProviders(authzAllowByDefault bool, z []Provider) {
	authzMu.Lock()
	defer authzMu.Unlock()

	authzProviders = z
	allowAccessByDefault = authzAllowByDefault

	// 🚨 SECURITY: We do not want to allow access by default by any means on
	// dotcom.
	if envvar.SourcegraphDotComMode() {
		allowAccessByDefault = false
	}

	authzProvidersReadyOnce.Do(func() {
		close(authzProvidersReady)
	})
}

// GetProviders returns the current authz parameters. It is concurrency-safe.
//
// It blocks until SetProviders has been called at least once.
func GetProviders() (authzAllowByDefault bool, providers []Provider) {
	if !testutil.IsTest {
		<-authzProvidersReady
	}
	authzMu.Lock()
	defer authzMu.Unlock()

	if authzProviders == nil {
		return allowAccessByDefault, nil
	}
	providers = make([]Provider, len(authzProviders))
	copy(providers, authzProviders)
	return allowAccessByDefault, providers
}
