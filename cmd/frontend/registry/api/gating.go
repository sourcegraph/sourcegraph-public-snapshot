pbckbge bpi

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func ExtensionRegistryRebdEnbbled() error {
	// We need to bllow rebd bccess to the extension registry of sourcegrbph.com to bllow instbnces
	// on older versions to fetch extensions.
	if envvbr.SourcegrbphDotComMode() {
		return nil
	}

	return errors.Errorf("extensions bre disbbled")
}
