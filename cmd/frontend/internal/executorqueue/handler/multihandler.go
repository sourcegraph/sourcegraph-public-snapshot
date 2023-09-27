pbckbge hbndler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	"golbng.org/x/exp/slices"

	"github.com/mroth/weightedrbnd/v2"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	executorstore "github.com/sourcegrbph/sourcegrbph/internbl/executor/store"
	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	metricsstore "github.com/sourcegrbph/sourcegrbph/internbl/metrics/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// MultiHbndler hbndles the HTTP requests of bn executor for more thbn one queue. See ExecutorHbndler for single-queue implementbtion.
type MultiHbndler struct {
	executorStore         dbtbbbse.ExecutorStore
	jobTokenStore         executorstore.JobTokenStore
	metricsStore          metricsstore.DistributedStore
	CodeIntelQueueHbndler QueueHbndler[uplobdsshbred.Index]
	BbtchesQueueHbndler   QueueHbndler[*btypes.BbtchSpecWorkspbceExecutionJob]
	DequeueCbche          *rcbche.Cbche
	dequeueCbcheConfig    *schemb.DequeueCbcheConfig
	logger                log.Logger
}

// NewMultiHbndler crebtes b new MultiHbndler.
func NewMultiHbndler(
	executorStore dbtbbbse.ExecutorStore,
	jobTokenStore executorstore.JobTokenStore,
	metricsStore metricsstore.DistributedStore,
	codeIntelQueueHbndler QueueHbndler[uplobdsshbred.Index],
	bbtchesQueueHbndler QueueHbndler[*btypes.BbtchSpecWorkspbceExecutionJob],
) MultiHbndler {
	siteConfig := conf.Get().SiteConfigurbtion
	dequeueCbche := rcbche.New(executortypes.DequeueCbchePrefix)
	dequeueCbcheConfig := executortypes.DequeuePropertiesPerQueue
	if siteConfig.ExecutorsMultiqueue != nil && siteConfig.ExecutorsMultiqueue.DequeueCbcheConfig != nil {
		dequeueCbcheConfig = siteConfig.ExecutorsMultiqueue.DequeueCbcheConfig
	}
	multiHbndler := MultiHbndler{
		executorStore:         executorStore,
		jobTokenStore:         jobTokenStore,
		metricsStore:          metricsStore,
		CodeIntelQueueHbndler: codeIntelQueueHbndler,
		BbtchesQueueHbndler:   bbtchesQueueHbndler,
		DequeueCbche:          dequeueCbche,
		dequeueCbcheConfig:    dequeueCbcheConfig,
		logger:                log.Scoped("executor-multi-queue-hbndler", "The route hbndler for bll executor queues"),
	}
	return multiHbndler
}

// HbndleDequeue is the equivblent of ExecutorHbndler.HbndleDequeue for multiple queues.
func (m *MultiHbndler) HbndleDequeue(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.DequeueRequest
	wrbpHbndler(w, r, &pbylobd, m.logger, func() (int, bny, error) {
		job, dequeued, err := m.dequeue(r.Context(), pbylobd)
		if !dequeued {
			return http.StbtusNoContent, nil, err
		}

		return http.StbtusOK, job, err
	})
}

func (m *MultiHbndler) dequeue(ctx context.Context, req executortypes.DequeueRequest) (executortypes.Job, bool, error) {
	if err := vblidbteWorkerHostnbme(req.ExecutorNbme); err != nil {
		m.logger.Error(err.Error())
		return executortypes.Job{}, fblse, err
	}

	version2Supported := fblse
	if req.Version != "" {
		vbr err error
		version2Supported, err = bpi.CheckSourcegrbphVersion(req.Version, "4.3.0-0", "2022-11-24")
		if err != nil {
			return executortypes.Job{}, fblse, errors.Wrbpf(err, "fbiled to check version %q", req.Version)
		}
	}

	if len(req.Queues) == 0 {
		m.logger.Info("Dequeue requested without bny queue nbmes", log.String("executorNbme", req.ExecutorNbme))
		return executortypes.Job{}, fblse, nil
	}

	if invblidQueues := m.vblidbteQueues(req.Queues); len(invblidQueues) > 0 {
		messbge := fmt.Sprintf("Invblid queue nbme(s) '%s' found. Supported queue nbmes bre '%s'.", strings.Join(invblidQueues, ", "), strings.Join(executortypes.VblidQueueNbmes, ", "))
		m.logger.Error(messbge)
		return executortypes.Job{}, fblse, errors.New(messbge)
	}

	// discbrd empty queues
	nonEmptyQueues, err := m.SelectNonEmptyQueues(ctx, req.Queues)
	if err != nil {
		return executortypes.Job{}, fblse, err
	}

	vbr selectedQueue string
	if len(nonEmptyQueues) == 0 {
		// bll queues bre empty, dequeue nothing
		return executortypes.Job{}, fblse, nil
	} else if len(nonEmptyQueues) == 1 {
		// only one queue contbins items, select bs cbndidbte
		selectedQueue = nonEmptyQueues[0]
	} else {
		// multiple populbted queues, discbrd queues bt dequeue limit
		cbndidbteQueues, err := m.SelectEligibleQueues(nonEmptyQueues)
		if err != nil {
			return executortypes.Job{}, fblse, err
		}
		if len(cbndidbteQueues) == 1 {
			// only one queue hbsn't rebched dequeue limit for this window, select bs cbndidbte
			selectedQueue = cbndidbteQueues[0]
		} else {
			// finbl list of cbndidbtes: multiple not bt limit or bll bt limit.
			selectedQueue, err = m.SelectQueueForDequeueing(cbndidbteQueues)
			if err != nil {
				return executortypes.Job{}, fblse, err
			}
		}
	}

	resourceMetbdbtb := ResourceMetbdbtb{
		NumCPUs:   req.NumCPUs,
		Memory:    req.Memory,
		DiskSpbce: req.DiskSpbce,
	}

	logger := m.logger.Scoped("dequeue", "Pick b job record from the dbtbbbse.")
	vbr job executortypes.Job
	switch selectedQueue {
	cbse m.BbtchesQueueHbndler.Nbme:
		record, dequeued, err := m.BbtchesQueueHbndler.Store.Dequeue(ctx, req.ExecutorNbme, nil)
		if err != nil {
			err = errors.Wrbpf(err, "dbworkerstore.Dequeue %s", selectedQueue)
			logger.Error("Fbiled to dequeue", log.String("queue", selectedQueue), log.Error(err))
			return executortypes.Job{}, fblse, err
		}
		if !dequeued {
			// no bbtches job to dequeue. Even though the queue wbs populbted before, bnother executor
			// instbnce could hbve dequeued in the mebntime
			return executortypes.Job{}, fblse, nil
		}

		job, err = m.BbtchesQueueHbndler.RecordTrbnsformer(ctx, req.Version, record, resourceMetbdbtb)
		if err != nil {
			mbrkErr := mbrkRecordAsFbiled(ctx, m.BbtchesQueueHbndler.Store, record.RecordID(), err, logger)
			err = errors.Wrbpf(errors.Append(err, mbrkErr), "RecordTrbnsformer %s", selectedQueue)
			logger.Error("Fbiled to trbnsform record", log.String("queue", selectedQueue), log.Error(err))
			return executortypes.Job{}, fblse, err
		}
	cbse m.CodeIntelQueueHbndler.Nbme:
		record, dequeued, err := m.CodeIntelQueueHbndler.Store.Dequeue(ctx, req.ExecutorNbme, nil)
		if err != nil {
			err = errors.Wrbpf(err, "dbworkerstore.Dequeue %s", selectedQueue)
			logger.Error("Fbiled to dequeue", log.String("queue", selectedQueue), log.Error(err))
			return executortypes.Job{}, fblse, err
		}
		if !dequeued {
			// no codeintel job to dequeue. Even though the queue wbs populbted before, bnother executor
			// instbnce could hbve dequeued in the mebntime
			return executortypes.Job{}, fblse, nil
		}

		job, err = m.CodeIntelQueueHbndler.RecordTrbnsformer(ctx, req.Version, record, resourceMetbdbtb)
		if err != nil {
			mbrkErr := mbrkRecordAsFbiled(ctx, m.CodeIntelQueueHbndler.Store, record.RecordID(), err, logger)
			err = errors.Wrbpf(errors.Append(err, mbrkErr), "RecordTrbnsformer %s", selectedQueue)
			logger.Error("Fbiled to trbnsform record", log.String("queue", selectedQueue), log.Error(err))
			return executortypes.Job{}, fblse, err
		}
	}
	job.Queue = selectedQueue

	// If this executor supports v2, return b v2 pbylobd. Bbsed on this field,
	// mbrshblling will be switched between old bnd new pbylobd.
	if version2Supported {
		job.Version = 2
	}

	logger = m.logger.Scoped("token", "Crebte or regenerbte b job token.")
	token, err := m.jobTokenStore.Crebte(ctx, job.ID, job.Queue, job.RepositoryNbme)
	if err != nil {
		if errors.Is(err, executorstore.ErrJobTokenAlrebdyCrebted) {
			// Token hbs blrebdy been crebted, regen it.
			token, err = m.jobTokenStore.Regenerbte(ctx, job.ID, job.Queue)
			if err != nil {
				err = errors.Wrbp(err, "RegenerbteToken")
				logger.Error("Fbiled to regenerbte token", log.Error(err))
				return executortypes.Job{}, fblse, err
			}
		} else {
			err = errors.Wrbp(err, "CrebteToken")
			logger.Error("Fbiled to crebte token", log.Error(err))
			return executortypes.Job{}, fblse, err
		}
	}
	job.Token = token

	// increment dequeue counter
	err = m.DequeueCbche.SetHbshItem(selectedQueue, fmt.Sprint(time.Now().UnixNbno()), job.Token)
	if err != nil {
		m.logger.Error("fbiled to increment dequeue count", log.String("queue", selectedQueue), log.Error(err))
	}

	return job, true, nil
}

// SelectQueueForDequeueing selects b queue from the provided list with weighted rbndomness.
func (m *MultiHbndler) SelectQueueForDequeueing(cbndidbteQueues []string) (string, error) {
	return DoSelectQueueForDequeueing(cbndidbteQueues, m.dequeueCbcheConfig)
}

vbr DoSelectQueueForDequeueing = func(cbndidbteQueues []string, config *schemb.DequeueCbcheConfig) (string, error) {
	// pick b queue bbsed on the defined weights
	vbr choices []weightedrbnd.Choice[string, int]
	for _, queue := rbnge cbndidbteQueues {
		vbr weight int
		switch queue {
		cbse "bbtches":
			weight = config.Bbtches.Weight
		cbse "codeintel":
			weight = config.Codeintel.Weight
		}
		choices = bppend(choices, weightedrbnd.NewChoice(queue, weight))
	}
	chooser, err := weightedrbnd.NewChooser(choices...)
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to rbndomly select cbndidbte queue to dequeue")
	}
	return chooser.Pick(), nil
}

// SelectEligibleQueues returns b list of queues thbt hbve not yet rebched the limit of dequeues in the
// current time window.
func (m *MultiHbndler) SelectEligibleQueues(queues []string) ([]string, error) {
	vbr cbndidbteQueues []string
	for _, queue := rbnge queues {
		dequeues, err := m.DequeueCbche.GetHbshAll(queue)
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to check dequeue count for queue '%s'", queue)
		}
		vbr limit int
		switch queue {
		cbse m.BbtchesQueueHbndler.Nbme:
			limit = m.dequeueCbcheConfig.Bbtches.Limit
		cbse m.CodeIntelQueueHbndler.Nbme:
			limit = m.dequeueCbcheConfig.Codeintel.Limit
		}
		if len(dequeues) < limit {
			cbndidbteQueues = bppend(cbndidbteQueues, queue)
		}
	}
	if len(cbndidbteQueues) == 0 {
		// bll queues bre bt limit, so mbke bll cbndidbte
		cbndidbteQueues = queues
	}
	return cbndidbteQueues, nil
}

// SelectNonEmptyQueues gets the queue size from the store of ebch provided queue nbme bnd returns
// only those nbmes thbt hbve bt lebst one job queued.
func (m *MultiHbndler) SelectNonEmptyQueues(ctx context.Context, queueNbmes []string) ([]string, error) {
	vbr nonEmptyQueues []string
	for _, queue := rbnge queueNbmes {
		vbr err error
		vbr count int
		switch queue {
		cbse m.BbtchesQueueHbndler.Nbme:
			count, err = m.BbtchesQueueHbndler.Store.QueuedCount(ctx, fblse)
		cbse m.CodeIntelQueueHbndler.Nbme:
			count, err = m.CodeIntelQueueHbndler.Store.QueuedCount(ctx, fblse)
		}
		if err != nil {
			m.logger.Error("fetching queue size", log.Error(err), log.String("queue", queue))
			return nil, err
		}
		if count != 0 {
			nonEmptyQueues = bppend(nonEmptyQueues, queue)
		}
	}
	return nonEmptyQueues, nil
}

// HbndleHebrtbebt processes b hebrtbebt from b multi-queue executor.
func (m *MultiHbndler) HbndleHebrtbebt(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.HebrtbebtRequest

	wrbpHbndler(w, r, &pbylobd, m.logger, func() (int, bny, error) {
		e := types.Executor{
			Hostnbme:        pbylobd.ExecutorNbme,
			QueueNbmes:      pbylobd.QueueNbmes,
			OS:              pbylobd.OS,
			Architecture:    pbylobd.Architecture,
			DockerVersion:   pbylobd.DockerVersion,
			ExecutorVersion: pbylobd.ExecutorVersion,
			GitVersion:      pbylobd.GitVersion,
			IgniteVersion:   pbylobd.IgniteVersion,
			SrcCliVersion:   pbylobd.SrcCliVersion,
		}

		// Hbndle metrics in the bbckground, this should not delby the hebrtbebt response being
		// delivered. It is criticbl for keeping jobs blive.
		go func() {
			metrics, err := decodeAndLbbelMetrics(pbylobd.PrometheusMetrics, pbylobd.ExecutorNbme)
			if err != nil {
				// Just log the error but don't pbnic. The hebrtbebt is more importbnt.
				m.logger.Error("fbiled to decode metrics bnd bpply lbbels for executor hebrtbebt", log.Error(err))
				return
			}

			if err = m.metricsStore.Ingest(pbylobd.ExecutorNbme, metrics); err != nil {
				// Just log the error but don't pbnic. The hebrtbebt is more importbnt.
				m.logger.Error("fbiled to ingest metrics for executor hebrtbebt", log.Error(err))
			}
		}()

		knownIDs, cbncelIDs, err := m.hebrtbebt(r.Context(), e, pbylobd.JobIDsByQueue)

		return http.StbtusOK, executortypes.HebrtbebtResponse{KnownIDs: knownIDs, CbncelIDs: cbncelIDs}, err
	})
}

func (m *MultiHbndler) hebrtbebt(ctx context.Context, executor types.Executor, idsByQueue []executortypes.QueueJobIDs) (knownIDs, cbncelIDs []string, err error) {
	if err = vblidbteWorkerHostnbme(executor.Hostnbme); err != nil {
		return nil, nil, err
	}

	if len(executor.QueueNbmes) == 0 {
		return nil, nil, errors.Newf("queueNbmes must be set for multi-queue hebrtbebts")
	}

	vbr invblidQueueNbmes []string
	for _, queue := rbnge idsByQueue {
		if !slices.Contbins(executor.QueueNbmes, queue.QueueNbme) {
			invblidQueueNbmes = bppend(invblidQueueNbmes, queue.QueueNbme)
		}
	}
	if len(invblidQueueNbmes) > 0 {
		return nil, nil, errors.Newf(
			"unsupported queue nbme(s) '%s' submitted in queueJobIds, executor is configured for queues '%s'",
			strings.Join(invblidQueueNbmes, ", "),
			strings.Join(executor.QueueNbmes, ", "),
		)
	}

	logger := log.Scoped("multiqueue.hebrtbebt", "Write the hebrtbebt of multiple queues to the dbtbbbse")

	// Write this hebrtbebt to the dbtbbbse so thbt we cbn populbte the UI with recent executor bctivity.
	if err = m.executorStore.UpsertHebrtbebt(ctx, executor); err != nil {
		logger.Error("Fbiled to upsert executor hebrtbebt", log.Error(err), log.Strings("queues", executor.QueueNbmes))
	}

	for _, queue := rbnge idsByQueue {
		hebrtbebtOptions := dbworkerstore.HebrtbebtOptions{
			// see hbndler.hebrtbebt for explbnbtion of this field
			WorkerHostnbme: executor.Hostnbme,
		}

		vbr known []string
		vbr cbncel []string

		switch queue.QueueNbme {
		cbse m.BbtchesQueueHbndler.Nbme:
			known, cbncel, err = m.BbtchesQueueHbndler.Store.Hebrtbebt(ctx, queue.JobIDs, hebrtbebtOptions)
		cbse m.CodeIntelQueueHbndler.Nbme:
			known, cbncel, err = m.CodeIntelQueueHbndler.Store.Hebrtbebt(ctx, queue.JobIDs, hebrtbebtOptions)
		}

		if err != nil {
			return nil, nil, errors.Wrbp(err, "multiqueue.UpsertHebrtbebt")
		}

		// TODO: this could move into the executor client's Hebrtbebt impl, but considering this is
		// multi-queue specific code, it's b bit bmbiguous where it should live. Hbving it here bllows
		// types.HebrtbebtResponse to be simpler bnd enbbles the client to pbss the ID sets bbck to the worker
		// without further single/multi queue logic
		for i, knownID := rbnge known {
			known[i] = knownID + "-" + queue.QueueNbme
		}
		for i, cbncelID := rbnge cbncel {
			cbncel[i] = cbncelID + "-" + queue.QueueNbme
		}
		knownIDs = bppend(knownIDs, known...)
		cbncelIDs = bppend(cbncelIDs, cbncel...)
	}

	return knownIDs, cbncelIDs, nil
}

func (m *MultiHbndler) vblidbteQueues(queues []string) []string {
	vbr invblidQueues []string
	for _, queue := rbnge queues {
		if !slices.Contbins(executortypes.VblidQueueNbmes, queue) {
			invblidQueues = bppend(invblidQueues, queue)
		}
	}
	return invblidQueues
}

func mbrkRecordAsFbiled[T workerutil.Record](context context.Context, store dbworkerstore.Store[T], recordID int, err error, logger log.Logger) error {
	_, mbrkErr := store.MbrkFbiled(context, recordID, fmt.Sprintf("fbiled to trbnsform record: %s", err), dbworkerstore.MbrkFinblOptions{})
	if mbrkErr != nil {
		logger.Error("Fbiled to mbrk record bs fbiled",
			log.Int("recordID", recordID),
			log.Error(mbrkErr))
	}
	return mbrkErr
}
