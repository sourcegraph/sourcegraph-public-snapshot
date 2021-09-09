package graphql

import (
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

const (
	DefaultUploadPageSize = 50
	DefaultIndexPageSize  = 50
)

var errAutoIndexingNotEnabled = errors.New("precise code intelligence auto indexing is not enabled")

// Resolver is the main interface to code intel-related operations exposted to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type Resolver struct {
	resolver         resolvers.Resolver
	locationResolver *CachedLocationResolver
}

// NewResolver creates a new Resolver with the given resolver that defines all code intel-specific behavior.
func NewResolver(db dbutil.DB, resolver resolvers.Resolver) gql.CodeIntelResolver {
	return &Resolver{
		resolver:         resolver,
		locationResolver: NewCachedLocationResolver(db),
	}
}

func (r *Resolver) NodeResolvers() map[string]gql.NodeByIDFunc {
	return map[string]gql.NodeByIDFunc{
		"LSIFUpload": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.LSIFUploadByID(ctx, id)
		},
		"LSIFIndex": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.LSIFIndexByID(ctx, id)
		},
		"CodeIntelligenceConfigurationPolicy": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.ConfigurationPolicyByID(ctx, id)
		},
	}
}

func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (gql.LSIFUploadResolver, error) {
	// 🚨 SECURITY: Only site admins may see LSIF upload data
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := NewPrefetcher(r.resolver)

	upload, exists, err := prefetcher.GetUploadByID(ctx, int(uploadID))
	if err != nil || !exists {
		return nil, err
	}

	return NewUploadResolver(r.resolver, upload, prefetcher, r.locationResolver), nil
}

func (r *Resolver) LSIFUploads(ctx context.Context, args *gql.LSIFUploadsQueryArgs) (gql.LSIFUploadConnectionResolver, error) {
	// Delegate behavior to LSIFUploadsByRepo with no specified repository identifier
	return r.LSIFUploadsByRepo(ctx, &gql.LSIFRepositoryUploadsQueryArgs{LSIFUploadsQueryArgs: args})
}

func (r *Resolver) LSIFUploadsByRepo(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (gql.LSIFUploadConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may see LSIF upload data
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	opts, err := makeGetUploadsOptions(ctx, args)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := NewPrefetcher(r.resolver)

	return NewUploadConnectionResolver(r.resolver, r.resolver.UploadConnectionResolver(opts), prefetcher, r.locationResolver), nil
}

func (r *Resolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may modify LSIF data
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	uploadID, err := unmarshalLSIFUploadGQLID(args.ID)
	if err != nil {
		return nil, err
	}

	if err := r.resolver.DeleteUploadByID(ctx, int(uploadID)); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

var autoIndexingEnabled = conf.CodeIntelAutoIndexingEnabled

func (r *Resolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (gql.LSIFIndexResolver, error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// 🚨 SECURITY: Only site admins may see LSIF index data
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	indexID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := NewPrefetcher(r.resolver)

	index, exists, err := prefetcher.GetIndexByID(ctx, int(indexID))
	if err != nil || !exists {
		return nil, err
	}

	return NewIndexResolver(r.resolver, index, prefetcher, r.locationResolver), nil
}

func (r *Resolver) LSIFIndexes(ctx context.Context, args *gql.LSIFIndexesQueryArgs) (gql.LSIFIndexConnectionResolver, error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// Delegate behavior to LSIFIndexesByRepo with no specified repository identifier
	return r.LSIFIndexesByRepo(ctx, &gql.LSIFRepositoryIndexesQueryArgs{LSIFIndexesQueryArgs: args})
}

func (r *Resolver) LSIFIndexesByRepo(ctx context.Context, args *gql.LSIFRepositoryIndexesQueryArgs) (gql.LSIFIndexConnectionResolver, error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// 🚨 SECURITY: Only site admins may see LSIF index data
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	opts, err := makeGetIndexesOptions(ctx, args)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := NewPrefetcher(r.resolver)

	return NewIndexConnectionResolver(r.resolver, r.resolver.IndexConnectionResolver(opts), prefetcher, r.locationResolver), nil
}

func (r *Resolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*gql.EmptyResponse, error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// 🚨 SECURITY: Only site admins may modify LSIF data
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	indexID, err := unmarshalLSIFIndexGQLID(args.ID)
	if err != nil {
		return nil, err
	}

	if err := r.resolver.DeleteIndexByID(ctx, int(indexID)); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) CommitGraph(ctx context.Context, id graphql.ID) (gql.CodeIntelligenceCommitGraphResolver, error) {
	repositoryID, err := gql.UnmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}

	return r.resolver.CommitGraph(ctx, int(repositoryID))
}

func (r *Resolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *gql.QueueAutoIndexJobsForRepoArgs) ([]gql.LSIFIndexResolver, error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// 🚨 SECURITY: Only site admins may queue indexing jobs
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	repositoryID, err := gql.UnmarshalRepositoryID(args.Repository)
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

	indexes, err := r.resolver.QueueAutoIndexJobsForRepo(ctx, int(repositoryID), rev, configuration)
	if err != nil {
		return nil, err
	}

	// Create a new prefetcher here as we only want to cache upload and index records in
	// the same graphQL request, not across different request.
	prefetcher := NewPrefetcher(r.resolver)

	resolvers := make([]gql.LSIFIndexResolver, 0, len(indexes))
	for i := range indexes {
		resolvers = append(resolvers, NewIndexResolver(r.resolver, indexes[i], prefetcher, r.locationResolver))
	}
	return resolvers, nil
}

func (r *Resolver) GitBlobLSIFData(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (gql.GitBlobLSIFDataResolver, error) {
	resolver, err := r.resolver.QueryResolver(ctx, args)
	if err != nil || resolver == nil {
		return nil, err
	}

	return NewQueryResolver(resolver, r.locationResolver), nil
}

func (r *Resolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (gql.CodeIntelligenceConfigurationPolicyResolver, error) {
	// 🚨 SECURITY: Only site admins may configure code intelligence
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	configurationPolicyID, err := unmarshalConfigurationPolicyGQLID(id)
	if err != nil {
		return nil, err
	}

	configurationPolicy, exists, err := r.resolver.GetConfigurationPolicyByID(ctx, int(configurationPolicyID))
	if err != nil || !exists {
		return nil, err
	}

	return NewConfigurationPolicyResolver(configurationPolicy), nil
}

func (r *Resolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *gql.CodeIntelligenceConfigurationPoliciesArgs) ([]gql.CodeIntelligenceConfigurationPolicyResolver, error) {
	// 🚨 SECURITY: Only site admins may configure code intelligence
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	opts := store.GetConfigurationPoliciesOptions{}
	if args.Repository != nil {
		id64, err := unmarshalRepositoryID(*args.Repository)
		if err != nil {
			return nil, err
		}

		opts.RepositoryID = int(id64)
	}

	policies, err := r.resolver.GetConfigurationPolicies(ctx, opts)
	if err != nil {
		return nil, err
	}

	resolvers := make([]gql.CodeIntelligenceConfigurationPolicyResolver, 0, len(policies))
	for _, policy := range policies {
		resolvers = append(resolvers, NewConfigurationPolicyResolver(policy))
	}

	return resolvers, nil
}

func (r *Resolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.CreateCodeIntelligenceConfigurationPolicyArgs) (gql.CodeIntelligenceConfigurationPolicyResolver, error) {
	// 🚨 SECURITY: Only site admins may configure code intelligence
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	if args.Type != gql.GitObjectTypeCommit && args.Type != gql.GitObjectTypeTag && args.Type != gql.GitObjectTypeTree {
		return nil, errors.Errorf("illegal git object type '%s', expected 'GIT_COMMIT', 'GIT_TAG', or 'GIT_TREE'", args.Type)
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

	configurationPolicy, err := r.resolver.CreateConfigurationPolicy(ctx, store.ConfigurationPolicy{
		RepositoryID:              repositoryID,
		Name:                      args.Name,
		Type:                      store.GitObjectType(args.Type),
		Pattern:                   args.Pattern,
		RetentionEnabled:          args.RetentionEnabled,
		RetentionDuration:         toDuration(args.RetentionDurationHours),
		RetainIntermediateCommits: args.RetainIntermediateCommits,
		IndexingEnabled:           args.IndexingEnabled,
		IndexCommitMaxAge:         toDuration(args.IndexCommitMaxAgeHours),
		IndexIntermediateCommits:  args.IndexIntermediateCommits,
	})
	if err != nil {
		return nil, err
	}

	return NewConfigurationPolicyResolver(configurationPolicy), nil
}

func toDuration(hours *int32) *time.Duration {
	if hours == nil {
		return nil
	}

	v := time.Duration(*hours) * time.Hour
	return &v
}

func (r *Resolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.UpdateCodeIntelligenceConfigurationPolicyArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may configure code intelligence
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	if args.Type != gql.GitObjectTypeCommit && args.Type != gql.GitObjectTypeTag && args.Type != gql.GitObjectTypeTree {
		return nil, errors.Errorf("illegal git object type '%s', expected 'GIT_COMMIT', 'GIT_TAG', or 'GIT_TREE'", args.Type)
	}

	id, err := unmarshalConfigurationPolicyGQLID(args.ID)
	if err != nil {
		return nil, err
	}

	if err := r.resolver.UpdateConfigurationPolicy(ctx, store.ConfigurationPolicy{
		ID:                        int(id),
		Name:                      args.Name,
		Type:                      store.GitObjectType(args.Type),
		Pattern:                   args.Pattern,
		RetentionEnabled:          args.RetentionEnabled,
		RetentionDuration:         toDuration(args.RetentionDurationHours),
		RetainIntermediateCommits: args.RetainIntermediateCommits,
		IndexingEnabled:           args.IndexingEnabled,
		IndexCommitMaxAge:         toDuration(args.IndexCommitMaxAgeHours),
		IndexIntermediateCommits:  args.IndexIntermediateCommits,
	}); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.DeleteCodeIntelligenceConfigurationPolicyArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may configure code intelligence
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	id, err := unmarshalConfigurationPolicyGQLID(args.Policy)
	if err != nil {
		return nil, err
	}

	if err := r.resolver.DeleteConfigurationPolicyByID(ctx, int(id)); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) IndexConfiguration(ctx context.Context, id graphql.ID) (gql.IndexConfigurationResolver, error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// 🚨 SECURITY: Only site admins may configure code intelligence
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	repositoryID, err := gql.UnmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}

	return NewIndexConfigurationResolver(r.resolver, int(repositoryID)), nil
}

func (r *Resolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *gql.UpdateRepositoryIndexConfigurationArgs) (*gql.EmptyResponse, error) {
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// 🚨 SECURITY: Only site admins may configure code intelligence
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, dbconn.Global); err != nil {
		return nil, err
	}

	repositoryID, err := unmarshalLSIFIndexGQLID(args.Repository)
	if err != nil {
		return nil, err
	}

	if err := r.resolver.UpdateIndexConfigurationByRepositoryID(ctx, int(repositoryID), args.Configuration); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

// makeGetUploadsOptions translates the given GraphQL arguments into options defined by the
// store.GetUploads operations.
func makeGetUploadsOptions(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (store.GetUploadsOptions, error) {
	repositoryID, err := resolveRepositoryID(ctx, args.RepositoryID)
	if err != nil {
		return store.GetUploadsOptions{}, err
	}

	var dependencyOf int64
	if args.DependencyOf != nil {
		dependencyOf, err = unmarshalLSIFUploadGQLID(*args.DependencyOf)
		if err != nil {
			return store.GetUploadsOptions{}, err
		}
	}

	var dependentOf int64
	if args.DependentOf != nil {
		dependentOf, err = unmarshalLSIFUploadGQLID(*args.DependentOf)
		if err != nil {
			return store.GetUploadsOptions{}, err
		}
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return store.GetUploadsOptions{}, err
	}

	return store.GetUploadsOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(derefString(args.State, "")),
		Term:         derefString(args.Query, ""),
		VisibleAtTip: derefBool(args.IsLatestForRepo, false),
		DependencyOf: int(dependencyOf),
		DependentOf:  int(dependentOf),
		Limit:        derefInt32(args.First, DefaultUploadPageSize),
		Offset:       offset,
		AllowExpired: true,
	}, nil
}

// makeGetIndexesOptions translates the given GraphQL arguments into options defined by the
// store.GetIndexes operations.
func makeGetIndexesOptions(ctx context.Context, args *gql.LSIFRepositoryIndexesQueryArgs) (store.GetIndexesOptions, error) {
	repositoryID, err := resolveRepositoryID(ctx, args.RepositoryID)
	if err != nil {
		return store.GetIndexesOptions{}, err
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return store.GetIndexesOptions{}, err
	}

	return store.GetIndexesOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(derefString(args.State, "")),
		Term:         derefString(args.Query, ""),
		Limit:        derefInt32(args.First, DefaultIndexPageSize),
		Offset:       offset,
	}, nil
}

// resolveRepositoryByID gets a repository's internal identifier from a GraphQL identifier.
func resolveRepositoryID(ctx context.Context, id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repoID, err := gql.UnmarshalRepositoryID(id)
	if err != nil {
		return 0, err
	}

	return int(repoID), nil
}
