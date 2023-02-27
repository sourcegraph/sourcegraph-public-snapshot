package ui

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func ShouldRedirectLegacyExtensionEndpoints() bool {
	if conf.ExperimentalFeatures().EnableLegacyExtensions {
		return false
	}
	return true
}
