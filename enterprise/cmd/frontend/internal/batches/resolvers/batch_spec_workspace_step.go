package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
)

type batchSpecWorkspaceStepResolver struct {
	store    *store.Store
	repo     *graphqlbackend.RepositoryResolver
	baseRev  string
	index    int
	step     batcheslib.Step
	stepInfo *btypes.StepInfo

	cachedResult *execution.AfterStepResult
}

func (r *batchSpecWorkspaceStepResolver) Number() int32 {
	return int32(r.index + 1)
}

func (r *batchSpecWorkspaceStepResolver) Run() string {
	return r.step.Run
}

func (r *batchSpecWorkspaceStepResolver) Container() string {
	return r.step.Container
}

func (r *batchSpecWorkspaceStepResolver) IfCondition() *string {
	cond := r.step.IfCondition()
	if cond == "" {
		return nil
	}
	return &cond
}

func (r *batchSpecWorkspaceStepResolver) CachedResultFound() bool {
	return r.stepInfo.StartedAt.IsZero() && r.cachedResult != nil
}

func (r *batchSpecWorkspaceStepResolver) Skipped() bool {
	return r.CachedResultFound() || r.stepInfo.Skipped
}

func (r *batchSpecWorkspaceStepResolver) OutputLines(ctx context.Context, args *graphqlbackend.BatchSpecWorkspaceStepOutputLinesArgs) (*[]string, error) {
	lines := r.stepInfo.OutputLines
	if args.After != nil {
		lines = lines[*args.After:]
	}
	if int(args.First) < len(lines) {
		lines = lines[:args.First]
	}
	// TODO: Return nil when execution not yet started.
	return &lines, nil
}

func (r *batchSpecWorkspaceStepResolver) StartedAt() *graphqlbackend.DateTime {
	if r.stepInfo.StartedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.stepInfo.StartedAt}
}

func (r *batchSpecWorkspaceStepResolver) FinishedAt() *graphqlbackend.DateTime {
	if r.stepInfo.FinishedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.stepInfo.FinishedAt}
}

func (r *batchSpecWorkspaceStepResolver) ExitCode() *int32 {
	if r.stepInfo.ExitCode == nil {
		return nil
	}
	code := int32(*r.stepInfo.ExitCode)
	return &code
}

func (r *batchSpecWorkspaceStepResolver) Environment() ([]graphqlbackend.BatchSpecWorkspaceEnvironmentVariableResolver, error) {
	// The environment is dependent on environment of the executor and template variables, that aren't
	// known at the time when we resolve the workspace. If the step already started, src cli has logged
	// the final env. Otherwise, we fall back to the preliminary set of env vars as determined by the
	// resolve workspaces step.

	var env = r.stepInfo.Environment

	// Not yet resolved, do a server-side pass.
	if env == nil {
		var err error
		env, err = r.step.Env.Resolve([]string{})
		if err != nil {
			return nil, err
		}
	}

	resolvers := make([]graphqlbackend.BatchSpecWorkspaceEnvironmentVariableResolver, 0, len(env))
	for k, v := range env {
		resolvers = append(resolvers, &batchSpecWorkspaceEnvironmentVariableResolver{key: k, value: v})
	}
	return resolvers, nil
}

func (r *batchSpecWorkspaceStepResolver) OutputVariables() *[]graphqlbackend.BatchSpecWorkspaceOutputVariableResolver {
	if r.CachedResultFound() {
		resolvers := make([]graphqlbackend.BatchSpecWorkspaceOutputVariableResolver, 0, len(r.cachedResult.Outputs))
		for k, v := range r.cachedResult.Outputs {
			resolvers = append(resolvers, &batchSpecWorkspaceOutputVariableResolver{key: k, value: v})
		}
		return &resolvers
	}

	if r.stepInfo.OutputVariables == nil {
		return nil
	}

	resolvers := make([]graphqlbackend.BatchSpecWorkspaceOutputVariableResolver, 0, len(r.stepInfo.OutputVariables))
	for k, v := range r.stepInfo.OutputVariables {
		resolvers = append(resolvers, &batchSpecWorkspaceOutputVariableResolver{key: k, value: v})
	}
	return &resolvers
}

func (r *batchSpecWorkspaceStepResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	diffRes, err := r.Diff(ctx)
	if err != nil {
		return nil, err
	}
	if diffRes != nil {
		fd, err := diffRes.FileDiffs(ctx, &graphqlbackend.FileDiffsConnectionArgs{})
		if err != nil {
			return nil, err
		}
		return fd.DiffStat(ctx)
	}
	return nil, nil
}

func (r *batchSpecWorkspaceStepResolver) Diff(ctx context.Context) (graphqlbackend.PreviewRepositoryComparisonResolver, error) {
	if r.CachedResultFound() {
		return graphqlbackend.NewPreviewRepositoryComparisonResolver(ctx, r.store.DatabaseDB(), r.repo, r.baseRev, r.cachedResult.Diff)
	}
	if r.stepInfo.Diff != nil {
		return graphqlbackend.NewPreviewRepositoryComparisonResolver(ctx, r.store.DatabaseDB(), r.repo, r.baseRev, *r.stepInfo.Diff)
	}
	return nil, nil
}

type batchSpecWorkspaceEnvironmentVariableResolver struct {
	key   string
	value string
}

var _ graphqlbackend.BatchSpecWorkspaceEnvironmentVariableResolver = &batchSpecWorkspaceEnvironmentVariableResolver{}

func (r *batchSpecWorkspaceEnvironmentVariableResolver) Name() string {
	return r.key
}
func (r *batchSpecWorkspaceEnvironmentVariableResolver) Value() string {
	return r.value
}

type batchSpecWorkspaceOutputVariableResolver struct {
	key   string
	value any
}

var _ graphqlbackend.BatchSpecWorkspaceOutputVariableResolver = &batchSpecWorkspaceOutputVariableResolver{}

func (r *batchSpecWorkspaceOutputVariableResolver) Name() string {
	return r.key
}
func (r *batchSpecWorkspaceOutputVariableResolver) Value() graphqlbackend.JSONValue {
	return graphqlbackend.JSONValue{Value: r.value}
}
