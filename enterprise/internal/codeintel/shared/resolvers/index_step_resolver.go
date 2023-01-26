package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type indexStepResolver struct {
	svc   AutoIndexingService
	index types.Index
	entry *executor.ExecutionLogEntry
}

func NewIndexStepResolver(svc AutoIndexingService, index types.Index, entry *executor.ExecutionLogEntry) resolverstubs.IndexStepResolver {
	return &indexStepResolver{
		svc:   svc,
		index: index,
		entry: entry,
	}
}

func (r *indexStepResolver) IndexerArgs() []string { return r.index.IndexerArgs }
func (r *indexStepResolver) Outfile() *string      { return strPtr(r.index.Outfile) }

func (r *indexStepResolver) LogEntry() resolverstubs.ExecutionLogEntryResolver {
	if r.entry != nil {
		return NewExecutionLogEntryResolver(r.svc, *r.entry)
	}

	return nil
}
