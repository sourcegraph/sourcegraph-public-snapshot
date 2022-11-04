package graphql

import (
	"context"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type RootResolver interface {
	CodeIntelligenceConfigurationPolicies(ctx context.Context, args *CodeIntelligenceConfigurationPoliciesArgs) (_ CodeIntelligenceConfigurationPolicyConnectionResolver, err error)
	ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ CodeIntelligenceConfigurationPolicyResolver, err error)
	CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *CreateCodeIntelligenceConfigurationPolicyArgs) (_ CodeIntelligenceConfigurationPolicyResolver, err error)
	UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error)
	DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error)
	PreviewRepositoryFilter(ctx context.Context, args *PreviewRepositoryFilterArgs) (_ RepositoryFilterPreviewResolver, err error)
	PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *PreviewGitObjectFilterArgs) (_ []GitObjectFilterPreviewResolver, err error)
}

type rootResolver struct {
	policySvc  *policies.Service
	operations *operations
}

func NewRootResolver(policySvc *policies.Service, observationContext *observation.Context) RootResolver {
	return &rootResolver{
		policySvc:  policySvc,
		operations: newOperations(observationContext),
	}
}

func (r *rootResolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ CodeIntelligenceConfigurationPolicyResolver, err error) {
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

type CodeIntelligenceConfigurationPoliciesArgs struct {
	ConnectionArgs
	Repository       *graphql.ID
	Query            *string
	ForDataRetention *bool
	ForIndexing      *bool
	After            *string
}

const DefaultConfigurationPolicyPageSize = 50

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicies
func (r *rootResolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *CodeIntelligenceConfigurationPoliciesArgs) (_ CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
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

	policies, totalCount, err := r.policySvc.GetConfigurationPolicies(ctx, opts)
	if err != nil {
		return nil, err
	}

	return NewCodeIntelligenceConfigurationPolicyConnectionResolver(r.policySvc, policies, totalCount, traceErrs), nil
}

type CodeIntelConfigurationPolicy struct {
	Name                      string
	RepositoryID              *int32
	RepositoryPatterns        *[]string
	Type                      types.GitObjectType
	Pattern                   string
	RetentionEnabled          bool
	RetentionDurationHours    *int32
	RetainIntermediateCommits bool
	IndexingEnabled           bool
	IndexCommitMaxAgeHours    *int32
	IndexIntermediateCommits  bool
}

type CreateCodeIntelligenceConfigurationPolicyArgs struct {
	Repository *graphql.ID
	CodeIntelConfigurationPolicy
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *CreateCodeIntelligenceConfigurationPolicyArgs) (_ CodeIntelligenceConfigurationPolicyResolver, err error) {
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
		Type:                      args.Type,
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

type UpdateCodeIntelligenceConfigurationPolicyArgs struct {
	ID         graphql.ID
	Repository *graphql.ID
	CodeIntelConfigurationPolicy
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error) {
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
		Type:                      args.Type,
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

	return &sharedresolvers.EmptyResponse{}, nil
}

type DeleteCodeIntelligenceConfigurationPolicyArgs struct {
	Policy graphql.ID
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *rootResolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error) {
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

	return &sharedresolvers.EmptyResponse{}, nil
}

type PreviewRepositoryFilterArgs struct {
	graphqlutil.ConnectionArgs
	Patterns []string
	After    *string
}

const DefaultRepositoryFilterPreviewPageSize = 50

func (r *rootResolver) PreviewRepositoryFilter(ctx context.Context, args *PreviewRepositoryFilterArgs) (_ RepositoryFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewRepoFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	pageSize := DefaultRepositoryFilterPreviewPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	ids, totalMatches, repositoryMatchLimit, err := r.policySvc.GetPreviewRepositoryFilter(ctx, args.Patterns, pageSize, offset)
	if err != nil {
		return nil, err
	}

	resv := make([]*sharedresolvers.RepositoryResolver, 0, len(ids))
	logger := sglog.Scoped("PreviewRepositoryFilter", "policies resolver")
	for _, id := range ids {
		db := r.policySvc.GetUnsafeDB()
		repo, err := backend.NewRepos(logger, db, gitserver.NewClient(db)).Get(ctx, api.RepoID(id))
		if err != nil {
			return nil, err
		}

		resv = append(resv, sharedresolvers.NewRepositoryResolver(r.policySvc.GetUnsafeDB(), repo))
	}

	limitedCount := totalMatches
	if repositoryMatchLimit != nil && *repositoryMatchLimit < limitedCount {
		limitedCount = *repositoryMatchLimit
	}

	return NewRepositoryFilterPreviewResolver(resv, limitedCount, totalMatches, offset, repositoryMatchLimit), nil
}

type PreviewGitObjectFilterArgs struct {
	Type    types.GitObjectType
	Pattern string
}

func (r *rootResolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *PreviewGitObjectFilterArgs) (_ []GitObjectFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewGitObjectFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repositoryID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	namesByRev, err := r.policySvc.GetPreviewGitObjectFilter(ctx, int(repositoryID), args.Type, args.Pattern)
	if err != nil {
		return nil, err
	}

	var previews []GitObjectFilterPreviewResolver
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
