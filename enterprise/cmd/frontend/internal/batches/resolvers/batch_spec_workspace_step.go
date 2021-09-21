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
		if l.Operation == batcheslib.LogEventOperationTaskSkippingSteps {
			if v, ok := l.Metadata["startStep"]; ok {
				if int(v.(float64)-1) > r.index {
					return true
				}
			}
		} else if l.Operation == batcheslib.LogEventOperationTaskStepSkipped {
			if v, ok := l.Metadata["step"]; ok {
				if int(v.(float64)-1) == r.index {
					return true
				}
			}
		}
	}

	return false
}

func (r *batchSpecWorkspaceStepResolver) OutputLines(ctx context.Context, args *graphqlbackend.BatchSpecWorkspaceStepOutputLinesArgs) (*[]string, error) {
	lines := []string{}
	for _, l := range r.logLines {
		if l.Operation == batcheslib.LogEventOperationTaskSkippingSteps && l.Status == batcheslib.LogEventStatusProgress {
			if v, ok := l.Metadata["step"]; !ok {
				continue
			} else {
				if v.(int)-1 != r.index {
					continue
				}
				out, ok := l.Metadata["out"]
				if !ok {
					continue
				}
				outputType, ok := l.Metadata["output_type"]
				if !ok {
					continue
				}
				if outputType == "stdout" {
					lines = append(lines, "stdout: "+out.(string))
				} else {
					lines = append(lines, "stderr: "+out.(string))
				}
			}
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
		if l.Operation == batcheslib.LogEventOperationTaskPreparingStep {
			if v, ok := l.Metadata["step"]; ok {
				if int(v.(float64)-1) == r.index {
					return &graphqlbackend.DateTime{Time: l.Timestamp}
				}
			}
		}
	}
	return nil
}

func (r *batchSpecWorkspaceStepResolver) FinishedAt() *graphqlbackend.DateTime {
	for _, l := range r.logLines {
		if l.Operation == batcheslib.LogEventOperationTaskStep && l.Status == batcheslib.LogEventStatusSuccess {
			if v, ok := l.Metadata["step"]; ok {
				if int(v.(float64)-1) == r.index {
					return &graphqlbackend.DateTime{Time: l.Timestamp}
				}
			}
		}
	}
	return nil
}

func (r *batchSpecWorkspaceStepResolver) ExitCode() *int32 {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceStepResolver) Environment() ([]graphqlbackend.BatchSpecWorkspaceEnvironmentVariableResolver, error) {
	// TODO: This should be precedented by env as printed by src-cli. It will do the evaluation of output vars etc.
	env, err := r.step.Env.Resolve([]string{})
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.BatchSpecWorkspaceEnvironmentVariableResolver, 0, len(env))
	for k, v := range env {
		resolvers = append(resolvers, &batchSpecWorkspaceEnvironmentVariableResolver{key: k, value: v})
	}
	return resolvers, nil
}

func (r *batchSpecWorkspaceStepResolver) OutputVariables() *[]graphqlbackend.BatchSpecWorkspaceOutputVariableResolver {
	for _, l := range r.logLines {
		if l.Operation == batcheslib.LogEventOperationTaskStep && l.Status == batcheslib.LogEventStatusSuccess {
			if v, ok := l.Metadata["step"]; ok {
				if int(v.(float64)-1) == r.index {
					if o, ok := l.Metadata["outputs"]; ok {
						om := o.(map[string]interface{})
						resolvers := make([]graphqlbackend.BatchSpecWorkspaceOutputVariableResolver, 0, len(om))
						for k, v := range om {
							resolvers = append(resolvers, &batchSpecWorkspaceOutputVariableResolver{key: k, value: v})
						}
						return &resolvers
					}
				}
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
		if l.Operation == batcheslib.LogEventOperationTaskStep && l.Status == batcheslib.LogEventStatusSuccess {
			if v, ok := l.Metadata["step"]; ok {
				if int(v.(float64)-1) == r.index {
					if v, ok := l.Metadata["diff"]; ok {
						diff := v.(string)
						return graphqlbackend.NewPreviewRepositoryComparisonResolver(ctx, r.store.DB(), r.repo, r.baseRev, diff)
					}
				}
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
