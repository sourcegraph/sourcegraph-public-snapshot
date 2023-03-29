package graphql

import (
	"context"
	"strings"
	"time"

	traceLog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *gitBlobLSIFDataResolver) Definitions(ctx context.Context, args *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character)}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.definitions, time.Second, observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", requestArgs.RepositoryID),
			traceLog.String("commit", requestArgs.Commit),
			traceLog.String("path", requestArgs.Path),
			traceLog.Int("line", requestArgs.Line),
			traceLog.Int("character", requestArgs.Character),
			traceLog.Int("limit", requestArgs.Limit),
		},
	})
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
