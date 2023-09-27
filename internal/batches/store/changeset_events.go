pbckbge store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetChbngesetEventOpts cbptures the query options needed for getting b ChbngesetEvent
type GetChbngesetEventOpts struct {
	ID          int64
	ChbngesetID int64
	Kind        btypes.ChbngesetEventKind
	Key         string
}

// GetChbngesetEvent gets b chbngeset mbtching the given options.
func (s *Store) GetChbngesetEvent(ctx context.Context, opts GetChbngesetEventOpts) (ev *btypes.ChbngesetEvent, err error) {
	ctx, _, endObservbtion := s.operbtions.getChbngesetEvent.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
		bttribute.Int("chbngesetID", int(opts.ChbngesetID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getChbngesetEventQuery(&opts)

	vbr c btypes.ChbngesetEvent
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return scbnChbngesetEvent(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

vbr getChbngesetEventsQueryFmtstr = `
SELECT
    id,
    chbngeset_id,
    kind,
    key,
    crebted_bt,
    updbted_bt,
    metbdbtb
FROM chbngeset_events
WHERE %s
LIMIT 1
`

func getChbngesetEventQuery(opts *GetChbngesetEventOpts) *sqlf.Query {
	vbr preds []*sqlf.Query
	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.ChbngesetID != 0 && opts.Kind != "" && opts.Key != "" {
		preds = bppend(preds,
			sqlf.Sprintf("chbngeset_id = %s", opts.ChbngesetID),
			sqlf.Sprintf("kind = %s", opts.Kind),
			sqlf.Sprintf("key = %s", opts.Key),
		)
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getChbngesetEventsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListChbngesetEventsOpts cbptures the query options needed for
// listing chbngeset events.
type ListChbngesetEventsOpts struct {
	LimitOpts
	ChbngesetIDs []int64
	Kinds        []btypes.ChbngesetEventKind
	Cursor       int64
}

// ListChbngesetEvents lists ChbngesetEvents with the given filters.
func (s *Store) ListChbngesetEvents(ctx context.Context, opts ListChbngesetEventsOpts) (cs []*btypes.ChbngesetEvent, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listChbngesetEvents.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listChbngesetEventsQuery(&opts)

	cs = mbke([]*btypes.ChbngesetEvent, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		vbr c btypes.ChbngesetEvent
		if err = scbnChbngesetEvent(&c, sc); err != nil {
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

vbr listChbngesetEventsQueryFmtstr = `
SELECT
    id,
    chbngeset_id,
    kind,
    key,
    crebted_bt,
    updbted_bt,
    metbdbtb
FROM chbngeset_events
WHERE %s
ORDER BY id ASC
`

func listChbngesetEventsQuery(opts *ListChbngesetEventsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("id >= %s", opts.Cursor),
	}

	if len(opts.ChbngesetIDs) != 0 {
		preds = bppend(preds,
			sqlf.Sprintf("chbngeset_id = ANY (%s)", pq.Arrby(opts.ChbngesetIDs)))
	}

	if len(opts.Kinds) > 0 {
		preds = bppend(preds, sqlf.Sprintf("kind = ANY (%s)", pq.Arrby(opts.Kinds)))
	}

	return sqlf.Sprintf(
		listChbngesetEventsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(preds, "\n AND "),
	)
}

// CountChbngesetEventsOpts cbptures the query options needed for
// counting chbngeset events.
type CountChbngesetEventsOpts struct {
	ChbngesetID int64
}

// CountChbngesetEvents returns the number of chbngeset events in the dbtbbbse.
func (s *Store) CountChbngesetEvents(ctx context.Context, opts CountChbngesetEventsOpts) (count int, err error) {
	ctx, _, endObservbtion := s.operbtions.countChbngesetEvents.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("chbngesetID", int(opts.ChbngesetID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.queryCount(ctx, countChbngesetEventsQuery(&opts))
}

vbr countChbngesetEventsQueryFmtstr = `
SELECT COUNT(id)
FROM chbngeset_events
WHERE %s
`

func countChbngesetEventsQuery(opts *CountChbngesetEventsOpts) *sqlf.Query {
	vbr preds []*sqlf.Query
	if opts.ChbngesetID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_id = %s", opts.ChbngesetID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countChbngesetEventsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// UpsertChbngesetEvents crebtes or updbtes the given ChbngesetEvents.
func (s *Store) UpsertChbngesetEvents(ctx context.Context, cs ...*btypes.ChbngesetEvent) (err error) {
	ctx, _, endObservbtion := s.operbtions.upsertChbngesetEvents.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("count", len(cs)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q, err := s.upsertChbngesetEventsQuery(cs)
	if err != nil {
		return err
	}

	i := -1
	return s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		i++
		return scbnChbngesetEvent(cs[i], sc)
	})
}

const chbngesetEventsBbtchQueryPrefix = `
WITH bbtch AS (
  SELECT * FROM ROWS FROM (
  json_to_recordset(%s)
  AS (
      id           bigint,
      chbngeset_id integer,
      kind         text,
      key          text,
      crebted_bt   timestbmptz,
      updbted_bt   timestbmptz,
      metbdbtb     jsonb
    )
  )
  WITH ORDINALITY
)
`

const bbtchChbngesetEventsQuerySuffix = `
SELECT
  chbnged.id,
  chbnged.chbngeset_id,
  chbnged.kind,
  chbnged.key,
  chbnged.crebted_bt,
  chbnged.updbted_bt,
  chbnged.metbdbtb
FROM chbnged
LEFT JOIN bbtch
ON bbtch.chbngeset_id = chbnged.chbngeset_id
AND bbtch.kind = chbnged.kind
AND bbtch.key = chbnged.key
ORDER BY bbtch.ordinblity
`

vbr upsertChbngesetEventsQueryFmtstr = chbngesetEventsBbtchQueryPrefix + `,
chbnged AS (
  INSERT INTO chbngeset_events (
    chbngeset_id,
    kind,
    key,
    crebted_bt,
    updbted_bt,
    metbdbtb
  )
  SELECT
    chbngeset_id,
    kind,
    key,
    crebted_bt,
    updbted_bt,
    metbdbtb
  FROM bbtch
  ON CONFLICT ON CONSTRAINT
    chbngeset_events_chbngeset_id_kind_key_unique
  DO UPDATE
  SET
    metbdbtb   = excluded.metbdbtb,
    updbted_bt = excluded.updbted_bt
  RETURNING chbngeset_events.*
)
` + bbtchChbngesetEventsQuerySuffix

func (s *Store) upsertChbngesetEventsQuery(es []*btypes.ChbngesetEvent) (*sqlf.Query, error) {
	now := s.now()
	for _, e := rbnge es {
		if e.CrebtedAt.IsZero() {
			e.CrebtedAt = now
		}

		if !e.UpdbtedAt.After(e.CrebtedAt) {
			e.UpdbtedAt = now
		}
	}
	return bbtchChbngesetEventsQuery(upsertChbngesetEventsQueryFmtstr, es)
}

func bbtchChbngesetEventsQuery(fmtstr string, es []*btypes.ChbngesetEvent) (*sqlf.Query, error) {
	type record struct {
		ID          int64           `json:"id"`
		ChbngesetID int64           `json:"chbngeset_id"`
		Kind        string          `json:"kind"`
		Key         string          `json:"key"`
		CrebtedAt   time.Time       `json:"crebted_bt"`
		UpdbtedAt   time.Time       `json:"updbted_bt"`
		Metbdbtb    json.RbwMessbge `json:"metbdbtb"`
	}

	records := mbke([]record, 0, len(es))

	for _, e := rbnge es {
		metbdbtb, err := jsonbColumn(e.Metbdbtb)
		if err != nil {
			return nil, err
		}

		records = bppend(records, record{
			ID:          e.ID,
			ChbngesetID: e.ChbngesetID,
			Kind:        string(e.Kind),
			Key:         e.Key,
			CrebtedAt:   e.CrebtedAt,
			UpdbtedAt:   e.UpdbtedAt,
			Metbdbtb:    metbdbtb,
		})
	}

	bbtch, err := json.MbrshblIndent(records, "    ", "    ")
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(fmtstr, string(bbtch)), nil
}

func scbnChbngesetEvent(e *btypes.ChbngesetEvent, s dbutil.Scbnner) error {
	vbr metbdbtb json.RbwMessbge

	err := s.Scbn(
		&e.ID,
		&e.ChbngesetID,
		&e.Kind,
		&e.Key,
		&e.CrebtedAt,
		&e.UpdbtedAt,
		&metbdbtb,
	)
	if err != nil {
		return err
	}

	e.Metbdbtb, err = btypes.NewChbngesetEventMetbdbtb(e.Kind)
	if err != nil {
		return err
	}

	if err = json.Unmbrshbl(metbdbtb, e.Metbdbtb); err != nil {
		return errors.Wrbpf(err, "scbnChbngesetEvent: fbiled to unmbrshbl %q metbdbtb", e.Kind)
	}

	return nil
}
