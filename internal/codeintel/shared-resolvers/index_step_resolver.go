package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
)

type IndexStepResolver interface {
	IndexerArgs() []string
	Outfile() *string
	LogEntry() ExecutionLogEntryResolver
}

type indexStepResolver struct {
	svc   AutoIndexingService
	index shared.Index
	entry *shared.ExecutionLogEntry
}

func NewIndexStepResolver(svc AutoIndexingService, index shared.Index, entry *shared.ExecutionLogEntry) IndexStepResolver {
	return &indexStepResolver{
		svc:   svc,
		index: index,
		entry: entry,
	}
}

func (r *indexStepResolver) IndexerArgs() []string { return r.index.IndexerArgs }
func (r *indexStepResolver) Outfile() *string      { return strPtr(r.index.Outfile) }

func (r *indexStepResolver) LogEntry() ExecutionLogEntryResolver {
	if r.entry != nil {
		return NewExecutionLogEntryResolver(r.svc, *r.entry)
	}

	return nil
}
