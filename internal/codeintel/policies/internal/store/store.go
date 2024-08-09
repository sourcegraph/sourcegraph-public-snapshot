package store

import (
	"context"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	// Global metadata
	RepoCount(ctx context.Context) (int, error)

	// Configurations
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]shared.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (shared.ConfigurationPolicy, bool, error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy shared.ConfigurationPolicy) (shared.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy shared.ConfigurationPolicyPatch) error
	DeleteConfigurationPolicyByID(ctx context.Context, id int) error

	// Repository matches
	GetRepoIDsByGlobPatterns(ctx context.Context, patterns []string, limit, offset int) ([]int, int, error)
	UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) error
	SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) ([]shared.ConfigurationPolicy, error)
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("policies.store"),
		operations: newOperations(observationCtx),
	}
}
