package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type AutoIndexingService interface {
	// Inference configuration
	GetInferenceScript(ctx context.Context) (string, error)
	SetInferenceScript(ctx context.Context, script string) error

	// Repository configuration
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (shared.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error

	// Inference
	QueueAutoIndexJobs(ctx context.Context, repositoryID int, rev, configuration string, force bool, bypassLimit bool) ([]uploadsshared.AutoIndexJob, error)
	InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) (*shared.InferenceResult, error)
	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) (*shared.InferenceResult, error)
}
