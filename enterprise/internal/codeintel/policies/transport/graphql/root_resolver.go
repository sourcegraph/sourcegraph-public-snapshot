package graphql

import (
	"context"
	"encoding/base64"
	"strconv"

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

	return NewConfigurationPolicyResolver(r.repoStore, configurationPolicy, traceErrs), nil
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

	offset, err := decodeIntCursor(args.After)
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

	return NewCodeIntelligenceConfigurationPolicyConnectionResolver(r.repoStore, configPolicies, totalCount, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.CreateCodeIntelligenceConfigurationPolicyArgs) (_ resolverstubs.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.createConfigurationPolicy.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

	return NewConfigurationPolicyResolver(r.repoStore, configurationPolicy, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(args.ID)),
	}})
	defer endObservation(1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("configPolicyID", string(args.Policy)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := unmarshalConfigurationPolicyGQLID(args.Policy)
	if err != nil {
		return nil, err
	}

	if err := r.policySvc.DeleteConfigurationPolicyByID(ctx, int(id)); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, nil
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

func decodeIntCursor(val *string) (int, error) {
	cursor, err := decodeCursor(val)
	if err != nil || cursor == "" {
		return 0, err
	}

	return strconv.Atoi(cursor)
}

func decodeCursor(val *string) (string, error) {
	if val == nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
