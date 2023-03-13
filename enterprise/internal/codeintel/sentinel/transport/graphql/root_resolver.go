package graphql

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	sentinelSvc  *sentinel.Service
	autoindexSvc sharedresolvers.AutoIndexingService
	uploadSvc    sharedresolvers.UploadsService
	policySvc    sharedresolvers.PolicyService
	operations   *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	sentinelSvc *sentinel.Service,
	autoindexSvc sharedresolvers.AutoIndexingService,
	uploadSvc sharedresolvers.UploadsService,
	policySvc sharedresolvers.PolicyService,
) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc:  sentinelSvc,
		autoindexSvc: autoindexSvc,
		uploadSvc:    uploadSvc,
		policySvc:    policySvc,
		operations:   newOperations(observationCtx),
	}
}

func (r *rootResolver) Vulnerabilities(ctx context.Context, args resolverstubs.GetVulnerabilitiesArgs) (_ resolverstubs.VulnerabilityConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.getVulnerabilities.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit := 50
	if args.First != nil {
		limit = int(*args.First)
	}

	offset := 0
	if args.After != nil {
		after, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}

		offset = after
	}

	vulnerabilities, totalCount, err := r.sentinelSvc.GetVulnerabilities(ctx, shared.GetVulnerabilitiesArgs{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return &vulnerabilityConnectionResolver{
		vulnerabilities: vulnerabilities,
		offset:          offset,
		totalCount:      totalCount,
	}, nil
}

func (r *rootResolver) VulnerabilityMatches(ctx context.Context, args resolverstubs.GetVulnerabilityMatchesArgs) (_ resolverstubs.VulnerabilityMatchConnectionResolver, err error) {
	ctx, errTracer, endObservation := r.operations.getMatches.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit := 50
	if args.First != nil {
		limit = int(*args.First)
	}

	offset := 0
	if args.After != nil {
		after, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}

		offset = after
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
		Limit:          limit,
		Offset:         offset,
		Language:       language,
		Severity:       severity,
		RepositoryName: repositoryName,
	})
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)
	db := r.autoindexSvc.GetUnsafeDB()
	locationResolver := sharedresolvers.NewCachedLocationResolver(db, gitserver.NewClient())
	bulkLoader := NewBulkLoader(r.sentinelSvc)

	for _, match := range matches {
		prefetcher.MarkUpload(match.UploadID)
		bulkLoader.MarkVulnerability(match.VulnerabilityID)
	}

	return &vulnerabilityMatchConnectionResolver{
		sentinelSvc:      r.sentinelSvc,
		autoindexSvc:     r.autoindexSvc,
		uploadSvc:        r.uploadSvc,
		policySvc:        r.policySvc,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		errTracer:        errTracer,
		bulkLoader:       bulkLoader,
		matches:          matches,
		offset:           offset,
		totalCount:       totalCount,
	}, nil
}

func (r *rootResolver) VulnerabilityByID(ctx context.Context, gqlID graphql.ID) (_ resolverstubs.VulnerabilityResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	id, err := unmarshalVulnerabilityGQLID(gqlID)
	if err != nil {
		return nil, err
	}

	vulnerability, ok, err := r.sentinelSvc.VulnerabilityByID(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	return &vulnerabilityResolver{vulnerability}, nil
}

func (r *rootResolver) VulnerabilityMatchByID(ctx context.Context, gqlID graphql.ID) (_ resolverstubs.VulnerabilityMatchResolver, err error) {
	ctx, errTracer, endObservation := r.operations.vulnerabilityMatchByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	id, err := unmarshalVulnerabilityMatchGQLID(gqlID)
	if err != nil {
		return nil, err
	}

	match, ok, err := r.sentinelSvc.VulnerabilityMatchByID(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)
	db := r.autoindexSvc.GetUnsafeDB()
	locationResolver := sharedresolvers.NewCachedLocationResolver(db, gitserver.NewClient())
	bulkLoader := NewBulkLoader(r.sentinelSvc)

	return &vulnerabilityMatchResolver{
		sentinelSvc:      r.sentinelSvc,
		autoindexSvc:     r.autoindexSvc,
		uploadSvc:        r.uploadSvc,
		policySvc:        r.policySvc,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		errTracer:        errTracer,
		bulkLoader:       bulkLoader,
		m:                match,
	}, nil
}

func (r *rootResolver) VulnerabilityMatchesGroupByRepository(ctx context.Context, args resolverstubs.GetVulnerabilityMatchGroupByRepositoryArgs) (_ resolverstubs.VulnerabilityMatchGroupByRepositoryConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.vulnerabilityMatchByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit := 50
	if args.First != nil {
		limit = int(*args.First)
	}

	offset := 0
	if args.After != nil {
		after, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}

		offset = after
	}

	repositoryName := ""
	if args.RepositoryName != nil {
		repositoryName = *args.RepositoryName
	}

	groupedMatches, totalCount, err := r.sentinelSvc.GetVulnerabilityMatchesCountByRepository(ctx, shared.GetVulnerabilityMatchesGroupByRepositoryArgs{
		Limit:          limit,
		Offset:         offset,
		RepositoryName: repositoryName,
	})
	if err != nil {
		return nil, err
	}

	return &vulnerabilityMatchGroupByRepositoryConnectionResolver{
		groupedMatches: groupedMatches,
		offset:         offset,
		totalCount:     totalCount,
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

func (r *vulnerabilityResolver) ID() graphql.ID     { return marshalVulnerabilityGQLID(r.v.ID) }
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

type bulkLoader struct {
	sync.RWMutex
	sentinelSvc *sentinel.Service
	ids         []int
	cache       map[int]shared.Vulnerability
}

func NewBulkLoader(sentinelSvc *sentinel.Service) *bulkLoader {
	return &bulkLoader{
		sentinelSvc: sentinelSvc,
		cache:       map[int]shared.Vulnerability{},
	}
}

func (l *bulkLoader) MarkVulnerability(id int) {
	l.Lock()
	l.ids = append(l.ids, id)
	l.Unlock()
}

func (l *bulkLoader) GetVulnerabilityByID(ctx context.Context, id int) (shared.Vulnerability, bool, error) {
	l.RLock()
	vulnerability, ok := l.cache[id]
	l.RUnlock()
	if ok {
		return vulnerability, true, nil
	}

	l.Lock()
	defer l.Unlock()

	if vulnerability, ok := l.cache[id]; ok {
		return vulnerability, true, nil
	}

	m := map[int]struct{}{}
	for _, x := range append(l.ids, id) {
		if _, ok := l.cache[x]; !ok {
			m[x] = struct{}{}
		}
	}
	ids := make([]int, 0, len(m))
	for x := range m {
		ids = append(ids, x)
	}
	sort.Ints(ids)

	vulnerabilities, err := l.sentinelSvc.GetVulnerabilitiesByIDs(ctx, ids...)
	if err != nil {
		return shared.Vulnerability{}, false, err
	}

	for _, vulnerability := range vulnerabilities {
		l.cache[vulnerability.ID] = vulnerability
	}
	l.ids = nil

	vulnerability, ok = l.cache[id]
	return vulnerability, ok, nil
}

type vulnerabilityMatchResolver struct {
	sentinelSvc      *sentinel.Service
	autoindexSvc     sharedresolvers.AutoIndexingService
	uploadSvc        sharedresolvers.UploadsService
	policySvc        sharedresolvers.PolicyService
	prefetcher       *sharedresolvers.Prefetcher
	locationResolver *sharedresolvers.CachedLocationResolver
	errTracer        *observation.ErrCollector
	bulkLoader       *bulkLoader
	m                shared.VulnerabilityMatch
}

func (r *vulnerabilityMatchResolver) ID() graphql.ID {
	return marshalVulnerabilityMatchGQLID(r.m.ID)
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
		r.autoindexSvc,
		r.uploadSvc,
		r.policySvc,
		r.prefetcher,
		r.locationResolver,
		r.errTracer,
		&upload,
		nil,
	)
}

//
//

func unmarshalVulnerabilityGQLID(id graphql.ID) (vulnerabilityID int, err error) {
	err = relay.UnmarshalSpec(id, &vulnerabilityID)
	return vulnerabilityID, err
}

func marshalVulnerabilityGQLID(vulnerabilityID int) graphql.ID {
	return relay.MarshalID("Vulnerability", vulnerabilityID)
}

func unmarshalVulnerabilityMatchGQLID(id graphql.ID) (vulnerabilityMatchID int, err error) {
	err = relay.UnmarshalSpec(id, &vulnerabilityMatchID)
	return vulnerabilityMatchID, err
}

func marshalVulnerabilityMatchGQLID(vulnerabilityMatchID int) graphql.ID {
	return relay.MarshalID("VulnerabilityMatch", vulnerabilityMatchID)
}

func marshalVulnerabilityMatchGroupByRepositoryGQLID(vulnerabilityMatchID int) graphql.ID {
	return relay.MarshalID("VulnerabilityMatchGroup", vulnerabilityMatchID)
}

//
//

type vulnerabilityConnectionResolver struct {
	vulnerabilities []shared.Vulnerability
	offset          int
	totalCount      int
}

func (r *vulnerabilityConnectionResolver) Nodes() []resolverstubs.VulnerabilityResolver {
	var resolvers []resolverstubs.VulnerabilityResolver
	for _, v := range r.vulnerabilities {
		resolvers = append(resolvers, &vulnerabilityResolver{v: v})
	}

	return resolvers
}

func (r *vulnerabilityConnectionResolver) TotalCount() *int32 {
	v := int32(r.totalCount)
	return &v
}

func (r *vulnerabilityConnectionResolver) PageInfo() resolverstubs.PageInfo {
	if r.offset+len(r.vulnerabilities) < r.totalCount {
		return sharedresolvers.NextPageCursor(strconv.Itoa(r.offset + len(r.vulnerabilities)))
	}

	return sharedresolvers.HasNextPage(false)
}

//
//

type vulnerabilityMatchConnectionResolver struct {
	sentinelSvc      *sentinel.Service
	autoindexSvc     sharedresolvers.AutoIndexingService
	uploadSvc        sharedresolvers.UploadsService
	policySvc        sharedresolvers.PolicyService
	prefetcher       *sharedresolvers.Prefetcher
	locationResolver *sharedresolvers.CachedLocationResolver
	errTracer        *observation.ErrCollector
	bulkLoader       *bulkLoader
	matches          []shared.VulnerabilityMatch
	offset           int
	totalCount       int
}

func (r *vulnerabilityMatchConnectionResolver) Nodes() []resolverstubs.VulnerabilityMatchResolver {
	var resolvers []resolverstubs.VulnerabilityMatchResolver
	for _, m := range r.matches {
		resolvers = append(resolvers, &vulnerabilityMatchResolver{
			sentinelSvc:      r.sentinelSvc,
			autoindexSvc:     r.autoindexSvc,
			uploadSvc:        r.uploadSvc,
			policySvc:        r.policySvc,
			prefetcher:       r.prefetcher,
			locationResolver: r.locationResolver,
			errTracer:        r.errTracer,
			bulkLoader:       r.bulkLoader,
			m:                m,
		})
	}

	return resolvers
}

func (r *vulnerabilityMatchConnectionResolver) TotalCount() *int32 {
	v := int32(r.totalCount)
	return &v
}

func (r *vulnerabilityMatchConnectionResolver) PageInfo() resolverstubs.PageInfo {
	if r.offset+len(r.matches) < r.totalCount {
		return sharedresolvers.NextPageCursor(strconv.Itoa(r.offset + len(r.matches)))
	}

	return sharedresolvers.HasNextPage(false)
}

//
//

type vulnerabilityMatchGroupByRepositoryResolver struct {
	v shared.VulnerabilityMatchesByRepository
}

func (v vulnerabilityMatchGroupByRepositoryResolver) ID() graphql.ID {
	return marshalVulnerabilityMatchGroupByRepositoryGQLID(v.v.ID)
}

func (v vulnerabilityMatchGroupByRepositoryResolver) RepositoryName() string {
	return v.v.RepositoryName
}

func (v vulnerabilityMatchGroupByRepositoryResolver) MatchCount() int32 {
	return v.v.MatchCount
}

//
//

type vulnerabilityMatchGroupByRepositoryConnectionResolver struct {
	groupedMatches []shared.VulnerabilityMatchesByRepository
	offset         int
	totalCount     int
}

func (v *vulnerabilityMatchGroupByRepositoryConnectionResolver) Nodes() []resolverstubs.VulnerabilityMatchGroupByRepositoryResolver {
	var resolvers []resolverstubs.VulnerabilityMatchGroupByRepositoryResolver
	for _, m := range v.groupedMatches {
		resolvers = append(resolvers, &vulnerabilityMatchGroupByRepositoryResolver{v: m})
	}

	return resolvers
}

func (v *vulnerabilityMatchGroupByRepositoryConnectionResolver) TotalCount() *int32 {
	c := int32(v.totalCount)
	return &c
}

func (v *vulnerabilityMatchGroupByRepositoryConnectionResolver) PageInfo() resolverstubs.PageInfo {
	if v.offset+len(v.groupedMatches) < v.totalCount {
		return sharedresolvers.NextPageCursor(strconv.Itoa(v.offset + len(v.groupedMatches)))
	}

	return sharedresolvers.HasNextPage(false)
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
