package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

type ResolvedCodeIntelligenceRange struct {
	Range       lsifstore.Range
	Definitions []ResolvedLocation
	References  []ResolvedLocation
	HoverText   string
}

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (api *codeIntelAPI) Ranges(ctx context.Context, file string, startLine, endLine, uploadID int) ([]ResolvedCodeIntelligenceRange, error) {
	dump, exists, err := api.store.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(file, dump.Root)
	ranges, err := api.lsifStore.Ranges(ctx, dump.ID, pathInBundle, startLine, endLine)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, nil
		}
		return nil, errors.Wrap(err, "bundleClient.Ranges")
	}

	var codeintelRanges []ResolvedCodeIntelligenceRange
	for _, r := range ranges {
		codeintelRanges = append(codeintelRanges, ResolvedCodeIntelligenceRange{
			Range:       r.Range,
			Definitions: resolveLocationsWithDump(dump, r.Definitions),
			References:  resolveLocationsWithDump(dump, r.References),
			HoverText:   r.HoverText,
		})
	}

	return codeintelRanges, nil
}
