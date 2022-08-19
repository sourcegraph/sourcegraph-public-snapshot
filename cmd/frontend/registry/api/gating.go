package api

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
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
	// @TODO(@philipp-spiess): Enable this once we shipped a fix for #40085 and make the flag dynamic
	// cfg := conf.Get()
	// if cfg.ExperimentalFeatures != nil && cfg.ExperimentalFeatures.EnableLegacyExtensions == false {
	// 	return errors.Errorf("Extensions are disabled. See https://docs.sourcegraph.com/extensions/deprecation")
	// }

	// @TODO(@philipp-spiess): Change the default when we roll out 4.0
	return nil
}
