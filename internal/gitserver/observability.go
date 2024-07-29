package gitserver

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type operations struct {
	archiveReader            *observation.Operation
	commits                  *observation.Operation
	contributorCount         *observation.Operation
	firstEverCommit          *observation.Operation
	behindAhead              *observation.Operation
	getCommit                *observation.Operation
	listRefs                 *observation.Operation
	lstat                    *observation.Operation
	mergeBase                *observation.Operation
	newFileReader            *observation.Operation
	readDir                  *observation.Operation
	resolveRevision          *observation.Operation
	revAtTime                *observation.Operation
	search                   *observation.Operation
	stat                     *observation.Operation
	streamBlameFile          *observation.Operation
	systemsInfo              *observation.Operation
	systemInfo               *observation.Operation
	isRepoCloneable          *observation.Operation
	repoCloneProgress        *observation.Operation
	isPerforcePathCloneable  *observation.Operation
	checkPerforceCredentials *observation.Operation
	perforceUsers            *observation.Operation
	perforceProtectsForUser  *observation.Operation
	perforceProtectsForDepot *observation.Operation
	perforceGroupMembers     *observation.Operation
	isPerforceSuperUser      *observation.Operation
	perforceGetChangelist    *observation.Operation
	createCommitFromPatch    *observation.Operation
	getObject                *observation.Operation
	getDefaultBranch         *observation.Operation
	diff                     *observation.Operation
	changedFiles             *observation.Operation
	mergeBaseOctopus         *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"gitserver_client",
		metrics.WithLabels("op", "scope"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("gitserver.client.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				return observation.EmitForAllExceptLogs
			},
		})
	}

	// suboperations do not have their own metrics but do have their own spans.
	// This allows us to more granularly track the latency for parts of a
	// request without noising up Prometheus.
	subOp := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name: fmt.Sprintf("gitserver.client.%s", name),
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				return observation.EmitForAllExceptLogs
			},
		})
	}

	// We don't want to send errors to sentry for `gitdomain.RevisionNotFoundError`
	// errors, as they should be actionable on the call site.
	resolveRevisionOperation := observationCtx.Operation(observation.Op{
		Name:              fmt.Sprintf("gitserver.client.%s", "ResolveRevision"),
		MetricLabelValues: []string{"ResolveRevision"},
		Metrics:           redMetrics,
		ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			if errors.HasType[*gitdomain.RevisionNotFoundError](err) {
				return observation.EmitForMetrics
			}
			return observation.EmitForSentry
		},
	})

	return &operations{
		archiveReader:            op("ArchiveReader"),
		commits:                  op("Commits"),
		contributorCount:         op("ContributorCount"),
		firstEverCommit:          op("FirstEverCommit"),
		behindAhead:              op("BehindAhead"),
		getCommit:                op("GetCommit"),
		listRefs:                 op("ListRefs"),
		lstat:                    subOp("lStat"),
		mergeBase:                op("MergeBase"),
		newFileReader:            op("NewFileReader"),
		readDir:                  op("ReadDir"),
		resolveRevision:          resolveRevisionOperation,
		revAtTime:                op("RevAtTime"),
		search:                   op("Search"),
		stat:                     op("Stat"),
		streamBlameFile:          op("StreamBlameFile"),
		systemsInfo:              op("SystemsInfo"),
		systemInfo:               op("SystemInfo"),
		isRepoCloneable:          op("IsRepoCloneable"),
		repoCloneProgress:        op("RepoCloneProgress"),
		isPerforcePathCloneable:  op("IsPerforcePathCloneable"),
		checkPerforceCredentials: op("CheckPerforceCredentials"),
		perforceUsers:            op("PerforceUsers"),
		perforceProtectsForUser:  op("PerforceProtectsForUser"),
		perforceProtectsForDepot: op("PerforceProtectsForDepot"),
		perforceGroupMembers:     op("PerforceGroupMembers"),
		isPerforceSuperUser:      op("IsPerforceSuperUser"),
		perforceGetChangelist:    op("PerforceGetChangelist"),
		createCommitFromPatch:    op("CreateCommitFromPatch"),
		getObject:                op("GetObject"),
		getDefaultBranch:         op("GetDefaultBranch"),
		diff:                     op("Diff"),
		changedFiles:             op("ChangedFiles"),
		mergeBaseOctopus:         op("MergeBaseOctopus"),
	}
}

var (
	operationsInst     *operations
	operationsInstOnce sync.Once
)

func getOperations() *operations {
	operationsInstOnce.Do(func() {
		observationCtx := observation.NewContext(log.Scoped("gitserver.client"))
		operationsInst = newOperations(observationCtx)
	})

	return operationsInst
}
