package sharedresolvers

import "github.com/sourcegraph/sourcegraph/internal/codeintel/types"

type PreIndexStepResolver interface {
	Root() string
	Image() string
	Commands() []string
	LogEntry() ExecutionLogEntryResolver
}

type preIndexStepResolver struct {
	svc   AutoIndexingService
	step  types.DockerStep
	entry *types.ExecutionLogEntry
}

func NewPreIndexStepResolver(svc AutoIndexingService, step types.DockerStep, entry *types.ExecutionLogEntry) PreIndexStepResolver {
	return &preIndexStepResolver{
		svc:   svc,
		step:  step,
		entry: entry,
	}
}

func (r *preIndexStepResolver) Root() string       { return r.step.Root }
func (r *preIndexStepResolver) Image() string      { return r.step.Image }
func (r *preIndexStepResolver) Commands() []string { return r.step.Commands }

func (r *preIndexStepResolver) LogEntry() ExecutionLogEntryResolver {
	if r.entry != nil {
		return NewExecutionLogEntryResolver(r.svc, *r.entry)
	}

	return nil
}
