package conf

import "github.com/sourcegraph/sourcegraph/schema"

// PlatformConfiguration contains site configuration for the Sourcegraph platform.
type PlatformConfiguration struct {
	RemoteRegistryURL string
}

// Platform returns the configuration for the Sourcegraph platform, or nil if it is disabled.
func Platform() *PlatformConfiguration {
	cfg := Get()
	if cfg.ExperimentalFeatures == nil || cfg.ExperimentalFeatures.Platform == nil || !*cfg.ExperimentalFeatures.Platform {
		return nil
	}

	p := cfg.Platform
	if p == nil {
		p = &schema.Platform{}
	}

	var pc PlatformConfiguration

	// If the "remoteRegistry" value is a string, use that. If false, then keep it empty. Otherwise
	// use the default.
	const defaultRemoteRegistry = "https://sourcegraph.com/.api/registry"
	if s, ok := p.RemoteRegistry.(string); ok {
		pc.RemoteRegistryURL = s
	} else if b, ok := p.RemoteRegistry.(bool); ok && !b {
		// Nothing to do.
	} else {
		pc.RemoteRegistryURL = defaultRemoteRegistry
	}

	return &pc
}
