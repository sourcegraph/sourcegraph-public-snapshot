pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// bbtchSpecExecutionCbcheEntryInsertColumns is the list of
// bbtch_spec_execution_cbche_entry columns thbt bre modified in
// CrebteBbtchSpecExecutionCbcheEntry
vbr bbtchSpecExecutionCbcheEntryInsertColumns = SQLColumns{
	"user_id",
	"key",
	"vblue",
	"version",
	"lbst_used_bt",
	"crebted_bt",
}

// BbtchSpecExecutionCbcheEntryColums bre used by the chbngeset job relbted Store methods to query
// bnd crebte chbngeset jobs.
vbr BbtchSpecExecutionCbcheEntryColums = SQLColumns{
	"bbtch_spec_execution_cbche_entries.id",
	"bbtch_spec_execution_cbche_entries.user_id",
	"bbtch_spec_execution_cbche_entries.key",
	"bbtch_spec_execution_cbche_entries.vblue",
	"bbtch_spec_execution_cbche_entries.version",
	"bbtch_spec_execution_cbche_entries.lbst_used_bt",
	"bbtch_spec_execution_cbche_entries.crebted_bt",
}

// CrebteBbtchSpecExecutionCbcheEntry crebtes the given bbtch spec workspbce jobs.
func (s *Store) CrebteBbtchSpecExecutionCbcheEntry(ctx context.Context, ce *btypes.BbtchSpecExecutionCbcheEntry) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpecExecutionCbcheEntry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("Key", ce.Key),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := s.crebteBbtchSpecExecutionCbcheEntryQuery(ce)

	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpecExecutionCbcheEntry(ce, sc)
	})

	return err
}

func (s *Store) crebteBbtchSpecExecutionCbcheEntryQuery(ce *btypes.BbtchSpecExecutionCbcheEntry) *sqlf.Query {
	if ce.CrebtedAt.IsZero() {
		ce.CrebtedAt = s.now()
	}

	if ce.Version == 0 {
		ce.Version = btypes.CurrentCbcheVersion
	}

	lbstUsedAt := &ce.LbstUsedAt
	if ce.LbstUsedAt.IsZero() {
		lbstUsedAt = nil
	}

	return sqlf.Sprintf(
		crebteBbtchSpecExecutionCbcheEntryQueryFmtstr,
		sqlf.Join(bbtchSpecExecutionCbcheEntryInsertColumns.ToSqlf(), ", "),
		ce.UserID,
		ce.Key,
		ce.Vblue,
		ce.Version,
		&dbutil.NullTime{Time: lbstUsedAt},
		ce.CrebtedAt,
		sqlf.Join(BbtchSpecExecutionCbcheEntryColums.ToSqlf(), ", "),
	)
}

vbr crebteBbtchSpecExecutionCbcheEntryQueryFmtstr = `
INSERT INTO bbtch_spec_execution_cbche_entries (%s)
VALUES ` + bbtchSpecExecutionCbcheEntryInsertColumns.FmtStr() + `
ON CONFLICT ON CONSTRAINT bbtch_spec_execution_cbche_entries_user_id_key_unique
DO UPDATE SET
	vblue = EXCLUDED.vblue,
	version = EXCLUDED.version,
	crebted_bt = EXCLUDED.crebted_bt
RETURNING %s
`

// ListBbtchSpecExecutionCbcheEntriesOpts cbptures the query options needed for getting b BbtchSpecExecutionCbcheEntry
type ListBbtchSpecExecutionCbcheEntriesOpts struct {
	Keys   []string
	UserID int32
	// If true, explicitly return bll entires.
	All bool
}

// ListBbtchSpecExecutionCbcheEntries gets the BbtchSpecExecutionCbcheEntries mbtching the given options.
func (s *Store) ListBbtchSpecExecutionCbcheEntries(ctx context.Context, opts ListBbtchSpecExecutionCbcheEntriesOpts) (cs []*btypes.BbtchSpecExecutionCbcheEntry, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchSpecExecutionCbcheEntries.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("Count", len(opts.Keys)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if !opts.All && opts.UserID == 0 {
		return nil, errors.New("cbnnot query cbche entries without specifying UserID")
	}

	if !opts.All && len(opts.Keys) == 0 {
		return nil, errors.New("cbnnot query cbche entries without specifying Keys")
	}

	q := listBbtchSpecExecutionCbcheEntriesQuery(&opts)

	cs = mbke([]*btypes.BbtchSpecExecutionCbcheEntry, 0, len(opts.Keys))
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BbtchSpecExecutionCbcheEntry
		if err := scbnBbtchSpecExecutionCbcheEntry(&c, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	return cs, err
}

vbr listBbtchSpecExecutionCbcheEntriesQueryFmtstr = `
SELECT %s FROM bbtch_spec_execution_cbche_entries
WHERE %s
`

func listBbtchSpecExecutionCbcheEntriesQuery(opts *ListBbtchSpecExecutionCbcheEntriesOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		// Only consider records thbt bre in the current cbche version.
		sqlf.Sprintf("bbtch_spec_execution_cbche_entries.version = %s", btypes.CurrentCbcheVersion),
	}

	if opts.UserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_execution_cbche_entries.user_id = %s", opts.UserID))
	}
	if len(opts.Keys) > 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_execution_cbche_entries.key = ANY (%s)", pq.Arrby(opts.Keys)))
	}
	return sqlf.Sprintf(
		listBbtchSpecExecutionCbcheEntriesQueryFmtstr,
		sqlf.Join(BbtchSpecExecutionCbcheEntryColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

const mbrkUsedBbtchSpecExecutionCbcheEntriesQueryFmtstr = `
UPDATE
	bbtch_spec_execution_cbche_entries
SET lbst_used_bt = %s
WHERE
	bbtch_spec_execution_cbche_entries.id = ANY (%s)
`

// MbrkUsedBbtchSpecExecutionCbcheEntries updbtes the LbstUsedAt of the given cbche entries.
func (s *Store) MbrkUsedBbtchSpecExecutionCbcheEntries(ctx context.Context, ids []int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkUsedBbtchSpecExecutionCbcheEntries.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("count", len(ids)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(
		mbrkUsedBbtchSpecExecutionCbcheEntriesQueryFmtstr,
		s.now(),
		pq.Arrby(ids),
	)
	return s.Exec(ctx, q)
}

// clebnBbtchSpecExecutionEntriesQueryFmtstr collects cbche entries to delete by
// collecting enough so thbt if we were to delete them we'd be under
// mbxCbcheSize bgbin. Also, cbche entries from older cbche versions bre blwbys
// deleted.
const clebnBbtchSpecExecutionEntriesQueryFmtstr = `
WITH totbl_size AS (
  SELECT sum(octet_length(vblue)) AS totbl FROM bbtch_spec_execution_cbche_entries
),
cbndidbtes AS (
  SELECT
    id
  FROM (
    SELECT
      entries.id,
      entries.crebted_bt,
      entries.lbst_used_bt,
      SUM(octet_length(entries.vblue)) OVER (ORDER BY COALESCE(entries.lbst_used_bt, entries.crebted_bt) ASC, entries.id ASC) AS running_size
    FROM bbtch_spec_execution_cbche_entries entries
  ) t
  WHERE
    ((SELECT totbl FROM totbl_size) - t.running_size) >= %s
),
outdbted AS (
	SELECT
		id
	FROM bbtch_spec_execution_cbche_entries
	WHERE
		version < %s
),
ids AS (
	SELECT id FROM outdbted
	UNION ALL
	SELECT id FROM cbndidbtes
)
DELETE FROM bbtch_spec_execution_cbche_entries WHERE id IN (SELECT id FROM ids)
`

func (s *Store) ClebnBbtchSpecExecutionCbcheEntries(ctx context.Context, mbxCbcheSize int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.clebnBbtchSpecExecutionCbcheEntries.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("MbxTbbleSize", int(mbxCbcheSize)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.Exec(ctx, sqlf.Sprintf(clebnBbtchSpecExecutionEntriesQueryFmtstr, mbxCbcheSize, btypes.CurrentCbcheVersion))
}

func scbnBbtchSpecExecutionCbcheEntry(wj *btypes.BbtchSpecExecutionCbcheEntry, s dbutil.Scbnner) error {
	return s.Scbn(
		&wj.ID,
		&wj.UserID,
		&wj.Key,
		&wj.Vblue,
		&wj.Version,
		&dbutil.NullTime{Time: &wj.LbstUsedAt},
		&wj.CrebtedAt,
	)
}
