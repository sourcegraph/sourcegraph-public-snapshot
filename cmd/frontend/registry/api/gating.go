package api

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ExtensionRegistryReadEnabled() error {
	// We need to allow read access to the extension registry of sourcegraph.com to allow instances
	// on older versions to fetch extensions.
	if envvar.SourcegraphDotComMode() {
		return nil
	}

	return errors.Errorf("extensions are disabled")
}
