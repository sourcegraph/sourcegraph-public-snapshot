package ui

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func ShouldRedirectLegacyExtensionEndpoints() bool {
	cfg := conf.Get()
	if cfg.ExperimentalFeatures != nil && cfg.ExperimentalFeatures.EnableLegacyExtensions != nil && *cfg.ExperimentalFeatures.EnableLegacyExtensions == true {
		return false
	}
	return true
}
