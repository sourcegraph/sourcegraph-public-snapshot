package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
)

type PolicyResolver interface {
	// Configurations
	GetConfigurationPolicies(ctx context.Context, opts shared.GetConfigurationPoliciesOptions) ([]shared.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (shared.ConfigurationPolicy, bool, error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy shared.ConfigurationPolicy) (shared.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy shared.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error)

	// Retention
	GetRetentionPolicyOverview(ctx context.Context, upload shared.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []shared.RetentionPolicyMatchCandidate, totalCount int, err error)

	// Previews
	GetPreviewRepositoryFilter(ctx context.Context, patterns []string, limit, offset int) (_ []int, totalCount int, repositoryMatchLimit *int, _ error)
	GetPreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType shared.GitObjectType, pattern string) (map[string][]string, error)
}

type policyResolver struct {
	svc        Service
	operations *operations
}

func NewPolicyResolver(svc Service, operations *operations) PolicyResolver {
	return &policyResolver{
		svc:        svc,
		operations: operations,
	}
}

func (p *policyResolver) GetConfigurationPolicies(ctx context.Context, opts shared.GetConfigurationPoliciesOptions) (_ []shared.ConfigurationPolicy, total int, err error) {
	ctx, _, endObservation := p.operations.getConfigurationPolicies.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.GetConfigurationPolicies(ctx, opts)
}

func (p *policyResolver) GetConfigurationPolicyByID(ctx context.Context, id int) (_ shared.ConfigurationPolicy, _ bool, err error) {
	ctx, _, endObservation := p.operations.getConfigurationPolicyByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.GetConfigurationPolicyByID(ctx, id)
}

func (p *policyResolver) CreateConfigurationPolicy(ctx context.Context, configurationPolicy shared.ConfigurationPolicy) (_ shared.ConfigurationPolicy, err error) {
	ctx, _, endObservation := p.operations.createConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.CreateConfigurationPolicy(ctx, configurationPolicy)
}

func (p *policyResolver) UpdateConfigurationPolicy(ctx context.Context, policy shared.ConfigurationPolicy) (err error) {
	ctx, _, endObservation := p.operations.updateConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.UpdateConfigurationPolicy(ctx, policy)
}

func (p *policyResolver) DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := p.operations.deleteConfigurationPolicyByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.DeleteConfigurationPolicyByID(ctx, id)
}

func (p *policyResolver) GetRetentionPolicyOverview(ctx context.Context, upload shared.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []shared.RetentionPolicyMatchCandidate, totalCount int, err error) {
	ctx, _, endObservation := p.operations.getRetentionPolicyOverview.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.GetRetentionPolicyOverview(ctx, upload, matchesOnly, first, after, query, now)
}

func (p *policyResolver) GetPreviewRepositoryFilter(ctx context.Context, patterns []string, limit, offset int) (_ []int, totalCount int, repositoryMatchLimit *int, err error) {
	ctx, _, endObservation := p.operations.getPreviewRepositoryFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.GetPreviewRepositoryFilter(ctx, patterns, limit, offset)
}

func (p *policyResolver) GetPreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType shared.GitObjectType, pattern string) (_ map[string][]string, err error) {
	ctx, _, endObservation := p.operations.getPreviewGitObjectFilter.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return p.svc.GetPreviewGitObjectFilter(ctx, repositoryID, gitObjectType, pattern)
}
