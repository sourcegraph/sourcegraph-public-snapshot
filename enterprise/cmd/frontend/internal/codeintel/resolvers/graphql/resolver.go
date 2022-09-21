package graphql

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	DefaultUploadPageSize                  = 50
	DefaultIndexPageSize                   = 50
	DefaultConfigurationPolicyPageSize     = 50
	DefaultRepositoryFilterPreviewPageSize = 50
	DefaultRetentionPolicyMatchesPageSize  = 50
)

var errAutoIndexingNotEnabled = errors.New("precise code intelligence auto-indexing is not enabled")

// Resolver is the main interface to code intel-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type Resolver struct {
	db                 database.DB
	gitserver          GitserverClient
	resolver           resolvers.Resolver
	locationResolver   *CachedLocationResolver
	observationContext *operations
}

// NewResolver creates a new Resolver with the given resolver that defines all code intel-specific behavior.
func NewResolver(db database.DB, gitserver GitserverClient, resolver resolvers.Resolver, observationContext *observation.Context) gql.CodeIntelResolver {
	baseResolver := &Resolver{
		db:                 db,
		gitserver:          gitserver,
		resolver:           resolver,
		locationResolver:   NewCachedLocationResolver(db),
		observationContext: newOperations(observationContext),
	}

	return &frankenResolver{
		Resolver: baseResolver,
	}
}

func (r *frankenResolver) NodeResolvers() map[string]gql.NodeByIDFunc {
	return map[string]gql.NodeByIDFunc{
		"LSIFUpload": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.Resolver.LSIFUploadByID(ctx, id)
		},
		"LSIFIndex": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.Resolver.LSIFIndexByID(ctx, id)
		},
		"CodeIntelligenceConfigurationPolicy": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.ConfigurationPolicyByID(ctx, id)
		},
	}
}

func (r *Resolver) ExecutorResolver() executor.Resolver {
	return r.resolver.ExecutorResolver()
}

// func (r *Resolver) CodeNavResolver() resolvers.CodeNavResolver {
// 	return r.resolver.CodeNavResolver()
// }

// func (r *Resolver) PoliciesResolver() resolvers.PoliciesResolver {
// 	return r.resolver.PoliciesResolver()
// }

// func (r *Resolver) AutoIndexingResolver() resolvers.AutoIndexingResolver {
// 	return r.resolver.AutoIndexingResolver()
// }

// 🚨 SECURITY: dbstore layer handles authz for GetUploadByID
func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFUploadResolver, err error) {
	return r.resolver.UploadRootResolver().LSIFUploadByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *Resolver) LSIFUploads(ctx context.Context, args *uploadsgraphql.LSIFUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.resolver.UploadRootResolver().LSIFUploads(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *Resolver) LSIFUploadsByRepo(ctx context.Context, args *uploadsgraphql.LSIFRepositoryUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.resolver.UploadRootResolver().LSIFUploadsByRepo(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *Resolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.UploadRootResolver().DeleteLSIFUpload(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexByID
func (r *Resolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFIndexResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().LSIFIndexByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *Resolver) LSIFIndexes(ctx context.Context, args *autoindexinggraphql.LSIFIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().LSIFIndexes(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *Resolver) LSIFIndexesByRepo(ctx context.Context, args *autoindexinggraphql.LSIFRepositoryIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().LSIFIndexesByRepo(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence index data
func (r *Resolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.AutoIndexingRootResolver().DeleteLSIFIndex(ctx, args)
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *Resolver) CommitGraph(ctx context.Context, id graphql.ID) (_ uploadsgraphql.CodeIntelligenceCommitGraphResolver, err error) {
	return r.resolver.UploadRootResolver().CommitGraph(ctx, id)
}

// 🚨 SECURITY: Only site admins may queue auto-index jobs
func (r *Resolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *autoindexinggraphql.QueueAutoIndexJobsForRepoArgs) (_ []sharedresolvers.LSIFIndexResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().QueueAutoIndexJobsForRepo(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for query resolution
func (r *Resolver) GitBlobLSIFData(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (_ gql.GitBlobLSIFDataResolver, err error) {
	ctx, errTracer, endObservation := r.observationContext.gitBlobLsifData.WithErrors(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	codenav := r.resolver.CodeNavResolver()
	gitBlobResolver, err := codenav.GitBlobLSIFDataResolverFactory(ctx, args.Repo, string(args.Commit), args.Path, args.ToolName, args.ExactPath)
	if err != nil || gitBlobResolver == nil {
		return nil, err
	}

	return NewQueryResolver(r.gitserver, gitBlobResolver, r.resolver, r.locationResolver, errTracer), nil
}

// move to autoindexing service
func (r *Resolver) GitBlobCodeIntelInfo(ctx context.Context, args *gql.GitTreeEntryCodeIntelInfoArgs) (_ gql.GitBlobCodeIntelSupportResolver, err error) {
	ctx, errTracer, endObservation := r.observationContext.gitBlobCodeIntelInfo.WithErrors(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	// move to autoindexing service
	return NewCodeIntelSupportResolver(r.resolver, args.Repo.Name, args.Path, errTracer), nil
}

// move to autoindexing service
func (r *Resolver) GitTreeCodeIntelInfo(ctx context.Context, args *gql.GitTreeEntryCodeIntelInfoArgs) (resolver gql.GitTreeCodeIntelSupportResolver, err error) {
	ctx, errTracer, endObservation := r.observationContext.gitBlobCodeIntelInfo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repoID", int(args.Repo.ID)),
		log.String("path", args.Path),
		log.String("commit", args.Commit),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	filesRegex, err := regexp.Compile("^" + regexp.QuoteMeta(args.Path) + "[^.]{1}[^/]*$")
	if err != nil {
		return nil, errors.Wrapf(err, "path '%s' caused invalid regex", args.Path)
	}

	files, err := r.gitserver.ListFiles(ctx, int(args.Repo.ID), args.Commit, filesRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "gitserver.ListFiles: error listing files at %s for repo %d", args.Path, args.Repo.ID)
	}

	return NewCodeIntelTreeInfoResolver(r.resolver, args.Repo, args.Commit, args.Path, files, errTracer), nil
}

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicyByID
func (r *Resolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ gql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, traceErrs, endObservation := r.observationContext.configurationPolicyByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	configurationPolicyID, err := unmarshalConfigurationPolicyGQLID(id)
	if err != nil {
		return nil, err
	}

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}
	configurationPolicy, exists, err := policyResolver.GetConfigurationPolicyByID(ctx, int(configurationPolicyID))
	if err != nil || !exists {
		return nil, err
	}
	cp := sharedConfigurationPoliciesToStoreConfigurationPolicies(configurationPolicy)

	return NewConfigurationPolicyResolver(r.db, cp, traceErrs), nil
}

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicies
func (r *Resolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *gql.CodeIntelligenceConfigurationPoliciesArgs) (_ gql.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	fields := []log.Field{}
	if args.Repository != nil {
		fields = append(fields, log.String("repoID", string(*args.Repository)))
	}
	ctx, traceErrs, endObservation := r.observationContext.configurationPolicies.WithErrors(ctx, &err, observation.Args{LogFields: fields})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	pageSize := DefaultConfigurationPolicyPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	opts := types.GetConfigurationPoliciesOptions{
		Limit:  pageSize,
		Offset: offset,
	}
	if args.Repository != nil {
		id64, err := unmarshalRepositoryID(*args.Repository)
		if err != nil {
			return nil, err
		}
		opts.RepositoryID = int(id64)
	}
	if args.Query != nil {
		opts.Term = *args.Query
	}
	if args.ForDataRetention != nil {
		opts.ForDataRetention = *args.ForDataRetention
	}
	if args.ForIndexing != nil {
		opts.ForIndexing = *args.ForIndexing
	}

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}
	policies, totalCount, err := policyResolver.GetConfigurationPolicies(ctx, opts)
	if err != nil {
		return nil, err
	}

	p := sharedConfigurationPoliciesListToStoreConfigurationPoliciesList(policies)
	return NewCodeIntelligenceConfigurationPolicyConnectionResolver(r.db, p, totalCount, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.CreateCodeIntelligenceConfigurationPolicyArgs) (_ gql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, traceErrs, endObservation := r.observationContext.createConfigurationPolicy.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if err := validateConfigurationPolicy(args.CodeIntelConfigurationPolicy); err != nil {
		return nil, err
	}

	var repositoryID *int
	if args.Repository != nil {
		id64, err := unmarshalRepositoryID(*args.Repository)
		if err != nil {
			return nil, err
		}

		id := int(id64)
		repositoryID = &id
	}

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}

	opts := types.ConfigurationPolicy{
		RepositoryID:              repositoryID,
		Name:                      args.Name,
		RepositoryPatterns:        args.RepositoryPatterns,
		Type:                      types.GitObjectType(args.Type),
		Pattern:                   args.Pattern,
		RetentionEnabled:          args.RetentionEnabled,
		RetentionDuration:         toDuration(args.RetentionDurationHours),
		RetainIntermediateCommits: args.RetainIntermediateCommits,
		IndexingEnabled:           args.IndexingEnabled,
		IndexCommitMaxAge:         toDuration(args.IndexCommitMaxAgeHours),
		IndexIntermediateCommits:  args.IndexIntermediateCommits,
	}
	configurationPolicy, err := policyResolver.CreateConfigurationPolicy(ctx, opts)
	if err != nil {
		return nil, err
	}

	cp := sharedConfigurationPoliciesToStoreConfigurationPolicies(configurationPolicy)
	return NewConfigurationPolicyResolver(r.db, cp, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.observationContext.updateConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(args.ID)),
	}})
	defer endObservation(1, observation.Args{})

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if err := validateConfigurationPolicy(args.CodeIntelConfigurationPolicy); err != nil {
		return nil, err
	}

	id, err := unmarshalConfigurationPolicyGQLID(args.ID)
	if err != nil {
		return nil, err
	}

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}
	opts := types.ConfigurationPolicy{
		ID:                        int(id),
		Name:                      args.Name,
		RepositoryPatterns:        args.RepositoryPatterns,
		Type:                      types.GitObjectType(args.Type),
		Pattern:                   args.Pattern,
		RetentionEnabled:          args.RetentionEnabled,
		RetentionDuration:         toDuration(args.RetentionDurationHours),
		RetainIntermediateCommits: args.RetainIntermediateCommits,
		IndexingEnabled:           args.IndexingEnabled,
		IndexCommitMaxAge:         toDuration(args.IndexCommitMaxAgeHours),
		IndexIntermediateCommits:  args.IndexIntermediateCommits,
	}
	if err := policyResolver.UpdateConfigurationPolicy(ctx, opts); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.observationContext.deleteConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(args.Policy)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmarshalConfigurationPolicyGQLID(args.Policy)
	if err != nil {
		return nil, err
	}

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}
	if err := policyResolver.DeleteConfigurationPolicyByID(ctx, int(id)); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) RepositorySummary(ctx context.Context, id graphql.ID) (_ sharedresolvers.CodeIntelRepositorySummaryResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().RepositorySummary(ctx, id)
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *Resolver) IndexConfiguration(ctx context.Context, id graphql.ID) (_ autoindexinggraphql.IndexConfigurationResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().IndexConfiguration(ctx, id)
}

// 🚨 SECURITY: Only site admins may modify code intelligence indexing configuration
func (r *Resolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *autoindexinggraphql.UpdateRepositoryIndexConfigurationArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.AutoIndexingRootResolver().UpdateRepositoryIndexConfiguration(ctx, args)
}

func (r *Resolver) PreviewRepositoryFilter(ctx context.Context, args *gql.PreviewRepositoryFilterArgs) (_ gql.RepositoryFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.observationContext.previewRepoFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	pageSize := DefaultRepositoryFilterPreviewPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}

	ids, totalCount, repositoryMatchLimit, err := policyResolver.GetPreviewRepositoryFilter(ctx, args.Patterns, pageSize, offset)
	if err != nil {
		return nil, err
	}

	resv := make([]*gql.RepositoryResolver, 0, len(ids))
	for _, id := range ids {
		repo, err := backend.NewRepos(r.locationResolver.logger, r.db).Get(ctx, api.RepoID(id))
		if err != nil {
			return nil, err
		}

		resv = append(resv, gql.NewRepositoryResolver(r.db, repo))
	}

	limitedCount := totalCount
	if repositoryMatchLimit != nil && *repositoryMatchLimit < limitedCount {
		limitedCount = *repositoryMatchLimit
	}

	return &repositoryFilterPreviewResolver{
		repositoryResolvers: resv,
		totalCount:          limitedCount,
		offset:              offset,
		totalMatches:        totalCount,
		limit:               repositoryMatchLimit,
	}, nil
}

func (r *Resolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *gql.PreviewGitObjectFilterArgs) (_ []gql.GitObjectFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.observationContext.previewGitObjectFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repositoryID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}

	namesByRev, err := policyResolver.GetPreviewGitObjectFilter(ctx, int(repositoryID), types.GitObjectType(args.Type), args.Pattern)
	if err != nil {
		return nil, err
	}

	var previews []gql.GitObjectFilterPreviewResolver
	for rev, names := range namesByRev {
		for _, name := range names {
			previews = append(previews, &gitObjectFilterPreviewResolver{
				name: name,
				rev:  rev,
			})
		}
	}

	sort.Slice(previews, func(i, j int) bool {
		return previews[i].Name() < previews[j].Name() || (previews[i].Name() == previews[j].Name() && previews[i].Rev() < previews[j].Rev())
	})

	return previews, nil
}

// makeGetUploadsOptions translates the given GraphQL arguments into options defined by the
// store.GetUploads operations.
func makeGetUploadsOptions(args *gql.LSIFRepositoryUploadsQueryArgs) (types.GetUploadsOptions, error) {
	repositoryID, err := resolveRepositoryID(args.RepositoryID)
	if err != nil {
		return types.GetUploadsOptions{}, err
	}

	var dependencyOf int64
	if args.DependencyOf != nil {
		dependencyOf, err = unmarshalLSIFUploadGQLID(*args.DependencyOf)
		if err != nil {
			return types.GetUploadsOptions{}, err
		}
	}

	var dependentOf int64
	if args.DependentOf != nil {
		dependentOf, err = unmarshalLSIFUploadGQLID(*args.DependentOf)
		if err != nil {
			return types.GetUploadsOptions{}, err
		}
	}

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return types.GetUploadsOptions{}, err
	}

	return types.GetUploadsOptions{
		RepositoryID:       repositoryID,
		State:              strings.ToLower(derefString(args.State, "")),
		Term:               derefString(args.Query, ""),
		VisibleAtTip:       derefBool(args.IsLatestForRepo, false),
		DependencyOf:       int(dependencyOf),
		DependentOf:        int(dependentOf),
		Limit:              derefInt32(args.First, DefaultUploadPageSize),
		Offset:             offset,
		AllowExpired:       true,
		AllowDeletedUpload: derefBool(args.IncludeDeleted, false),
	}, nil
}

// makeGetIndexesOptions translates the given GraphQL arguments into options defined by the
// store.GetIndexes operations.
func makeGetIndexesOptions(args *gql.LSIFRepositoryIndexesQueryArgs) (types.GetIndexesOptions, error) {
	repositoryID, err := resolveRepositoryID(args.RepositoryID)
	if err != nil {
		return types.GetIndexesOptions{}, err
	}

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return types.GetIndexesOptions{}, err
	}

	return types.GetIndexesOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(derefString(args.State, "")),
		Term:         derefString(args.Query, ""),
		Limit:        derefInt32(args.First, DefaultIndexPageSize),
		Offset:       offset,
	}, nil
}

// resolveRepositoryByID gets a repository's internal identifier from a GraphQL identifier.
func resolveRepositoryID(id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repoID, err := gql.UnmarshalRepositoryID(id)
	if err != nil {
		return 0, err
	}

	return int(repoID), nil
}

func validateConfigurationPolicy(policy gql.CodeIntelConfigurationPolicy) error {
	switch policy.Type {
	case gql.GitObjectTypeCommit:
	case gql.GitObjectTypeTag:
	case gql.GitObjectTypeTree:
	default:
		return errors.Errorf("illegal git object type '%s', expected 'GIT_COMMIT', 'GIT_TAG', or 'GIT_TREE'", policy.Type)
	}

	if policy.Name == "" {
		return errors.Errorf("no name supplied")
	}
	if policy.Pattern == "" {
		return errors.Errorf("no pattern supplied")
	}
	if policy.Type == gql.GitObjectTypeCommit && policy.Pattern != "HEAD" {
		return errors.Errorf("pattern must be HEAD for policy type 'GIT_COMMIT'")
	}
	if policy.RetentionDurationHours != nil && *policy.RetentionDurationHours <= 0 {
		return errors.Errorf("illegal retention duration '%d'", *policy.RetentionDurationHours)
	}
	if policy.IndexCommitMaxAgeHours != nil && *policy.IndexCommitMaxAgeHours <= 0 {
		return errors.Errorf("illegal index commit max age '%d'", *policy.IndexCommitMaxAgeHours)
	}

	return nil
}

func toDuration(hours *int32) *time.Duration {
	if hours == nil {
		return nil
	}

	v := time.Duration(*hours) * time.Hour
	return &v
}
