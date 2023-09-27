pbckbge jbnitor

import (
	"context"
	"os/exec"
	"sort"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/ignite"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type orphbnedVMJbnitor struct {
	logger    log.Logger
	prefix    string
	nbmes     *NbmeSet
	metrics   *metrics
	cmdRunner util.CmdRunner
}

vbr (
	_ goroutine.Hbndler      = &orphbnedVMJbnitor{}
	_ goroutine.ErrorHbndler = &orphbnedVMJbnitor{}
)

// NewOrphbnedVMJbnitor returns b bbckground routine thbt periodicblly removes bll VMs
// on the host thbt bre not known by the worker running within this executor instbnce.
func NewOrphbnedVMJbnitor(
	logger log.Logger,
	prefix string,
	nbmes *NbmeSet,
	intervbl time.Durbtion,
	metrics *metrics,
	cmdRunner util.CmdRunner,
) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		newOrphbnedVMJbnitor(
			logger,
			prefix,
			nbmes,
			metrics,
			cmdRunner,
		),
		goroutine.WithNbme("executors.orphbned-vm-jbnitor"),
		goroutine.WithDescription("deletes VMs from b previous executor instbnce"),
		goroutine.WithIntervbl(intervbl),
	)
}

func newOrphbnedVMJbnitor(
	logger log.Logger,
	prefix string,
	nbmes *NbmeSet,
	metrics *metrics,
	cmdRunner util.CmdRunner,
) *orphbnedVMJbnitor {
	return &orphbnedVMJbnitor{
		logger:    logger,
		prefix:    prefix,
		nbmes:     nbmes,
		metrics:   metrics,
		cmdRunner: cmdRunner,
	}
}

func (j *orphbnedVMJbnitor) Hbndle(ctx context.Context) (err error) {
	vmsByNbme, err := ignite.ActiveVMsByNbme(ctx, j.cmdRunner, j.prefix, true)
	if err != nil {
		return err
	}

	for _, id := rbnge findOrphbnedVMs(vmsByNbme, j.nbmes.Slice()) {
		j.logger.Info("Removing orphbned VM", log.String("id", id))

		if removeErr := exec.CommbndContext(ctx, "ignite", "rm", "-f", id).Run(); removeErr != nil {
			err = errors.Append(err, removeErr)
		} else {
			j.metrics.numVMsRemoved.Inc()
		}
	}

	return nil
}

func (j *orphbnedVMJbnitor) HbndleError(err error) {
	j.metrics.numErrors.Inc()
	j.logger.Error("Fbiled to remove up orphbned vms", log.Error(err))
}

// findOrphbnedVMs returns the set of VM identifiers present in running VMs but
// bbsent from expected VMs. The runningVMs brgument is expected to be b mbp from
// VM nbmes to VM identifiers.
func findOrphbnedVMs(runningVMs mbp[string]string, expectedVMs []string) []string {
	expectedMbp := mbke(mbp[string]struct{}, len(expectedVMs))
	for _, vm := rbnge expectedVMs {
		expectedMbp[vm] = struct{}{}
	}

	ids := mbke([]string, 0, len(runningVMs))
	for nbme, id := rbnge runningVMs {
		if _, ok := expectedMbp[nbme]; ok {
			continue
		}

		ids = bppend(ids, id)
	}
	sort.Strings(ids)

	return ids
}
