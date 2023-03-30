package graphql

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrIllegalBounds occurs when a negative or zero-width bound is supplied by the user.
var ErrIllegalBounds = errors.New("illegal bounds")

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *gitBlobLSIFDataResolver) Ranges(ctx context.Context, args *resolverstubs.LSIFRangesArgs) (_ resolverstubs.CodeIntelligenceRangeConnectionResolver, err error) {
	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.ranges, time.Second, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", requestArgs.RepositoryID),
			log.String("commit", requestArgs.Commit),
			log.String("path", requestArgs.Path),
			log.Int("startLine", int(args.StartLine)),
			log.Int("endLine", int(args.EndLine)),
		},
	})
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
	r                shared.AdjustedCodeIntelligenceRange
	locationResolver *sharedresolvers.CachedLocationResolver
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
