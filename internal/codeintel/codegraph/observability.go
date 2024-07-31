package codegraph

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	insertMetadata              *observation.Operation
	idsWithMeta                 *observation.Operation
	reconcileCandidates         *observation.Operation
	deleteLsifDataByUploadIds   *observation.Operation
	deleteUnreferencedDocuments *observation.Operation
	writerOperations            writerOperations
}

type writerOperations struct {
	insertDocuments       *observation.Operation
	insertDocumentLookups *observation.Operation
	constructTrie         *observation.Operation
	insertSymbolNames     *observation.Operation
	insertSymbols         *observation.Operation
	flushSymbolNames      *observation.Operation
	flushSymbols          *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_uploads_lsifstore",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.lsifstore.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &operations{
		insertMetadata:              op("InsertMetadata"),
		idsWithMeta:                 op("IDsWithMeta"),
		reconcileCandidates:         op("ReconcileCandidates"),
		deleteLsifDataByUploadIds:   op("DeleteLsifDataByUploadIds"),
		deleteUnreferencedDocuments: op("DeleteUnreferencedDocuments"),
		writerOperations: writerOperations{
			insertDocuments:       op("InsertDocuments"),
			insertDocumentLookups: op("InsertDocumentLookups"),
			constructTrie:         op("ConstructTrie"),
			insertSymbols:         op("InsertSymbols"),
			insertSymbolNames:     op("InsertSymbolNames"),
			flushSymbolNames:      op("flushSymbolNames"),
			flushSymbols:          op("flushSymbols"),
		},
	}
}
