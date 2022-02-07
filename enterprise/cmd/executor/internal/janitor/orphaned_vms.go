package janitor

import (
	"context"
	"os/exec"
	"sort"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type orphanedVMJanitor struct {
	prefix  string
	names   *NameSet
	metrics *metrics
}

var _ goroutine.Handler = &orphanedVMJanitor{}
var _ goroutine.ErrorHandler = &orphanedVMJanitor{}

// NewOrphanedVMJanitor returns a background routine that periodically removes all VMs
// on the host that are not known by the worker running within this executor instance.
func NewOrphanedVMJanitor(
	prefix string,
	names *NameSet,
	interval time.Duration,
	metrics *metrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, newOrphanedVMJanitor(
		prefix,
		names,
		metrics,
	))
}

func newOrphanedVMJanitor(
	prefix string,
	names *NameSet,
	metrics *metrics,
) *orphanedVMJanitor {
	return &orphanedVMJanitor{
		prefix:  prefix,
		names:   names,
		metrics: metrics,
	}
}

func (j *orphanedVMJanitor) Handle(ctx context.Context) (err error) {
	vmsByName, err := ignite.ActiveVMsByName(ctx, j.prefix, true)
	if err != nil {
		return err
	}

	for _, id := range findOrphanedVMs(vmsByName, j.names.Slice()) {
		log15.Info("Removing orphaned VM", "id", id)

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
	log15.Error("Failed to remove up orphaned vms", "error", err)
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
