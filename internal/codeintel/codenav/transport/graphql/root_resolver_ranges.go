package graphql

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrIllegalBounds occurs when a negative or zero-width bound is supplied by the user.
var ErrIllegalBounds = errors.New("illegal bounds")

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *gitBlobLSIFDataResolver) Ranges(ctx context.Context, args *resolverstubs.LSIFRangesArgs) (_ resolverstubs.CodeIntelligenceRangeConnectionResolver, err error) {
	requestArgs := codenav.PositionalRequestArgs{
		RequestArgs: codenav.RequestArgs{
			RepositoryID: r.requestState.RepositoryID,
			Commit:       r.requestState.Commit,
		},
		Path: r.requestState.Path,
	}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.ranges, time.Second, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", requestArgs.RepositoryID),
		attribute.String("commit", requestArgs.Commit),
		attribute.String("path", requestArgs.Path),
		attribute.Int("startLine", int(args.StartLine)),
		attribute.Int("endLine", int(args.EndLine)),
	}})
	defer endObservation()

	if args.StartLine < 0 || args.EndLine < args.StartLine {
		return nil, ErrIllegalBounds
	}

	ranges, err := r.codeNavSvc.GetRanges(ctx, requestArgs, r.requestState, int(args.StartLine), int(args.EndLine))
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.CodeIntelligenceRangeResolver
	for _, rn := range ranges {
		resolvers = append(resolvers, &codeIntelligenceRangeResolver{
			r:                rn,
			locationResolver: r.locationResolver,
		})
	}

	return resolverstubs.NewConnectionResolver(resolvers), nil
}

//
//

type codeIntelligenceRangeResolver struct {
	r                codenav.AdjustedCodeIntelligenceRange
	locationResolver *gitresolvers.CachedLocationResolver
}

func (r *codeIntelligenceRangeResolver) Range(ctx context.Context) (resolverstubs.RangeResolver, error) {
	return newRangeResolver(convertRange(r.r.Range)), nil
}

func (r *codeIntelligenceRangeResolver) Definitions(ctx context.Context) (resolverstubs.LocationConnectionResolver, error) {
	return newLocationConnectionResolver(r.r.Definitions, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) References(ctx context.Context) (resolverstubs.LocationConnectionResolver, error) {
	return newLocationConnectionResolver(r.r.References, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) Implementations(ctx context.Context) (resolverstubs.LocationConnectionResolver, error) {
	return newLocationConnectionResolver(r.r.Implementations, nil, r.locationResolver), nil
}

func (r *codeIntelligenceRangeResolver) Hover(ctx context.Context) (resolverstubs.HoverResolver, error) {
	return newHoverResolver(r.r.HoverText, convertRange(r.r.Range)), nil
}
