package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

type batchSpecWorkspaceStepResolver struct {
	store    *store.Store
	repo     *graphqlbackend.RepositoryResolver
	baseRev  string
	index    int
	step     batcheslib.Step
	logLines []batcheslib.LogEvent
}

func (r *batchSpecWorkspaceStepResolver) Run() string {
	return r.step.Run
}

func (r *batchSpecWorkspaceStepResolver) Container() string {
	return r.step.Container
}

func (r *batchSpecWorkspaceStepResolver) CachedResultFound() bool {
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecWorkspaceStepResolver) Skipped() bool {
	for _, l := range r.logLines {
		if m, ok := l.Metadata.(*batcheslib.TaskSkippingStepsMetadata); ok {
			if m.StartStep-1 > r.index {
				return true
			}
		}
		if m, ok := l.Metadata.(*batcheslib.TaskStepSkippedMetadata); ok {
			if m.Step-1 == r.index {
				return true
			}
		}
	}

	return false
}

func (r *batchSpecWorkspaceStepResolver) OutputLines(ctx context.Context, args *graphqlbackend.BatchSpecWorkspaceStepOutputLinesArgs) (*[]string, error) {
	lines := []string{}
	for _, l := range r.logLines {
		if l.Status != batcheslib.LogEventStatusProgress {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.TaskStepMetadata); ok {
			if m.Step-1 != r.index {
				continue
			}
			if m.Out == "" {
				continue
			}
			lines = append(lines, m.Out)
		}
	}
	if args.After != nil {
		lines = lines[*args.After:]
	}
	if int(args.First) < len(lines) {
		lines = lines[:args.First]
	}
	// TODO: Should sometimes return nil.
	return &lines, nil
}

func (r *batchSpecWorkspaceStepResolver) StartedAt() *graphqlbackend.DateTime {
	for _, l := range r.logLines {
		if l.Status != batcheslib.LogEventStatusStarted {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.TaskPreparingStepMetadata); ok {
			if m.Step-1 == r.index {
				return &graphqlbackend.DateTime{Time: l.Timestamp}
			}
		}
	}
	return nil
}

func (r *batchSpecWorkspaceStepResolver) FinishedAt() *graphqlbackend.DateTime {
	for _, l := range r.logLines {
		if l.Status != batcheslib.LogEventStatusSuccess && l.Status != batcheslib.LogEventStatusFailure {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.TaskStepMetadata); ok {
			if m.Step-1 == r.index {
				return &graphqlbackend.DateTime{Time: l.Timestamp}
			}
		}
	}
	return nil
}

func (r *batchSpecWorkspaceStepResolver) ExitCode() *int32 {
	for _, l := range r.logLines {
		if l.Status != batcheslib.LogEventStatusSuccess && l.Status != batcheslib.LogEventStatusFailure {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.TaskStepMetadata); ok {
			if m.Step-1 == r.index {
				code := int32(m.ExitCode)
				return &code
			}
		}
	}
	return nil
}

func (r *batchSpecWorkspaceStepResolver) Environment() ([]graphqlbackend.BatchSpecWorkspaceEnvironmentVariableResolver, error) {
	// The environment is dependent on environment of the executor and template variables, that aren't
	// known at the time when we resolve the workspace. If the step already started, src cli has logged
	// the final env. Otherwise, we fall back to the preliminary set of env vars as determined by the
	// resolve workspaces step.
	found := false
	var env map[string]string
	for _, l := range r.logLines {
		if l.Status != batcheslib.LogEventStatusStarted {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.TaskStepMetadata); ok {
			if m.Step-1 == r.index {
				if m.Env != nil {
					found = true
					env = m.Env
				}
			}
		}
	}

	if !found {
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
	for _, l := range r.logLines {
		if l.Status != batcheslib.LogEventStatusSuccess {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.TaskStepMetadata); ok {
			if m.Step-1 == r.index {
				resolvers := make([]graphqlbackend.BatchSpecWorkspaceOutputVariableResolver, 0, len(m.Outputs))
				for k, v := range m.Outputs {
					resolvers = append(resolvers, &batchSpecWorkspaceOutputVariableResolver{key: k, value: v})
				}
				return &resolvers
			}
		}
	}
	return nil
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
	for _, l := range r.logLines {
		if l.Status != batcheslib.LogEventStatusSuccess {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.TaskStepMetadata); ok {
			if m.Step-1 == r.index {
				return graphqlbackend.NewPreviewRepositoryComparisonResolver(ctx, r.store.DB(), r.repo, r.baseRev, m.Diff)
			}
		}
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
	value interface{}
}

var _ graphqlbackend.BatchSpecWorkspaceOutputVariableResolver = &batchSpecWorkspaceOutputVariableResolver{}

func (r *batchSpecWorkspaceOutputVariableResolver) Name() string {
	return r.key
}
func (r *batchSpecWorkspaceOutputVariableResolver) Value() graphqlbackend.JSONValue {
	return graphqlbackend.JSONValue{Value: r.value}
}
