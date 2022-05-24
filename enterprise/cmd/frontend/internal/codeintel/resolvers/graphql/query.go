package graphql

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DefaultReferencesPageSize is the reference result page size when no limit is supplied.
const DefaultReferencesPageSize = 100

// DefaultReferencesPageSize is the implementation result page size when no limit is supplied.
const DefaultImplementationsPageSize = 100

// DefaultDiagnosticsPageSize is the diagnostic result page size when no limit is supplied.
const DefaultDiagnosticsPageSize = 100

// ErrIllegalLimit occurs when the user requests less than one object per page.
var ErrIllegalLimit = errors.New("illegal limit")

// ErrIllegalBounds occurs when a negative or zero-width bound is supplied by the user.
var ErrIllegalBounds = errors.New("illegal bounds")

// QueryResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type QueryResolver struct {
	queryResolver    resolvers.QueryResolver
	resolver         resolvers.Resolver
	gitserver        policies.GitserverClient
	locationResolver *CachedLocationResolver
	errTracer        *observation.ErrCollector
}

// NewQueryResolver creates a new QueryResolver with the given resolver that defines all code intel-specific
// behavior. A cached location resolver instance is also given to the query resolver, which should be used
// to resolve all location-related values.
func NewQueryResolver(gitserver policies.GitserverClient, queryResolver resolvers.QueryResolver, resolver resolvers.Resolver, locationResolver *CachedLocationResolver, errTracer *observation.ErrCollector) gql.GitBlobLSIFDataResolver {
	return &QueryResolver{
		queryResolver:    queryResolver,
		resolver:         resolver,
		gitserver:        gitserver,
		locationResolver: locationResolver,
		errTracer:        errTracer,
	}
}

func (r *QueryResolver) ToGitTreeLSIFData() (gql.GitTreeLSIFDataResolver, bool) { return r, true }
func (r *QueryResolver) ToGitBlobLSIFData() (gql.GitBlobLSIFDataResolver, bool) { return r, true }

func (r *QueryResolver) Stencil(ctx context.Context) (_ []gql.RangeResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "stencil"))

	ranges, err := r.queryResolver.Stencil(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]gql.RangeResolver, 0, len(ranges))
	for _, r := range ranges {
		resolvers = append(resolvers, gql.NewRangeResolver(convertRange(r)))
	}

	return resolvers, nil
}

func (r *QueryResolver) Ranges(ctx context.Context, args *gql.LSIFRangesArgs) (_ gql.CodeIntelligenceRangeConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "ranges"))

	if args.StartLine < 0 || args.EndLine < args.StartLine {
		return nil, ErrIllegalBounds
	}

	ranges, err := r.queryResolver.Ranges(ctx, int(args.StartLine), int(args.EndLine))
	if err != nil {
		return nil, err
	}

	return &CodeIntelligenceRangeConnectionResolver{
		ranges:           ranges,
		locationResolver: r.locationResolver,
	}, nil
}

func (r *QueryResolver) Definitions(ctx context.Context, args *gql.LSIFQueryPositionArgs) (_ gql.LocationConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "definitions"))

	locations, err := r.queryResolver.Definitions(ctx, int(args.Line), int(args.Character))
	if err != nil {
		return nil, err
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := locations[:0]
		for _, loc := range locations {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		locations = filtered
	}

	return NewLocationConnectionResolver(locations, nil, r.locationResolver), nil
}

func (r *QueryResolver) References(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) (_ gql.LocationConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "references"))

	limit := derefInt32(args.First, DefaultReferencesPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	cursor, err := graphqlutil.DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	locations, cursor, err := r.queryResolver.References(ctx, int(args.Line), int(args.Character), limit, cursor)
	if err != nil {
		return nil, err
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := locations[:0]
		for _, loc := range locations {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		locations = filtered
	}

	return NewLocationConnectionResolver(locations, strPtr(cursor), r.locationResolver), nil
}

func (r *QueryResolver) Implementations(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) (_ gql.LocationConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "implementations"))

	limit := derefInt32(args.First, DefaultImplementationsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	cursor, err := graphqlutil.DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	locations, cursor, err := r.queryResolver.Implementations(ctx, int(args.Line), int(args.Character), limit, cursor)
	if err != nil {
		return nil, err
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := locations[:0]
		for _, loc := range locations {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		locations = filtered
	}

	return NewLocationConnectionResolver(locations, strPtr(cursor), r.locationResolver), nil
}

func (r *QueryResolver) Hover(ctx context.Context, args *gql.LSIFQueryPositionArgs) (_ gql.HoverResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "hover"))

	text, rx, exists, err := r.queryResolver.Hover(ctx, int(args.Line), int(args.Character))
	if err != nil || !exists {
		return nil, err
	}

	return NewHoverResolver(text, convertRange(rx)), nil
}

func (r *QueryResolver) LSIFUploads(ctx context.Context) (_ []gql.LSIFUploadResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "lsifUploads"))

	uploads, err := r.queryResolver.LSIFUploads(ctx)
	if err != nil {
		return nil, err
	}

	prefetcher := NewPrefetcher(r.resolver)

	resolvers := make([]gql.LSIFUploadResolver, 0, len(uploads))
	for _, upload := range uploads {
		resolvers = append(resolvers, NewUploadResolver(r.locationResolver.db, r.gitserver, r.resolver, upload, prefetcher, r.locationResolver, r.errTracer))
	}

	return resolvers, nil
}

func (r *QueryResolver) Diagnostics(ctx context.Context, args *gql.LSIFDiagnosticsArgs) (_ gql.DiagnosticConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "diagnostics"))

	limit := derefInt32(args.First, DefaultDiagnosticsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	diagnostics, totalCount, err := r.queryResolver.Diagnostics(ctx, limit)
	if err != nil {
		return nil, err
	}

	return NewDiagnosticConnectionResolver(diagnostics, totalCount, r.locationResolver), nil
}
