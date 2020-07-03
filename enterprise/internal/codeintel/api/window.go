package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
)

type ResolvedAggregateCodeIntelligence struct {
	Range       bundles.Range
	Definitions []ResolvedLocation
	References  []ResolvedLocation
	HoverText   string
}

// Window returns definition, reference, and hover data for each range within the given span of lines.
func (api *codeIntelAPI) Window(ctx context.Context, file string, startLine, endLine, uploadID int) ([]ResolvedAggregateCodeIntelligence, error) {
	dump, exists, err := api.store.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(file, dump.Root)
	bundleClient := api.bundleManagerClient.BundleClient(dump.ID)

	ranges, err := bundleClient.Window(ctx, pathInBundle, startLine, endLine)
	if err != nil {
		if err == bundles.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, nil
		}
		return nil, errors.Wrap(err, "bundleClient.Window")
	}

	var aggregatedRanges []ResolvedAggregateCodeIntelligence
	for _, r := range ranges {
		aggregatedRanges = append(aggregatedRanges, ResolvedAggregateCodeIntelligence{
			Range:       r.Range,
			Definitions: resolveLocationsWithDump(dump, r.Definitions),
			References:  resolveLocationsWithDump(dump, r.References),
			HoverText:   r.HoverText,
		})
	}

	return aggregatedRanges, nil
}
