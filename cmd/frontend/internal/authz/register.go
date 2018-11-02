package authz

import "sync"

var (
	// allowAccessByDefault, if set to true, grants all users access to repositories that are
	// not matched by any authz provider. The default value is true. It is only set to false in
	// error modes (when the configuration is in a state where interpreting it literally could lead
	// to leakage of private repositories).
	allowAccessByDefault bool = true

	// authzProviders is the currently registered list of authorization providers.
	authzProviders []AuthzProvider

	// authzMu protects access to both allowAccessByDefault and authzProviders
	authzMu sync.RWMutex
)

// SetProviders sets the current authz parameters. It is concurrency-safe.
func SetProviders(authzAllowByDefault bool, z []AuthzProvider) {
	authzMu.Lock()
	defer authzMu.Unlock()

	authzProviders = z
	allowAccessByDefault = authzAllowByDefault
}

// DoWithAuthzProviders provides concurrency-safe access to the authz providers currently registered.
func DoWithAuthzProviders(f func(p []AuthzProvider) error) error {
	authzMu.RLock()
	defer authzMu.RUnlock()
	return f(authzProviders)
}

// NumAuthzProviders returns the number of authz providers currently registered
func NumAuthzProviders() int {
	authzMu.RLock()
	defer authzMu.RUnlock()
	return len(authzProviders)
}

// AllowByDefault returns true if and only if the current authz behavior is to allow access by
// default to a repository that isn't covered by a registered authz provider.
func AllowByDefault() bool {
	return allowAccessByDefault
}
