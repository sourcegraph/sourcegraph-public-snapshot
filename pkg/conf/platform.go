package conf

import "github.com/sourcegraph/sourcegraph/schema"

// PlatformConfiguration contains site configuration for the Sourcegraph platform.
type PlatformConfiguration struct {
	RemoteRegistryURL string
}

// DefaultRemoteRegistry is the default value for the site configuration field
// "remoteRegistry" when unspecified.
var DefaultRemoteRegistry string

// Extensions returns the configuration for the Sourcegraph platform, or nil if it is disabled.
func Extensions() *PlatformConfiguration {
	cfg := Get()

	x := cfg.Extensions
	if x == nil {
		x = &schema.Extensions{}
	}

	if x.Disabled != nil && *x.Disabled {
		return nil
	}

	var pc PlatformConfiguration

	// If the "remoteRegistry" value is a string, use that. If false, then keep it empty. Otherwise
	// use the default.
	if s, ok := x.RemoteRegistry.(string); ok {
		pc.RemoteRegistryURL = s
	} else if b, ok := x.RemoteRegistry.(bool); ok && !b {
		// Nothing to do.
	} else {
		pc.RemoteRegistryURL = DefaultRemoteRegistry
	}

	return &pc
}
