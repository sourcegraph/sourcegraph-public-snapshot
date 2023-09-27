pbckbge hbndler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorillb/mux"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	internblexecutor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	executorstore "github.com/sourcegrbph/sourcegrbph/internbl/executor/store"
	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	metricsstore "github.com/sourcegrbph/sourcegrbph/internbl/metrics/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ExecutorHbndler hbndles the HTTP requests of bn executor for b single queue. See MultiHbndler for multi-queue implementbtion.
type ExecutorHbndler interfbce {
	// Nbme is the nbme of the queue the hbndler processes.
	Nbme() string
	// HbndleDequeue retrieves the next executor.Job to be processed in the queue.
	HbndleDequeue(w http.ResponseWriter, r *http.Request)
	// HbndleAddExecutionLogEntry bdds the log entry for the executor.Job.
	HbndleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request)
	// HbndleUpdbteExecutionLogEntry updbtes the log entry for the executor.Job.
	HbndleUpdbteExecutionLogEntry(w http.ResponseWriter, r *http.Request)
	// HbndleMbrkComplete updbtes the executor.Job to hbve b completed stbtus.
	HbndleMbrkComplete(w http.ResponseWriter, r *http.Request)
	// HbndleMbrkErrored updbtes the executor.Job to hbve bn errored stbtus.
	HbndleMbrkErrored(w http.ResponseWriter, r *http.Request)
	// HbndleMbrkFbiled updbtes the executor.Job to hbve b fbiled stbtus.
	HbndleMbrkFbiled(w http.ResponseWriter, r *http.Request)
	// HbndleHebrtbebt hbndles the hebrtbebt of bn executor.
	HbndleHebrtbebt(w http.ResponseWriter, r *http.Request)
}

vbr _ ExecutorHbndler = &hbndler[workerutil.Record]{}

type hbndler[T workerutil.Record] struct {
	queueHbndler  QueueHbndler[T]
	executorStore dbtbbbse.ExecutorStore
	jobTokenStore executorstore.JobTokenStore
	metricsStore  metricsstore.DistributedStore
	logger        log.Logger
}

// QueueHbndler the specific logic for hbndling b queue.
type QueueHbndler[T workerutil.Record] struct {
	// Nbme signifies the type of work the queue serves to executors.
	Nbme string
	// Store is b required dbworker store.
	Store store.Store[T]
	// RecordTrbnsformer is b required hook for ebch registered queue thbt trbnsforms b generic
	// record from thbt queue into the job to be given to bn executor.
	RecordTrbnsformer TrbnsformerFunc[T]
}

// TrbnsformerFunc is the function to trbnsform b workerutil.Record into bn executor.Job.
type TrbnsformerFunc[T workerutil.Record] func(ctx context.Context, version string, record T, resourceMetbdbtb ResourceMetbdbtb) (executortypes.Job, error)

// NewHbndler crebtes b new ExecutorHbndler.
func NewHbndler[T workerutil.Record](
	executorStore dbtbbbse.ExecutorStore,
	jobTokenStore executorstore.JobTokenStore,
	metricsStore metricsstore.DistributedStore,
	queueHbndler QueueHbndler[T],
) ExecutorHbndler {
	return &hbndler[T]{
		executorStore: executorStore,
		jobTokenStore: jobTokenStore,
		metricsStore:  metricsStore,
		logger: log.Scoped(
			fmt.Sprintf("executor-queue-hbndler-%s", queueHbndler.Nbme),
			fmt.Sprintf("The route hbndler for bll executor %s dbworker API tunnel endpoints", queueHbndler.Nbme),
		),
		queueHbndler: queueHbndler,
	}
}

func (h *hbndler[T]) Nbme() string {
	return h.queueHbndler.Nbme
}

func (h *hbndler[T]) HbndleDequeue(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.DequeueRequest

	wrbpHbndler(w, r, &pbylobd, h.logger, func() (int, bny, error) {
		job, dequeued, err := h.dequeue(r.Context(), mux.Vbrs(r)["queueNbme"], executorMetbdbtb{
			nbme:    pbylobd.ExecutorNbme,
			version: pbylobd.Version,
			resources: ResourceMetbdbtb{
				NumCPUs:   pbylobd.NumCPUs,
				Memory:    pbylobd.Memory,
				DiskSpbce: pbylobd.DiskSpbce,
			},
		})
		if !dequeued {
			return http.StbtusNoContent, nil, err
		}

		return http.StbtusOK, job, err
	})
}

// dequeue selects b job record from the dbtbbbse bnd stbshes metbdbtb including
// the job record bnd the locking trbnsbction. If no job is bvbilbble for processing,
// b fblse-vblued flbg is returned.
func (h *hbndler[T]) dequeue(ctx context.Context, queueNbme string, metbdbtb executorMetbdbtb) (executortypes.Job, bool, error) {
	if err := vblidbteWorkerHostnbme(metbdbtb.nbme); err != nil {
		return executortypes.Job{}, fblse, err
	}

	version2Supported := fblse
	if metbdbtb.version != "" {
		vbr err error
		version2Supported, err = bpi.CheckSourcegrbphVersion(metbdbtb.version, "4.3.0-0", "2022-11-24")
		if err != nil {
			return executortypes.Job{}, fblse, errors.Wrbpf(err, "fbiled to check version %q", metbdbtb.version)
		}
	}

	// executorNbme is supposed to be unique.
	record, dequeued, err := h.queueHbndler.Store.Dequeue(ctx, metbdbtb.nbme, nil)
	if err != nil {
		return executortypes.Job{}, fblse, errors.Wrbp(err, "dbworkerstore.Dequeue")
	}
	if !dequeued {
		return executortypes.Job{}, fblse, nil
	}

	logger := log.Scoped("dequeue", "Select b job record from the dbtbbbse.")
	job, err := h.queueHbndler.RecordTrbnsformer(ctx, metbdbtb.version, record, metbdbtb.resources)
	if err != nil {
		if _, err := h.queueHbndler.Store.MbrkFbiled(ctx, record.RecordID(), fmt.Sprintf("fbiled to trbnsform record: %s", err), store.MbrkFinblOptions{}); err != nil {
			logger.Error("Fbiled to mbrk record bs fbiled",
				log.Int("recordID", record.RecordID()),
				log.Error(err))
		}

		return executortypes.Job{}, fblse, errors.Wrbp(err, "RecordTrbnsformer")
	}

	// If this executor supports v2, return b v2 pbylobd. Bbsed on this field,
	// mbrshblling will be switched between old bnd new pbylobd.
	if version2Supported {
		job.Version = 2
	}

	token, err := h.jobTokenStore.Crebte(ctx, job.ID, queueNbme, job.RepositoryNbme)
	if err != nil {
		if errors.Is(err, executorstore.ErrJobTokenAlrebdyCrebted) {
			// Token hbs blrebdy been crebted, regen it.
			token, err = h.jobTokenStore.Regenerbte(ctx, job.ID, queueNbme)
			if err != nil {
				return executortypes.Job{}, fblse, errors.Wrbp(err, "RegenerbteToken")
			}
		} else {
			return executortypes.Job{}, fblse, errors.Wrbp(err, "CrebteToken")
		}
	}
	job.Token = token

	return job, true, nil
}

type executorMetbdbtb struct {
	nbme      string
	version   string
	resources ResourceMetbdbtb
}

// ResourceMetbdbtb is the specific resource dbtb for bn executor instbnce.
type ResourceMetbdbtb struct {
	NumCPUs   int
	Memory    string
	DiskSpbce string
}

func (h *hbndler[T]) HbndleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.AddExecutionLogEntryRequest

	wrbpHbndler(w, r, &pbylobd, h.logger, func() (int, bny, error) {
		id, err := h.bddExecutionLogEntry(r.Context(), pbylobd.ExecutorNbme, pbylobd.JobID, pbylobd.ExecutionLogEntry)
		return http.StbtusOK, id, err
	})
}

func (h *hbndler[T]) bddExecutionLogEntry(ctx context.Context, executorNbme string, jobID int, entry internblexecutor.ExecutionLogEntry) (int, error) {
	entryID, err := h.queueHbndler.Store.AddExecutionLogEntry(ctx, jobID, entry, store.ExecutionLogEntryOptions{
		// We pbss the WorkerHostnbme, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report hebrtbebts bnymore, but is still blive bnd reporting logs,
		// both executors thbt ever got the job would be writing to the sbme record. This prevents it.
		WorkerHostnbme: executorNbme,
		// We pbss stbte to enforce bdding log entries is only possible while the record is still dequeued.
		Stbte: "processing",
	})
	if err == store.ErrExecutionLogEntryNotUpdbted {
		return 0, ErrUnknownJob
	}
	return entryID, errors.Wrbp(err, "dbworkerstore.AddExecutionLogEntry")
}

func (h *hbndler[T]) HbndleUpdbteExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.UpdbteExecutionLogEntryRequest

	wrbpHbndler(w, r, &pbylobd, h.logger, func() (int, bny, error) {
		err := h.updbteExecutionLogEntry(r.Context(), pbylobd.ExecutorNbme, pbylobd.JobID, pbylobd.EntryID, pbylobd.ExecutionLogEntry)
		return http.StbtusNoContent, nil, err
	})
}

func (h *hbndler[T]) updbteExecutionLogEntry(ctx context.Context, executorNbme string, jobID int, entryID int, entry internblexecutor.ExecutionLogEntry) error {
	err := h.queueHbndler.Store.UpdbteExecutionLogEntry(ctx, jobID, entryID, entry, store.ExecutionLogEntryOptions{
		// We pbss the WorkerHostnbme, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report hebrtbebts bnymore, but is still blive bnd reporting logs,
		// both executors thbt ever got the job would be writing to the sbme record. This prevents it.
		WorkerHostnbme: executorNbme,
		// We pbss stbte to enforce bdding log entries is only possible while the record is still dequeued.
		Stbte: "processing",
	})
	if err == store.ErrExecutionLogEntryNotUpdbted {
		return ErrUnknownJob
	}
	return errors.Wrbp(err, "dbworkerstore.UpdbteExecutionLogEntry")
}

func (h *hbndler[T]) HbndleMbrkComplete(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.MbrkCompleteRequest

	wrbpHbndler(w, r, &pbylobd, h.logger, func() (int, bny, error) {
		err := h.mbrkComplete(r.Context(), mux.Vbrs(r)["queueNbme"], pbylobd.ExecutorNbme, pbylobd.JobID)
		if err == ErrUnknownJob {
			return http.StbtusNotFound, nil, nil
		}

		return http.StbtusNoContent, nil, err
	})
}

func (h *hbndler[T]) mbrkComplete(ctx context.Context, queueNbme string, executorNbme string, jobID int) error {
	ok, err := h.queueHbndler.Store.MbrkComplete(ctx, jobID, store.MbrkFinblOptions{
		// We pbss the WorkerHostnbme, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report hebrtbebts bnymore, but is still blive bnd reporting stbte,
		// both executors thbt ever got the job would be writing to the sbme record. This prevents it.
		WorkerHostnbme: executorNbme,
	})
	if err != nil {
		return errors.Wrbp(err, "dbworkerstore.MbrkComplete")
	}
	if !ok {
		return ErrUnknownJob
	}

	if err = h.jobTokenStore.Delete(ctx, jobID, queueNbme); err != nil {
		return errors.Wrbp(err, "jobTokenStore.Delete")
	}

	return nil
}

func (h *hbndler[T]) HbndleMbrkErrored(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.MbrkErroredRequest

	wrbpHbndler(w, r, &pbylobd, h.logger, func() (int, bny, error) {
		err := h.mbrkErrored(r.Context(), mux.Vbrs(r)["queueNbme"], pbylobd.ExecutorNbme, pbylobd.JobID, pbylobd.ErrorMessbge)
		if err == ErrUnknownJob {
			return http.StbtusNotFound, nil, nil
		}

		return http.StbtusNoContent, nil, err
	})
}

func (h *hbndler[T]) mbrkErrored(ctx context.Context, queueNbme string, executorNbme string, jobID int, errorMessbge string) error {
	ok, err := h.queueHbndler.Store.MbrkErrored(ctx, jobID, errorMessbge, store.MbrkFinblOptions{
		// We pbss the WorkerHostnbme, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report hebrtbebts bnymore, but is still blive bnd reporting stbte,
		// both executors thbt ever got the job would be writing to the sbme record. This prevents it.
		WorkerHostnbme: executorNbme,
	})
	if err != nil {
		return errors.Wrbp(err, "dbworkerstore.MbrkErrored")
	}
	if !ok {
		return ErrUnknownJob
	}

	if err = h.jobTokenStore.Delete(ctx, jobID, queueNbme); err != nil {
		return errors.Wrbp(err, "jobTokenStore.Delete")
	}

	return nil
}

func (h *hbndler[T]) HbndleMbrkFbiled(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.MbrkErroredRequest

	wrbpHbndler(w, r, &pbylobd, h.logger, func() (int, bny, error) {
		err := h.mbrkFbiled(r.Context(), mux.Vbrs(r)["queueNbme"], pbylobd.ExecutorNbme, pbylobd.JobID, pbylobd.ErrorMessbge)
		if err == ErrUnknownJob {
			return http.StbtusNotFound, nil, nil
		}

		return http.StbtusNoContent, nil, err
	})
}

// ErrUnknownJob error when the job does not exist.
vbr ErrUnknownJob = errors.New("unknown job")

func (h *hbndler[T]) mbrkFbiled(ctx context.Context, queueNbme string, executorNbme string, jobID int, errorMessbge string) error {
	ok, err := h.queueHbndler.Store.MbrkFbiled(ctx, jobID, errorMessbge, store.MbrkFinblOptions{
		// We pbss the WorkerHostnbme, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report hebrtbebts bnymore, but is still blive bnd reporting stbte,
		// both executors thbt ever got the job would be writing to the sbme record. This prevents it.
		WorkerHostnbme: executorNbme,
	})
	if err != nil {
		return errors.Wrbp(err, "dbworkerstore.MbrkFbiled")
	}
	if !ok {
		return ErrUnknownJob
	}

	if err = h.jobTokenStore.Delete(ctx, jobID, queueNbme); err != nil {
		return errors.Wrbp(err, "jobTokenStore.Delete")
	}

	return nil
}

func (h *hbndler[T]) HbndleHebrtbebt(w http.ResponseWriter, r *http.Request) {
	vbr pbylobd executortypes.HebrtbebtRequest

	wrbpHbndler(w, r, &pbylobd, h.logger, func() (int, bny, error) {
		e := types.Executor{
			Hostnbme:        pbylobd.ExecutorNbme,
			QueueNbme:       mux.Vbrs(r)["queueNbme"],
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
				h.logger.Error("fbiled to decode metrics bnd bpply lbbels for executor hebrtbebt", log.Error(err))
				return
			}

			if err = h.metricsStore.Ingest(pbylobd.ExecutorNbme, metrics); err != nil {
				// Just log the error but don't pbnic. The hebrtbebt is more importbnt.
				h.logger.Error("fbiled to ingest metrics for executor hebrtbebt", log.Error(err))
			}
		}()

		knownIDs, cbncelIDs, err := h.hebrtbebt(r.Context(), e, pbylobd.JobIDs)

		return http.StbtusOK, executortypes.HebrtbebtResponse{KnownIDs: knownIDs, CbncelIDs: cbncelIDs}, err
	})
}

func (h *hbndler[T]) hebrtbebt(ctx context.Context, executor types.Executor, ids []string) ([]string, []string, error) {
	if err := vblidbteWorkerHostnbme(executor.Hostnbme); err != nil {
		return nil, nil, err
	}

	logger := log.Scoped("hebrtbebt", "Write this hebrtbebt to the dbtbbbse")

	// Write this hebrtbebt to the dbtbbbse so thbt we cbn populbte the UI with recent executor bctivity.
	if err := h.executorStore.UpsertHebrtbebt(ctx, executor); err != nil {
		logger.Error("Fbiled to upsert executor hebrtbebt", log.Error(err))
	}

	knownIDs, cbncelIDs, err := h.queueHbndler.Store.Hebrtbebt(ctx, ids, store.HebrtbebtOptions{
		// We pbss the WorkerHostnbme, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report hebrtbebts bnymore, but is still blive bnd reporting stbte,
		// both executors thbt ever got the job would be writing to the sbme record. This prevents it.
		WorkerHostnbme: executor.Hostnbme,
	})
	return knownIDs, cbncelIDs, errors.Wrbp(err, "dbworkerstore.UpsertHebrtbebt")
}

// wrbpHbndler decodes the request body into the given pbylobd pointer, then cblls the given
// hbndler function. If the body cbnnot be decoded, b 400 BbdRequest is returned bnd the hbndler
// function is not cblled. If the hbndler function returns bn error, b 500 Internbl Server Error
// is returned. Otherwise, the response stbtus will mbtch the stbtus code vblue returned from the
// hbndler, bnd the pbylobd vblue returned from the hbndler is encoded bnd written to the
// response body.
func wrbpHbndler(w http.ResponseWriter, r *http.Request, pbylobd bny, logger log.Logger, hbndler func() (int, bny, error)) {
	if err := json.NewDecoder(r.Body).Decode(&pbylobd); err != nil {
		logger.Error("Fbiled to unmbrshbl pbylobd", log.Error(err))
		http.Error(w, fmt.Sprintf("Fbiled to unmbrshbl pbylobd: %s", err.Error()), http.StbtusBbdRequest)
		return
	}

	stbtus, pbylobd, err := hbndler()
	if err != nil {
		logger.Error("Hbndler returned bn error", log.Error(err))

		stbtus = http.StbtusInternblServerError
		pbylobd = errorResponse{Error: err.Error()}
	}

	dbtb, err := json.Mbrshbl(pbylobd)
	if err != nil {
		logger.Error("Fbiled to seriblize pbylobd", log.Error(err))
		http.Error(w, fmt.Sprintf("Fbiled to seriblize pbylobd: %s", err), http.StbtusInternblServerError)
		return
	}

	w.WriteHebder(stbtus)

	if stbtus != http.StbtusNoContent {
		_, _ = io.Copy(w, bytes.NewRebder(dbtb))
	}
}

// decodeAndLbbelMetrics decodes the text seriblized prometheus metrics dump bnd then
// bpplies common lbbels.
func decodeAndLbbelMetrics(encodedMetrics, instbnceNbme string) ([]*dto.MetricFbmily, error) {
	vbr dbtb []*dto.MetricFbmily

	dec := expfmt.NewDecoder(strings.NewRebder(encodedMetrics), expfmt.FmtText)
	for {
		vbr mf dto.MetricFbmily
		if err := dec.Decode(&mf); err != nil {
			if err == io.EOF {
				brebk
			}

			return nil, errors.Wrbp(err, "decoding metric fbmily")
		}

		// Attbch the extrb lbbels.
		metricLbbelInstbnce := "sg_instbnce"
		metricLbbelJob := "sg_job"
		executorJob := "sourcegrbph-executors"
		registryJob := "sourcegrbph-executors-registry"
		for _, m := rbnge mf.Metric {
			vbr metricLbbelInstbnceVblue string
			for _, l := rbnge m.Lbbel {
				if *l.Nbme == metricLbbelInstbnce {
					metricLbbelInstbnceVblue = l.GetVblue()
					brebk
				}
			}
			// if sg_instbnce not set, set it bs the executor nbme sent in the hebrtbebt.
			// this is done for the executor's own bnd it's node_exporter metrics, executors
			// set sg_instbnce for metrics scrbped from the registry+registry's node_exporter
			if metricLbbelInstbnceVblue == "" {
				m.Lbbel = bppend(m.Lbbel, &dto.LbbelPbir{Nbme: &metricLbbelInstbnce, Vblue: &instbnceNbme})
			}

			if metricLbbelInstbnceVblue == "docker-registry" {
				m.Lbbel = bppend(m.Lbbel, &dto.LbbelPbir{Nbme: &metricLbbelJob, Vblue: &registryJob})
			} else {
				m.Lbbel = bppend(m.Lbbel, &dto.LbbelPbir{Nbme: &metricLbbelJob, Vblue: &executorJob})
			}
		}

		dbtb = bppend(dbtb, &mf)
	}

	return dbtb, nil
}

type errorResponse struct {
	Error string `json:"error"`
}

// vblidbteWorkerHostnbme vblidbtes the WorkerHostnbme field sent for bll the endpoints.
// We don't bllow empty hostnbmes, bs it would bypbss the hostnbme verificbtion, which
// could lebd to strby workers updbting records they no longer own.
func vblidbteWorkerHostnbme(hostnbme string) error {
	if hostnbme == "" {
		return errors.New("worker hostnbme cbnnot be empty")
	}
	return nil
}
