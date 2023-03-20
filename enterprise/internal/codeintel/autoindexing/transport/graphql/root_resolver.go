package graphql

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type rootResolver struct {
	autoindexSvc            AutoIndexingService
	uploadSvc               UploadsService
	policySvc               PolicyService
	gitserverClient         gitserver.Client
	operations              *operations
	siteAdminChecker        sharedresolvers.SiteAdminChecker
	repoStore               database.RepoStore
	prefetcherFactory       *sharedresolvers.PrefetcherFactory
	locationResolverFactory *sharedresolvers.CachedLocationResolverFactory
}

func NewRootResolver(
	observationCtx *observation.Context,
	autoindexSvc AutoIndexingService,
	uploadSvc UploadsService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	repoStore database.RepoStore,
	prefetcherFactory *sharedresolvers.PrefetcherFactory,
	locationResolverFactory *sharedresolvers.CachedLocationResolverFactory,
) resolverstubs.AutoindexingServiceResolver {
	return &rootResolver{
		autoindexSvc:            autoindexSvc,
		uploadSvc:               uploadSvc,
		policySvc:               policySvc,
		gitserverClient:         gitserverClient,
		operations:              newOperations(observationCtx),
		siteAdminChecker:        siteAdminChecker,
		repoStore:               repoStore,
		prefetcherFactory:       prefetcherFactory,
		locationResolverFactory: locationResolverFactory,
	}
}

var (
	autoIndexingEnabled       = conf.CodeIntelAutoIndexingEnabled
	errAutoIndexingNotEnabled = errors.New("precise code intelligence auto-indexing is not enabled")
)

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *rootResolver) IndexConfiguration(ctx context.Context, repoID graphql.ID) (_ resolverstubs.IndexConfigurationResolver, err error) {
	_, traceErrs, endObservation := r.operations.indexConfiguration.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(repoID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := resolverstubs.UnmarshalID[api.RepoID](repoID)
	if err != nil {
		return nil, err
	}

	return NewIndexConfigurationResolver(r.autoindexSvc, r.siteAdminChecker, int(repositoryID), traceErrs), nil
}

// 🚨 SECURITY: Only site admins may infer auto-index jobs
func (r *rootResolver) InferAutoIndexJobsForRepo(ctx context.Context, args *resolverstubs.InferAutoIndexJobsForRepoArgs) (_ []resolverstubs.AutoIndexJobDescriptionResolver, err error) {
	ctx, _, endObservation := r.operations.inferAutoIndexJobsForRepo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repository", string(args.Repository)),
		log.String("rev", resolverstubs.Deref(args.Rev, "")),
		log.String("script", resolverstubs.Deref(args.Script, "")),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := resolverstubs.UnmarshalID[api.RepoID](args.Repository)
	if err != nil {
		return nil, err
	}

	rev := "HEAD"
	if args.Rev != nil {
		rev = *args.Rev
	}

	localOverrideScript := ""
	if args.Script != nil {
		localOverrideScript = *args.Script
	}

	// TODO - expose hints
	config, _, err := r.autoindexSvc.InferIndexConfiguration(ctx, int(repositoryID), rev, localOverrideScript, false)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return newDescriptionResolvers(r.siteAdminChecker, config)
}

type autoIndexJobDescriptionResolver struct {
	siteAdminChecker sharedresolvers.SiteAdminChecker
	indexJob         config.IndexJob
	steps            []types.DockerStep
}

func (r *autoIndexJobDescriptionResolver) Root() string {
	return r.indexJob.Root
}

func (r *autoIndexJobDescriptionResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	return types.NewCodeIntelIndexerResolver(r.indexJob.Indexer, r.indexJob.Indexer)
}

func (r *autoIndexJobDescriptionResolver) ComparisonKey() string {
	return comparisonKey(r.indexJob.Root, r.Indexer().Name())
}

func comparisonKey(root, indexer string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(strings.Join([]string{root, indexer}, "\x00")))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func (r *autoIndexJobDescriptionResolver) Steps() resolverstubs.IndexStepsResolver {
	return sharedresolvers.NewIndexStepsResolver(r.siteAdminChecker, types.Index{
		DockerSteps:      r.steps,
		LocalSteps:       r.indexJob.LocalSteps,
		Root:             r.indexJob.Root,
		Indexer:          r.indexJob.Indexer,
		IndexerArgs:      r.indexJob.IndexerArgs,
		Outfile:          r.indexJob.Outfile,
		RequestedEnvVars: r.indexJob.RequestedEnvVars,
	})
}

// 🚨 SECURITY: Only site admins may queue auto-index jobs
func (r *rootResolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *resolverstubs.QueueAutoIndexJobsForRepoArgs) (_ []resolverstubs.PreciseIndexResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.queueAutoIndexJobsForRepo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repository", string(args.Repository)),
		log.String("rev", resolverstubs.Deref(args.Rev, "")),
		log.String("configuration", resolverstubs.Deref(args.Configuration, "")),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := resolverstubs.UnmarshalID[api.RepoID](args.Repository)
	if err != nil {
		return nil, err
	}

	rev := "HEAD"
	if args.Rev != nil {
		rev = *args.Rev
	}

	configuration := ""
	if args.Configuration != nil {
		configuration = *args.Configuration
	}

	indexes, err := r.autoindexSvc.QueueIndexes(ctx, int(repositoryID), rev, configuration, true, true)
	if err != nil {
		return nil, err
	}

	prefetcher := r.prefetcherFactory.Create()

	for _, index := range indexes {
		prefetcher.MarkIndex(index.ID)
	}

	resolvers := make([]resolverstubs.PreciseIndexResolver, 0, len(indexes))
	for _, index := range indexes {
		index := index
		resolver, err := sharedresolvers.NewPreciseIndexResolver(ctx, r.uploadSvc, r.policySvc, r.gitserverClient, prefetcher, r.siteAdminChecker, r.repoStore, r.locationResolverFactory.Create(), traceErrs, nil, &index)
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, resolver)
	}

	return resolvers, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence indexing configuration
func (r *rootResolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *resolverstubs.UpdateRepositoryIndexConfigurationArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateIndexConfiguration.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repository", string(args.Repository)),
		log.String("configuration", args.Configuration),
	}})
	defer endObservation(1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := resolverstubs.UnmarshalID[int](args.Repository)
	if err != nil {
		return nil, err
	}

	if _, err := config.UnmarshalJSON([]byte(args.Configuration)); err != nil {
		return nil, err
	}

	if err := r.autoindexSvc.UpdateIndexConfigurationByRepositoryID(ctx, repositoryID, []byte(args.Configuration)); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, nil
}

func (r *rootResolver) CodeIntelligenceInferenceScript(ctx context.Context) (script string, err error) {
	ctx, _, endObservation := r.operations.codeIntelligenceInferenceScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.GetInferenceScript(ctx)
}

func (r *rootResolver) UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *resolverstubs.UpdateCodeIntelligenceInferenceScriptArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateCodeIntelligenceInferenceScript.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("script", args.Script),
	}})
	defer endObservation(1, observation.Args{})

	return resolverstubs.Empty, r.autoindexSvc.SetInferenceScript(ctx, args.Script)
}

func (r *rootResolver) GitBlobCodeIntelInfo(ctx context.Context, args *resolverstubs.GitTreeEntryCodeIntelInfoArgs) (_ resolverstubs.GitBlobCodeIntelSupportResolver, err error) {
	ctx, errTracer, endObservation := r.operations.gitBlobCodeIntelInfo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repoID", int(args.Repo.ID)),
		log.String("commit", args.Commit),
		log.String("path", args.Path),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	return NewCodeIntelSupportResolver(r.autoindexSvc, args.Repo.Name, args.Path, errTracer), nil
}

func (r *rootResolver) GitTreeCodeIntelInfo(ctx context.Context, args *resolverstubs.GitTreeEntryCodeIntelInfoArgs) (resolver resolverstubs.GitTreeCodeIntelSupportResolver, err error) {
	ctx, errTracer, endObservation := r.operations.gitBlobCodeIntelInfo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repoID", int(args.Repo.ID)),
		log.String("commit", args.Commit),
		log.String("path", args.Path),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	filesRegex, err := regexp.Compile("^" + regexp.QuoteMeta(args.Path) + "[^.]{1}[^/]*$")
	if err != nil {
		return nil, errors.Wrapf(err, "path '%s' caused invalid regex", args.Path)
	}

	files, err := r.gitserverClient.ListFiles(ctx, authz.DefaultSubRepoPermsChecker, args.Repo.Name, api.CommitID(args.Commit), filesRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "gitserver.ListFiles: error listing files at %s for repo %d", args.Path, args.Repo.ID)
	}

	return NewCodeIntelTreeInfoResolver(r.autoindexSvc, args.Repo, args.Commit, args.Path, files, errTracer), nil
}

func (r *rootResolver) GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error) {
	ctx, _, endObservation := r.operations.getRecentIndexesSummary.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.GetRecentIndexesSummary(ctx, repositoryID)
}

func (r *rootResolver) GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := r.operations.getLastIndexScanForRepository.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.GetLastIndexScanForRepository(ctx, repositoryID)
}

func (r *rootResolver) CodeIntelSummary(ctx context.Context) (_ resolverstubs.CodeIntelSummaryResolver, err error) {
	ctx, _, endObservation := r.operations.summary.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	return sharedresolvers.NewSummaryResolver(r.autoindexSvc, r.locationResolverFactory.Create()), nil
}

func (r *rootResolver) RepositorySummary(ctx context.Context, repoID graphql.ID) (_ resolverstubs.CodeIntelRepositorySummaryResolver, err error) {
	ctx, errTracer, endObservation := r.operations.repositorySummary.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(repoID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	id, err := resolverstubs.UnmarshalID[int](repoID)
	if err != nil {
		return nil, err
	}

	lastUploadRetentionScan, err := r.uploadSvc.GetLastUploadRetentionScanForRepository(ctx, id)
	if err != nil {
		return nil, err
	}

	lastIndexScan, err := r.autoindexSvc.GetLastIndexScanForRepository(ctx, id)
	if err != nil {
		return nil, err
	}

	recentUploads, err := r.uploadSvc.GetRecentUploadsSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	recentIndexes, err := r.autoindexSvc.GetRecentIndexesSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create blocklist for indexes that have already been uploaded.
	blocklist := map[string]struct{}{}
	for _, u := range recentUploads {
		key := shared.GetKeyForLookup(u.Indexer, u.Root)
		blocklist[key] = struct{}{}
	}
	for _, u := range recentIndexes {
		key := shared.GetKeyForLookup(u.Indexer, u.Root)
		blocklist[key] = struct{}{}
	}

	commit := "HEAD"
	var limitErr error

	indexJobs, err := r.autoindexSvc.InferIndexJobsFromRepositoryStructure(ctx, id, commit, "", false)
	if err != nil {
		if !errors.As(err, &inference.LimitError{}) {
			return nil, err
		}

		limitErr = errors.Append(limitErr, err)
	}
	// indexJobHints, err := r.autoindexSvc.InferIndexJobHintsFromRepositoryStructure(ctx, repoID, commit)
	// if err != nil {
	// 	if !errors.As(err, &inference.LimitError{}) {
	// 		return nil, err
	// 	}

	// 	limitErr = errors.Append(limitErr, err)
	// }

	inferredAvailableIndexers := map[string]shared.AvailableIndexer{}
	inferredAvailableIndexers = shared.PopulateInferredAvailableIndexers(indexJobs, blocklist, inferredAvailableIndexers)
	// inferredAvailableIndexers = shared.PopulateInferredAvailableIndexers(indexJobHints, blocklist, inferredAvailableIndexers)

	inferredAvailableIndexersResolver := make([]sharedresolvers.InferredAvailableIndexers, 0, len(inferredAvailableIndexers))
	for _, indexer := range inferredAvailableIndexers {
		inferredAvailableIndexersResolver = append(inferredAvailableIndexersResolver,
			sharedresolvers.InferredAvailableIndexers{
				Indexer: indexer.Indexer,
				Roots:   indexer.Roots,
			},
		)
	}

	summary := sharedresolvers.RepositorySummary{
		RecentUploads:           recentUploads,
		RecentIndexes:           recentIndexes,
		LastUploadRetentionScan: lastUploadRetentionScan,
		LastIndexScan:           lastIndexScan,
	}

	return sharedresolvers.NewRepositorySummaryResolver(
		r.uploadSvc,
		r.policySvc,
		r.gitserverClient,
		r.siteAdminChecker,
		r.repoStore,
		r.locationResolverFactory.Create(),
		summary,
		inferredAvailableIndexersResolver,
		limitErr,
		r.prefetcherFactory.Create(),
		errTracer,
	), nil
}
