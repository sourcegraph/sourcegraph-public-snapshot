package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ResolvedCodeIntelligenceRange struct {
	Range       lsifstore.Range
	Definitions []ResolvedLocation
	References  []ResolvedLocation
	HoverText   string
}

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (api *CodeIntelAPI) Ranges(ctx context.Context, file string, startLine, endLine, uploadID int) (_ []ResolvedCodeIntelligenceRange, err error) {
	ctx, endObservation := api.operations.ranges.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("file", file),
		log.Int("startLine", startLine),
		log.Int("endLine", endLine),
		log.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
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
