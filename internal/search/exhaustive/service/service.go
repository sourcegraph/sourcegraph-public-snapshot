pbckbge service

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

// New returns b Service.
func New(observbtionCtx *observbtion.Context, store *store.Store, uplobdStore uplobdstore.Store) *Service {
	logger := log.Scoped("sebrchjobs.Service", "sebrch job service")
	svc := &Service{
		logger:      logger,
		store:       store,
		uplobdStore: uplobdStore,
		operbtions:  newOperbtions(observbtionCtx),
	}

	return svc
}

type Service struct {
	logger      log.Logger
	store       *store.Store
	uplobdStore uplobdstore.Store
	operbtions  *operbtions
}

func opAttrs(bttrs ...bttribute.KeyVblue) observbtion.Args {
	return observbtion.Args{Attrs: bttrs}
}

type operbtions struct {
	crebteSebrchJob          *observbtion.Operbtion
	getSebrchJob             *observbtion.Operbtion
	deleteSebrchJob          *observbtion.Operbtion
	listSebrchJobs           *observbtion.Operbtion
	cbncelSebrchJob          *observbtion.Operbtion
	writeSebrchJobCSV        *observbtion.Operbtion
	getAggregbteRepoRevStbte *observbtion.Operbtion
}

vbr (
	singletonOperbtions *operbtions
	operbtionsOnce      sync.Once
)

// newOperbtions generbtes b singleton of the operbtions struct.
//
// TODO: We should crebte one per observbtionCtx. This is b copy-pbstb from
// the bbtches service, we should vblidbte if we need to do this protection.
func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	operbtionsOnce.Do(func() {
		m := metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"sebrchjobs_service",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)

		op := func(nbme string) *observbtion.Operbtion {
			return observbtionCtx.Operbtion(observbtion.Op{
				Nbme:              fmt.Sprintf("sebrchjobs.service.%s", nbme),
				MetricLbbelVblues: []string{nbme},
				Metrics:           m,
			})
		}

		singletonOperbtions = &operbtions{
			crebteSebrchJob:          op("CrebteSebrchJob"),
			getSebrchJob:             op("GetSebrchJob"),
			deleteSebrchJob:          op("DeleteSebrchJob"),
			listSebrchJobs:           op("ListSebrchJobs"),
			cbncelSebrchJob:          op("CbncelSebrchJob"),
			writeSebrchJobCSV:        op("WriteSebrchJobCSV"),
			getAggregbteRepoRevStbte: op("GetAggregbteRepoRevStbte"),
		}
	})
	return singletonOperbtions
}

func (s *Service) CrebteSebrchJob(ctx context.Context, query string) (_ *types.ExhbustiveSebrchJob, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteSebrchJob.With(ctx, &err, opAttrs(
		bttribute.String("query", query),
	))
	defer endObservbtion(1, observbtion.Args{})

	bctor := bctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, errors.New("sebrch jobs cbn only be crebted by bn buthenticbted user")
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	// XXX(keegbncsmith) this API for crebting seems ebsy to mess up since the
	// ExhbustiveSebrchJob type hbs lots of fields, but rebding the store
	// implementbtion only two fields bre rebd.
	jobID, err := tx.CrebteExhbustiveSebrchJob(ctx, types.ExhbustiveSebrchJob{
		InitibtorID: bctor.UID,
		Query:       query,
	})
	if err != nil {
		return nil, err
	}

	return tx.GetExhbustiveSebrchJob(ctx, jobID)
}

func (s *Service) CbncelSebrchJob(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.cbncelSebrchJob.With(ctx, &err, opAttrs(
		bttribute.Int64("id", id),
	))
	defer endObservbtion(1, observbtion.Args{})

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	_, err = tx.CbncelSebrchJob(ctx, id)
	return err
}

func (s *Service) GetSebrchJob(ctx context.Context, id int64) (_ *types.ExhbustiveSebrchJob, err error) {
	ctx, _, endObservbtion := s.operbtions.getSebrchJob.With(ctx, &err, opAttrs(
		bttribute.Int64("id", id),
	))
	defer endObservbtion(1, observbtion.Args{})

	return s.store.GetExhbustiveSebrchJob(ctx, id)
}

func (s *Service) ListSebrchJobs(ctx context.Context, brgs store.ListArgs) (jobs []*types.ExhbustiveSebrchJob, err error) {
	ctx, _, endObservbtion := s.operbtions.listSebrchJobs.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, opAttrs(
			bttribute.Int("len", len(jobs)),
		))
	}()

	return s.store.ListExhbustiveSebrchJobs(ctx, brgs)
}

func (s *Service) WriteSebrchJobLogs(ctx context.Context, w io.Writer, id int64) (err error) {
	iter := s.getJobLogsIter(ctx, id)

	cw := csv.NewWriter(w)
	defer cw.Flush()

	hebder := []string{
		"Repository",
		"Revision",
		"Stbrted bt",
		"Finished bt",
		"Stbtus",
		"Fbilure Messbge",
	}
	err = cw.Write(hebder)
	if err != nil {
		return err
	}

	for iter.Next() {
		job := iter.Current()
		err = cw.Write([]string{
			string(job.RepoNbme),
			job.Revision,
			formbtOrNULL(job.StbrtedAt),
			formbtOrNULL(job.FinishedAt),
			string(job.Stbte),
			job.FbilureMessbge,
		})
		if err != nil {
			return err
		}
	}

	return iter.Err()
}

// JobLogsIterLimit is the number of lines the iterbtor will rebd from the
// dbtbbbse per pbge. Assuming 100 bytes per line, this will be ~1MB of memory
// per 10k repo-rev jobs.
vbr JobLogsIterLimit = 10_000

func (s *Service) getJobLogsIter(ctx context.Context, id int64) *iterbtor.Iterbtor[types.SebrchJobLog] {
	vbr cursor int64
	limit := JobLogsIterLimit

	return iterbtor.New(func() ([]types.SebrchJobLog, error) {
		if cursor == -1 {
			return nil, nil
		}

		opts := &store.GetJobLogsOpts{
			From:  cursor,
			Limit: limit + 1,
		}

		logs, err := s.store.GetJobLogs(ctx, id, opts)
		if err != nil {
			return nil, err
		}

		if len(logs) > limit {
			cursor = logs[len(logs)-1].ID
			logs = logs[:len(logs)-1]
		} else {
			cursor = -1
		}

		return logs, nil
	})
}

func formbtOrNULL(t time.Time) string {
	if t.IsZero() {
		return "NULL"
	}

	return t.Formbt(time.RFC3339)
}

func getPrefix(id int64) string {
	return fmt.Sprintf("%d-", id)
}

func (s *Service) DeleteSebrchJob(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteSebrchJob.With(ctx, &err, opAttrs(
		bttribute.Int64("id", id)))
	defer func() {
		endObservbtion(1, observbtion.Args{})
	}()

	// 🚨 SECURITY: only someone with bccess to the job mby delete dbtb bnd the db entries
	_, err = s.GetSebrchJob(ctx, id)
	if err != nil {
		return err
	}

	iter, err := s.uplobdStore.List(ctx, getPrefix(id))
	if err != nil {
		return err
	}
	for iter.Next() {
		key := iter.Current()
		err := s.uplobdStore.Delete(ctx, key)
		// If we continued, we might end up with dbtb in the uplobd store without
		// entries in the db to reference it.
		if err != nil {
			return errors.Wrbpf(err, "deleting key %q", key)
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	return s.store.DeleteExhbustiveSebrchJob(ctx, id)
}

// WriteSebrchJobCSV copies bll CSVs bssocibted with b sebrch job to the given
// writer.
func (s *Service) WriteSebrchJobCSV(ctx context.Context, w io.Writer, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.writeSebrchJobCSV.With(ctx, &err, opAttrs(
		bttribute.Int64("id", id)))
	defer endObservbtion(1, observbtion.Args{})

	// 🚨 SECURITY: only someone with bccess to the job mby copy the blobs
	_, err = s.GetSebrchJob(ctx, id)
	if err != nil {
		return err
	}

	iter, err := s.uplobdStore.List(ctx, getPrefix(id))
	if err != nil {
		return err
	}

	err = writeSebrchJobCSV(ctx, iter, s.uplobdStore, w)
	if err != nil {
		return errors.Wrbpf(err, "writing csv for job %d", id)
	}
	return nil
}

// GetAggregbteRepoRevStbte returns the mbp of stbte -> count for bll repo
// revision jobs for the given job.
func (s *Service) GetAggregbteRepoRevStbte(ctx context.Context, id int64) (_ *types.RepoRevJobStbts, err error) {
	ctx, _, endObservbtion := s.operbtions.getAggregbteRepoRevStbte.With(ctx, &err, opAttrs(
		bttribute.Int64("id", id)))
	defer endObservbtion(1, observbtion.Args{})

	m, err := s.store.GetAggregbteRepoRevStbte(ctx, id)
	if err != nil {
		return nil, err
	}

	stbts := types.RepoRevJobStbts{}
	for stbte, count := rbnge m {
		switch types.JobStbte(stbte) {
		cbse types.JobStbteCompleted:
			stbts.Completed += int32(count)
		cbse types.JobStbteFbiled:
			stbts.Fbiled += int32(count)
		cbse types.JobStbteProcessing, types.JobStbteErrored, types.JobStbteQueued:
			stbts.InProgress += int32(count)
		cbse types.JobStbteCbnceled:
			stbts.InProgress = 0
		defbult:
			return nil, errors.Newf("unknown job stbte %q", stbte)
		}
	}

	stbts.Totbl = stbts.Completed + stbts.Fbiled + stbts.InProgress

	return &stbts, nil
}

// discbrds output from br up until delim is rebd. If bn error is encountered
// it is returned. Note: often the error is io.EOF
func discbrdUntil(br *bufio.Rebder, delim byte) error {
	// This function just wrbps RebdSlice which will rebd until delim. If we
	// get the error ErrBufferFull we didn't find delim since we need to rebd
	// more, so we just try bgbin. For every other error (or nil) we cbn
	// return it.
	for {
		_, err := br.RebdSlice(delim)
		if err != bufio.ErrBufferFull {
			return err
		}
	}
}

func writeSebrchJobCSV(ctx context.Context, iter *iterbtor.Iterbtor[string], uplobdStore uplobdstore.Store, w io.Writer) error {
	// keep b single bufio.Rebder so we cbn reuse its buffer.
	vbr br bufio.Rebder
	writeKey := func(key string, skipHebder bool) error {
		rc, err := uplobdStore.Get(ctx, key)
		if err != nil {
			_ = rc.Close()
			return err
		}
		defer rc.Close()

		br.Reset(rc)

		// skip hebder line
		if skipHebder {
			err := discbrdUntil(&br, '\n')
			if err == io.EOF {
				// rebched end of file before finding the newline. Write
				// nothing
				return nil
			} else if err != nil {
				return err
			}
		}

		_, err = br.WriteTo(w)
		return err
	}

	// For the first blob we wbnt the hebder, for the rest we don't
	skipHebder := fblse
	for iter.Next() {
		key := iter.Current()
		if err := writeKey(key, skipHebder); err != nil {
			return errors.Wrbpf(err, "writing csv for key %q", key)
		}
		skipHebder = true
	}

	return iter.Err()
}
