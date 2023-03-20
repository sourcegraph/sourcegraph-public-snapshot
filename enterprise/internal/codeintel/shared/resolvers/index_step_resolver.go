package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type indexStepResolver struct {
	siteAdminChecker SiteAdminChecker
	index            types.Index
	entry            *executor.ExecutionLogEntry
}

func NewIndexStepResolver(siteAdminChecker SiteAdminChecker, index types.Index, entry *executor.ExecutionLogEntry) resolverstubs.IndexStepResolver {
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
		return NewExecutionLogEntryResolver(r.siteAdminChecker, *r.entry)
	}

	return nil
}
