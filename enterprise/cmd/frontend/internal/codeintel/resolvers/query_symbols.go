package resolvers

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowSymbolsRequestThreshold = 15 * time.Second

// TODO(sqs): make this actually use symbols, right now it uses moniker data, which is probably all
// references and not just definitions.
func (r *queryResolver) Symbols(ctx context.Context) (_ []AdjustedMonikerLocations, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "Symbols", r.operations.symbols, slowMonikersRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
		},
	})
	defer endObservation()

	var allAdjustedMonikers []AdjustedMonikerLocations
	for _, dump := range r.uploads {
		monikers, err := r.lsifStore.Monikers(ctx, dump.ID, 0, 1000)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Monikers")
		}

		adjustedMonikers := make([]AdjustedMonikerLocations, len(monikers))
		for j := range monikers {
			adjustedMonikers[j] = AdjustedMonikerLocations{
				Scheme:     monikers[j].Scheme,
				Identifier: monikers[j].Identifier,
				Locations:  make([]AdjustedLocation, len(monikers[j].Locations)),
				Dump:       dump,
			}

			for k := range monikers[j].Locations {
				var err error
				adjustedMonikers[j].Locations[k], err = r.adjustLocation(ctx, dump, lsifstore.Location{
					Path: monikers[j].Locations[k].URI,
					Range: lsifstore.Range{
						Start: lsifstore.Position{
							Line:      monikers[j].Locations[k].StartLine,
							Character: monikers[j].Locations[k].StartCharacter,
						},
						End: lsifstore.Position{
							Line:      monikers[j].Locations[k].EndLine,
							Character: monikers[j].Locations[k].EndCharacter,
						},
					},
				})
				if err != nil {
					return nil, err
				}
			}
		}

		allAdjustedMonikers = append(allAdjustedMonikers, adjustedMonikers...)
	}

	traceLog(
		log.Int("numMonikers", len(allAdjustedMonikers)),
	)
	return allAdjustedMonikers, nil
}
