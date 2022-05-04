package graphql

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver struct {
	svc        *policies.Service
	operations *operations
}

func newResolver(svc *policies.Service, observationContext *observation.Context) *Resolver {
	return &Resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *Resolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ gql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, _, endObservation := r.operations.configurationPolicyByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in - https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _ = ctx, id
	return nil, errors.New("unimplemented: ConfigurationPolicyByID")
}

func (r *Resolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *gql.CodeIntelligenceConfigurationPoliciesArgs) (_ gql.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.codeIntelligenceConfiogurationPolicies.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in - https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _ = ctx, args
	return nil, errors.New("unimplemented: CodeIntelligenceConfiogurationPolicies")
}

func (r *Resolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.CreateCodeIntelligenceConfigurationPolicyArgs) (_ gql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	ctx, _, endObservation := r.operations.createCodeIntelligenceConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in - https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _ = ctx, args
	return nil, errors.New("unimplemented: CreateCodeIntelligenceConfigurationPolicy")
}

func (r *Resolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateCodeIntelligenceConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in - https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _ = ctx, args
	return nil, errors.New("unimplemented: UpdateCodeIntelligenceConfigurationPolicy")
}

func (r *Resolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *gql.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.deleteCodeIntelligenceConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in - https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _ = ctx, args
	return nil, errors.New("unimplemented: DeleteCodeIntelligenceConfigurationPolicy")
}

func (r *Resolver) PreviewRepositoryFilter(ctx context.Context, args *gql.PreviewRepositoryFilterArgs) (_ gql.RepositoryFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewRepositoryFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in - https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _ = ctx, args
	return nil, errors.New("unimplemented: PreviewRepositoryFilter")
}

func (r *Resolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *gql.PreviewGitObjectFilterArgs) (_ []gql.GitObjectFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewGitObjectFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in - https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _, _ = ctx, id, args
	return nil, errors.New("unimplemented: PreviewGitObjectFilter")
}
