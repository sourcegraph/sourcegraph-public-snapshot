package graphql

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultDefinitionsPageSize = 100

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *gitBlobLSIFDataResolver) Definitions(ctx context.Context, args *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	requestArgs := codenav.PositionalRequestArgs{
		RequestArgs: codenav.RequestArgs{
			RepositoryID: r.requestState.RepositoryID,
			Commit:       r.requestState.Commit,
			Limit:        DefaultDefinitionsPageSize,
		},
		Path:      r.requestState.Path,
		Line:      int(args.Line),
		Character: int(args.Character),
	}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.definitions, time.Second, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", requestArgs.RepositoryID),
		attribute.String("commit", requestArgs.Commit),
		attribute.String("path", requestArgs.Path),
		attribute.Int("line", requestArgs.Line),
		attribute.Int("character", requestArgs.Character),
		attribute.Int("limit", requestArgs.Limit),
	}})
	defer endObservation()

	def, err := r.codeNavSvc.GetDefinitions(ctx, requestArgs, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetDefinitions")
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := def[:0]
		for _, loc := range def {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		def = filtered
	}

	return newLocationConnectionResolver(def, nil, r.locationResolver), nil
}
