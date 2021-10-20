package gitserver

import (
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	commitDate            *observation.Operation
	commitExists          *observation.Operation
	commitGraph           *observation.Operation
	commitsUniqueToBranch *observation.Operation
	directoryChildren     *observation.Operation
	fileExists            *observation.Operation
	head                  *observation.Operation
	listFiles             *observation.Operation
	rawContents           *observation.Operation
	refDescriptions       *observation.Operation
	repoInfo              *observation.Operation
	resolveRevision       *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_gitserver",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.gitserver.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
					return observation.EmitForNone
				}
				return observation.EmitForAll
			},
		})
	}

	return &operations{
		commitDate:            op("CommitDate"),
		commitExists:          op("CommitExists"),
		commitGraph:           op("CommitGraph"),
		commitsUniqueToBranch: op("CommitsUniqueToBranch"),
		directoryChildren:     op("DirectoryChildren"),
		fileExists:            op("FileExists"),
		head:                  op("Head"),
		listFiles:             op("ListFiles"),
		rawContents:           op("RawContents"),
		refDescriptions:       op("RefDescriptions"),
		repoInfo:              op("RepoInfo"),
		resolveRevision:       op("ResolveRevision"),
	}
}
