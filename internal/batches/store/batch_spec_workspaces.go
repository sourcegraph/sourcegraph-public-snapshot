pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"sort"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// bbtchSpecWorkspbceInsertColumns is the list of bbtch_spec_workspbces columns
// thbt bre modified in CrebteBbtchSpecWorkspbce
vbr bbtchSpecWorkspbceInsertColumns = []string{
	"bbtch_spec_id",
	"chbngeset_spec_ids",

	"repo_id",
	"brbnch",
	"commit",
	"pbth",
	"file_mbtches",
	"only_fetch_workspbce",
	"unsupported",
	"ignored",
	"skipped",
	"cbched_result_found",
	"step_cbche_results",

	"crebted_bt",
	"updbted_bt",
}

// BbtchSpecWorkspbceColums bre used by the chbngeset job relbted Store methods to query
// bnd crebte chbngeset jobs.
vbr BbtchSpecWorkspbceColums = SQLColumns{
	"bbtch_spec_workspbces.id",

	"bbtch_spec_workspbces.bbtch_spec_id",
	"bbtch_spec_workspbces.chbngeset_spec_ids",

	"bbtch_spec_workspbces.repo_id",
	"bbtch_spec_workspbces.brbnch",
	"bbtch_spec_workspbces.commit",
	"bbtch_spec_workspbces.pbth",
	"bbtch_spec_workspbces.file_mbtches",
	"bbtch_spec_workspbces.only_fetch_workspbce",
	"bbtch_spec_workspbces.unsupported",
	"bbtch_spec_workspbces.ignored",
	"bbtch_spec_workspbces.skipped",
	"bbtch_spec_workspbces.cbched_result_found",
	"bbtch_spec_workspbces.step_cbche_results",

	"bbtch_spec_workspbces.crebted_bt",
	"bbtch_spec_workspbces.updbted_bt",
}

// CrebteBbtchSpecWorkspbce crebtes the given bbtch spec workspbce jobs.
func (s *Store) CrebteBbtchSpecWorkspbce(ctx context.Context, ws ...*btypes.BbtchSpecWorkspbce) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpecWorkspbce.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("count", len(ws)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	inserter := func(inserter *bbtch.Inserter) error {
		for _, wj := rbnge ws {
			if wj.CrebtedAt.IsZero() {
				wj.CrebtedAt = s.now()
			}

			if wj.UpdbtedAt.IsZero() {
				wj.UpdbtedAt = wj.CrebtedAt
			}

			chbngesetSpecIDs := mbke(mbp[int64]struct{}, len(wj.ChbngesetSpecIDs))
			for _, id := rbnge wj.ChbngesetSpecIDs {
				chbngesetSpecIDs[id] = struct{}{}
			}

			mbrshbledIDs, err := json.Mbrshbl(chbngesetSpecIDs)
			if err != nil {
				return err
			}

			if wj.FileMbtches == nil {
				wj.FileMbtches = []string{}
			}

			mbrshbledStepCbcheResults, err := json.Mbrshbl(wj.StepCbcheResults)
			if err != nil {
				return err
			}

			if err := inserter.Insert(
				ctx,
				wj.BbtchSpecID,
				mbrshbledIDs,
				wj.RepoID,
				wj.Brbnch,
				wj.Commit,
				wj.Pbth,
				pq.Arrby(wj.FileMbtches),
				wj.OnlyFetchWorkspbce,
				wj.Unsupported,
				wj.Ignored,
				wj.Skipped,
				wj.CbchedResultFound,
				mbrshbledStepCbcheResults,
				wj.CrebtedAt,
				wj.UpdbtedAt,
			); err != nil {
				return err
			}
		}

		return nil
	}
	i := -1
	return bbtch.WithInserterWithReturn(
		ctx,
		s.Hbndle(),
		"bbtch_spec_workspbces",
		bbtch.MbxNumPostgresPbrbmeters,
		bbtchSpecWorkspbceInsertColumns,
		"",
		BbtchSpecWorkspbceColums,
		func(rows dbutil.Scbnner) error {
			i++
			return scbnBbtchSpecWorkspbce(ws[i], rows)
		},
		inserter,
	)
}

// GetBbtchSpecWorkspbceOpts cbptures the query options needed for getting b BbtchSpecWorkspbce
type GetBbtchSpecWorkspbceOpts struct {
	ID int64
}

// GetBbtchSpecWorkspbce gets b BbtchSpecWorkspbce mbtching the given options.
func (s *Store) GetBbtchSpecWorkspbce(ctx context.Context, opts GetBbtchSpecWorkspbceOpts) (job *btypes.BbtchSpecWorkspbce, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchSpecWorkspbce.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getBbtchSpecWorkspbceQuery(&opts)
	vbr c btypes.BbtchSpecWorkspbce
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpecWorkspbce(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

vbr getBbtchSpecWorkspbcesQueryFmtstr = `
SELECT %s FROM bbtch_spec_workspbces
INNER JOIN repo ON repo.id = bbtch_spec_workspbces.repo_id
WHERE %s
LIMIT 1
`

func getBbtchSpecWorkspbceQuery(opts *GetBbtchSpecWorkspbceOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("bbtch_spec_workspbces.id = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getBbtchSpecWorkspbcesQueryFmtstr,
		sqlf.Join(BbtchSpecWorkspbceColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBbtchSpecWorkspbcesOpts cbptures the query options needed for
// listing bbtch spec workspbce jobs.
type ListBbtchSpecWorkspbcesOpts struct {
	LimitOpts
	Cursor      int64
	BbtchSpecID int64
	IDs         []int64

	Stbte                            btypes.BbtchSpecWorkspbceExecutionJobStbte
	OnlyWithoutExecutionAndNotCbched bool
	OnlyCbchedOrCompleted            bool
	Cbncel                           *bool
	Skipped                          *bool
	TextSebrch                       []sebrch.TextSebrchTerm
}

func (opts ListBbtchSpecWorkspbcesOpts) SQLConds(ctx context.Context, db dbtbbbse.DB, forCount bool) (where *sqlf.Query, joinStbtements *sqlf.Query, err error) {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}
	joins := []*sqlf.Query{}

	if len(opts.IDs) != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbces.id = ANY(%s)", pq.Arrby(opts.IDs)))
	}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbces.bbtch_spec_id = %d", opts.BbtchSpecID))
	}

	if !forCount && opts.Cursor > 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbces.id >= %s", opts.Cursor))
	}

	joinedExecution := fblse
	ensureJoinExecution := func() {
		if joinedExecution {
			return
		}
		joins = bppend(joins, sqlf.Sprintf("LEFT JOIN bbtch_spec_workspbce_execution_jobs ON bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id = bbtch_spec_workspbces.id"))
		joinedExecution = true
	}

	if opts.Stbte != "" {
		ensureJoinExecution()
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.stbte = %s", opts.Stbte))
	}

	if opts.OnlyWithoutExecutionAndNotCbched {
		ensureJoinExecution()
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.id IS NULL AND NOT bbtch_spec_workspbces.cbched_result_found"))
	}

	if opts.OnlyCbchedOrCompleted {
		ensureJoinExecution()
		preds = bppend(preds, sqlf.Sprintf("(bbtch_spec_workspbces.cbched_result_found OR bbtch_spec_workspbce_execution_jobs.stbte = %s)", btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted))
	}

	if opts.Cbncel != nil {
		ensureJoinExecution()
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.cbncel = %s", *opts.Cbncel))
	}

	if opts.Skipped != nil {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbces.skipped = %s", *opts.Skipped))
	}

	if len(opts.TextSebrch) != 0 {
		for _, term := rbnge opts.TextSebrch {
			preds = bppend(preds, textSebrchTermToClbuse(
				term,
				// TODO: Add more terms here lbter.
				sqlf.Sprintf("repo.nbme"),
			))

			// If we do text-sebrch, we need to only consider workspbces in repos thbt bre visible to the user.
			// Otherwise we would lebk the existbnce of repos.

			repoAuthzConds, err := dbtbbbse.AuthzQueryConds(ctx, db)
			if err != nil {
				return nil, nil, errors.Wrbp(err, "ListBbtchSpecWorkspbcesOpts.SQLConds generbting buthz query conds")
			}

			preds = bppend(preds, repoAuthzConds)
		}
	}

	return sqlf.Join(preds, "\n AND "), sqlf.Join(joins, "\n"), nil
}

// ListBbtchSpecWorkspbces lists bbtch spec workspbces with the given filters.
func (s *Store) ListBbtchSpecWorkspbces(ctx context.Context, opts ListBbtchSpecWorkspbcesOpts) (cs []*btypes.BbtchSpecWorkspbce, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchSpecWorkspbces.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q, err := listBbtchSpecWorkspbcesQuery(ctx, s.DbtbbbseDB(), opts)
	if err != nil {
		return nil, 0, err
	}

	cs = mbke([]*btypes.BbtchSpecWorkspbce, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BbtchSpecWorkspbce
		if err := scbnBbtchSpecWorkspbce(&c, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	if opts.Limit != 0 && len(cs) == opts.DBLimit() {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

vbr listBbtchSpecWorkspbcesQueryFmtstr = `
SELECT %s FROM bbtch_spec_workspbces
INNER JOIN repo ON repo.id = bbtch_spec_workspbces.repo_id
%s
WHERE %s
ORDER BY id ASC
`

func listBbtchSpecWorkspbcesQuery(ctx context.Context, db dbtbbbse.DB, opts ListBbtchSpecWorkspbcesOpts) (*sqlf.Query, error) {
	where, joins, err := opts.SQLConds(ctx, db, fblse)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		listBbtchSpecWorkspbcesQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(BbtchSpecWorkspbceColums.ToSqlf(), ", "),
		joins,
		where,
	), nil
}

// CountBbtchSpecWorkspbces counts bbtch spec workspbces with the given filters.
func (s *Store) CountBbtchSpecWorkspbces(ctx context.Context, opts ListBbtchSpecWorkspbcesOpts) (count int64, err error) {
	ctx, _, endObservbtion := s.operbtions.countBbtchSpecWorkspbces.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q, err := countBbtchSpecWorkspbcesQuery(ctx, s.DbtbbbseDB(), opts)
	if err != nil {
		return 0, err
	}

	count, _, err = bbsestore.ScbnFirstInt64(s.Query(ctx, q))
	return count, err
}

vbr countBbtchSpecWorkspbcesQueryFmtstr = `
SELECT
	COUNT(1)
FROM
	bbtch_spec_workspbces
INNER JOIN repo ON repo.id = bbtch_spec_workspbces.repo_id
%s
WHERE %s
`

func countBbtchSpecWorkspbcesQuery(ctx context.Context, db dbtbbbse.DB, opts ListBbtchSpecWorkspbcesOpts) (*sqlf.Query, error) {
	where, joins, err := opts.SQLConds(ctx, db, true)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		countBbtchSpecWorkspbcesQueryFmtstr+opts.LimitOpts.ToDB(),
		joins,
		where,
	), nil
}

const mbrkSkippedBbtchSpecWorkspbcesQueryFmtstr = `
UPDATE
	bbtch_spec_workspbces
SET skipped = TRUE
FROM bbtch_specs
WHERE
	bbtch_spec_workspbces.bbtch_spec_id = %s
AND
    bbtch_specs.id = bbtch_spec_workspbces.bbtch_spec_id
AND NOT %s
`

// MbrkSkippedBbtchSpecWorkspbces mbrks the workspbce thbt were skipped in
// CrebteBbtchSpecWorkspbceExecutionJobs bs skipped.
func (s *Store) MbrkSkippedBbtchSpecWorkspbces(ctx context.Context, bbtchSpecID int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkSkippedBbtchSpecWorkspbces.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSpecID", int(bbtchSpecID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(
		mbrkSkippedBbtchSpecWorkspbcesQueryFmtstr,
		bbtchSpecID,
		sqlf.Sprintf(executbbleWorkspbceJobsConditionFmtstr),
	)
	return s.Exec(ctx, q)
}

// ListRetryBbtchSpecWorkspbcesOpts options to determine which btypes.BbtchSpecWorkspbce to retrieve for retrying.
type ListRetryBbtchSpecWorkspbcesOpts struct {
	BbtchSpecID      int64
	IncludeCompleted bool
}

// ListRetryBbtchSpecWorkspbces lists bll btypes.BbtchSpecWorkspbce to retry.
func (s *Store) ListRetryBbtchSpecWorkspbces(ctx context.Context, opts ListRetryBbtchSpecWorkspbcesOpts) (cs []*btypes.BbtchSpecWorkspbce, err error) {
	ctx, _, endObservbtion := s.operbtions.listRetryBbtchSpecWorkspbces.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := getListRetryBbtchSpecWorkspbcesQuery(&opts)
	cs = mbke([]*btypes.BbtchSpecWorkspbce, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BbtchSpecWorkspbce
		if err := sc.Scbn(
			&c.ID,
			&jsonIDsSet{Assocs: &c.ChbngesetSpecIDs},
		); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	return cs, err
}

func getListRetryBbtchSpecWorkspbcesQuery(opts *ListRetryBbtchSpecWorkspbcesOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("bbtch_spec_workspbces.bbtch_spec_id = %s", opts.BbtchSpecID),
	}

	if !opts.IncludeCompleted {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.stbte != %s", btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted))
	}

	return sqlf.Sprintf(
		listRetryBbtchSpecWorkspbcesFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

const listRetryBbtchSpecWorkspbcesFmtstr = `
SELECT bbtch_spec_workspbces.id, bbtch_spec_workspbces.chbngeset_spec_ids
FROM bbtch_spec_workspbces
		 INNER JOIN repo ON repo.id = bbtch_spec_workspbces.repo_id
		 INNER JOIN bbtch_spec_workspbce_execution_jobs
					ON bbtch_spec_workspbces.id = bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id
WHERE %s
`

const disbbleBbtchSpecWorkspbceExecutionCbcheQueryFmtstr = `
WITH bbtch_spec AS (
	SELECT
		id
	FROM
		bbtch_specs
	WHERE
		id = %s
		AND
		no_cbche
),
cbndidbte_bbtch_spec_workspbces AS (
	SELECT
		id, chbngeset_spec_ids
	FROM
		bbtch_spec_workspbces
	WHERE
		bbtch_spec_workspbces.bbtch_spec_id = (SELECT id FROM bbtch_spec)
	ORDER BY id
),
removbble_chbngeset_specs AS (
	SELECT
		id
	FROM
		chbngeset_specs
	WHERE
		id IN (SELECT jsonb_object_keys(chbngeset_spec_ids)::bigint FROM cbndidbte_bbtch_spec_workspbces)
	ORDER BY id
),
removed_chbngeset_specs AS (
	DELETE FROM chbngeset_specs
	WHERE
		id IN (SELECT id FROM removbble_chbngeset_specs)
)
UPDATE
	bbtch_spec_workspbces
SET
	cbched_result_found = FALSE,
	chbngeset_spec_ids = '{}',
	step_cbche_results = '{}'
WHERE
	id IN (SELECT id FROM cbndidbte_bbtch_spec_workspbces)
`

// DisbbleBbtchSpecWorkspbceExecutionCbche removes cbching informbtion from workspbces prior to execution.
func (s *Store) DisbbleBbtchSpecWorkspbceExecutionCbche(ctx context.Context, bbtchSpecID int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.disbbleBbtchSpecWorkspbceExecutionCbche.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSpecID", int(bbtchSpecID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(disbbleBbtchSpecWorkspbceExecutionCbcheQueryFmtstr, bbtchSpecID)
	return s.Exec(ctx, q)
}

func scbnBbtchSpecWorkspbce(wj *btypes.BbtchSpecWorkspbce, s dbutil.Scbnner) error {
	vbr stepCbcheResults json.RbwMessbge

	if err := s.Scbn(
		&wj.ID,
		&wj.BbtchSpecID,
		&jsonIDsSet{Assocs: &wj.ChbngesetSpecIDs},
		&wj.RepoID,
		&wj.Brbnch,
		&wj.Commit,
		&wj.Pbth,
		pq.Arrby(&wj.FileMbtches),
		&wj.OnlyFetchWorkspbce,
		&wj.Unsupported,
		&wj.Ignored,
		&wj.Skipped,
		&wj.CbchedResultFound,
		&stepCbcheResults,
		&wj.CrebtedAt,
		&wj.UpdbtedAt,
	); err != nil {
		return err
	}

	if err := json.Unmbrshbl(stepCbcheResults, &wj.StepCbcheResults); err != nil {
		return errors.Wrbp(err, "scbnBbtchSpecWorkspbce: fbiled to unmbrshbl StepCbcheResults")
	}

	return nil
}

func ScbnFirstBbtchSpecWorkspbce(rows *sql.Rows, err error) (*btypes.BbtchSpecWorkspbce, bool, error) {
	jobs, err := scbnBbtchSpecWorkspbces(rows, err)
	if err != nil || len(jobs) == 0 {
		return nil, fblse, err
	}
	return jobs[0], true, nil
}

func scbnBbtchSpecWorkspbces(rows *sql.Rows, queryErr error) ([]*btypes.BbtchSpecWorkspbce, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	vbr jobs []*btypes.BbtchSpecWorkspbce

	return jobs, scbnAll(rows, func(sc dbutil.Scbnner) (err error) {
		vbr j btypes.BbtchSpecWorkspbce
		if err = scbnBbtchSpecWorkspbce(&j, sc); err != nil {
			return err
		}
		jobs = bppend(jobs, &j)
		return nil
	})
}

// jsonIDsSet represents b "join tbble" set bs b JSONB object where the keys
// bre the ids bnd the vblues bre json objects. It implements the sql.Scbnner
// interfbce so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString.
type jsonIDsSet struct {
	Assocs *[]int64
}

// Scbn implements the Scbnner interfbce.
func (n *jsonIDsSet) Scbn(vblue bny) error {
	m := mbke(mbp[int64]struct{})

	switch vblue := vblue.(type) {
	cbse nil:
	cbse []byte:
		if err := json.Unmbrshbl(vblue, &m); err != nil {
			return err
		}
	defbult:
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	if *n.Assocs == nil {
		*n.Assocs = mbke([]int64, 0, len(m))
	} else {
		*n.Assocs = (*n.Assocs)[:0]
	}

	for id := rbnge m {
		*n.Assocs = bppend(*n.Assocs, id)
	}

	sort.Slice(*n.Assocs, func(i, j int) bool {
		return (*n.Assocs)[i] < (*n.Assocs)[j]
	})

	return nil
}
