package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type IndexStepResolver interface {
	IndexerArgs() []string
	Outfile() *string
	LogEntry() ExecutionLogEntryResolver
}

type indexStepResolver struct {
	svc   AutoIndexingService
	index types.Index
	entry *workerutil.ExecutionLogEntry
}

func NewIndexStepResolver(svc AutoIndexingService, index types.Index, entry *workerutil.ExecutionLogEntry) IndexStepResolver {
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
