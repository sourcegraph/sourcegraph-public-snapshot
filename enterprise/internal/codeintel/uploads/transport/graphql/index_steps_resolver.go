package graphql

import (
	"fmt"
	"strings"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

// indexStepsResolver resolves the steps of an index record.
//
// Index jobs are broken into three parts:
//   - pre-index steps; all but the last docker step
//   - index step; the last docker step
//   - upload step; the only src-cli step
//
// The setup and teardown steps match the executor setup and teardown.
type indexStepsResolver struct {
	siteAdminChecker sharedresolvers.SiteAdminChecker
	index            types.Index
}

func NewIndexStepsResolver(siteAdminChecker sharedresolvers.SiteAdminChecker, index types.Index) resolverstubs.IndexStepsResolver {
	return &indexStepsResolver{siteAdminChecker: siteAdminChecker, index: index}
}

func (r *indexStepsResolver) Setup() []resolverstubs.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix("setup.")
}

func (r *indexStepsResolver) PreIndex() []resolverstubs.PreIndexStepResolver {
	var resolvers []resolverstubs.PreIndexStepResolver
	for i, step := range r.index.DockerSteps {
		if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.pre-index.%d", i)); ok {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.siteAdminChecker, step, &entry))
			// This is here for backwards compatibility for records that were created before
			// named keys for steps existed.
		} else if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.%d", i)); ok {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.siteAdminChecker, step, &entry))
		} else {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.siteAdminChecker, step, nil))
		}
	}

	return resolvers
}

func (r *indexStepsResolver) Index() resolverstubs.IndexStepResolver {
	if entry, ok := r.findExecutionLogEntry("step.docker.indexer"); ok {
		return NewIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.%d", len(r.index.DockerSteps))); ok {
		return NewIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	return NewIndexStepResolver(r.siteAdminChecker, r.index, nil)
}

func (r *indexStepsResolver) Upload() resolverstubs.ExecutionLogEntryResolver {
	if entry, ok := r.findExecutionLogEntry("step.docker.upload"); ok {
		return NewExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	// This is here for backwards compatibility for records that were created before
	// src became a docker step.
	if entry, ok := r.findExecutionLogEntry("step.src.upload"); ok {
		return NewExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	if entry, ok := r.findExecutionLogEntry("step.src.0"); ok {
		return NewExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	return nil
}

func (r *indexStepsResolver) Teardown() []resolverstubs.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix("teardown.")
}

func (r *indexStepsResolver) findExecutionLogEntry(key string) (executor.ExecutionLogEntry, bool) {
	for _, entry := range r.index.ExecutionLogs {
		if entry.Key == key {
			return entry, true
		}
	}

	return executor.ExecutionLogEntry{}, false
}

func (r *indexStepsResolver) executionLogEntryResolversWithPrefix(prefix string) []resolverstubs.ExecutionLogEntryResolver {
	var resolvers []resolverstubs.ExecutionLogEntryResolver
	for _, entry := range r.index.ExecutionLogs {
		if !strings.HasPrefix(entry.Key, prefix) {
			continue
		}
		r := NewExecutionLogEntryResolver(r.siteAdminChecker, entry)
		resolvers = append(resolvers, r)
	}

	return resolvers
}
