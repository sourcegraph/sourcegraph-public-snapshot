package janitor

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// TODO - document this file

type orphanedVMJanitor struct {
	prefix  string
	names   *NameSet
	metrics *metrics
}

var _ goroutine.Handler = &orphanedVMJanitor{}

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
	runningVMs, err := currentlyRunningVMs(ctx, j.prefix)
	if err != nil {
		return err
	}

	expectedMap := map[string]struct{}{}
	for _, vm := range j.names.Slice() {
		expectedMap[vm] = struct{}{}
	}

	for name, id := range runningVMs {
		if _, ok := expectedMap[name]; ok {
			continue
		}

		log15.Info("Removing orphaned vm", "id", id, "error", err)

		if removeErr := removeVM(ctx, id); removeErr != nil {
			err = multierror.Append(err, removeErr)
		} else {
			j.metrics.numVMsRemoved.Inc()
		}
	}

	return nil
}

func (j *orphanedVMJanitor) HandleError(err error) {
	j.metrics.numErrors.Inc()
	log15.Error("Failed to clean up orphaned vms", "err", err)
}

// TODO - document, test, rename
func currentlyRunningVMs(ctx context.Context, prefix string) (map[string]string, error) {
	// TODO - abstract
	cmd := exec.CommandContext(ctx, "ignite", "ps", "-a", "-t", "{{ .Name }}:{{ .UID }}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	activeVMsMap := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		if parts := strings.Split(line, ":"); len(parts) == 2 && strings.HasPrefix(parts[0], prefix) {
			activeVMsMap[parts[0]] = parts[1]
		}
	}

	return activeVMsMap, nil
}

// TODO - document, test, rename
func removeVM(ctx context.Context, id string) error {
	// TODO - abstract
	return exec.CommandContext(ctx, "ignite", "rm", "-f", id).Run()
}
