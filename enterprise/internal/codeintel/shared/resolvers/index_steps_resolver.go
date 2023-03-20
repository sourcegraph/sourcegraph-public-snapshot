package sharedresolvers

import (
	"fmt"
	"regexp"

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
	siteAdminChecker SiteAdminChecker
	index            types.Index
}

func NewIndexStepsResolver(siteAdminChecker SiteAdminChecker, index types.Index) resolverstubs.IndexStepsResolver {
	return &indexStepsResolver{siteAdminChecker: siteAdminChecker, index: index}
}

func (r *indexStepsResolver) Setup() []resolverstubs.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix(logKeyPrefixSetup)
}

var logKeyPrefixSetup = regexp.MustCompile("^setup\\.")

func (r *indexStepsResolver) PreIndex() []resolverstubs.PreIndexStepResolver {
	var resolvers []resolverstubs.PreIndexStepResolver
	for i, step := range r.index.DockerSteps {
		keyPreIndex := regexp.MustCompile(fmt.Sprintf("^step\\.(kubernetes|docker)\\.pre-index\\.%d$", i))
		if entry, ok := r.findExecutionLogEntry(keyPreIndex); ok {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.siteAdminChecker, step, &entry))
			// This is here for backwards compatibility for records that were created before
			// named keys for steps existed.
		} else if entry, ok := r.findExecutionLogEntry(regexp.MustCompile(fmt.Sprintf("^step\\.(kubernetes|docker)\\.%d$", i))); ok {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.siteAdminChecker, step, &entry))
		} else {
			resolvers = append(resolvers, NewPreIndexStepResolver(r.siteAdminChecker, step, nil))
		}
	}

	return resolvers
}

func (r *indexStepsResolver) Index() resolverstubs.IndexStepResolver {
	if entry, ok := r.findExecutionLogEntry(logKeyIndexer); ok {
		return NewIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	keyDockerStep := regexp.MustCompile(fmt.Sprintf("^step\\.(kubernetes|docker)\\.%d$", len(r.index.DockerSteps)))
	if entry, ok := r.findExecutionLogEntry(keyDockerStep); ok {
		return NewIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	return NewIndexStepResolver(r.siteAdminChecker, r.index, nil)
}

var logKeyIndexer = regexp.MustCompile("^step\\.(kubernetes|docker)\\.indexer$")

func (r *indexStepsResolver) Upload() resolverstubs.ExecutionLogEntryResolver {
	if entry, ok := r.findExecutionLogEntry(logKeyUpload); ok {
		return NewExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	// This is here for backwards compatibility for records that were created before
	// named keys for steps existed.
	if entry, ok := r.findExecutionLogEntry(logKeySrcFirst); ok {
		return NewExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	return nil
}

var (
	// This is here for backwards compatibility for records that were created before
	// src became a docker step.
	logKeyUpload   = regexp.MustCompile("^step\\.(kubernetes|docker|src)\\.upload$")
	logKeySrcFirst = regexp.MustCompile("^step\\.src\\.0$")
)

func (r *indexStepsResolver) Teardown() []resolverstubs.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix(logKeyPrefixTeardown)
}

var logKeyPrefixTeardown = regexp.MustCompile("^teardown\\.")

func (r *indexStepsResolver) findExecutionLogEntry(key *regexp.Regexp) (executor.ExecutionLogEntry, bool) {
	for _, entry := range r.index.ExecutionLogs {
		if key.MatchString(entry.Key) {
			return entry, true
		}
	}

	return executor.ExecutionLogEntry{}, false
}

func (r *indexStepsResolver) executionLogEntryResolversWithPrefix(prefix *regexp.Regexp) []resolverstubs.ExecutionLogEntryResolver {
	var resolvers []resolverstubs.ExecutionLogEntryResolver
	for _, entry := range r.index.ExecutionLogs {
		if prefix.MatchString(entry.Key) {
			continue
		}
		r := NewExecutionLogEntryResolver(r.siteAdminChecker, entry)
		resolvers = append(resolvers, r)
	}

	return resolvers
}
