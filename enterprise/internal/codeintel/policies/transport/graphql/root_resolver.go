package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	policySvc  *policies.Service
	operations *operations
}

func NewRootResolver(observationCtx *observation.Context, policySvc *policies.Service) resolverstubs.PoliciesServiceResolver {
	return &rootResolver{
		policySvc:  policySvc,
		operations: newOperations(observationCtx),
	}
}

func (r *rootResolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ resolverstubs.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.configurationPolicyByID.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(id)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	configurationPolicyID, err := unmarshalConfigurationPolicyGQLID(id)
	if err != nil {
		return nil, err
	}

	configurationPolicy, exists, err := r.policySvc.GetConfigurationPolicyByID(ctx, int(configurationPolicyID))
	if err != nil || !exists {
		return nil, err
	}

	return NewConfigurationPolicyResolver(r.policySvc, configurationPolicy, traceErrs), nil
}

const DefaultConfigurationPolicyPageSize = 50

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicies
func (r *rootResolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *resolverstubs.CodeIntelligenceConfigurationPoliciesArgs) (_ resolverstubs.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	fields := []log.Field{}
	if args.Repository != nil {
		fields = append(fields, log.String("repoID", string(*args.Repository)))
	}
	ctx, traceErrs, endObservation := r.operations.configurationPolicies.WithErrors(ctx, &err, observation.Args{LogFields: fields})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	pageSize := DefaultConfigurationPolicyPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	opts := policiesshared.GetConfigurationPoliciesOptions{
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
	if args.Protected != nil {
		opts.Protected = args.Protected
	}

	configPolicies, totalCount, err := r.policySvc.GetConfigurationPolicies(ctx, opts)
	if err != nil {
		return nil, err
	}

	return NewCodeIntelligenceConfigurationPolicyConnectionResolver(r.policySvc, configPolicies, totalCount, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.CreateCodeIntelligenceConfigurationPolicyArgs) (_ resolverstubs.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.createConfigurationPolicy.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.policySvc.GetUnsafeDB()); err != nil {
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

	return NewConfigurationPolicyResolver(r.policySvc, configurationPolicy, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(args.ID)),
	}})
	defer endObservation(1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.policySvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	if err := validateConfigurationPolicy(args.CodeIntelConfigurationPolicy); err != nil {
		return nil, err
	}

	id, err := unmarshalConfigurationPolicyGQLID(args.ID)
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
	if err := r.policySvc.UpdateConfigurationPolicy(ctx, opts); err != nil {
		return nil, err
	}

	return &resolverstubs.EmptyResponse{}, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(args.Policy)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.policySvc.GetUnsafeDB()); err != nil {
		return nil, err
	}

	id, err := unmarshalConfigurationPolicyGQLID(args.Policy)
	if err != nil {
		return nil, err
	}

	if err := r.policySvc.DeleteConfigurationPolicyByID(ctx, int(id)); err != nil {
		return nil, err
	}

	return &resolverstubs.EmptyResponse{}, nil
}

const DefaultRepositoryFilterPreviewPageSize = 15 // TEMP: 50

func (r *rootResolver) PreviewRepositoryFilter(ctx context.Context, args *resolverstubs.PreviewRepositoryFilterArgs) (_ resolverstubs.RepositoryFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewRepoFilter.With(ctx, &err, observation.Args{})
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
	logger := sglog.Scoped("PreviewRepositoryFilter", "policies resolver")
	for _, id := range ids {
		db := r.policySvc.GetUnsafeDB()
		repo, err := backend.NewRepos(logger, db, gitserver.NewClient()).Get(ctx, api.RepoID(id))
		if err != nil {
			return nil, err
		}

		resv = append(resv, sharedresolvers.NewRepositoryResolver(r.policySvc.GetUnsafeDB(), repo))
	}

	limitedCount := totalMatches
	if repositoryMatchLimit != nil && *repositoryMatchLimit < limitedCount {
		limitedCount = *repositoryMatchLimit
	}

	return NewRepositoryFilterPreviewResolver(resv, limitedCount, totalMatches, matchesAll, repositoryMatchLimit), nil
}

const DefaultGitObjectFilterPreviewPageSize = 15 // TEMP: 100

func (r *rootResolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *resolverstubs.PreviewGitObjectFilterArgs) (_ resolverstubs.GitObjectFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewGitObjectFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repositoryID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	pageSize := DefaultGitObjectFilterPreviewPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	gitObjects, totalCount, totalCountYoungerThanThreshold, err := r.policySvc.GetPreviewGitObjectFilter(
		ctx,
		int(repositoryID),
		types.GitObjectType(args.Type),
		args.Pattern,
		pageSize,
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
