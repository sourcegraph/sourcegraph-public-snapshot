package resolvers

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Resolver is the main interface to code intel-related operations exposed to the GraphQL API.
// This resolver consolidates the logic for code intel operations and is not itself concerned
// with GraphQL/API specifics (auth, validation, marshaling, etc.). This resolver is wrapped
// by a symmetrics resolver in this package's graphql subpackage, which is exposed directly
// by the API.
type Resolver interface {
	GetUploadByID(ctx context.Context, id int) (store.Upload, bool, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) ([]store.Upload, error)
	DeleteUploadByID(ctx context.Context, uploadID int) error

	GetIndexByID(ctx context.Context, id int) (store.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]store.Index, error)
	DeleteIndexByID(ctx context.Context, id int) error

	GetConfigurationPolicies(ctx context.Context, opts store.GetConfigurationPoliciesOptions) ([]store.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (store.ConfigurationPolicy, bool, error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy store.ConfigurationPolicy) (store.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy store.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error)

	IndexConfiguration(ctx context.Context, repositoryID int) ([]byte, bool, error)
	InferredIndexConfiguration(ctx context.Context, repositoryID int) (*config.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, configuration string) error

	CommitGraph(ctx context.Context, repositoryID int) (gql.CodeIntelligenceCommitGraphResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, repositoryID int, rev, configuration string) ([]store.Index, error)
	PreviewRepositoryFilter(ctx context.Context, patterns []string, limit, offset int) (_ []int, totalCount int, repositoryMatchLimit *int, _ error)
	PreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType store.GitObjectType, pattern string) (map[string][]string, error)
	DocumentationSearch(ctx context.Context, query string, repos []string) ([]precise.DocumentationSearchResult, error)

	UploadConnectionResolver(opts store.GetUploadsOptions) *UploadsResolver
	IndexConnectionResolver(opts store.GetIndexesOptions) *IndexesResolver
	QueryResolver(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (QueryResolver, error)

	ExecutorResolver() executor.Resolver
}

type resolver struct {
	dbStore          DBStore
	lsifStore        LSIFStore
	gitserverClient  GitserverClient
	policyMatcher    *policies.Matcher
	indexEnqueuer    IndexEnqueuer
	hunkCache        HunkCache
	operations       *operations
	executorResolver executor.Resolver
}

// NewResolver creates a new resolver with the given services.
func NewResolver(
	dbStore DBStore,
	lsifStore LSIFStore,
	gitserverClient GitserverClient,
	policyMatcher *policies.Matcher,
	indexEnqueuer IndexEnqueuer,
	hunkCache HunkCache,
	observationContext *observation.Context,
	dbConn dbutil.DB,
) Resolver {
	return newResolver(dbStore, lsifStore, gitserverClient, policyMatcher, indexEnqueuer, hunkCache, observationContext, dbConn)
}

func newResolver(
	dbStore DBStore,
	lsifStore LSIFStore,
	gitserverClient GitserverClient,
	policyMatcher *policies.Matcher,
	indexEnqueuer IndexEnqueuer,
	hunkCache HunkCache,
	observationContext *observation.Context,
	dbConn dbutil.DB,
) *resolver {
	return &resolver{
		dbStore:          dbStore,
		lsifStore:        lsifStore,
		gitserverClient:  gitserverClient,
		policyMatcher:    policyMatcher,
		indexEnqueuer:    indexEnqueuer,
		hunkCache:        hunkCache,
		operations:       newOperations(observationContext),
		executorResolver: executor.New(dbConn),
	}
}

func (r *resolver) ExecutorResolver() executor.Resolver {
	return r.executorResolver
}

func (r *resolver) GetUploadByID(ctx context.Context, id int) (store.Upload, bool, error) {
	return r.dbStore.GetUploadByID(ctx, id)
}

func (r *resolver) GetIndexByID(ctx context.Context, id int) (store.Index, bool, error) {
	return r.dbStore.GetIndexByID(ctx, id)
}

func (r *resolver) GetUploadsByIDs(ctx context.Context, ids ...int) ([]store.Upload, error) {
	return r.dbStore.GetUploadsByIDs(ctx, ids...)
}

func (r *resolver) GetIndexesByIDs(ctx context.Context, ids ...int) ([]store.Index, error) {
	return r.dbStore.GetIndexesByIDs(ctx, ids...)
}

func (r *resolver) UploadConnectionResolver(opts store.GetUploadsOptions) *UploadsResolver {
	return NewUploadsResolver(r.dbStore, opts)
}

func (r *resolver) IndexConnectionResolver(opts store.GetIndexesOptions) *IndexesResolver {
	return NewIndexesResolver(r.dbStore, opts)
}

func (r *resolver) DeleteUploadByID(ctx context.Context, uploadID int) error {
	_, err := r.dbStore.DeleteUploadByID(ctx, uploadID)
	return err
}

func (r *resolver) DeleteIndexByID(ctx context.Context, id int) error {
	_, err := r.dbStore.DeleteIndexByID(ctx, id)
	return err
}

func (r *resolver) CommitGraph(ctx context.Context, repositoryID int) (gql.CodeIntelligenceCommitGraphResolver, error) {
	stale, updatedAt, err := r.dbStore.CommitGraphMetadata(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return NewCommitGraphResolver(stale, updatedAt), nil
}

func (r *resolver) QueueAutoIndexJobsForRepo(ctx context.Context, repositoryID int, rev, configuration string) ([]store.Index, error) {
	return r.indexEnqueuer.QueueIndexes(ctx, repositoryID, rev, configuration, true)
}

const slowQueryResolverRequestThreshold = time.Second

// QueryResolver determines the set of dumps that can answer code intel queries for the
// given repository, commit, and path, then constructs a new query resolver instance which
// can be used to answer subsequent queries.
func (r *resolver) QueryResolver(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (_ QueryResolver, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, "QueryResolver", r.operations.queryResolver, slowQueryResolverRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", int(args.Repo.ID)),
			log.String("commit", string(args.Commit)),
			log.String("path", args.Path),
			log.Bool("exactPath", args.ExactPath),
			log.String("indexer", args.ToolName),
		},
	})
	defer endObservation()

	cachedCommitChecker := newCachedCommitChecker(r.gitserverClient)
	cachedCommitChecker.set(int(args.Repo.ID), string(args.Commit))

	dumps, err := r.findClosestDumps(
		ctx,
		cachedCommitChecker,
		int(args.Repo.ID),
		string(args.Commit),
		args.Path,
		args.ExactPath,
		args.ToolName,
	)
	if err != nil || len(dumps) == 0 {
		return nil, err
	}

	return NewQueryResolver(
		r.dbStore,
		r.lsifStore,
		cachedCommitChecker,
		NewPositionAdjuster(args.Repo, string(args.Commit), r.hunkCache),
		int(args.Repo.ID),
		string(args.Commit),
		args.Path,
		dumps,
		r.operations,
		authz.DefaultSubRepoPermsChecker,
	), nil
}

func (r *resolver) GetConfigurationPolicies(ctx context.Context, opts store.GetConfigurationPoliciesOptions) ([]store.ConfigurationPolicy, int, error) {
	return r.dbStore.GetConfigurationPolicies(ctx, opts)
}

func (r *resolver) GetConfigurationPolicyByID(ctx context.Context, id int) (store.ConfigurationPolicy, bool, error) {
	return r.dbStore.GetConfigurationPolicyByID(ctx, id)
}

func (r *resolver) CreateConfigurationPolicy(ctx context.Context, configurationPolicy store.ConfigurationPolicy) (store.ConfigurationPolicy, error) {
	return r.dbStore.CreateConfigurationPolicy(ctx, configurationPolicy)
}

func (r *resolver) UpdateConfigurationPolicy(ctx context.Context, policy store.ConfigurationPolicy) (err error) {
	return r.dbStore.UpdateConfigurationPolicy(ctx, policy)
}

func (r *resolver) DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error) {
	return r.dbStore.DeleteConfigurationPolicyByID(ctx, id)
}

func (r *resolver) IndexConfiguration(ctx context.Context, repositoryID int) ([]byte, bool, error) {
	configuration, exists, err := r.dbStore.GetIndexConfigurationByRepositoryID(ctx, repositoryID)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	return configuration.Data, true, nil
}

func (r *resolver) InferredIndexConfiguration(ctx context.Context, repositoryID int) (*config.IndexConfiguration, bool, error) {
	maybeConfig, err := r.indexEnqueuer.InferIndexConfiguration(ctx, repositoryID)
	if err != nil || maybeConfig == nil {
		return nil, false, err
	}

	return maybeConfig, true, nil
}

func (r *resolver) UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, configuration string) error {
	if _, err := config.UnmarshalJSON([]byte(configuration)); err != nil {
		return err
	}

	return r.dbStore.UpdateIndexConfigurationByRepositoryID(ctx, repositoryID, []byte(configuration))
}

func (r *resolver) PreviewRepositoryFilter(ctx context.Context, patterns []string, limit, offset int) (_ []int, totalCount int, repositoryMatchLimit *int, _ error) {
	if val := conf.CodeIntelAutoIndexingPolicyRepositoryMatchLimit(); val != -1 {
		repositoryMatchLimit = &val

		if offset+limit > *repositoryMatchLimit {
			limit = *repositoryMatchLimit - offset
		}
	}

	ids, totalCount, err := r.dbStore.RepoIDsByGlobPatterns(ctx, patterns, limit, offset)
	if err != nil {
		return nil, 0, nil, err
	}

	return ids, totalCount, repositoryMatchLimit, nil
}

func (r *resolver) PreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType store.GitObjectType, pattern string) (map[string][]string, error) {
	policyMatches, err := r.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, []store.ConfigurationPolicy{{Type: gitObjectType, Pattern: pattern}}, timeutil.Now())
	if err != nil {
		return nil, err
	}

	namesByCommit := make(map[string][]string, len(policyMatches))
	for commit, policyMatches := range policyMatches {
		names := make([]string, 0, len(policyMatches))
		for _, policyMatch := range policyMatches {
			names = append(names, policyMatch.Name)
		}

		namesByCommit[commit] = names
	}

	return namesByCommit, nil
}
