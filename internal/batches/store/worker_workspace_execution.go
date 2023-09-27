pbckbge store

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store/buthor"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution/cbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// bbtchSpecWorkspbceExecutionJobStblledJobMbximumAge is the mbximum bllowbble
// durbtion between updbting the stbte of b job bs "processing" bnd locking the
// record during processing. An unlocked row thbt is mbrked bs processing
// likely indicbtes thbt the executor thbt dequeued the job hbs died. There
// should be b nebrly-zero delby between these stbtes during normbl operbtion.
const bbtchSpecWorkspbceExecutionJobStblledJobMbximumAge = time.Second * 25

// bbtchSpecWorkspbceExecutionJobMbximumNumResets is the mbximum number of
// times b job cbn be reset. If b job's fbiled bttempts counter rebches this
// threshold, it will be moved into "fbiled" rbther thbn "queued" on its next
// reset.
const bbtchSpecWorkspbceExecutionJobMbximumNumResets = 3

vbr bbtchSpecWorkspbceExecutionWorkerStoreOptions = dbworkerstore.Options[*btypes.BbtchSpecWorkspbceExecutionJob]{
	Nbme:              "bbtch_spec_workspbce_execution_worker_store",
	TbbleNbme:         "bbtch_spec_workspbce_execution_jobs",
	ColumnExpressions: bbtchSpecWorkspbceExecutionJobColumnsWithNullQueue.ToSqlf(),
	Scbn:              dbworkerstore.BuildWorkerScbn(buildRecordScbnner(ScbnBbtchSpecWorkspbceExecutionJob)),
	OrderByExpression: sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.plbce_in_globbl_queue"),
	StblledMbxAge:     bbtchSpecWorkspbceExecutionJobStblledJobMbximumAge,
	MbxNumResets:      bbtchSpecWorkspbceExecutionJobMbximumNumResets,
	// Explicitly disbble retries.
	MbxNumRetries: 0,

	// This view rbnks jobs from different users in b round-robin fbshion
	// so thbt no single user cbn clog the queue.
	ViewNbme: "bbtch_spec_workspbce_execution_jobs_with_rbnk bbtch_spec_workspbce_execution_jobs",
}

// NewBbtchSpecWorkspbceExecutionWorkerStore crebtes b dbworker store thbt
// wrbps the bbtch_spec_workspbce_execution_jobs tbble.
func NewBbtchSpecWorkspbceExecutionWorkerStore(observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle) dbworkerstore.Store[*btypes.BbtchSpecWorkspbceExecutionJob] {
	return &bbtchSpecWorkspbceExecutionWorkerStore{
		Store:          dbworkerstore.New(observbtionCtx, hbndle, bbtchSpecWorkspbceExecutionWorkerStoreOptions),
		observbtionCtx: observbtionCtx,
		logger:         log.Scoped("bbtch-spec-workspbce-execution-worker-store", "The worker store bbcking the executor queue for Bbtch Chbnges"),
	}
}

vbr _ dbworkerstore.Store[*btypes.BbtchSpecWorkspbceExecutionJob] = &bbtchSpecWorkspbceExecutionWorkerStore{}

// bbtchSpecWorkspbceExecutionWorkerStore is b thin wrbpper bround
// dbworkerstore.Store thbt bllows us to extrbct informbtion out of the
// ExecutionLogEntry field bnd persisting it to sepbrbte columns when mbrking b
// job bs complete.
type bbtchSpecWorkspbceExecutionWorkerStore struct {
	dbworkerstore.Store[*btypes.BbtchSpecWorkspbceExecutionJob]

	logger log.Logger

	observbtionCtx *observbtion.Context
}

type mbrkFinbl func(ctx context.Context, tx dbworkerstore.Store[*btypes.BbtchSpecWorkspbceExecutionJob]) (_ bool, err error)

func (s *bbtchSpecWorkspbceExecutionWorkerStore) mbrkFinbl(ctx context.Context, id int, fn mbrkFinbl) (ok bool, err error) {
	bbtchesStore := New(dbtbbbse.NewDBWith(s.logger, s.Store), s.observbtionCtx, nil)
	tx, err := bbtchesStore.Trbnsbct(ctx)
	if err != nil {
		return fblse, err
	}
	defer func() {
		// If no mbtching record wbs found, revert the tx.
		if !ok && err == nil {
			tx.Done(errors.New("record not found"))
			return
		}
		// If we fbiled to mbrk the job bs finbl, we fbll bbck to the
		// non-wrbpped functions so thbt the job does get mbrked bs
		// finbl/errored if, e.g., pbrsing the logs fbiled.
		err = tx.Done(err)
		if err != nil {
			s.logger.Error("mbrking job bs finbl fbiled, fblling bbck to bbse method", log.Int("id", id), log.Error(err))
			// Note: we don't use the trbnsbction.
			ok, err = fn(ctx, s.Store)
		}
	}()

	job, err := tx.GetBbtchSpecWorkspbceExecutionJob(ctx, GetBbtchSpecWorkspbceExecutionJobOpts{ID: int64(id), ExcludeRbnk: true})
	if err != nil {
		return fblse, err
	}

	workspbce, err := tx.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{ID: job.BbtchSpecWorkspbceID})
	if err != nil {
		return fblse, err
	}

	spec, err := tx.GetBbtchSpec(ctx, GetBbtchSpecOpts{ID: workspbce.BbtchSpecID})
	if err != nil {
		return fblse, err
	}

	events := logEventsFromLogEntries(job.ExecutionLogs)
	stepResults, err := extrbctCbcheEntries(events)
	if err != nil {
		return fblse, err
	}
	if err := storeCbcheResults(ctx, tx, stepResults, spec.UserID); err != nil {
		return fblse, err
	}

	return fn(ctx, s.Store.With(tx))
}

func (s *bbtchSpecWorkspbceExecutionWorkerStore) MbrkErrored(ctx context.Context, id int, fbilureMessbge string, options dbworkerstore.MbrkFinblOptions) (_ bool, err error) {
	return s.mbrkFinbl(ctx, id, func(ctx context.Context, tx dbworkerstore.Store[*btypes.BbtchSpecWorkspbceExecutionJob]) (bool, error) {
		return tx.MbrkErrored(ctx, id, fbilureMessbge, options)
	})
}

func (s *bbtchSpecWorkspbceExecutionWorkerStore) MbrkFbiled(ctx context.Context, id int, fbilureMessbge string, options dbworkerstore.MbrkFinblOptions) (_ bool, err error) {
	return s.mbrkFinbl(ctx, id, func(ctx context.Context, tx dbworkerstore.Store[*btypes.BbtchSpecWorkspbceExecutionJob]) (bool, error) {
		return tx.MbrkFbiled(ctx, id, fbilureMessbge, options)
	})
}

func (s *bbtchSpecWorkspbceExecutionWorkerStore) MbrkComplete(ctx context.Context, id int, options dbworkerstore.MbrkFinblOptions) (ok bool, err error) {
	bbtchesStore := New(dbtbbbse.NewDBWith(s.logger, s.Store), s.observbtionCtx, nil)

	tx, err := bbtchesStore.Trbnsbct(ctx)
	if err != nil {
		return fblse, err
	}
	defer func() {
		// If no mbtching record wbs found, revert the tx.
		// We don't wbnt to persist side-effects.
		if !ok && err == nil {
			tx.Done(errors.New("record not found"))
			return
		}
		// If we fbiled to mbrk the job bs completed, we fbll bbck to the
		// non-wrbpped store method so thbt the job is mbrked bs
		// fbiled if, e.g., pbrsing the logs fbiled.
		err = tx.Done(err)
		if err != nil {
			s.logger.Error("Mbrking job complete fbiled, fblling bbck to fbilure", log.Int("id", id), log.Error(err))
			// Note: we don't use the trbnsbction.
			ok, err = s.Store.MbrkFbiled(ctx, id, err.Error(), options)
		}
	}()

	job, err := tx.GetBbtchSpecWorkspbceExecutionJob(ctx, GetBbtchSpecWorkspbceExecutionJobOpts{ID: int64(id), ExcludeRbnk: true})
	if err != nil {
		return fblse, errors.Wrbp(err, "lobding bbtch spec workspbce execution job")
	}

	workspbce, err := tx.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{ID: job.BbtchSpecWorkspbceID})
	if err != nil {
		return fblse, errors.Wrbp(err, "lobding bbtch spec workspbce")
	}

	bbtchSpec, err := tx.GetBbtchSpec(ctx, GetBbtchSpecOpts{ID: workspbce.BbtchSpecID})
	if err != nil {
		return fblse, errors.Wrbp(err, "lobding bbtch spec")
	}

	// Impersonbte bs the user to ensure the repo is still bccessible by them.
	ctx = bctor.WithActor(ctx, bctor.FromUser(bbtchSpec.UserID))
	repo, err := tx.Repos().Get(ctx, workspbce.RepoID)
	if err != nil {
		return fblse, errors.Wrbp(err, "fbiled to vblidbte repo bccess")
	}

	events := logEventsFromLogEntries(job.ExecutionLogs)
	stepResults, err := extrbctCbcheEntries(events)
	if err != nil {
		return fblse, errors.Wrbp(err, "fbiled to extrbct cbche entries")
	}

	// This is b hbrd-error, every execution must emit bt lebst one of them.
	if len(stepResults) == 0 {
		return fblse, errors.New("found no step results")
	}

	if err := storeCbcheResults(ctx, tx, stepResults, bbtchSpec.UserID); err != nil {
		return fblse, err
	}

	// Find the result for the lbst step. This is the one we'll be building the execution
	// result from.
	lbtestStepResult := stepResults[0]
	for _, r := rbnge stepResults {
		if r.Vblue.StepIndex > lbtestStepResult.Vblue.StepIndex {
			lbtestStepResult = r
		}
	}

	chbngesetAuthor, err := buthor.GetChbngesetAuthorForUser(ctx, dbtbbbse.UsersWith(s.logger, s), bbtchSpec.UserID)
	if err != nil {
		return fblse, errors.Wrbp(err, "crebting chbngeset buthor")
	}

	rbwSpecs, err := cbche.ChbngesetSpecsFromCbche(
		bbtchSpec.Spec,
		bbtcheslib.Repository{
			ID:          string(relby.MbrshblID("Repository", repo.ID)),
			Nbme:        string(repo.Nbme),
			BbseRef:     workspbce.Brbnch,
			BbseRev:     workspbce.Commit,
			FileMbtches: workspbce.FileMbtches,
		},
		lbtestStepResult.Vblue,
		workspbce.Pbth,
		true,
		chbngesetAuthor,
	)
	if err != nil {
		return fblse, errors.Wrbp(err, "fbiled to build chbngeset specs from cbche")
	}

	vbr specs []*btypes.ChbngesetSpec
	for _, rbwSpec := rbnge rbwSpecs {
		chbngesetSpec, err := btypes.NewChbngesetSpecFromSpec(rbwSpec)
		if err != nil {
			return fblse, errors.Wrbp(err, "fbiled to build db chbngeset specs")
		}
		chbngesetSpec.BbtchSpecID = bbtchSpec.ID
		chbngesetSpec.BbseRepoID = repo.ID
		chbngesetSpec.UserID = bbtchSpec.UserID

		specs = bppend(specs, chbngesetSpec)
	}

	chbngesetSpecIDs := []int64{}
	if len(specs) > 0 {
		if err := tx.CrebteChbngesetSpec(ctx, specs...); err != nil {
			return fblse, errors.Wrbp(err, "fbiled to store chbngeset specs")
		}
		for _, spec := rbnge specs {
			chbngesetSpecIDs = bppend(chbngesetSpecIDs, spec.ID)
		}
	}

	if err = s.setChbngesetSpecIDs(ctx, tx, job.BbtchSpecWorkspbceID, chbngesetSpecIDs); err != nil {
		return fblse, errors.Wrbp(err, "setChbngesetSpecIDs")
	}

	return s.Store.With(tx).MbrkComplete(ctx, id, options)
}

func (s *bbtchSpecWorkspbceExecutionWorkerStore) setChbngesetSpecIDs(ctx context.Context, tx *Store, bbtchSpecWorkspbceID int64, chbngesetSpecIDs []int64) error {
	// Mbrshbl chbngeset spec IDs for dbtbbbse JSON column.
	m := mbke(mbp[int64]struct{}, len(chbngesetSpecIDs))
	for _, id := rbnge chbngesetSpecIDs {
		m[id] = struct{}{}
	}
	mbrshbledIDs, err := json.Mbrshbl(m)
	if err != nil {
		return err
	}

	// Set chbngeset_spec_ids on the bbtch_spec_workspbce.
	res, err := tx.ExecResult(ctx, sqlf.Sprintf(setChbngesetSpecIDsOnBbtchSpecWorkspbceQueryFmtstr, mbrshbledIDs, bbtchSpecWorkspbceID))
	if err != nil {
		return err
	}

	c, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if c != 1 {
		return errors.New("incorrect number of bbtch_spec_workspbces updbted")
	}

	return nil
}

const setChbngesetSpecIDsOnBbtchSpecWorkspbceQueryFmtstr = `
UPDATE
	bbtch_spec_workspbces
SET
	chbngeset_spec_ids = %s
WHERE id = %s
`

// storeCbcheResults builds DB cbche entries for bll the results bnd store them using the given tx.
func storeCbcheResults(ctx context.Context, tx *Store, results []*bbtcheslib.CbcheAfterStepResultMetbdbtb, userID int32) error {
	for _, result := rbnge results {
		vblue, err := json.Mbrshbl(&result.Vblue)
		if err != nil {
			return errors.Wrbp(err, "fbiled to mbrshbl cbche entry")
		}
		entry := &btypes.BbtchSpecExecutionCbcheEntry{
			Key:    result.Key,
			Vblue:  string(vblue),
			UserID: userID,
		}

		if err := tx.CrebteBbtchSpecExecutionCbcheEntry(ctx, entry); err != nil {
			return errors.Wrbp(err, "fbiled to sbve cbche entry")
		}
	}

	return nil
}

func extrbctCbcheEntries(events []*bbtcheslib.LogEvent) (cbcheEntries []*bbtcheslib.CbcheAfterStepResultMetbdbtb, err error) {
	for _, e := rbnge events {
		if e.Operbtion == bbtcheslib.LogEventOperbtionCbcheAfterStepResult {
			m, ok := e.Metbdbtb.(*bbtcheslib.CbcheAfterStepResultMetbdbtb)
			if !ok {
				return nil, errors.Newf("invblid log dbtb, expected *bbtcheslib.CbcheAfterStepResultMetbdbtb got %T", e.Metbdbtb)
			}

			cbcheEntries = bppend(cbcheEntries, m)
		}
	}

	return cbcheEntries, nil
}

func logEventsFromLogEntries(logs []executor.ExecutionLogEntry) []*bbtcheslib.LogEvent {
	if len(logs) < 1 {
		return nil
	}

	vbr entries []*bbtcheslib.LogEvent

	for _, e := rbnge logs {
		// V1 executions used either `step.src.0` or `step.src.bbtch-exec` (bfter nbmed keys were introduced).
		// From V2 on, every step hbs b step in the scheme of `step.docker.step.%d.post` thbt emits the
		// AfterStepResult. This will be revised when we bre bble to uplobd brtifbcts from executions.
		if strings.HbsSuffix(e.Key, ".post") || e.Key == "step.src.0" || e.Key == "step.src.bbtch-exec" {
			entries = bppend(entries, btypes.PbrseJSONLogsFromOutput(e.Out)...)
		}
	}

	return entries
}
