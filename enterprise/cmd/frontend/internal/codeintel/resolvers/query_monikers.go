package resolvers

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowMonikersRequestThreshold = time.Second

func (r *queryResolver) MonikersAtPosition(ctx context.Context, line, character int) (_ []AdjustedMonikerData, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "Monikers", r.operations.monikers, slowMonikersRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation()

	adjustedUploads, err := r.adjustUploadPaths(ctx)
	if err != nil {
		return nil, err
	}

	var adjustedMonikers []AdjustedMonikerData
	for i := range adjustedUploads {
		rangeMonikers, err := r.lsifStore.MonikersByPosition(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].AdjustedPathInBundle,
			adjustedUploads[i].AdjustedPosition.Line,
			adjustedUploads[i].AdjustedPosition.Character,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.MonikersByPosition")
		}

		for _, monikers := range rangeMonikers {
			for _, moniker := range monikers {
				adjustedMonikers = append(adjustedMonikers, AdjustedMonikerData{
					MonikerData: moniker,
					Dump:        adjustedUploads[i].Upload,
				})
			}
		}
	}
	traceLog(
		log.Int("numMonikers", len(adjustedMonikers)),
	)
	return adjustedMonikers, nil
}
