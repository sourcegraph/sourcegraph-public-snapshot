package graphql

import (
	"context"
	"fmt"
	"strings"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
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
			resolvers = append(resolvers, newPreIndexStepResolver(r.siteAdminChecker, step, &entry))
			// This is here for backwards compatibility for records that were created before
			// named keys for steps existed.
		} else if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.%d", i)); ok {
			resolvers = append(resolvers, newPreIndexStepResolver(r.siteAdminChecker, step, &entry))
		} else {
			resolvers = append(resolvers, newPreIndexStepResolver(r.siteAdminChecker, step, nil))
		}
	}

	return resolvers
}

func (r *indexStepsResolver) Index() resolverstubs.IndexStepResolver {
	if entry, ok := r.findExecutionLogEntry("step.docker.indexer"); ok {
		return newIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	if entry, ok := r.findExecutionLogEntry(fmt.Sprintf("step.docker.%d", len(r.index.DockerSteps))); ok {
		return newIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	return newIndexStepResolver(r.siteAdminChecker, r.index, nil)
}

func (r *indexStepsResolver) Upload() resolverstubs.ExecutionLogEntryResolver {
	if entry, ok := r.findExecutionLogEntry("step.docker.upload"); ok {
		return newExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	// This is here for backwards compatibility for records that were created before
	// src became a docker step.
	if entry, ok := r.findExecutionLogEntry("step.src.upload"); ok {
		return newExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	if entry, ok := r.findExecutionLogEntry("step.src.0"); ok {
		return newExecutionLogEntryResolver(r.siteAdminChecker, entry)
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
		r := newExecutionLogEntryResolver(r.siteAdminChecker, entry)
		resolvers = append(resolvers, r)
	}

	return resolvers
}

//
//

type preIndexStepResolver struct {
	siteAdminChecker sharedresolvers.SiteAdminChecker
	step             types.DockerStep
	entry            *executor.ExecutionLogEntry
}

func newPreIndexStepResolver(siteAdminChecker sharedresolvers.SiteAdminChecker, step types.DockerStep, entry *executor.ExecutionLogEntry) resolverstubs.PreIndexStepResolver {
	return &preIndexStepResolver{
		siteAdminChecker: siteAdminChecker,
		step:             step,
		entry:            entry,
	}
}

func (r *preIndexStepResolver) Root() string       { return r.step.Root }
func (r *preIndexStepResolver) Image() string      { return r.step.Image }
func (r *preIndexStepResolver) Commands() []string { return r.step.Commands }

func (r *preIndexStepResolver) LogEntry() resolverstubs.ExecutionLogEntryResolver {
	if r.entry != nil {
		return newExecutionLogEntryResolver(r.siteAdminChecker, *r.entry)
	}

	return nil
}

//
//

type indexStepResolver struct {
	siteAdminChecker sharedresolvers.SiteAdminChecker
	index            types.Index
	entry            *executor.ExecutionLogEntry
}

func newIndexStepResolver(siteAdminChecker sharedresolvers.SiteAdminChecker, index types.Index, entry *executor.ExecutionLogEntry) resolverstubs.IndexStepResolver {
	return &indexStepResolver{
		siteAdminChecker: siteAdminChecker,
		index:            index,
		entry:            entry,
	}
}

func (r *indexStepResolver) Commands() []string    { return r.index.LocalSteps }
func (r *indexStepResolver) IndexerArgs() []string { return r.index.IndexerArgs }
func (r *indexStepResolver) Outfile() *string      { return resolverstubs.NonZeroPtr(r.index.Outfile) }

func (r *indexStepResolver) RequestedEnvVars() *[]string {
	if len(r.index.RequestedEnvVars) == 0 {
		return nil
	}
	return &r.index.RequestedEnvVars
}

func (r *indexStepResolver) LogEntry() resolverstubs.ExecutionLogEntryResolver {
	if r.entry != nil {
		return newExecutionLogEntryResolver(r.siteAdminChecker, *r.entry)
	}

	return nil
}

//
//

type executionLogEntryResolver struct {
	entry            executor.ExecutionLogEntry
	siteAdminChecker sharedresolvers.SiteAdminChecker
}

func newExecutionLogEntryResolver(siteAdminChecker sharedresolvers.SiteAdminChecker, entry executor.ExecutionLogEntry) resolverstubs.ExecutionLogEntryResolver {
	return &executionLogEntryResolver{
		entry:            entry,
		siteAdminChecker: siteAdminChecker,
	}
}

func (r *executionLogEntryResolver) Key() string       { return r.entry.Key }
func (r *executionLogEntryResolver) Command() []string { return r.entry.Command }

func (r *executionLogEntryResolver) ExitCode() *int32 {
	if r.entry.ExitCode == nil {
		return nil
	}
	val := int32(*r.entry.ExitCode)
	return &val
}

func (r *executionLogEntryResolver) StartTime() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.entry.StartTime}
}

func (r *executionLogEntryResolver) DurationMilliseconds() *int32 {
	if r.entry.DurationMs == nil {
		return nil
	}
	val := int32(*r.entry.DurationMs)
	return &val
}

func (r *executionLogEntryResolver) Out(ctx context.Context) (string, error) {
	// 🚨 SECURITY: Only site admins can view executor log contents.
	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		if err != auth.ErrMustBeSiteAdmin {
			return "", err
		}

		return "", nil
	}

	return r.entry.Out, nil
}
