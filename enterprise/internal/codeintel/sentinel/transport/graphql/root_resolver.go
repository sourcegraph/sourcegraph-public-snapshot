package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	sentinelSvc             *sentinel.Service
	uploadSvc               sharedresolvers.UploadsService
	policySvc               sharedresolvers.PolicyService
	gitserverClient         gitserver.Client
	siteAdminChecker        sharedresolvers.SiteAdminChecker
	repoStore               database.RepoStore
	prefetcherFactory       *sharedresolvers.PrefetcherFactory
	bulkLoaderFactory       *bulkLoaderFactory
	locationResolverFactory *sharedresolvers.CachedLocationResolverFactory
	operations              *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	sentinelSvc *sentinel.Service,
	uploadSvc sharedresolvers.UploadsService,
	policySvc sharedresolvers.PolicyService,
	gitserverClient gitserver.Client,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	repoStore database.RepoStore,
	prefetcherFactory *sharedresolvers.PrefetcherFactory,
	locationResolverFactory *sharedresolvers.CachedLocationResolverFactory,
) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc:             sentinelSvc,
		uploadSvc:               uploadSvc,
		policySvc:               policySvc,
		gitserverClient:         gitserverClient,
		siteAdminChecker:        siteAdminChecker,
		repoStore:               repoStore,
		prefetcherFactory:       prefetcherFactory,
		bulkLoaderFactory:       &bulkLoaderFactory{sentinelSvc},
		locationResolverFactory: locationResolverFactory,
		operations:              newOperations(observationCtx),
	}
}

func (r *rootResolver) Vulnerabilities(ctx context.Context, args resolverstubs.GetVulnerabilitiesArgs) (_ resolverstubs.VulnerabilityConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.getVulnerabilities.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("first", resolverstubs.Deref(args.First, 0)),
		log.String("after", resolverstubs.Deref(args.After, "")),
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
	ctx, errTracer, endObservation := r.operations.getMatches.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("first", resolverstubs.Deref(args.First, 0)),
		log.String("after", resolverstubs.Deref(args.After, "")),
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

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := r.prefetcherFactory.Create()
	bulkLoader := r.bulkLoaderFactory.Create()
	locationResolver := r.locationResolverFactory.Create()

	for _, match := range matches {
		prefetcher.MarkUpload(match.UploadID)
		bulkLoader.MarkVulnerability(match.VulnerabilityID)
	}

	var resolvers []resolverstubs.VulnerabilityMatchResolver
	for _, m := range matches {
		resolvers = append(resolvers, &vulnerabilityMatchResolver{
			uploadsSvc:       r.uploadSvc,
			sentinelSvc:      r.sentinelSvc,
			policySvc:        r.policySvc,
			gitserverClient:  r.gitserverClient,
			siteAdminChecker: r.siteAdminChecker,
			prefetcher:       prefetcher,
			locationResolver: locationResolver,
			errTracer:        errTracer,
			bulkLoader:       bulkLoader,
			m:                m,
		})
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, offset, int32(totalCount)), nil
}

func (r *rootResolver) VulnerabilityMatchesCountByRepository(ctx context.Context, args resolverstubs.GetVulnerabilityMatchCountByRepositoryArgs) (_ resolverstubs.VulnerabilityMatchCountByRepositoryConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityMatchesCountByRepository.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit, offset, err := args.ParseLimitOffset(50)
	if err != nil {
		return nil, err
	}

	repositoryName := ""
	if args.RepositoryName != nil {
		repositoryName = *args.RepositoryName
	}

	vulerabilityCounts, totalCount, err := r.sentinelSvc.GetVulnerabilityMatchesCountByRepository(ctx, shared.GetVulnerabilityMatchesCountByRepositoryArgs{
		Limit:          int(limit),
		Offset:         int(offset),
		RepositoryName: repositoryName,
	})
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.VulnerabilityMatchCountByRepositoryResolver
	for _, v := range vulerabilityCounts {
		resolvers = append(resolvers, &vulnerabilityMatchCountByRepositoryResolver{v: v})
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, offset, int32(totalCount)), nil
}

func (r *rootResolver) VulnerabilityByID(ctx context.Context, vulnerabilityID graphql.ID) (_ resolverstubs.VulnerabilityResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("vulnerabilityID", string(vulnerabilityID)),
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
	ctx, errTracer, endObservation := r.operations.vulnerabilityMatchByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("vulnerabilityMatchID", string(vulnerabilityMatchID)),
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

	return &vulnerabilityMatchResolver{
		uploadsSvc:       r.uploadSvc,
		sentinelSvc:      r.sentinelSvc,
		policySvc:        r.policySvc,
		gitserverClient:  r.gitserverClient,
		siteAdminChecker: r.siteAdminChecker,
		repoStore:        r.repoStore,
		prefetcher:       r.prefetcherFactory.Create(),
		locationResolver: r.locationResolverFactory.Create(),
		errTracer:        errTracer,
		bulkLoader:       r.bulkLoaderFactory.Create(),
		m:                match,
	}, nil
}

func (r *rootResolver) VulnerabilityMatchesSummaryCounts(ctx context.Context) (_ resolverstubs.VulnerabilityMatchesSummaryCountResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityMatchByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
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

//
//

type bulkLoaderFactory struct {
	sentinelSvc SentinelService
}

func (f *bulkLoaderFactory) Create() *bulkLoader {
	return NewBulkLoader(f.sentinelSvc)
}

type bulkLoader struct {
	loader *sharedresolvers.DataLoader[int, shared.Vulnerability]
}

func NewBulkLoader(sentinelSvc SentinelService) *bulkLoader {
	return &bulkLoader{
		loader: sharedresolvers.NewDataLoader[int, shared.Vulnerability](sharedresolvers.DataLoaderBackingServiceFunc[int, shared.Vulnerability](func(ctx context.Context, ids ...int) ([]shared.Vulnerability, error) {
			return sentinelSvc.GetVulnerabilitiesByIDs(ctx, ids...)
		})),
	}
}

func (l *bulkLoader) MarkVulnerability(id int) {
	l.loader.Presubmit(id)
}

func (l *bulkLoader) GetVulnerabilityByID(ctx context.Context, id int) (shared.Vulnerability, bool, error) {
	return l.loader.GetByID(ctx, id)
}

type vulnerabilityMatchResolver struct {
	sentinelSvc      SentinelService
	uploadsSvc       sharedresolvers.UploadsService
	policySvc        sharedresolvers.PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker sharedresolvers.SiteAdminChecker
	repoStore        database.RepoStore
	prefetcher       *sharedresolvers.Prefetcher
	locationResolver *sharedresolvers.CachedLocationResolver
	errTracer        *observation.ErrCollector
	bulkLoader       *bulkLoader
	m                shared.VulnerabilityMatch
}

func (r *vulnerabilityMatchResolver) ID() graphql.ID {
	return resolverstubs.MarshalID("VulnerabilityMatch", r.m.ID)
}

func (r *vulnerabilityMatchResolver) Vulnerability(ctx context.Context) (resolverstubs.VulnerabilityResolver, error) {
	vulnerability, ok, err := r.bulkLoader.GetVulnerabilityByID(ctx, r.m.VulnerabilityID)
	if err != nil || !ok {
		return nil, err
	}

	return &vulnerabilityResolver{v: vulnerability}, nil
}

func (r *vulnerabilityMatchResolver) AffectedPackage(ctx context.Context) (resolverstubs.VulnerabilityAffectedPackageResolver, error) {
	return &vulnerabilityAffectedPackageResolver{r.m.AffectedPackage}, nil
}

func (r *vulnerabilityMatchResolver) PreciseIndex(ctx context.Context) (resolverstubs.PreciseIndexResolver, error) {
	upload, ok, err := r.prefetcher.GetUploadByID(ctx, r.m.UploadID)
	if err != nil || !ok {
		return nil, err
	}

	return sharedresolvers.NewPreciseIndexResolver(
		ctx,
		r.uploadsSvc,
		r.policySvc,
		r.gitserverClient,
		r.prefetcher,
		r.siteAdminChecker,
		r.repoStore,
		r.locationResolver,
		r.errTracer,
		&upload,
		nil,
	)
}

//
//

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
