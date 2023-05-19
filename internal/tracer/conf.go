package tracer

import "github.com/sourcegraph/sourcegraph/internal/conf/conftypes"

// ConfConfigurationSource wraps a basic conf-based configuration source in
// a tracing-specific configuration source.
type ConfConfigurationSource struct{ conftypes.WatchableSiteConfig }

var _ WatchableConfigurationSource = ConfConfigurationSource{}

func (c ConfConfigurationSource) Config() Configuration {
	s := c.SiteConfig()
	return Configuration{
		ExternalURL:          s.ExternalURL,
		ObservabilityTracing: s.ObservabilityTracing,
	}
}
