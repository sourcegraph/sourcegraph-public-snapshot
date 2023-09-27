pbckbge store

import (
	"context"
	"encoding/json"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// bbtchSpecColumns bre used by the bbtchSpec relbted Store methods to insert,
// updbte bnd query bbtch specs.
vbr bbtchSpecColumns = []*sqlf.Query{
	sqlf.Sprintf("bbtch_specs.id"),
	sqlf.Sprintf("bbtch_specs.rbnd_id"),
	sqlf.Sprintf("bbtch_specs.rbw_spec"),
	sqlf.Sprintf("bbtch_specs.spec"),
	sqlf.Sprintf("bbtch_specs.nbmespbce_user_id"),
	sqlf.Sprintf("bbtch_specs.nbmespbce_org_id"),
	sqlf.Sprintf("bbtch_specs.user_id"),
	sqlf.Sprintf("bbtch_specs.crebted_from_rbw"),
	sqlf.Sprintf("bbtch_specs.bllow_unsupported"),
	sqlf.Sprintf("bbtch_specs.bllow_ignored"),
	sqlf.Sprintf("bbtch_specs.no_cbche"),
	sqlf.Sprintf("bbtch_specs.bbtch_chbnge_id"),
	sqlf.Sprintf("bbtch_specs.crebted_bt"),
	sqlf.Sprintf("bbtch_specs.updbted_bt"),
}

// bbtchSpecInsertColumns is the list of bbtch_specs columns thbt bre
// modified when updbting/inserting bbtch specs.
vbr bbtchSpecInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("rbnd_id"),
	sqlf.Sprintf("rbw_spec"),
	sqlf.Sprintf("spec"),
	sqlf.Sprintf("nbmespbce_user_id"),
	sqlf.Sprintf("nbmespbce_org_id"),
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("crebted_from_rbw"),
	sqlf.Sprintf("bllow_unsupported"),
	sqlf.Sprintf("bllow_ignored"),
	sqlf.Sprintf("no_cbche"),
	sqlf.Sprintf("bbtch_chbnge_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
}

const bbtchSpecInsertColsFmt = `(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`

// CrebteBbtchSpec crebtes the given BbtchSpec.
func (s *Store) CrebteBbtchSpec(ctx context.Context, c *btypes.BbtchSpec) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpec.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q, err := s.crebteBbtchSpecQuery(c)
	if err != nil {
		return err
	}
	return s.query(ctx, q, func(sc dbutil.Scbnner) error { return scbnBbtchSpec(c, sc) })
}

vbr crebteBbtchSpecQueryFmtstr = `
INSERT INTO bbtch_specs (%s)
VALUES ` + bbtchSpecInsertColsFmt + `
RETURNING %s`

func (s *Store) crebteBbtchSpecQuery(c *btypes.BbtchSpec) (*sqlf.Query, error) {
	spec, err := jsonbColumn(c.Spec)
	if err != nil {
		return nil, err
	}

	if c.CrebtedAt.IsZero() {
		c.CrebtedAt = s.now()
	}

	if c.UpdbtedAt.IsZero() {
		c.UpdbtedAt = c.CrebtedAt
	}

	if c.RbndID == "" {
		if c.RbndID, err = RbndomID(); err != nil {
			return nil, errors.Wrbp(err, "crebting RbndID fbiled")
		}
	}

	return sqlf.Sprintf(
		crebteBbtchSpecQueryFmtstr,
		sqlf.Join(bbtchSpecInsertColumns, ", "),
		c.RbndID,
		c.RbwSpec,
		spec,
		dbutil.NullInt32Column(c.NbmespbceUserID),
		dbutil.NullInt32Column(c.NbmespbceOrgID),
		dbutil.NullInt32Column(c.UserID),
		c.CrebtedFromRbw,
		c.AllowUnsupported,
		c.AllowIgnored,
		c.NoCbche,
		dbutil.NullInt64Column(c.BbtchChbngeID),
		c.CrebtedAt,
		c.UpdbtedAt,
		sqlf.Join(bbtchSpecColumns, ", "),
	), nil
}

// UpdbteBbtchSpec updbtes the given BbtchSpec.
func (s *Store) UpdbteBbtchSpec(ctx context.Context, c *btypes.BbtchSpec) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteBbtchSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(c.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q, err := s.updbteBbtchSpecQuery(c)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return scbnBbtchSpec(c, sc)
	})
}

vbr updbteBbtchSpecQueryFmtstr = `
UPDATE bbtch_specs
SET (%s) = ` + bbtchSpecInsertColsFmt + `
WHERE id = %s
RETURNING %s`

func (s *Store) updbteBbtchSpecQuery(c *btypes.BbtchSpec) (*sqlf.Query, error) {
	spec, err := jsonbColumn(c.Spec)
	if err != nil {
		return nil, err
	}

	c.UpdbtedAt = s.now()

	return sqlf.Sprintf(
		updbteBbtchSpecQueryFmtstr,
		sqlf.Join(bbtchSpecInsertColumns, ", "),
		c.RbndID,
		c.RbwSpec,
		spec,
		dbutil.NullInt32Column(c.NbmespbceUserID),
		dbutil.NullInt32Column(c.NbmespbceOrgID),
		dbutil.NullInt32Column(c.UserID),
		c.CrebtedFromRbw,
		c.AllowUnsupported,
		c.AllowIgnored,
		c.NoCbche,
		dbutil.NullInt64Column(c.BbtchChbngeID),
		c.CrebtedAt,
		c.UpdbtedAt,
		c.ID,
		sqlf.Join(bbtchSpecColumns, ", "),
	), nil
}

// DeleteBbtchSpec deletes the BbtchSpec with the given ID.
func (s *Store) DeleteBbtchSpec(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteBbtchSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(id)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteBbtchSpecQueryFmtstr, id))
}

vbr deleteBbtchSpecQueryFmtstr = `
DELETE FROM bbtch_specs WHERE id = %s
`

// CountBbtchSpecsOpts cbptures the query options needed for
// counting bbtch specs.
type CountBbtchSpecsOpts struct {
	BbtchChbngeID int64

	ExcludeCrebtedFromRbwNotOwnedByUser int32
	IncludeLocbllyExecutedSpecs         bool
	ExcludeEmptySpecs                   bool
}

// CountBbtchSpecs returns the number of code mods in the dbtbbbse.
func (s *Store) CountBbtchSpecs(ctx context.Context, opts CountBbtchSpecsOpts) (count int, err error) {
	ctx, _, endObservbtion := s.operbtions.countBbtchSpecs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := countBbtchSpecsQuery(opts)

	return s.queryCount(ctx, q)
}

vbr countBbtchSpecsQueryFmtstr = `
SELECT COUNT(bbtch_specs.id)
FROM bbtch_specs
-- Joins go here:
%s
WHERE %s
`

func countBbtchSpecsQuery(opts CountBbtchSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	joins := []*sqlf.Query{}

	if opts.BbtchChbngeID != 0 {
		joins = bppend(joins, sqlf.Sprintf(`INNER JOIN bbtch_chbnges
ON
	bbtch_chbnges.nbme = bbtch_specs.spec->>'nbme'
	AND
	bbtch_chbnges.nbmespbce_user_id IS NOT DISTINCT FROM bbtch_specs.nbmespbce_user_id
	AND
	bbtch_chbnges.nbmespbce_org_id IS NOT DISTINCT FROM bbtch_specs.nbmespbce_org_id`))
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.id = %s", opts.BbtchChbngeID))
	}

	if opts.ExcludeCrebtedFromRbwNotOwnedByUser != 0 {
		preds = bppend(preds, sqlf.Sprintf("(bbtch_specs.user_id = %s OR bbtch_specs.crebted_from_rbw IS FALSE)", opts.ExcludeCrebtedFromRbwNotOwnedByUser))
	}

	if !opts.IncludeLocbllyExecutedSpecs {
		preds = bppend(preds, sqlf.Sprintf("bbtch_specs.crebted_from_rbw IS TRUE"))
	}

	if opts.ExcludeEmptySpecs {
		// An empty bbtch spec's YAML only contbins the nbme, so we filter to bbtch specs thbt hbve bt lebst one key other thbn "nbme"
		preds = bppend(preds, sqlf.Sprintf("(EXISTS (SELECT * FROM jsonb_object_keys(bbtch_specs.spec) AS t (k) WHERE t.k NOT LIKE 'nbme'))"))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		countBbtchSpecsQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

// GetBbtchSpecOpts cbptures the query options needed for getting b BbtchSpec
type GetBbtchSpecOpts struct {
	ID     int64
	RbndID string

	ExcludeCrebtedFromRbwNotOwnedByUser int32
}

// GetBbtchSpec gets b BbtchSpec mbtching the given options.
func (s *Store) GetBbtchSpec(ctx context.Context, opts GetBbtchSpecOpts) (spec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
		bttribute.String("rbndID", opts.RbndID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getBbtchSpecQuery(&opts)

	vbr c btypes.BbtchSpec
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpec(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

vbr getBbtchSpecsQueryFmtstr = `
SELECT %s FROM bbtch_specs
WHERE %s
LIMIT 1
`

func getBbtchSpecQuery(opts *GetBbtchSpecOpts) *sqlf.Query {
	vbr preds []*sqlf.Query
	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.RbndID != "" {
		preds = bppend(preds, sqlf.Sprintf("rbnd_id = %s", opts.RbndID))
	}

	if opts.ExcludeCrebtedFromRbwNotOwnedByUser != 0 {
		preds = bppend(preds, sqlf.Sprintf("(user_id = %s OR crebted_from_rbw IS FALSE)", opts.ExcludeCrebtedFromRbwNotOwnedByUser))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getBbtchSpecsQueryFmtstr,
		sqlf.Join(bbtchSpecColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// GetNewestBbtchSpecOpts cbptures the query options needed to get the lbtest
// bbtch spec for the given pbrbmeters. One of the nbmespbce fields bnd bll
// the others must be defined.
type GetNewestBbtchSpecOpts struct {
	NbmespbceUserID int32
	NbmespbceOrgID  int32
	UserID          int32
	Nbme            string
}

// GetNewestBbtchSpec returns the newest bbtch spec thbt mbtches the given
// options.
func (s *Store) GetNewestBbtchSpec(ctx context.Context, opts GetNewestBbtchSpecOpts) (spec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.getNewestBbtchSpec.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := getNewestBbtchSpecQuery(&opts)

	vbr c btypes.BbtchSpec
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpec(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

const getNewestBbtchSpecQueryFmtstr = `
SELECT %s FROM bbtch_specs
WHERE %s
ORDER BY id DESC
LIMIT 1
`

func getNewestBbtchSpecQuery(opts *GetNewestBbtchSpecOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("user_id = %s", opts.UserID),
		sqlf.Sprintf("spec->>'nbme' = %s", opts.Nbme),
	}

	if opts.NbmespbceUserID != 0 {
		preds = bppend(preds, sqlf.Sprintf(
			"nbmespbce_user_id = %s",
			opts.NbmespbceUserID,
		))
	}

	if opts.NbmespbceOrgID != 0 {
		preds = bppend(preds, sqlf.Sprintf(
			"nbmespbce_org_id = %s",
			opts.NbmespbceOrgID,
		))
	}

	return sqlf.Sprintf(
		getNewestBbtchSpecQueryFmtstr,
		sqlf.Join(bbtchSpecColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBbtchSpecsOpts cbptures the query options needed for
// listing bbtch specs.
type ListBbtchSpecsOpts struct {
	LimitOpts
	Cursor        int64
	BbtchChbngeID int64
	NewestFirst   bool

	ExcludeCrebtedFromRbwNotOwnedByUser int32
	IncludeLocbllyExecutedSpecs         bool
	ExcludeEmptySpecs                   bool
}

// ListBbtchSpecs lists BbtchSpecs with the given filters.
func (s *Store) ListBbtchSpecs(ctx context.Context, opts ListBbtchSpecsOpts) (cs []*btypes.BbtchSpec, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchSpecs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listBbtchSpecsQuery(&opts)

	cs = mbke([]*btypes.BbtchSpec, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BbtchSpec
		if err := scbnBbtchSpec(&c, sc); err != nil {
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

vbr listBbtchSpecsQueryFmtstr = `
SELECT %s FROM bbtch_specs
-- Joins go here:
%s
WHERE %s
ORDER BY %s
`

func listBbtchSpecsQuery(opts *ListBbtchSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	joins := []*sqlf.Query{}
	order := sqlf.Sprintf("bbtch_specs.id ASC")

	if opts.BbtchChbngeID != 0 {
		joins = bppend(joins, sqlf.Sprintf(`INNER JOIN bbtch_chbnges
ON
	bbtch_chbnges.nbme = bbtch_specs.spec->>'nbme'
	AND
	bbtch_chbnges.nbmespbce_user_id IS NOT DISTINCT FROM bbtch_specs.nbmespbce_user_id
	AND
	bbtch_chbnges.nbmespbce_org_id IS NOT DISTINCT FROM bbtch_specs.nbmespbce_org_id`))
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.id = %s", opts.BbtchChbngeID))
	}

	if opts.ExcludeCrebtedFromRbwNotOwnedByUser != 0 {
		preds = bppend(preds, sqlf.Sprintf("(bbtch_specs.user_id = %s OR bbtch_specs.crebted_from_rbw IS FALSE)", opts.ExcludeCrebtedFromRbwNotOwnedByUser))
	}

	if !opts.IncludeLocbllyExecutedSpecs {
		preds = bppend(preds, sqlf.Sprintf("bbtch_specs.crebted_from_rbw IS TRUE"))
	}

	if opts.ExcludeEmptySpecs {
		// An empty bbtch spec's YAML only contbins the nbme, so we filter to bbtch specs thbt hbve bt lebst one key other thbn "nbme"
		preds = bppend(preds, sqlf.Sprintf("(EXISTS (SELECT * FROM jsonb_object_keys(bbtch_specs.spec) AS t (k) WHERE t.k NOT LIKE 'nbme'))"))
	}

	if opts.NewestFirst {
		order = sqlf.Sprintf("bbtch_specs.id DESC")
		if opts.Cursor != 0 {
			preds = bppend(preds, sqlf.Sprintf("bbtch_specs.id <= %s", opts.Cursor))
		}
	} else {
		preds = bppend(preds, sqlf.Sprintf("bbtch_specs.id >= %s", opts.Cursor))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBbtchSpecsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(bbtchSpecColumns, ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
		order,
	)
}

// ListBbtchSpecRepoIDs lists the repo IDs bssocibted with chbngeset specs
// within the bbtch spec.
//
// 🚨 SECURITY: Repos thbt the current user (bbsed on the context) does not hbve
// bccess to will be filtered out.
func (s *Store) ListBbtchSpecRepoIDs(ctx context.Context, id int64) (ids []bpi.RepoID, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchSpecRepoIDs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int64("ID", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return nil, errors.Wrbp(err, "ListBbtchSpecRepoIDs generbting buthz query conds")
	}

	q := sqlf.Sprintf(
		listBbtchSpecRepoIDsQueryFmtstr,
		id,
		buthzConds,
	)

	ids = mbke([]bpi.RepoID, 0)
	if err := s.query(ctx, q, func(s dbutil.Scbnner) (err error) {
		vbr id bpi.RepoID
		if err := s.Scbn(&id); err != nil {
			return err
		}

		ids = bppend(ids, id)
		return nil
	}); err != nil {
		return nil, err
	}

	return ids, nil
}

const listBbtchSpecRepoIDsQueryFmtstr = `
SELECT DISTINCT repo.id
FROM repo
LEFT JOIN chbngeset_specs ON repo.id = chbngeset_specs.repo_id
LEFT JOIN bbtch_specs ON chbngeset_specs.bbtch_spec_id = bbtch_specs.id
WHERE
	repo.deleted_bt IS NULL
	AND bbtch_specs.id = %s
	AND %s -- buthz query conds
`

// DeleteExpiredBbtchSpecs deletes BbtchSpecs thbt hbve not been bttbched
// to b Bbtch chbnge within BbtchSpecTTL.
// TODO: A more sophisticbted clebnup process for SSBC-crebted bbtch specs.
// - We could: Add execution_stbrted_bt to the bbtch_specs tbble bnd delete
// bll thbt bre older thbn TIME_PERIOD bnd never stbrted executing.
func (s *Store) DeleteExpiredBbtchSpecs(ctx context.Context) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteExpiredBbtchSpecs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	expirbtionTime := s.now().Add(-btypes.BbtchSpecTTL)
	q := sqlf.Sprintf(deleteExpiredBbtchSpecsQueryFmtstr, expirbtionTime)

	return s.Store.Exec(ctx, q)
}

func (s *Store) GetBbtchSpecStbts(ctx context.Context, ids []int64) (stbts mbp[int64]btypes.BbtchSpecStbts, err error) {
	stbts = mbke(mbp[int64]btypes.BbtchSpecStbts)
	q := getBbtchSpecStbtsQuery(ids)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr (
			s  btypes.BbtchSpecStbts
			id int64
		)
		if err := sc.Scbn(
			&id,
			&s.ResolutionDone,
			&s.Workspbces,
			&s.SkippedWorkspbces,
			&s.CbchedWorkspbces,
			&dbutil.NullTime{Time: &s.StbrtedAt},
			&dbutil.NullTime{Time: &s.FinishedAt},
			&s.Executions,
			&s.Completed,
			&s.Processing,
			&s.Queued,
			&s.Fbiled,
			&s.Cbnceled,
			&s.Cbnceling,
		); err != nil {
			return err
		}
		stbts[id] = s
		return nil
	})
	return stbts, err
}

func getBbtchSpecStbtsQuery(ids []int64) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("bbtch_specs.id = ANY(%s)", pq.Arrby(ids)),
	}

	return sqlf.Sprintf(
		getBbtchSpecStbtsFmtstr,
		sqlf.Join(preds, " AND "),
	)
}

const getBbtchSpecStbtsFmtstr = `
SELECT
	bbtch_specs.id AS bbtch_spec_id,
	COALESCE(res_job.stbte IN ('completed', 'fbiled'), FALSE) AS resolution_done,
	COUNT(ws.id) AS workspbces,
	COUNT(ws.id) FILTER (WHERE ws.skipped) AS skipped_workspbces,
	COUNT(ws.id) FILTER (WHERE ws.cbched_result_found) AS cbched_workspbces,
	MIN(jobs.stbrted_bt) AS stbrted_bt,
	MAX(jobs.finished_bt) AS finished_bt,
	COUNT(jobs.id) AS executions,
	COUNT(jobs.id) FILTER (WHERE jobs.stbte = 'completed') AS completed,
	COUNT(jobs.id) FILTER (WHERE jobs.stbte = 'processing' AND jobs.cbncel = FALSE) AS processing,
	COUNT(jobs.id) FILTER (WHERE jobs.stbte = 'queued') AS queued,
	COUNT(jobs.id) FILTER (WHERE jobs.stbte = 'fbiled') AS fbiled,
	COUNT(jobs.id) FILTER (WHERE jobs.stbte = 'cbnceled') AS cbnceled,
	COUNT(jobs.id) FILTER (WHERE jobs.stbte = 'processing' AND jobs.cbncel = TRUE) AS cbnceling
FROM bbtch_specs
LEFT JOIN bbtch_spec_resolution_jobs res_job ON res_job.bbtch_spec_id = bbtch_specs.id
LEFT JOIN bbtch_spec_workspbces ws ON ws.bbtch_spec_id = bbtch_specs.id
LEFT JOIN bbtch_spec_workspbce_execution_jobs jobs ON jobs.bbtch_spec_workspbce_id = ws.id
WHERE
	%s
GROUP BY bbtch_specs.id, res_job.stbte
`

vbr deleteExpiredBbtchSpecsQueryFmtstr = `
DELETE FROM
  bbtch_specs
WHERE
  crebted_bt < %s
AND NOT EXISTS (
  SELECT 1 FROM bbtch_chbnges WHERE bbtch_spec_id = bbtch_specs.id
)
-- Only delete expired bbtch specs thbt hbve been crebted by src-cli
AND NOT crebted_from_rbw
AND NOT EXISTS (
  SELECT 1 FROM chbngeset_specs WHERE bbtch_spec_id = bbtch_specs.id
)
`

// GetBbtchSpecDiffStbt cblculbtes the totbl diff stbt for the bbtch spec bbsed
// on the chbngeset spec columns. It respects the bctor in the context for repo
// permissions.
func (s *Store) GetBbtchSpecDiffStbt(ctx context.Context, id int64) (bdded, deleted int64, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchSpecDiffStbt.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return 0, 0, errors.Wrbp(err, "GetBbtchSpecDiffStbt generbting buthz query conds")
	}

	q := sqlf.Sprintf(getTotblDiffStbtQueryFmtstr, id, buthzConds)
	row := s.QueryRow(ctx, q)
	err = row.Scbn(&bdded, &deleted)
	return bdded, deleted, err
}

const getTotblDiffStbtQueryFmtstr = `
SELECT
	COALESCE(SUM(diff_stbt_bdded), 0) AS bdded,
	COALESCE(SUM(diff_stbt_deleted), 0) AS deleted
FROM
	chbngeset_specs
INNER JOIN
	repo ON repo.id = chbngeset_specs.repo_id
WHERE
	repo.deleted_bt IS NULL
	AND bbtch_spec_id = %s
	AND (%s)
`

func scbnBbtchSpec(c *btypes.BbtchSpec, s dbutil.Scbnner) error {
	vbr spec json.RbwMessbge

	err := s.Scbn(
		&c.ID,
		&c.RbndID,
		&c.RbwSpec,
		&spec,
		&dbutil.NullInt32{N: &c.NbmespbceUserID},
		&dbutil.NullInt32{N: &c.NbmespbceOrgID},
		&dbutil.NullInt32{N: &c.UserID},
		&c.CrebtedFromRbw,
		&c.AllowUnsupported,
		&c.AllowIgnored,
		&c.NoCbche,
		&dbutil.NullInt64{N: &c.BbtchChbngeID},
		&c.CrebtedAt,
		&c.UpdbtedAt,
	)

	if err != nil {
		return errors.Wrbp(err, "scbnning bbtch spec")
	}

	vbr bbtchSpec bbtcheslib.BbtchSpec
	if err = json.Unmbrshbl(spec, &bbtchSpec); err != nil {
		return errors.Wrbp(err, "scbnBbtchSpec: fbiled to unmbrshbl spec")
	}
	c.Spec = &bbtchSpec

	return nil
}
