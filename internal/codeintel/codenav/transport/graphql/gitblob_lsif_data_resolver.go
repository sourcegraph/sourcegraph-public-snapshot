package graphql

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go/log"

	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitBlobLSIFDataResolver interface {
	GitTreeLSIFDataResolver
	ToGitTreeLSIFData() (GitTreeLSIFDataResolver, bool)
	ToGitBlobLSIFData() (GitBlobLSIFDataResolver, bool)

	Stencil(ctx context.Context) ([]RangeResolver, error)
	Ranges(ctx context.Context, args *LSIFRangesArgs) (CodeIntelligenceRangeConnectionResolver, error)
	Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (LocationConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Implementations(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFQueryPositionArgs) (HoverResolver, error)
}

// gitBlobLSIFDataResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type gitBlobLSIFDataResolverQueryResolver struct {
	autoindexingSvc         AutoIndexingService
	uploadSvc               UploadsService
	policiesSvc             PolicyService
	gitBlobLSIFDataResolver GitBlobResolver
	locationResolver        *sharedresolvers.CachedLocationResolver
	errTracer               *observation.ErrCollector
}

// NewQueryResolver creates a new QueryResolver with the given resolver that defines all code intel-specific
// behavior. A cached location resolver instance is also given to the query resolver, which should be used
// to resolve all location-related values.
func NewGitBlobLSIFDataResolverQueryResolver(autoindexSvc AutoIndexingService, uploadSvc UploadsService, policiesSvc PolicyService, gitBlobResolver GitBlobResolver, errTracer *observation.ErrCollector) GitBlobLSIFDataResolver {
	return &gitBlobLSIFDataResolverQueryResolver{
		gitBlobLSIFDataResolver: gitBlobResolver,
		autoindexingSvc:         autoindexSvc,
		uploadSvc:               uploadSvc,
		policiesSvc:             policiesSvc,
		locationResolver:        sharedresolvers.NewCachedLocationResolver(autoindexSvc.GetUnsafeDB()),
		errTracer:               errTracer,
	}
}

func (r *gitBlobLSIFDataResolverQueryResolver) ToGitTreeLSIFData() (GitTreeLSIFDataResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDataResolverQueryResolver) ToGitBlobLSIFData() (GitBlobLSIFDataResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDataResolverQueryResolver) Stencil(ctx context.Context) (_ []RangeResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "stencil"))

	ranges, err := r.gitBlobLSIFDataResolver.Stencil(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]RangeResolver, 0, len(ranges))
	for _, r := range ranges {
		resolvers = append(resolvers, NewRangeResolver(convertRange(r)))
	}

	return resolvers, nil
}

type LSIFRangesArgs struct {
	StartLine int32
	EndLine   int32
}

// ErrIllegalBounds occurs when a negative or zero-width bound is supplied by the user.
var ErrIllegalBounds = errors.New("illegal bounds")

func (r *gitBlobLSIFDataResolverQueryResolver) Ranges(ctx context.Context, args *LSIFRangesArgs) (_ CodeIntelligenceRangeConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "ranges"))

	if args.StartLine < 0 || args.EndLine < args.StartLine {
		return nil, ErrIllegalBounds
	}

	ranges, err := r.gitBlobLSIFDataResolver.Ranges(ctx, int(args.StartLine), int(args.EndLine))
	if err != nil {
		return nil, err
	}

	return NewCodeIntelligenceRangeConnectionResolver(ranges, r.locationResolver), nil
}

func (r *gitBlobLSIFDataResolverQueryResolver) Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (_ LocationConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "definitions"))

	locations, err := r.gitBlobLSIFDataResolver.Definitions(ctx, int(args.Line), int(args.Character))
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

const DefaultReferencesPageSize = 100

func (r *gitBlobLSIFDataResolverQueryResolver) References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (_ LocationConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "references"))

	limit := derefInt32(args.First, DefaultReferencesPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	cursor, err := DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	locations, cursor, err := r.gitBlobLSIFDataResolver.References(ctx, int(args.Line), int(args.Character), limit, cursor)
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

// DefaultReferencesPageSize is the implementation result page size when no limit is supplied.
const DefaultImplementationsPageSize = 100

// ErrIllegalLimit occurs when the user requests less than one object per page.
var ErrIllegalLimit = errors.New("illegal limit")

type LSIFPagedQueryPositionArgs struct {
	LSIFQueryPositionArgs
	ConnectionArgs
	After  *string
	Filter *string
}

type LSIFQueryPositionArgs struct {
	Line      int32
	Character int32
	Filter    *string
}

func (r *gitBlobLSIFDataResolverQueryResolver) Implementations(ctx context.Context, args *LSIFPagedQueryPositionArgs) (_ LocationConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "implementations"))

	limit := derefInt32(args.First, DefaultImplementationsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	cursor, err := DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	locations, cursor, err := r.gitBlobLSIFDataResolver.Implementations(ctx, int(args.Line), int(args.Character), limit, cursor)
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

func (r *gitBlobLSIFDataResolverQueryResolver) Hover(ctx context.Context, args *LSIFQueryPositionArgs) (_ HoverResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "hover"))

	text, rx, exists, err := r.gitBlobLSIFDataResolver.Hover(ctx, int(args.Line), int(args.Character))
	if err != nil || !exists {
		return nil, err
	}

	return NewHoverResolver(text, sharedRangeTolspRange(rx)), nil
}

func (r *gitBlobLSIFDataResolverQueryResolver) LSIFUploads(ctx context.Context) (_ []sharedresolvers.LSIFUploadResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "lsifUploads"))

	uploads, err := r.gitBlobLSIFDataResolver.LSIFUploads(ctx)
	if err != nil {
		return nil, err
	}

	dbUploads := []types.Upload{}
	for _, u := range uploads {
		dbUploads = append(dbUploads, sharedDumpToDbstoreUpload(u))
	}

	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexingSvc, r.uploadSvc)

	resolvers := make([]sharedresolvers.LSIFUploadResolver, 0, len(uploads))
	for _, upload := range dbUploads {
		resolvers = append(resolvers, sharedresolvers.NewUploadResolver(r.uploadSvc, r.autoindexingSvc, r.policiesSvc, upload, prefetcher, r.errTracer))
	}

	return resolvers, nil
}

// DefaultDiagnosticsPageSize is the diagnostic result page size when no limit is supplied.
const DefaultDiagnosticsPageSize = 100

type LSIFDiagnosticsArgs struct {
	ConnectionArgs
}

func (r *gitBlobLSIFDataResolverQueryResolver) Diagnostics(ctx context.Context, args *LSIFDiagnosticsArgs) (_ DiagnosticConnectionResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("queryResolver.field", "diagnostics"))

	limit := derefInt32(args.First, DefaultDiagnosticsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	diagnostics, totalCount, err := r.gitBlobLSIFDataResolver.Diagnostics(ctx, limit)
	if err != nil {
		return nil, err
	}

	return NewDiagnosticConnectionResolver(diagnostics, totalCount, r.locationResolver), nil
}
