package graphql

import (
	"context"
	"time"

	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RootResolver interface {
	// Mirrors AutoindexingServiceResolver in graphqlbackend
	IndexConfiguration(ctx context.Context, id graphql.ID) (IndexConfigurationResolver, error) // TODO - rename ...ForRepo
	DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*sharedresolvers.EmptyResponse, error)
	DeleteLSIFIndexes(ctx context.Context, args *DeleteLSIFIndexesArgs) (*sharedresolvers.EmptyResponse, error)
	LSIFIndexByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFIndexResolver, err error)
	LSIFIndexes(ctx context.Context, args *LSIFIndexesQueryArgs) (sharedresolvers.LSIFIndexConnectionResolver, error)
	LSIFIndexesByRepo(ctx context.Context, args *LSIFRepositoryIndexesQueryArgs) (sharedresolvers.LSIFIndexConnectionResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, args *QueueAutoIndexJobsForRepoArgs) ([]sharedresolvers.LSIFIndexResolver, error)
	UpdateRepositoryIndexConfiguration(ctx context.Context, args *UpdateRepositoryIndexConfigurationArgs) (*sharedresolvers.EmptyResponse, error)
	RepositorySummary(ctx context.Context, id graphql.ID) (_ sharedresolvers.CodeIntelRepositorySummaryResolver, err error)
	GitBlobCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (_ GitBlobCodeIntelSupportResolver, err error)
	GitTreeCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (resolver GitTreeCodeIntelSupportResolver, err error)
	RequestLanguageSupport(ctx context.Context, args *RequestLanguageSupportArgs) (_ *sharedresolvers.EmptyResponse, err error)
	RequestedLanguageSupport(ctx context.Context) (_ []string, err error)

	// AutoIndexing
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	InferedIndexConfiguration(ctx context.Context, repositoryID int, commit string) (_ *config.IndexConfiguration, _ bool, err error)
	InferedIndexConfigurationHints(ctx context.Context, repositoryID int, commit string) (_ []config.IndexJobHint, err error)
	CodeIntelligenceInferenceScript(ctx context.Context) (script string, err error)
	UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *UpdateCodeIntelligenceInferenceScriptArgs) (err error)

	// Symbols client
	GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error)
}

type rootResolver struct {
	autoindexSvc AutoIndexingService
	uploadSvc    UploadsService
	policySvc    PolicyService
	operations   *operations
}

func NewRootResolver(autoindexSvc AutoIndexingService, uploadSvc UploadsService, policySvc PolicyService, observationContext *observation.Context) RootResolver {
	return &rootResolver{
		autoindexSvc: autoindexSvc,
		uploadSvc:    uploadSvc,
		policySvc:    policySvc,
		operations:   newOperations(observationContext),
	}
}

var (
	autoIndexingEnabled       = conf.CodeIntelAutoIndexingEnabled
	errAutoIndexingNotEnabled = errors.New("precise code intelligence auto-indexing is not enabled")
)

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *rootResolver) IndexConfiguration(ctx context.Context, id graphql.ID) (_ IndexConfigurationResolver, err error) {
	_, traceErrs, endObservation := r.operations.indexConfiguration.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := UnmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}

	return NewIndexConfigurationResolver(r.autoindexSvc, int(repositoryID), traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence index data
func (r *rootResolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteLsifIndex.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("indexID", string(args.ID)),
	}})
	defer endObservation(1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	indexID, err := unmarshalLSIFIndexGQLID(args.ID)
	if err != nil {
		return nil, err
	}

	if _, err := r.autoindexSvc.DeleteIndexByID(ctx, int(indexID)); err != nil {
		return nil, err
	}

	return &sharedresolvers.EmptyResponse{}, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *rootResolver) DeleteLSIFIndexes(ctx context.Context, args *DeleteLSIFIndexesArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteLsifIndexes.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	opts, err := makeDeleteIndexesOptions(args)
	if err != nil {
		return nil, err
	}
	if err := r.autoindexSvc.DeleteIndexes(ctx, opts); err != nil {
		return nil, err
	}

	return &sharedresolvers.EmptyResponse{}, nil
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexByID
func (r *rootResolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFIndexResolver, err error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	ctx, traceErrs, endObservation := r.operations.lsifIndexByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("indexID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	indexID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)

	index, exists, err := prefetcher.GetIndexByID(ctx, int(indexID))
	if err != nil || !exists {
		return nil, err
	}

	return sharedresolvers.NewIndexResolver(r.autoindexSvc, r.uploadSvc, r.policySvc, index, prefetcher, traceErrs), nil
}

type LSIFIndexesQueryArgs struct {
	ConnectionArgs
	Query *string
	State *string
	After *string
}

type LSIFRepositoryIndexesQueryArgs struct {
	*LSIFIndexesQueryArgs
	RepositoryID graphql.ID
}

type DeleteLSIFIndexesArgs struct {
	Query      *string
	State      *string
	Repository *graphql.ID
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *rootResolver) LSIFIndexes(ctx context.Context, args *LSIFIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	ctx, _, endObservation := r.operations.lsifIndexes.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	// Delegate behavior to LSIFIndexesByRepo with no specified repository identifier
	return r.LSIFIndexesByRepo(ctx, &LSIFRepositoryIndexesQueryArgs{LSIFIndexesQueryArgs: args})
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *rootResolver) LSIFIndexesByRepo(ctx context.Context, args *LSIFRepositoryIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	ctx, traceErrs, endObservation := r.operations.lsifIndexesByRepo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(args.RepositoryID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	opts, err := makeGetIndexesOptions(args)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)

	// Create a new indexConnectionResolver here as we only want to index records in
	// the same graphQL request, not across different request.
	// indexConnectionResolver := r.resolver.AutoIndexingResolver().IndexConnectionResolverFromFactory(opts)
	indexConnectionResolver := sharedresolvers.NewIndexesResolver(r.autoindexSvc, opts)

	return sharedresolvers.NewIndexConnectionResolver(r.autoindexSvc, r.uploadSvc, r.policySvc, indexConnectionResolver, prefetcher, traceErrs), nil
}

type QueueAutoIndexJobsForRepoArgs struct {
	Repository    graphql.ID
	Rev           *string
	Configuration *string
}

// 🚨 SECURITY: Only site admins may queue auto-index jobs
func (r *rootResolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *QueueAutoIndexJobsForRepoArgs) (_ []sharedresolvers.LSIFIndexResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.queueAutoIndexJobsForRepo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(args.Repository)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := UnmarshalRepositoryID(args.Repository)
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

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)

	lsifIndexResolvers := make([]sharedresolvers.LSIFIndexResolver, 0, len(indexes))
	for i := range indexes {
		lsifIndexResolvers = append(lsifIndexResolvers, sharedresolvers.NewIndexResolver(r.autoindexSvc, r.uploadSvc, r.policySvc, indexes[i], prefetcher, traceErrs))
	}

	return lsifIndexResolvers, nil
}

type UpdateRepositoryIndexConfigurationArgs struct {
	Repository    graphql.ID
	Configuration string
}

// 🚨 SECURITY: Only site admins may modify code intelligence indexing configuration
func (r *rootResolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *UpdateRepositoryIndexConfigurationArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateIndexConfiguration.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(args.Repository)),
	}})
	defer endObservation(1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.autoindexSvc.GetUnsafeDB()); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := unmarshalLSIFIndexGQLID(args.Repository)
	if err != nil {
		return nil, err
	}

	if _, err := config.UnmarshalJSON([]byte(args.Configuration)); err != nil {
		return nil, err
	}

	if err := r.autoindexSvc.UpdateIndexConfigurationByRepositoryID(ctx, int(repositoryID), []byte(args.Configuration)); err != nil {
		return nil, err
	}

	return &sharedresolvers.EmptyResponse{}, nil
}

func (r *rootResolver) CodeIntelligenceInferenceScript(ctx context.Context) (script string, err error) {
	return r.autoindexSvc.GetInferenceScript(ctx)
}

type UpdateCodeIntelligenceInferenceScriptArgs struct {
	Script string
}

func (r *rootResolver) UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *UpdateCodeIntelligenceInferenceScriptArgs) (err error) {
	return r.autoindexSvc.SetInferenceScript(ctx, args.Script)
}

type GitTreeEntryCodeIntelInfoArgs struct {
	Repo   *types.Repo
	Path   string
	Commit string
}

func (r *rootResolver) GitBlobCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (_ GitBlobCodeIntelSupportResolver, err error) {
	ctx, errTracer, endObservation := r.operations.gitBlobCodeIntelInfo.WithErrors(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	return NewCodeIntelSupportResolver(r.autoindexSvc, args.Repo.Name, args.Path, errTracer), nil
}

func (r *rootResolver) GitTreeCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (resolver GitTreeCodeIntelSupportResolver, err error) {
	ctx, errTracer, endObservation := r.operations.gitBlobCodeIntelInfo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repoID", int(args.Repo.ID)),
		log.String("path", args.Path),
		log.String("commit", args.Commit),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	filesRegex, err := regexp.Compile("^" + regexp.QuoteMeta(args.Path) + "[^.]{1}[^/]*$")
	if err != nil {
		return nil, errors.Wrapf(err, "path '%s' caused invalid regex", args.Path)
	}

	files, err := r.autoindexSvc.ListFiles(ctx, int(args.Repo.ID), args.Commit, filesRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "gitserver.ListFiles: error listing files at %s for repo %d", args.Path, args.Repo.ID)
	}

	return NewCodeIntelTreeInfoResolver(r.autoindexSvc, args.Repo, args.Commit, args.Path, files, errTracer), nil
}

func (r *rootResolver) GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error) {
	ctx, _, endObservation := r.operations.getRecentIndexesSummary.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.GetRecentIndexesSummary(ctx, repositoryID)
}

func (r *rootResolver) GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := r.operations.getLastIndexScanForRepository.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.GetLastIndexScanForRepository(ctx, repositoryID)
}

func (r *rootResolver) InferedIndexConfiguration(ctx context.Context, repositoryID int, commit string) (_ *config.IndexConfiguration, _ bool, err error) {
	ctx, _, endObservation := r.operations.inferedIndexConfiguration.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID), log.String("commit", commit)},
	})
	defer endObservation(1, observation.Args{})

	maybeConfig, _, err := r.autoindexSvc.InferIndexConfiguration(ctx, repositoryID, commit, true)
	if err != nil || maybeConfig == nil {
		return nil, false, err
	}

	return maybeConfig, true, nil
}

func (r *rootResolver) InferedIndexConfigurationHints(ctx context.Context, repositoryID int, commit string) (_ []config.IndexJobHint, err error) {
	ctx, _, endObservation := r.operations.inferedIndexConfigurationHints.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID), log.String("commit", commit)},
	})
	defer endObservation(1, observation.Args{})

	_, hints, err := r.autoindexSvc.InferIndexConfiguration(ctx, repositoryID, commit, true)
	if err != nil {
		return nil, err
	}

	return hints, nil
}

func (r *rootResolver) RepositorySummary(ctx context.Context, id graphql.ID) (_ sharedresolvers.CodeIntelRepositorySummaryResolver, err error) {
	ctx, errTracer, endObservation := r.operations.repositorySummary.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	repositoryID, err := UnmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}
	repoID := int(repositoryID)

	// uploadResolver := r.resolver.UploadsResolver()
	recentUploads, err := r.uploadSvc.GetRecentUploadsSummary(ctx, repoID)
	if err != nil {
		return nil, err
	}

	lastUploadRetentionScan, err := r.uploadSvc.GetLastUploadRetentionScanForRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	// recentIndexes, err := r.resolver.AutoIndexingRootResolver().GetRecentIndexesSummary(ctx, repoID)
	recentIndexes, err := r.autoindexSvc.GetRecentIndexesSummary(ctx, repoID)
	if err != nil {
		return nil, err
	}

	// lastIndexScan, err := r.resolver.AutoIndexingRootResolver().GetLastIndexScanForRepository(ctx, repoID)
	lastIndexScan, err := r.autoindexSvc.GetLastIndexScanForRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	summary := sharedresolvers.RepositorySummary{
		RecentUploads:           recentUploads,
		RecentIndexes:           recentIndexes,
		LastUploadRetentionScan: lastUploadRetentionScan,
		LastIndexScan:           lastIndexScan,
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := sharedresolvers.NewPrefetcher(r.autoindexSvc, r.uploadSvc)

	return sharedresolvers.NewRepositorySummaryResolver(r.autoindexSvc, r.uploadSvc, r.policySvc, summary, prefetcher, errTracer), nil
}

// HERE HERE HERE
func (r *rootResolver) GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (_ bool, _ string, err error) {
	ctx, _, endObservation := r.operations.getSupportedByCtags.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("repoName", string(repoName))},
	})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.GetSupportedByCtags(ctx, filepath, repoName)
}

type RequestLanguageSupportArgs struct {
	Language string
}

func (r *rootResolver) RequestLanguageSupport(ctx context.Context, args *RequestLanguageSupportArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.requestLanguageSupport.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	userID := int(actor.FromContext(ctx).UID)
	if userID == 0 {
		return nil, errors.Newf("language support requests only logged for authenticated users")
	}

	if err := r.autoindexSvc.SetRequestLanguageSupport(ctx, userID, args.Language); err != nil {
		return nil, err
	}

	return &sharedresolvers.EmptyResponse{}, nil
}

func (r *rootResolver) SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error) {
	ctx, _, endObservation := r.operations.setRequestLanguageSupport.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("userID", userID), log.String("language", language)},
	})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.SetRequestLanguageSupport(ctx, userID, language)
}

func (r *rootResolver) RequestedLanguageSupport(ctx context.Context) (_ []string, err error) {
	ctx, _, endObservation := r.operations.requestedLanguageSupport.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	userID := int(actor.FromContext(ctx).UID)
	if userID == 0 {
		return nil, errors.Newf("language support requests only logged for authenticated users")
	}

	return r.autoindexSvc.GetLanguagesRequestedBy(ctx, userID)
}
