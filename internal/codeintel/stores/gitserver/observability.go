package gitserver

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type operations struct {
	commitDate            *observation.Operation
	commitExists          *observation.Operation
	commitGraph           *observation.Operation
	commitsExist          *observation.Operation
	commitsUniqueToBranch *observation.Operation
	directoryChildren     *observation.Operation
	fileExists            *observation.Operation
	head                  *observation.Operation
	listFiles             *observation.Operation
	rawContents           *observation.Operation
	refDescriptions       *observation.Operation
	repoInfo              *observation.Operation
	resolveRevision       *observation.Operation
	listTags              *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
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

				if gitdomain.IsCloneInProgress(err) {
					return observation.EmitForDefault ^ observation.EmitForLogs
				}

				return observation.EmitForDefault
			},
		})
	}

	return &operations{
		commitDate:            op("CommitDate"),
		commitExists:          op("CommitExists"),
		commitGraph:           op("CommitGraph"),
		commitsExist:          op("CommitsExist"),
		commitsUniqueToBranch: op("CommitsUniqueToBranch"),
		directoryChildren:     op("DirectoryChildren"),
		fileExists:            op("FileExists"),
		head:                  op("Head"),
		listFiles:             op("ListFiles"),
		rawContents:           op("RawContents"),
		refDescriptions:       op("RefDescriptions"),
		repoInfo:              op("RepoInfo"),
		resolveRevision:       op("ResolveRevision"),
		listTags:              op("ListTags"),
	}
}
