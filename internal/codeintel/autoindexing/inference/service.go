package inference

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type Service struct {
	sandboxService SandboxService
	gitService     GitService
	operations     *operations
}

func newService(
	sandboxService SandboxService,
	gitService GitService,
	observationContext *observation.Context,
) *Service {
	return &Service{
		sandboxService: sandboxService,
		gitService:     gitService,
		operations:     newOperations(observationContext),
	}
}

// InferIndexJobs invokes the given script in a fresh Lua sandbox. The return value of this script
// is assumed to be a table of recognizer instances. Keys conflicting with the default recognizers
// will overwrite them (to disable or change default behavior).
func (s *Service) InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) (_ []config.IndexJob, err error) {
	ctx, endObservation := s.operations.inferIndexJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return nil, fmt.Errorf("unimplemented")
}
