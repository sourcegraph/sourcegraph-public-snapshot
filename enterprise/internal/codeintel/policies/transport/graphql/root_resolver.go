package graphql

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	policySvc        PoliciesService
	repoStore        database.RepoStore
	operations       *operations
	siteAdminChecker sharedresolvers.SiteAdminChecker
}

func NewRootResolver(observationCtx *observation.Context, policySvc *policies.Service, repoStore database.RepoStore, siteAdminChecker sharedresolvers.SiteAdminChecker) resolverstubs.PoliciesServiceResolver {
	return &rootResolver{
		policySvc:        policySvc,
		repoStore:        repoStore,
		operations:       newOperations(observationCtx),
		siteAdminChecker: siteAdminChecker,
	}
}

func (r *rootResolver) ConfigurationPolicyByID(ctx context.Context, policyID graphql.ID) (_ resolverstubs.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.configurationPolicyByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("policyID", string(policyID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	configurationPolicyID, err := resolverstubs.UnmarshalID[int](policyID)
	if err != nil {
		return nil, err
	}

	configurationPolicy, exists, err := r.policySvc.GetConfigurationPolicyByID(ctx, configurationPolicyID)
	if err != nil || !exists {
		return nil, err
	}

	return NewConfigurationPolicyResolver(r.repoStore, configurationPolicy, traceErrs), nil
}

const DefaultConfigurationPolicyPageSize = 50

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicies
func (r *rootResolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *resolverstubs.CodeIntelligenceConfigurationPoliciesArgs) (_ resolverstubs.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.configurationPolicies.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("first", resolverstubs.Deref(args.First, 0)),
		log.String("after", resolverstubs.Deref(args.After, "")),
		log.String("repository", string(resolverstubs.Deref(args.Repository, ""))),
		log.String("query", resolverstubs.Deref(args.Query, "")),
		log.Bool("forDataRetention", resolverstubs.Deref(args.ForDataRetention, false)),
		log.Bool("forIndexing", resolverstubs.Deref(args.ForIndexing, false)),
		log.Bool("protected", resolverstubs.Deref(args.Protected, false)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	limit, offset, err := args.ParseLimitOffset(DefaultConfigurationPolicyPageSize)
	if err != nil {
		return nil, err
	}

	opts := policiesshared.GetConfigurationPoliciesOptions{
		Limit:  int(limit),
		Offset: int(offset),
	}
	if args.Repository != nil {
		id64, err := resolverstubs.UnmarshalID[int64](*args.Repository)
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
	if args.Protected != nil {
		opts.Protected = args.Protected
	}

	configPolicies, totalCount, err := r.policySvc.GetConfigurationPolicies(ctx, opts)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.CodeIntelligenceConfigurationPolicyResolver, 0, len(configPolicies))
	for _, policy := range configPolicies {
		resolvers = append(resolvers, NewConfigurationPolicyResolver(r.repoStore, policy, traceErrs))
	}

	return resolverstubs.NewTotalCountConnectionResolver(resolvers, 0, int32(totalCount)), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.CreateCodeIntelligenceConfigurationPolicyArgs) (_ resolverstubs.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.createConfigurationPolicy.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repository", string(resolverstubs.Deref(args.Repository, ""))),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	if err := validateConfigurationPolicy(args.CodeIntelConfigurationPolicy); err != nil {
		return nil, err
	}

	var repositoryID *int
	if args.Repository != nil {
		id64, err := resolverstubs.UnmarshalID[int64](*args.Repository)
		if err != nil {
			return nil, err
		}

		id := int(id64)
		repositoryID = &id
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
	configurationPolicy, err := r.policySvc.CreateConfigurationPolicy(ctx, opts)
	if err != nil {
		return nil, err
	}

	return NewConfigurationPolicyResolver(r.repoStore, configurationPolicy, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("policyID", string(args.ID)),
	}})
	defer endObservation(1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	if err := validateConfigurationPolicy(args.CodeIntelConfigurationPolicy); err != nil {
		return nil, err
	}

	id, err := resolverstubs.UnmarshalID[int](args.ID)
	if err != nil {
		return nil, err
	}

	opts := types.ConfigurationPolicy{
		ID:                        id,
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
	if err := r.policySvc.UpdateConfigurationPolicy(ctx, opts); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("policyID", string(args.Policy)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := resolverstubs.UnmarshalID[int](args.Policy)
	if err != nil {
		return nil, err
	}

	if err := r.policySvc.DeleteConfigurationPolicyByID(ctx, id); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, nil
}

const DefaultRepositoryFilterPreviewPageSize = 15 // TEMP: 50

func (r *rootResolver) PreviewRepositoryFilter(ctx context.Context, args *resolverstubs.PreviewRepositoryFilterArgs) (_ resolverstubs.RepositoryFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("first", resolverstubs.Deref(args.First, 0)),
		log.String("patterns", strings.Join(args.Patterns, ", ")),
	}})
	defer endObservation(1, observation.Args{})

	pageSize := DefaultRepositoryFilterPreviewPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	ids, totalMatches, matchesAll, repositoryMatchLimit, err := r.policySvc.GetPreviewRepositoryFilter(ctx, args.Patterns, pageSize)
	if err != nil {
		return nil, err
	}

	resv := make([]resolverstubs.RepositoryResolver, 0, len(ids))
	for _, id := range ids {
		res, err := sharedresolvers.NewRepositoryFromID(ctx, r.repoStore, id)
		if err != nil {
			return nil, err
		}

		resv = append(resv, res)
	}

	limitedCount := totalMatches
	if repositoryMatchLimit != nil && *repositoryMatchLimit < limitedCount {
		limitedCount = *repositoryMatchLimit
	}

	return NewRepositoryFilterPreviewResolver(resv, limitedCount, totalMatches, matchesAll, repositoryMatchLimit), nil
}

const DefaultGitObjectFilterPreviewPageSize = 15 // TEMP: 100

func (r *rootResolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *resolverstubs.PreviewGitObjectFilterArgs) (_ resolverstubs.GitObjectFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewGitObjectFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("first", resolverstubs.Deref(args.First, 0)),
		log.String("type", string(args.Type)),
		log.String("pattern", args.Pattern),
	}})
	defer endObservation(1, observation.Args{})

	repositoryID, err := resolverstubs.UnmarshalID[int](id)
	if err != nil {
		return nil, err
	}

	gitObjects, totalCount, totalCountYoungerThanThreshold, err := r.policySvc.GetPreviewGitObjectFilter(
		ctx,
		repositoryID,
		types.GitObjectType(args.Type),
		args.Pattern,
		int(args.Limit(DefaultGitObjectFilterPreviewPageSize)),
		args.CountObjectsYoungerThanHours,
	)
	if err != nil {
		return nil, err
	}

	var gitObjectResolvers []resolverstubs.CodeIntelGitObjectResolver
	for _, gitObject := range gitObjects {
		gitObjectResolvers = append(gitObjectResolvers, NewGitObjectResolver(gitObject.Name, gitObject.Rev, gitObject.CommittedAt))
	}

	return NewGitObjectFilterPreviewResolver(gitObjectResolvers, totalCount, totalCountYoungerThanThreshold), nil
}
