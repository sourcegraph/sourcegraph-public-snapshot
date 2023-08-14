package janitor

import (
	"context"
	"os/exec"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type orphanedVMJanitor struct {
	logger    log.Logger
	prefix    string
	names     *NameSet
	metrics   *metrics
	cmdRunner util.CmdRunner
}

var (
	_ goroutine.Handler      = &orphanedVMJanitor{}
	_ goroutine.ErrorHandler = &orphanedVMJanitor{}
)

// NewOrphanedVMJanitor returns a background routine that periodically removes all VMs
// on the host that are not known by the worker running within this executor instance.
func NewOrphanedVMJanitor(
	logger log.Logger,
	prefix string,
	names *NameSet,
	interval time.Duration,
	metrics *metrics,
	cmdRunner util.CmdRunner,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		newOrphanedVMJanitor(
			logger,
			prefix,
			names,
			metrics,
			cmdRunner,
		),
		goroutine.WithName("executors.orphaned-vm-janitor"),
		goroutine.WithDescription("deletes VMs from a previous executor instance"),
		goroutine.WithInterval(interval),
	)
}

func newOrphanedVMJanitor(
	logger log.Logger,
	prefix string,
	names *NameSet,
	metrics *metrics,
	cmdRunner util.CmdRunner,
) *orphanedVMJanitor {
	return &orphanedVMJanitor{
		logger:    logger,
		prefix:    prefix,
		names:     names,
		metrics:   metrics,
		cmdRunner: cmdRunner,
	}
}

func (j *orphanedVMJanitor) Handle(ctx context.Context) (err error) {
	vmsByName, err := ignite.ActiveVMsByName(ctx, j.cmdRunner, j.prefix, true)
	if err != nil {
		return err
	}

	for _, id := range findOrphanedVMs(vmsByName, j.names.Slice()) {
		j.logger.Info("Removing orphaned VM", log.String("id", id))

		if removeErr := exec.CommandContext(ctx, "ignite", "rm", "-f", id).Run(); removeErr != nil {
			err = errors.Append(err, removeErr)
		} else {
			j.metrics.numVMsRemoved.Inc()
		}
	}

	return nil
}

func (j *orphanedVMJanitor) HandleError(err error) {
	j.metrics.numErrors.Inc()
	j.logger.Error("Failed to remove up orphaned vms", log.Error(err))
}

// findOrphanedVMs returns the set of VM identifiers present in running VMs but
// absent from expected VMs. The runningVMs argument is expected to be a map from
// VM names to VM identifiers.
func findOrphanedVMs(runningVMs map[string]string, expectedVMs []string) []string {
	expectedMap := make(map[string]struct{}, len(expectedVMs))
	for _, vm := range expectedVMs {
		expectedMap[vm] = struct{}{}
	}

	ids := make([]string, 0, len(runningVMs))
	for name, id := range runningVMs {
		if _, ok := expectedMap[name]; ok {
			continue
		}

		ids = append(ids, id)
	}
	sort.Strings(ids)

	return ids
}
