package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"go.opentelemetry.io/otel/attribute"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type rootResolver struct {
	sentinelSvc                 SentinelService
	vulnerabilityLoaderFactory  VulnerabilityLoaderFactory
	uploadLoaderFactory         uploadsgraphql.UploadLoaderFactory
	indexLoaderFactory          uploadsgraphql.IndexLoaderFactory
	locationResolverFactory     *gitresolvers.CachedLocationResolverFactory
	preciseIndexResolverFactory *uploadsgraphql.PreciseIndexResolverFactory
	operations                  *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	sentinelSvc SentinelService,
	uploadLoaderFactory uploadsgraphql.UploadLoaderFactory,
	indexLoaderFactory uploadsgraphql.IndexLoaderFactory,
	locationResolverFactory *gitresolvers.CachedLocationResolverFactory,
	preciseIndexResolverFactory *uploadsgraphql.PreciseIndexResolverFactory,
) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc:                 sentinelSvc,
		vulnerabilityLoaderFactory:  NewVulnerabilityLoaderFactory(sentinelSvc),
		uploadLoaderFactory:         uploadLoaderFactory,
		indexLoaderFactory:          indexLoaderFactory,
		locationResolverFactory:     locationResolverFactory,
		preciseIndexResolverFactory: preciseIndexResolverFactory,
		operations:                  newOperations(observationCtx),
	}
}

func (r *rootResolver) Vulnerabilities(ctx context.Context, args resolverstubs.GetVulnerabilitiesArgs) (_ resolverstubs.VulnerabilityConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.getVulnerabilities.WithErrors(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("first", int(pointers.Deref(args.First, 0))),
		attribute.String("after", pointers.Deref(args.After, "")),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit, offset, err := args.ParseLimitOffset(50)
	if err != nil {
		return nil, err
	}

	vulnerabilities, totalCount, err := r.sentinelSvc.GetVulnerabilities(ctx, shared.GetVulnerabilitiesArgs{
		Limit:  int(limit),
		Offset: int(offset),
	})
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.VulnerabilityResolver
	for _, v := range vulnerabilities {
		resolvers = append(resolvers, &vulnerabilityResolver{v: v})
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, offset, int32(totalCount)), nil
}

func (r *rootResolver) VulnerabilityMatches(ctx context.Context, args resolverstubs.GetVulnerabilityMatchesArgs) (_ resolverstubs.VulnerabilityMatchConnectionResolver, err error) {
	ctx, errTracer, endObservation := r.operations.getMatches.WithErrors(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("first", int(pointers.Deref(args.First, 0))),
		attribute.String("after", pointers.Deref(args.After, "")),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit, offset, err := args.ParseLimitOffset(50)
	if err != nil {
		return nil, err
	}

	language := ""
	if args.Language != nil {
		language = *args.Language
	}

	severity := ""
	if args.Severity != nil {
		severity = *args.Severity
	}

	repositoryName := ""
	if args.RepositoryName != nil {
		repositoryName = *args.RepositoryName
	}

	matches, totalCount, err := r.sentinelSvc.GetVulnerabilityMatches(ctx, shared.GetVulnerabilityMatchesArgs{
		Limit:          int(limit),
		Offset:         int(offset),
		Language:       language,
		Severity:       severity,
		RepositoryName: repositoryName,
	})
	if err != nil {
		return nil, err
	}

	// Pre-submit vulnerability and upload ids for loading
	vulnerabilityLoader := r.vulnerabilityLoaderFactory.Create()
	uploadLoader := r.uploadLoaderFactory.Create()
	PresubmitMatches(vulnerabilityLoader, uploadLoader, matches...)

	// No data to load for associated indexes or git data (yet)
	indexLoader := r.indexLoaderFactory.Create()
	locationResolver := r.locationResolverFactory.Create()

	var resolvers []resolverstubs.VulnerabilityMatchResolver
	for _, m := range matches {
		resolvers = append(resolvers, &vulnerabilityMatchResolver{
			uploadLoader:        uploadLoader,
			indexLoader:         indexLoader,
			locationResolver:    locationResolver,
			errTracer:           errTracer,
			vulnerabilityLoader: vulnerabilityLoader,
			m:                   m,
		})
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, offset, int32(totalCount)), nil
}

func (r *rootResolver) VulnerabilityMatchesCountByRepository(ctx context.Context, args resolverstubs.GetVulnerabilityMatchCountByRepositoryArgs) (_ resolverstubs.VulnerabilityMatchCountByRepositoryConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityMatchesCountByRepository.WithErrors(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit, offset, err := args.ParseLimitOffset(50)
	if err != nil {
		return nil, err
	}

	repositoryName := ""
	if args.RepositoryName != nil {
		repositoryName = *args.RepositoryName
	}

	vulnerabilityCounts, totalCount, err := r.sentinelSvc.GetVulnerabilityMatchesCountByRepository(ctx, shared.GetVulnerabilityMatchesCountByRepositoryArgs{
		Limit:          int(limit),
		Offset:         int(offset),
		RepositoryName: repositoryName,
	})
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.VulnerabilityMatchCountByRepositoryResolver
	for _, v := range vulnerabilityCounts {
		resolvers = append(resolvers, &vulnerabilityMatchCountByRepositoryResolver{v: v})
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, offset, int32(totalCount)), nil
}

func (r *rootResolver) VulnerabilityByID(ctx context.Context, vulnerabilityID graphql.ID) (_ resolverstubs.VulnerabilityResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityByID.WithErrors(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("vulnerabilityID", string(vulnerabilityID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	id, err := resolverstubs.UnmarshalID[int](vulnerabilityID)
	if err != nil {
		return nil, err
	}

	vulnerability, ok, err := r.sentinelSvc.VulnerabilityByID(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	return &vulnerabilityResolver{vulnerability}, nil
}

func (r *rootResolver) VulnerabilityMatchByID(ctx context.Context, vulnerabilityMatchID graphql.ID) (_ resolverstubs.VulnerabilityMatchResolver, err error) {
	ctx, errTracer, endObservation := r.operations.vulnerabilityMatchByID.WithErrors(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("vulnerabilityMatchID", string(vulnerabilityMatchID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	id, err := resolverstubs.UnmarshalID[int](vulnerabilityMatchID)
	if err != nil {
		return nil, err
	}

	match, ok, err := r.sentinelSvc.VulnerabilityMatchByID(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	// Pre-submit vulnerability and upload ids for loading
	vulnerabilityLoader := r.vulnerabilityLoaderFactory.Create()
	uploadLoader := r.uploadLoaderFactory.Create()
	PresubmitMatches(vulnerabilityLoader, uploadLoader, match)

	// No data to load for associated indexes or git data (yet)
	indexLoader := r.indexLoaderFactory.Create()
	locationResolver := r.locationResolverFactory.Create()

	return &vulnerabilityMatchResolver{
		uploadLoader:     uploadLoader,
		indexLoader:      indexLoader,
		locationResolver: locationResolver,

		errTracer:                   errTracer,
		vulnerabilityLoader:         vulnerabilityLoader,
		m:                           match,
		preciseIndexResolverFactory: r.preciseIndexResolverFactory,
	}, nil
}

func (r *rootResolver) VulnerabilityMatchesSummaryCounts(ctx context.Context) (_ resolverstubs.VulnerabilityMatchesSummaryCountResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityMatchesSummaryCounts.WithErrors(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	counts, err := r.sentinelSvc.GetVulnerabilityMatchesSummaryCounts(ctx)
	if err != nil {
		return nil, err
	}

	return &vulnerabilityMatchesSummaryCountResolver{
		critical:   counts.Critical,
		high:       counts.High,
		medium:     counts.Medium,
		low:        counts.Low,
		repository: counts.Repositories,
	}, nil
}

//
//

type vulnerabilityResolver struct {
	v shared.Vulnerability
}

func (r *vulnerabilityResolver) ID() graphql.ID {
	return resolverstubs.MarshalID("Vulnerability", r.v.ID)
}
func (r *vulnerabilityResolver) SourceID() string   { return r.v.SourceID }
func (r *vulnerabilityResolver) Summary() string    { return r.v.Summary }
func (r *vulnerabilityResolver) Details() string    { return r.v.Details }
func (r *vulnerabilityResolver) CPEs() []string     { return r.v.CPEs }
func (r *vulnerabilityResolver) CWEs() []string     { return r.v.CWEs }
func (r *vulnerabilityResolver) Aliases() []string  { return r.v.Aliases }
func (r *vulnerabilityResolver) Related() []string  { return r.v.Related }
func (r *vulnerabilityResolver) DataSource() string { return r.v.DataSource }
func (r *vulnerabilityResolver) URLs() []string     { return r.v.URLs }
func (r *vulnerabilityResolver) Severity() string   { return r.v.Severity }
func (r *vulnerabilityResolver) CVSSVector() string { return r.v.CVSSVector }
func (r *vulnerabilityResolver) CVSSScore() string  { return r.v.CVSSScore }

func (r *vulnerabilityResolver) Published() gqlutil.DateTime {
	return *gqlutil.DateTimeOrNil(&r.v.PublishedAt)
}

func (r *vulnerabilityResolver) Modified() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.ModifiedAt)
}

func (r *vulnerabilityResolver) Withdrawn() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.WithdrawnAt)
}

func (r *vulnerabilityResolver) AffectedPackages() []resolverstubs.VulnerabilityAffectedPackageResolver {
	var resolvers []resolverstubs.VulnerabilityAffectedPackageResolver
	for _, p := range r.v.AffectedPackages {
		resolvers = append(resolvers, &vulnerabilityAffectedPackageResolver{
			p: p,
		})
	}

	return resolvers
}

type vulnerabilityAffectedPackageResolver struct {
	p shared.AffectedPackage
}

func (r *vulnerabilityAffectedPackageResolver) PackageName() string { return r.p.PackageName }
func (r *vulnerabilityAffectedPackageResolver) Language() string    { return r.p.Language }
func (r *vulnerabilityAffectedPackageResolver) Namespace() string   { return r.p.Namespace }
func (r *vulnerabilityAffectedPackageResolver) VersionConstraint() []string {
	return r.p.VersionConstraint
}
func (r *vulnerabilityAffectedPackageResolver) Fixed() bool      { return r.p.Fixed }
func (r *vulnerabilityAffectedPackageResolver) FixedIn() *string { return r.p.FixedIn }

func (r *vulnerabilityAffectedPackageResolver) AffectedSymbols() []resolverstubs.VulnerabilityAffectedSymbolResolver {
	var resolvers []resolverstubs.VulnerabilityAffectedSymbolResolver
	for _, s := range r.p.AffectedSymbols {
		resolvers = append(resolvers, &vulnerabilityAffectedSymbolResolver{
			s: s,
		})
	}

	return resolvers
}

type vulnerabilityAffectedSymbolResolver struct {
	s shared.AffectedSymbol
}

func (r *vulnerabilityAffectedSymbolResolver) Path() string      { return r.s.Path }
func (r *vulnerabilityAffectedSymbolResolver) Symbols() []string { return r.s.Symbols }

type vulnerabilityMatchResolver struct {
	uploadLoader                uploadsgraphql.UploadLoader
	indexLoader                 uploadsgraphql.IndexLoader
	locationResolver            *gitresolvers.CachedLocationResolver
	errTracer                   *observation.ErrCollector
	vulnerabilityLoader         VulnerabilityLoader
	m                           shared.VulnerabilityMatch
	preciseIndexResolverFactory *uploadsgraphql.PreciseIndexResolverFactory
}

func (r *vulnerabilityMatchResolver) ID() graphql.ID {
	return resolverstubs.MarshalID("VulnerabilityMatch", r.m.ID)
}

func (r *vulnerabilityMatchResolver) Vulnerability(ctx context.Context) (resolverstubs.VulnerabilityResolver, error) {
	vulnerability, ok, err := r.vulnerabilityLoader.GetByID(ctx, r.m.VulnerabilityID)
	if err != nil || !ok {
		return nil, err
	}

	return &vulnerabilityResolver{v: vulnerability}, nil
}

func (r *vulnerabilityMatchResolver) AffectedPackage(ctx context.Context) (resolverstubs.VulnerabilityAffectedPackageResolver, error) {
	return &vulnerabilityAffectedPackageResolver{r.m.AffectedPackage}, nil
}

func (r *vulnerabilityMatchResolver) PreciseIndex(ctx context.Context) (resolverstubs.PreciseIndexResolver, error) {
	upload, ok, err := r.uploadLoader.GetByID(ctx, r.m.UploadID)
	if err != nil || !ok {
		return nil, err
	}

	return r.preciseIndexResolverFactory.Create(ctx, r.uploadLoader, r.indexLoader, r.locationResolver, r.errTracer, &upload, nil)
}

//
//

type vulnerabilityMatchesSummaryCountResolver struct {
	critical   int32
	high       int32
	medium     int32
	low        int32
	repository int32
}

func (v *vulnerabilityMatchesSummaryCountResolver) Critical() int32 { return v.critical }
func (v *vulnerabilityMatchesSummaryCountResolver) High() int32     { return v.high }
func (v *vulnerabilityMatchesSummaryCountResolver) Medium() int32   { return v.medium }
func (v *vulnerabilityMatchesSummaryCountResolver) Low() int32      { return v.low }
func (v *vulnerabilityMatchesSummaryCountResolver) Repository() int32 {
	return v.repository
}

type vulnerabilityMatchCountByRepositoryResolver struct {
	v shared.VulnerabilityMatchesByRepository
}

func (v vulnerabilityMatchCountByRepositoryResolver) ID() graphql.ID {
	return resolverstubs.MarshalID("VulnerabilityMatchCountByRepository", v.v.ID)
}

func (v vulnerabilityMatchCountByRepositoryResolver) RepositoryName() string {
	return v.v.RepositoryName
}

func (v vulnerabilityMatchCountByRepositoryResolver) MatchCount() int32 {
	return v.v.MatchCount
}
