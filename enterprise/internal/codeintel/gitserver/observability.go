package gitserver

import (
	"errors"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	commitDate        *observation.Operation
	commitGraph       *observation.Operation
	directoryChildren *observation.Operation
	fileExists        *observation.Operation
	head              *observation.Operation
	listFiles         *observation.Operation
	rawContents       *observation.Operation
}

func makeOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_gitserver",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         fmt.Sprintf("codeintel.gitserver.%s", name),
			MetricLabels: []string{name},
			Metrics:      metrics,
			ErrorFilter: func(err error) bool {
				for ex := err; ex != nil; ex = errors.Unwrap(ex) {
					if gitserver.IsRevisionNotFound(ex) {
						return true
					}
				}

				return false
			},
		})
	}

	return &operations{
		commitDate:        op("CommitDate"),
		commitGraph:       op("CommitGraph"),
		directoryChildren: op("DirectoryChildren"),
		fileExists:        op("FileExists"),
		head:              op("Head"),
		listFiles:         op("ListFiles"),
		rawContents:       op("RawContents"),
	}
}
