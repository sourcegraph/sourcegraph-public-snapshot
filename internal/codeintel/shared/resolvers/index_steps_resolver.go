package sharedresolvers

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// indexStepsResolver resolves the steps of an index record.
//
// Index jobs are broken into three parts:
//   - pre-index steps; all but the last docker step
//   - index step; the last docker step
//   - upload step; the only src-cli step
//
// The setup and teardown steps match the executor setup and teardown.
type IndexStepsResolver interface {
	Setup() []ExecutionLogEntryResolver
	PreIndex() []PreIndexStepResolver
	Index() IndexStepResolver
	Upload() ExecutionLogEntryResolver
	Teardown() []ExecutionLogEntryResolver
}

type indexStepsResolver struct {
	svc   AutoIndexingService
	index types.Index
}

func NewIndexStepsResolver(svc AutoIndexingService, index types.Index) IndexStepsResolver {
	return &indexStepsResolver{svc: svc, index: index}
}

func (r *indexStepsResolver) Setup() []ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix("setup.")
}

func (r *indexStepsResolver) PreIndex() []PreIndexStepResolver {
	var resolvers []PreIndexStepResolver
	for i, step := range r.index.DockerSteps {
		if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.pre-index.%d", i)); ok {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.svc, step, &entry))
			// This is here for backwards compatibility for records that were created before
			// named keys for steps existed.
		} else if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.%d", i)); ok {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.svc, step, &entry))
		} else {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.svc, step, nil))
		}
	}

	return resolvers
}

func (r *indexStepsResolver) Index() IndexStepResolver {
	if entry, ok := r.findExecutionLogEntry("step.docker.indexer"); ok {
		return NewIndexStepResolver(r.svc, r.index, &entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.%d", len(r.index.DockerSteps))); ok {
		return NewIndexStepResolver(r.svc, r.index, &entry)
	}

	return NewIndexStepResolver(r.svc, r.index, nil)
}

func (r *indexStepsResolver) Upload() ExecutionLogEntryResolver {
	if entry, ok := r.findExecutionLogEntry("step.src.upload"); ok {
		return NewExecutionLogEntryResolver(r.svc, entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	if entry, ok := r.findExecutionLogEntry("step.src.0"); ok {
		return NewExecutionLogEntryResolver(r.svc, entry)
	}

	return nil
}

func (r *indexStepsResolver) Teardown() []ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix("teardown.")
}

func (r *indexStepsResolver) findExecutionLogEntry(key string) (workerutil.ExecutionLogEntry, bool) {
	for _, entry := range r.index.ExecutionLogs {
		if entry.Key == key {
			return entry, true
		}
	}

	return workerutil.ExecutionLogEntry{}, false
}

func (r *indexStepsResolver) executionLogEntryResolversWithPrefix(prefix string) []ExecutionLogEntryResolver {
	var resolvers []ExecutionLogEntryResolver
	for _, entry := range r.index.ExecutionLogs {
		if !strings.HasPrefix(entry.Key, prefix) {
			continue
		}
		r := NewExecutionLogEntryResolver(r.svc, entry)
		resolvers = append(resolvers, r)
	}

	return resolvers
}
