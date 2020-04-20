package api

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/bundles"
)

func (api *codeIntelAPI) Hover(file string, line, character, uploadID int) (string, bundles.Range, bool, error) {
	dump, exists, err := api.db.GetDumpByID(context.Background(), uploadID)
	if err != nil {
		return "", bundles.Range{}, false, err
	}
	if !exists {
		return "", bundles.Range{}, false, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(file, dump.Root)
	bundleClient := api.bundleManagerClient.BundleClient(dump.ID)

	text, rn, exists, err := bundleClient.Hover(context.Background(), pathInBundle, line, character)
	if err != nil {
		return "", bundles.Range{}, false, err
	}
	if exists {
		return text, rn, true, nil
	}

	definition, exists, err := api.definitionRaw(dump, bundleClient, pathInBundle, line, character)
	if err != nil || !exists {
		return "", bundles.Range{}, false, err
	}

	pathInDefinitionBundle := strings.TrimPrefix(definition.Path, definition.Dump.Root)
	definitionBundleClient := api.bundleManagerClient.BundleClient(definition.Dump.ID)

	return definitionBundleClient.Hover(context.Background(), pathInDefinitionBundle, definition.Range.Start.Line, definition.Range.Start.Character)
}
