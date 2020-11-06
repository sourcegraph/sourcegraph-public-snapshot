package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

// Hover returns the hover text and range for the symbol at the given position.
func (api *CodeIntelAPI) Hover(ctx context.Context, file string, line, character, uploadID int) (string, lsifstore.Range, bool, error) {
	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
	if err != nil {
		return "", lsifstore.Range{}, false, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return "", lsifstore.Range{}, false, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(file, dump.Root)
	text, rn, exists, err := api.lsifStore.Hover(ctx, dump.ID, pathInBundle, line, character)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return "", lsifstore.Range{}, false, nil
		}
		return "", lsifstore.Range{}, false, errors.Wrap(err, "bundleClient.Hover")
	}
	if exists {
		return text, rn, true, nil
	}

	definition, exists, err := api.definitionRaw(ctx, dump, pathInBundle, line, character)
	if err != nil || !exists {
		return "", lsifstore.Range{}, false, errors.Wrap(err, "api.definitionRaw")
	}

	pathInDefinitionBundle := strings.TrimPrefix(definition.Path, definition.Dump.Root)

	text, rn, exists, err = api.lsifStore.Hover(ctx, definition.Dump.ID, pathInDefinitionBundle, definition.Range.Start.Line, definition.Range.Start.Character)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return "", lsifstore.Range{}, false, nil
		}
		return "", lsifstore.Range{}, false, errors.Wrap(err, "definitionBundleClient.Hover")
	}

	return text, rn, exists, nil
}
