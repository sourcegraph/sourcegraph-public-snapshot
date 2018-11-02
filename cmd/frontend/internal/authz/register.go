package authz

import "sync"

var (
	// allowAccessByDefault, if set to true, grants all users access to repositories that are
	// not matched by any authz provider. The default value is true. It is only set to false in
	// error modes (when the configuration is in a state where interpreting it literally could lead
	// to leakage of private repositories).
	allowAccessByDefault bool = true

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
}

// GetProviders returns the current authz parameters. It is concurrency-safe.
func GetProviders() (authzAllowByDefault bool, providers []Provider) {
	authzMu.Lock()
	defer authzMu.Unlock()

	if authzProviders == nil {
		return allowAccessByDefault, nil
	}
	providers = make([]Provider, len(authzProviders))
	for i, p := range authzProviders {
		providers[i] = p
	}
	return allowAccessByDefault, providers
}
