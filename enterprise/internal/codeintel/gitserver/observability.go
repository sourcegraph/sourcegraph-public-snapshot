package gitserver

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
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
		})
	}

	return &operations{
		commitGraph:       op("CommitGraph"),
		directoryChildren: op("DirectoryChildren"),
		fileExists:        op("FileExists"),
		head:              op("Head"),
		listFiles:         op("ListFiles"),
		rawContents:       op("RawContents"),
	}
}
