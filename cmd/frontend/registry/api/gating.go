package api

import (
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ExtensionRegistryReadEnabled() error {
	// We need to allow read access to the extension registry of sourcegraph.com to allow instances
	// on older versions to fetch extensions.
	if dotcom.SourcegraphDotComMode() {
		return nil
	}

	return errors.Errorf("extensions are disabled")
}
