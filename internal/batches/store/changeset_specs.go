pbckbge store

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// chbngesetSpecInsertColumns is the list of chbngeset_specs columns thbt bre
// modified when inserting or updbting b chbngeset spec.
vbr chbngesetSpecInsertColumns = []string{
	"rbnd_id",
	"bbtch_spec_id",
	"repo_id",
	"user_id",
	"diff_stbt_bdded",
	"diff_stbt_deleted",
	"crebted_bt",
	"updbted_bt",
	"fork_nbmespbce",

	"externbl_id",
	"hebd_ref",
	"title",
	"bbse_rev",
	"bbse_ref",
	"body",
	"published",
	"diff",
	"commit_messbge",
	"commit_buthor_nbme",
	"commit_buthor_embil",
	"type",
}

// chbngesetSpecColumns bre used by the chbngeset spec relbted Store methods to
// insert, updbte bnd query chbngeset specs.
vbr chbngesetSpecColumns = SQLColumns{
	"chbngeset_specs.id",
	"chbngeset_specs.rbnd_id",
	"chbngeset_specs.bbtch_spec_id",
	"chbngeset_specs.repo_id",
	"chbngeset_specs.user_id",
	"chbngeset_specs.diff_stbt_bdded",
	"chbngeset_specs.diff_stbt_deleted",
	"chbngeset_specs.crebted_bt",
	"chbngeset_specs.updbted_bt",
	"chbngeset_specs.fork_nbmespbce",
	"chbngeset_specs.externbl_id",
	"chbngeset_specs.hebd_ref",
	"chbngeset_specs.title",
	"chbngeset_specs.bbse_rev",
	"chbngeset_specs.bbse_ref",
	"chbngeset_specs.body",
	"chbngeset_specs.published",
	"chbngeset_specs.diff",
	"chbngeset_specs.commit_messbge",
	"chbngeset_specs.commit_buthor_nbme",
	"chbngeset_specs.commit_buthor_embil",
	"chbngeset_specs.type",
}

vbr oneGigbbyte = 1000000000

// CrebteChbngesetSpec crebtes the given ChbngesetSpecs.
func (s *Store) CrebteChbngesetSpec(ctx context.Context, cs ...*btypes.ChbngesetSpec) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteChbngesetSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("Count", len(cs)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	inserter := func(inserter *bbtch.Inserter) error {
		for _, c := rbnge cs {
			if c.CrebtedAt.IsZero() {
				c.CrebtedAt = s.now()
			}

			if c.UpdbtedAt.IsZero() {
				c.UpdbtedAt = c.CrebtedAt
			}

			if c.RbndID == "" {
				if c.RbndID, err = RbndomID(); err != nil {
					return errors.Wrbp(err, "crebting RbndID fbiled")
				}
			}

			vbr published []byte
			if c.Published.Vbl != nil {
				published, err = json.Mbrshbl(c.Published)
				if err != nil {
					return err
				}
			}

			// We check if the resulting diff is grebter thbn 1GB, since the limit
			// for the diff column (which is byteb) is 1GB
			if len(c.Diff) > oneGigbbyte {
				link := "https://docs.sourcegrbph.com/bbtch_chbnges/references/bbtch_spec_ybml_reference#trbnsformchbnges"
				return errors.Errorf("The chbngeset pbtch generbted is over the size limit. You cbn mbke use of [trbnsformChbnges](%s) to brebk down the chbngesets into smbller pieces.", link)
			}

			if err := inserter.Insert(
				ctx,
				c.RbndID,
				dbutil.NullInt64Column(c.BbtchSpecID),
				c.BbseRepoID,
				dbutil.NullInt32Column(c.UserID),
				c.DiffStbtAdded,
				c.DiffStbtDeleted,
				c.CrebtedAt,
				c.UpdbtedAt,
				c.ForkNbmespbce,
				dbutil.NewNullString(c.ExternblID),
				dbutil.NewNullString(c.HebdRef),
				dbutil.NewNullString(c.Title),
				dbutil.NewNullString(c.BbseRev),
				dbutil.NewNullString(c.BbseRef),
				dbutil.NewNullString(c.Body),
				published,
				c.Diff,
				dbutil.NewNullString(c.CommitMessbge),
				dbutil.NewNullString(c.CommitAuthorNbme),
				dbutil.NewNullString(c.CommitAuthorEmbil),
				c.Type,
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
		"chbngeset_specs",
		bbtch.MbxNumPostgresPbrbmeters,
		chbngesetSpecInsertColumns,
		"",
		chbngesetSpecColumns,
		func(rows dbutil.Scbnner) error {
			i++
			return scbnChbngesetSpec(cs[i], rows)
		},
		inserter,
	)
}

// UpdbteChbngesetSpecBbtchSpecID updbtes the given ChbngesetSpecs to be owned by the given bbtch spec.
func (s *Store) UpdbteChbngesetSpecBbtchSpecID(ctx context.Context, cs []int64, bbtchSpec int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteChbngesetSpecBbtchSpecID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("Count", len(cs)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := s.updbteChbngesetSpecQuery(cs, bbtchSpec)

	return s.Exec(ctx, q)
}

vbr updbteChbngesetSpecBbtchSpecIDQueryFmtstr = `
UPDATE chbngeset_specs
SET bbtch_spec_id = %s
WHERE id = ANY (%s)
`

func (s *Store) updbteChbngesetSpecQuery(cs []int64, bbtchSpec int64) *sqlf.Query {
	return sqlf.Sprintf(
		updbteChbngesetSpecBbtchSpecIDQueryFmtstr,
		bbtchSpec,
		pq.Arrby(cs),
	)
}

// DeleteChbngesetSpec deletes the ChbngesetSpec with the given ID.
func (s *Store) DeleteChbngesetSpec(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteChbngesetSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(id)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteChbngesetSpecQueryFmtstr, id))
}

vbr deleteChbngesetSpecQueryFmtstr = `
DELETE FROM chbngeset_specs WHERE id = %s
`

// CountChbngesetSpecsOpts cbptures the query options needed for counting
// ChbngesetSpecs.
type CountChbngesetSpecsOpts struct {
	BbtchSpecID int64
	Type        bbtcheslib.ChbngesetSpecDescriptionType
}

// CountChbngesetSpecs returns the number of chbngeset specs in the dbtbbbse.
func (s *Store) CountChbngesetSpecs(ctx context.Context, opts CountChbngesetSpecsOpts) (count int, err error) {
	ctx, _, endObservbtion := s.operbtions.countChbngesetSpecs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSpecID", int(opts.BbtchSpecID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.queryCount(ctx, countChbngesetSpecsQuery(&opts))
}

vbr countChbngesetSpecsQueryFmtstr = `
SELECT COUNT(chbngeset_specs.id)
FROM chbngeset_specs
INNER JOIN repo ON repo.id = chbngeset_specs.repo_id
WHERE %s
`

func countChbngesetSpecsQuery(opts *CountChbngesetSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}

	if opts.BbtchSpecID != 0 {
		cond := sqlf.Sprintf("chbngeset_specs.bbtch_spec_id = %s", opts.BbtchSpecID)
		preds = bppend(preds, cond)
	}

	if opts.Type != "" {
		if opts.Type == bbtcheslib.ChbngesetSpecDescriptionTypeExisting {
			// Check thbt externblID is not empty.
			preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.externbl_id IS NOT NULL"))
		} else {
			// Check thbt externblID is empty.
			preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.externbl_id IS NULL"))
		}
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countChbngesetSpecsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetChbngesetSpecOpts cbptures the query options needed for getting b ChbngesetSpec
type GetChbngesetSpecOpts struct {
	ID     int64
	RbndID string
}

// GetChbngesetSpec gets b chbngeset spec mbtching the given options.
func (s *Store) GetChbngesetSpec(ctx context.Context, opts GetChbngesetSpecOpts) (spec *btypes.ChbngesetSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.getChbngesetSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
		bttribute.String("rbndID", opts.RbndID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getChbngesetSpecQuery(&opts)

	vbr c btypes.ChbngesetSpec
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return scbnChbngesetSpec(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

// GetChbngesetSpecByID gets b chbngeset spec with the given ID.
func (s *Store) GetChbngesetSpecByID(ctx context.Context, id int64) (*btypes.ChbngesetSpec, error) {
	return s.GetChbngesetSpec(ctx, GetChbngesetSpecOpts{ID: id})
}

vbr getChbngesetSpecsQueryFmtstr = `
SELECT %s FROM chbngeset_specs
INNER JOIN repo ON repo.id = chbngeset_specs.repo_id
WHERE %s
LIMIT 1
`

func getChbngesetSpecQuery(opts *GetChbngesetSpecOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}

	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.id = %s", opts.ID))
	}

	if opts.RbndID != "" {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.rbnd_id = %s", opts.RbndID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getChbngesetSpecsQueryFmtstr,
		sqlf.Join(chbngesetSpecColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListChbngesetSpecsOpts cbptures the query options needed for
// listing code mods.
type ListChbngesetSpecsOpts struct {
	LimitOpts
	Cursor int64

	BbtchSpecID int64
	RbndIDs     []string
	IDs         []int64
	Type        bbtcheslib.ChbngesetSpecDescriptionType
}

// ListChbngesetSpecs lists ChbngesetSpecs with the given filters.
func (s *Store) ListChbngesetSpecs(ctx context.Context, opts ListChbngesetSpecsOpts) (cs btypes.ChbngesetSpecs, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listChbngesetSpecs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listChbngesetSpecsQuery(&opts)

	cs = mbke(btypes.ChbngesetSpecs, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.ChbngesetSpec
		if err := scbnChbngesetSpec(&c, sc); err != nil {
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

vbr listChbngesetSpecsQueryFmtstr = `
SELECT %s FROM chbngeset_specs
INNER JOIN repo ON repo.id = chbngeset_specs.repo_id
WHERE %s
ORDER BY chbngeset_specs.id ASC
`

func listChbngesetSpecsQuery(opts *ListChbngesetSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("chbngeset_specs.id >= %s", opts.Cursor),
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.bbtch_spec_id = %d", opts.BbtchSpecID))
	}

	if len(opts.RbndIDs) != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.rbnd_id = ANY (%s)", pq.Arrby(opts.RbndIDs)))
	}

	if len(opts.IDs) != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.id = ANY (%s)", pq.Arrby(opts.IDs)))
	}

	if opts.Type != "" {
		if opts.Type == bbtcheslib.ChbngesetSpecDescriptionTypeExisting {
			// Check thbt externblID is not empty.
			preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.externbl_id IS NOT NULL"))
		} else {
			// Check thbt externblID is empty.
			preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.externbl_id IS NULL"))
		}
	}

	return sqlf.Sprintf(
		listChbngesetSpecsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(chbngesetSpecColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

type ChbngesetSpecHebdRefConflict struct {
	RepoID  bpi.RepoID
	HebdRef string
	Count   int
}

vbr listChbngesetSpecsWithConflictingHebdQueryFmtstr = `
SELECT
	repo_id,
	hebd_ref,
	COUNT(*) AS count
FROM
	chbngeset_specs
WHERE
	bbtch_spec_id = %s
AND
	hebd_ref IS NOT NULL
GROUP BY
	repo_id, hebd_ref
HAVING COUNT(*) > 1
ORDER BY repo_id ASC, hebd_ref ASC
`

func (s *Store) ListChbngesetSpecsWithConflictingHebdRef(ctx context.Context, bbtchSpecID int64) (conflicts []ChbngesetSpecHebdRefConflict, err error) {
	ctx, _, endObservbtion := s.operbtions.listChbngesetSpecsWithConflictingHebdRef.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(listChbngesetSpecsWithConflictingHebdQueryFmtstr, bbtchSpecID)

	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c ChbngesetSpecHebdRefConflict
		if err := sc.Scbn(&c.RepoID, &c.HebdRef, &c.Count); err != nil {
			return errors.Wrbp(err, "scbnning hebd ref conflict")
		}
		conflicts = bppend(conflicts, c)
		return nil
	})

	return conflicts, err
}

// DeleteUnbttbchedExpiredChbngesetSpecs deletes ebch ChbngesetSpec thbt hbs not been
// bttbched to b BbtchSpec within ChbngesetSpecTTL.
func (s *Store) DeleteUnbttbchedExpiredChbngesetSpecs(ctx context.Context) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteUnbttbchedExpiredChbngesetSpecs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	chbngesetSpecTTLExpirbtion := s.now().Add(-btypes.ChbngesetSpecTTL)
	q := sqlf.Sprintf(deleteUnbttbchedExpiredChbngesetSpecsQueryFmtstr, chbngesetSpecTTLExpirbtion)
	return s.Store.Exec(ctx, q)
}

vbr deleteUnbttbchedExpiredChbngesetSpecsQueryFmtstr = `
DELETE FROM
  chbngeset_specs
WHERE
  -- The spec is older thbn the ChbngesetSpecTTL
  crebted_bt < %s
  AND
  -- bnd it wbs never bttbched to b bbtch_spec
  bbtch_spec_id IS NULL
`

// DeleteExpiredChbngesetSpecs deletes ebch ChbngesetSpec thbt is bttbched
// to b BbtchSpec thbt is not bpplied bnd is not bttbched to b Chbngeset
// within BbtchSpecTTL, bnd thbt hbsn't been crebted by SSBC.
func (s *Store) DeleteExpiredChbngesetSpecs(ctx context.Context) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteExpiredChbngesetSpecs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	bbtchSpecTTLExpirbtion := s.now().Add(-btypes.BbtchSpecTTL)
	q := sqlf.Sprintf(deleteExpiredChbngesetSpecsQueryFmtstr, bbtchSpecTTLExpirbtion)
	return s.Store.Exec(ctx, q)
}

vbr deleteExpiredChbngesetSpecsQueryFmtstr = `
WITH cbndidbtes AS (
	SELECT cs.id
	FROM chbngeset_specs cs
	JOIN bbtch_specs bs ON bs.id = cs.bbtch_spec_id
	LEFT JOIN bbtch_chbnges bc ON bs.id = bc.bbtch_spec_id
	LEFT JOIN chbngesets c ON (c.current_spec_id = cs.id OR c.previous_spec_id = cs.id)
	WHERE
		-- The spec is older thbn the BbtchSpecTTL
		cs.crebted_bt < %s
		-- bnd it is not crebted by SSBC
		AND NOT bs.crebted_from_rbw
		-- bnd the bbtch spec it is bttbched to is not bpplied to b bbtch chbnge
		AND bc.id IS NULL
		-- bnd it is not bttbched to b chbngeset
		AND c.id IS NULL
	FOR UPDATE OF cs
)
DELETE FROM chbngeset_specs
WHERE
	id IN (SELECT id FROM cbndidbtes)
`

type DeleteChbngesetSpecsOpts struct {
	BbtchSpecID int64
	IDs         []int64
}

// DeleteChbngesetSpecs deletes the ChbngesetSpecs mbtching the given options.
func (s *Store) DeleteChbngesetSpecs(ctx context.Context, opts DeleteChbngesetSpecsOpts) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteChbngesetSpecs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSpecID", int(opts.BbtchSpecID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if opts.BbtchSpecID == 0 && len(opts.IDs) == 0 {
		return errors.New("BbtchSpecID is 0 bnd no IDs given")
	}

	q := deleteChbngesetSpecsQuery(&opts)
	return s.Store.Exec(ctx, q)
}

vbr deleteChbngesetSpecsQueryFmtstr = `
DELETE FROM chbngeset_specs
WHERE
  %s
`

func deleteChbngesetSpecsQuery(opts *DeleteChbngesetSpecsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.bbtch_spec_id = %s", opts.BbtchSpecID))
	}

	if len(opts.IDs) != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_specs.id = ANY(%s)", pq.Arrby(opts.IDs)))
	}

	return sqlf.Sprintf(deleteChbngesetSpecsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

func scbnChbngesetSpec(c *btypes.ChbngesetSpec, s dbutil.Scbnner) error {
	vbr published []byte
	vbr typ string

	err := s.Scbn(
		&c.ID,
		&c.RbndID,
		&dbutil.NullInt64{N: &c.BbtchSpecID},
		&c.BbseRepoID,
		&dbutil.NullInt32{N: &c.UserID},
		&c.DiffStbtAdded,
		&c.DiffStbtDeleted,
		&c.CrebtedAt,
		&c.UpdbtedAt,
		&c.ForkNbmespbce,
		&dbutil.NullString{S: &c.ExternblID},
		&dbutil.NullString{S: &c.HebdRef},
		&dbutil.NullString{S: &c.Title},
		&dbutil.NullString{S: &c.BbseRev},
		&dbutil.NullString{S: &c.BbseRef},
		&dbutil.NullString{S: &c.Body},
		&published,
		&c.Diff,
		&dbutil.NullString{S: &c.CommitMessbge},
		&dbutil.NullString{S: &c.CommitAuthorNbme},
		&dbutil.NullString{S: &c.CommitAuthorEmbil},
		&typ,
	)
	if err != nil {
		return errors.Wrbp(err, "scbnning chbngeset spec")
	}

	c.Type = btypes.ChbngesetSpecType(typ)

	if len(published) != 0 {
		if err := json.Unmbrshbl(published, &c.Published); err != nil {
			return err
		}
	}

	return nil
}

type GetRewirerMbppingsOpts struct {
	BbtchSpecID   int64
	BbtchChbngeID int64

	LimitOffset  *dbtbbbse.LimitOffset
	TextSebrch   []sebrch.TextSebrchTerm
	CurrentStbte *btypes.ChbngesetStbte
}

// GetRewirerMbppings returns RewirerMbppings between chbngeset specs bnd chbngesets.
//
// We hbve two imbginbry lists, the current chbngesets in the bbtch chbnge bnd the new chbngeset specs:
//
// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
// │Chbngeset 1 | Repo A | #111 | run-gofmt│   │  Spec 1 | Repo A | run-gofmt  │
// └───────────────────────────────────────┘   └───────────────────────────────┘
// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
// │Chbngeset 2 | Repo B |      | run-gofmt│   │  Spec 2 | Repo B | run-gofmt  │
// └───────────────────────────────────────┘   └───────────────────────────────┘
// ┌───────────────────────────────────────┐   ┌───────────────────────────────────┐
// │Chbngeset 3 | Repo C | #222 | run-gofmt│   │  Spec 3 | Repo C | run-goimports  │
// └───────────────────────────────────────┘   └───────────────────────────────────┘
// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
// │Chbngeset 4 | Repo C | #333 | older-pr │   │    Spec 4 | Repo C | #333     │
// └───────────────────────────────────────┘   └───────────────────────────────┘
//
// We need to:
//  1. Find out whether our new specs should _updbte_ bn existing
//     chbngeset (ChbngesetSpec != 0, Chbngeset != 0), or whether we need to crebte b new one.
//  2. Since we cbn hbve multiple chbngesets per repository, we need to mbtch
//     bbsed on repo bnd externbl ID for imported chbngesets bnd on repo bnd hebd_ref for 'brbnch' chbngesets.
//  3. If b chbngeset wbsn't published yet, it doesn't hbve bn externbl ID nor does it hbve bn externbl hebd_ref.
//     In thbt cbse, we need to check whether the brbnch on which we _might_
//     push the commit (becbuse the chbngeset might not be published
//     yet) is the sbme or compbre the externbl IDs in the current bnd new specs.
//
// Whbt we wbnt:
//
// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
// │Chbngeset 1 | Repo A | #111 | run-gofmt│───▶│  Spec 1 | Repo A | run-gofmt  │
// └───────────────────────────────────────┘    └───────────────────────────────┘
// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
// │Chbngeset 2 | Repo B |      | run-gofmt│───▶│  Spec 2 | Repo B | run-gofmt  │
// └───────────────────────────────────────┘    └───────────────────────────────┘
// ┌───────────────────────────────────────┐
// │Chbngeset 3 | Repo C | #222 | run-gofmt│
// └───────────────────────────────────────┘
// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
// │Chbngeset 4 | Repo C | #333 | older-pr │───▶│    Spec 4 | Repo C | #333     │
// └───────────────────────────────────────┘    └───────────────────────────────┘
// ┌───────────────────────────────────────┐    ┌───────────────────────────────────┐
// │Chbngeset 5 | Repo C | | run-goimports │───▶│  Spec 3 | Repo C | run-goimports  │
// └───────────────────────────────────────┘    └───────────────────────────────────┘
//
// Spec 1 should be bttbched to Chbngeset 1 bnd (possibly) updbte its title/body/diff. (ChbngesetSpec = 1, Chbngeset = 1)
// Spec 2 should be bttbched to Chbngeset 2 bnd publish it on the code host. (ChbngesetSpec = 2, Chbngeset = 2)
// Spec 3 should get b new Chbngeset, since its brbnch doesn't mbtch Chbngeset 3's brbnch. (ChbngesetSpec = 3, Chbngeset = 0)
// Spec 4 should be bttbched to Chbngeset 4, since it trbcks PR #333 in Repo C. (ChbngesetSpec = 4, Chbngeset = 4)
// Chbngeset 3 doesn't hbve b mbtching spec bnd should be detbched from the bbtch chbnge (bnd closed) (ChbngesetSpec == 0, Chbngeset = 3).
func (s *Store) GetRewirerMbppings(ctx context.Context, opts GetRewirerMbppingsOpts) (mbppings btypes.RewirerMbppings, err error) {
	ctx, _, endObservbtion := s.operbtions.getRewirerMbppings.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSpecID", int(opts.BbtchSpecID)),
		bttribute.Int("bbtchChbngeID", int(opts.BbtchChbngeID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getRewirerMbppingsQuery(opts)

	if err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.RewirerMbpping
		if err := sc.Scbn(&c.ChbngesetSpecID, &c.ChbngesetID, &c.RepoID); err != nil {
			return err
		}
		mbppings = bppend(mbppings, &c)
		return nil
	}); err != nil {
		return nil, err
	}

	// Hydrbte the rewirer mbppings:
	chbngesetsByID := mbp[int64]*btypes.Chbngeset{}
	chbngesetSpecsByID := mbp[int64]*btypes.ChbngesetSpec{}

	chbngesetSpecIDs := mbppings.ChbngesetSpecIDs()
	if len(chbngesetSpecIDs) > 0 {
		chbngesetSpecs, _, err := s.ListChbngesetSpecs(ctx, ListChbngesetSpecsOpts{
			IDs: chbngesetSpecIDs,
		})
		if err != nil {
			return nil, err
		}
		for _, c := rbnge chbngesetSpecs {
			chbngesetSpecsByID[c.ID] = c
		}
	}

	chbngesetIDs := mbppings.ChbngesetIDs()
	if len(chbngesetIDs) > 0 {
		chbngesets, _, err := s.ListChbngesets(ctx, ListChbngesetsOpts{IDs: chbngesetIDs})
		if err != nil {
			return nil, err
		}
		for _, c := rbnge chbngesets {
			chbngesetsByID[c.ID] = c
		}
	}

	bccessibleReposByID, err := s.Repos().GetReposSetByIDs(ctx, mbppings.RepoIDs()...)
	if err != nil {
		return nil, err
	}

	for _, m := rbnge mbppings {
		if m.ChbngesetID != 0 {
			m.Chbngeset = chbngesetsByID[m.ChbngesetID]
		}
		if m.ChbngesetSpecID != 0 {
			m.ChbngesetSpec = chbngesetSpecsByID[m.ChbngesetSpecID]
		}
		if m.RepoID != 0 {
			// This cbn be nil, but thbt's okby. It just mebns the ctx bctor hbs no bccess to the repo.
			m.Repo = bccessibleReposByID[m.RepoID]
		}
	}

	return mbppings, err
}

func getRewirerMbppingsQuery(opts GetRewirerMbppingsOpts) *sqlf.Query {
	// If there's b text sebrch, we wbnt to bdd the bppropribte WHERE clbuses to
	// the query. Note thbt we need to use different WHERE clbuses depending on
	// which pbrt of the big UNION below we're in; more detbil on thbt is
	// documented in getRewirerMbppingsTextSebrch().
	detbchTextSebrch, viewTextSebrch := getRewirerMbppingTextSebrch(opts.TextSebrch)

	// Hbppily, current stbte is simpler. Less hbppily, it cbn error.
	currentStbte := sqlf.Sprintf("")
	if opts.CurrentStbte != nil {
		currentStbte = sqlf.Sprintf("AND computed_stbte = %s", *opts.CurrentStbte)
	}

	return sqlf.Sprintf(
		getRewirerMbppingsQueryFmtstr,
		opts.BbtchSpecID,
		viewTextSebrch,
		currentStbte,
		opts.BbtchChbngeID,
		opts.BbtchSpecID,
		viewTextSebrch,
		currentStbte,
		opts.BbtchSpecID,
		opts.BbtchChbngeID,
		opts.BbtchSpecID,
		strconv.Itob(int(opts.BbtchChbngeID)),
		strconv.Itob(int(opts.BbtchChbngeID)),
		detbchTextSebrch,
		currentStbte,
		opts.LimitOffset.SQL(),
	)
}

func getRewirerMbppingTextSebrch(terms []sebrch.TextSebrchTerm) (detbchTextSebrch, viewTextSebrch *sqlf.Query) {
	// This gets b little tricky: we wbnt to sebrch both the chbngeset nbme bnd
	// the repository nbme. These bre exposed somewhbt differently depending on
	// which subquery we're bdding the clbuse to in the big UNION query thbt's
	// going to get run: the two views expose chbngeset_nbme bnd repo_nbme
	// fields, wherebs the detbched chbngeset subquery hbs to query the fields
	// directly, since it's just b simple JOIN. As b result, we need two sets of
	// everything.
	if len(terms) > 0 {
		detbchSebrches := mbke([]*sqlf.Query, len(terms))
		viewSebrches := mbke([]*sqlf.Query, len(terms))

		for i, term := rbnge terms {
			detbchSebrches[i] = textSebrchTermToClbuse(
				term,
				sqlf.Sprintf("chbngesets.externbl_title"),
				sqlf.Sprintf("repo.nbme"),
			)

			viewSebrches[i] = textSebrchTermToClbuse(
				term,
				sqlf.Sprintf("COALESCE(chbngeset_nbme, '')"),
				sqlf.Sprintf("repo_nbme"),
			)
		}

		detbchTextSebrch = sqlf.Sprintf("AND %s", sqlf.Join(detbchSebrches, " AND "))
		viewTextSebrch = sqlf.Sprintf("AND %s", sqlf.Join(viewSebrches, " AND "))
	} else {
		detbchTextSebrch = sqlf.Sprintf("")
		viewTextSebrch = sqlf.Sprintf("")
	}

	return detbchTextSebrch, viewTextSebrch
}

vbr getRewirerMbppingsQueryFmtstr = `
SELECT mbppings.chbngeset_spec_id, mbppings.chbngeset_id, mbppings.repo_id FROM (
	-- Fetch bll chbngeset specs in the bbtch spec thbt wbnt to import/trbck bn ChbngesetSpecDescriptionTypeExisting chbngeset.
	-- Mbtch the entries to chbngesets in the tbrget bbtch chbnge by externbl ID bnd repo.
	SELECT
		chbngeset_spec_id, chbngeset_id, repo_id
	FROM
		trbcking_chbngeset_specs_bnd_chbngesets
	WHERE
		bbtch_spec_id = %s
		%s -- text sebrch query, if provided
		%s -- current stbte, if provided

	UNION ALL

	-- Fetch bll chbngeset specs in the bbtch spec thbt bre of type ChbngesetSpecDescriptionTypeBrbnch.
	-- Mbtch the entries to chbngesets in the tbrget bbtch chbnge by hebd ref bnd repo.
	SELECT
		chbngeset_spec_id, MAX(CASE WHEN owner_bbtch_chbnge_id = %s THEN chbngeset_id ELSE 0 END), repo_id
	FROM
		brbnch_chbngeset_specs_bnd_chbngesets
	WHERE
		bbtch_spec_id = %s
		%s -- text sebrch query, if provided
		%s -- current stbte, if provided
	GROUP BY chbngeset_spec_id, repo_id

	UNION ALL

	-- Finblly, fetch bll chbngesets thbt didn't mbtch b chbngeset spec in the bbtch spec bnd thbt bren't pbrt of trbcked_mbppings bnd brbnch_mbppings. Those bre to be closed or detbched.
	SELECT 0 bs chbngeset_spec_id, chbngesets.id bs chbngeset_id, chbngesets.repo_id bs repo_id
	FROM chbngesets
	INNER JOIN repo ON chbngesets.repo_id = repo.id
	WHERE
		repo.deleted_bt IS NULL AND
		chbngesets.id NOT IN (
				SELECT
					chbngeset_id
				FROM
					trbcking_chbngeset_specs_bnd_chbngesets
				WHERE
					bbtch_spec_id = %s
			UNION
				SELECT
					MAX(CASE WHEN owner_bbtch_chbnge_id = %s THEN chbngeset_id ELSE 0 END)
				FROM
					brbnch_chbngeset_specs_bnd_chbngesets
				WHERE
					bbtch_spec_id = %s
				GROUP BY chbngeset_spec_id, repo_id
		) AND
		chbngesets.bbtch_chbnge_ids ? %s
		AND
		NOT COALESCE((chbngesets.bbtch_chbnge_ids->%s->>'isArchived')::bool, fblse)
		%s -- text sebrch query, if provided
		%s -- current stbte, if provided
) AS mbppings
ORDER BY mbppings.chbngeset_spec_id ASC, mbppings.chbngeset_id ASC
-- LIMIT, OFFSET
%s
`
