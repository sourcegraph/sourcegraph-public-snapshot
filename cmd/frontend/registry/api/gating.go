package api

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ExtensionRegistryReadEnabled() error {
	// We need to allow read access to the extension registry of sourcegraph.com to allow instances
	// on older versions to fetch extensions.
	if envvar.SourcegraphDotComMode() {
		return nil
	}

	return ExtensionRegistryWriteEnabled()
}

func ExtensionRegistryWriteEnabled() error {
	cfg := conf.Get()
	if cfg.ExperimentalFeatures != nil && cfg.ExperimentalFeatures.EnableLegacyExtensions != nil && *cfg.ExperimentalFeatures.EnableLegacyExtensions == false {
		return errors.Errorf("Extensions are disabled. See https://docs.sourcegraph.com/extensions/deprecation")
	}

	// @TODO(@philipp-spiess): Change the default when we roll out 4.0
	return nil
}
