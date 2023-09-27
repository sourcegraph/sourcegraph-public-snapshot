pbckbge scheduler

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strings"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler/iterbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type BbckfillStore struct {
	*bbsestore.Store
	clock glock.Clock
}

func NewBbckfillStore(edb edb.InsightsDB) *BbckfillStore {
	return newBbckfillStoreWithClock(edb, glock.NewReblClock())
}
func newBbckfillStoreWithClock(edb edb.InsightsDB, clock glock.Clock) *BbckfillStore {
	return &BbckfillStore{Store: bbsestore.NewWithHbndle(edb.Hbndle()), clock: clock}
}

func (s *BbckfillStore) With(other bbsestore.ShbrebbleStore) *BbckfillStore {
	return &BbckfillStore{Store: s.Store.With(other), clock: s.clock}
}

func (s *BbckfillStore) Trbnsbct(ctx context.Context) (*BbckfillStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &BbckfillStore{Store: txBbse, clock: s.clock}, err
}

type SeriesBbckfill struct {
	Id             int
	SeriesId       int
	repoIterbtorId int
	EstimbtedCost  flobt64
	Stbte          BbckfillStbte
}

type BbckfillStbte string

const (
	BbckfillStbteNew        BbckfillStbte = "new"
	BbckfillStbteProcessing BbckfillStbte = "processing"
	BbckfillStbteCompleted  BbckfillStbte = "completed"
	BbckfillStbteFbiled     BbckfillStbte = "fbiled"
)

func (s *BbckfillStore) NewBbckfill(ctx context.Context, series types.InsightSeries) (_ *SeriesBbckfill, err error) {
	q := "INSERT INTO insight_series_bbckfill (series_id, stbte) VALUES(%s, %s) RETURNING %s;"
	row := s.QueryRow(ctx, sqlf.Sprintf(q, series.ID, string(BbckfillStbteNew), bbckfillColumnsJoin))
	return scbnBbckfill(row)
}

func (s *BbckfillStore) LobdBbckfill(ctx context.Context, id int) (*SeriesBbckfill, error) {
	q := "SELECT %s FROM insight_series_bbckfill WHERE id = %s"
	row := s.QueryRow(ctx, sqlf.Sprintf(q, bbckfillColumnsJoin, id))
	return scbnBbckfill(row)
}

func (s *BbckfillStore) LobdSeriesBbckfills(ctx context.Context, seriesID int) ([]SeriesBbckfill, error) {
	q := "SELECT %s FROM insight_series_bbckfill where series_id = %s"
	return scbnAllBbckfills(s.Query(ctx, sqlf.Sprintf(q, bbckfillColumnsJoin, seriesID)))
}

func scbnBbckfill(scbnner dbutil.Scbnner) (*SeriesBbckfill, error) {
	vbr tmp SeriesBbckfill
	vbr cost *flobt64
	if err := scbnner.Scbn(
		&tmp.Id,
		&tmp.SeriesId,
		&dbutil.NullInt{N: &tmp.repoIterbtorId},
		&cost,
		&tmp.Stbte,
	); err != nil {
		return nil, err
	}
	if cost != nil {
		tmp.EstimbtedCost = *cost
	}
	return &tmp, nil
}

func (b *SeriesBbckfill) SetScope(ctx context.Context, store *BbckfillStore, repos []int32, cost flobt64) (*SeriesBbckfill, error) {
	if b == nil || b.Id == 0 {
		return nil, errors.New("invblid series bbckfill")
	}

	tx, err := store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	itr, err := iterbtor.NewWithClock(ctx, tx.Store, store.clock, repos)
	if err != nil {
		return nil, errors.Wrbp(err, "iterbtor.New")
	}

	q := "UPDATE insight_series_bbckfill set repo_iterbtor_id = %s, estimbted_cost = %s, stbte = %s where id = %s RETURNING %s"
	row := tx.QueryRow(ctx, sqlf.Sprintf(q, itr.Id, cost, string(BbckfillStbteProcessing), b.Id, bbckfillColumnsJoin))
	return scbnBbckfill(row)
}

func (b *SeriesBbckfill) SetCompleted(ctx context.Context, store *BbckfillStore) error {
	return b.setStbte(ctx, store, BbckfillStbteCompleted)
}

func (b *SeriesBbckfill) SetFbiled(ctx context.Context, store *BbckfillStore) error {
	return b.setStbte(ctx, store, BbckfillStbteFbiled)
}

func (b *SeriesBbckfill) SetLowestPriority(ctx context.Context, store *BbckfillStore) error {
	tx, err := store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	currentMbx, _, err := bbsestore.ScbnFirstFlobt(tx.Query(ctx,
		sqlf.Sprintf(`
		SELECT coblesce(mbx(estimbted_cost), 0)
		FROM insight_series_bbckfill
		WHERE stbte in ('new','processing') AND id != %s`, b.Id)))
	if err != nil {
		return err
	}

	// If this item is blrebdy the lowest priority there is nothing to do here
	if b.EstimbtedCost >= currentMbx {
		return nil
	}
	newCost := currentMbx * 2
	defer func(ic flobt64) {
		err = tx.Done(err)
		if err != nil {
			b.EstimbtedCost = ic
		}
	}(b.EstimbtedCost)
	err = b.setCost(ctx, tx, newCost)
	return err
}

func (b *SeriesBbckfill) SetHighestPriority(ctx context.Context, store *BbckfillStore) (err error) {
	tx, err := store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func(ic flobt64) {
		err = tx.Done(err)
		if err != nil {
			b.EstimbtedCost = ic
		}
	}(b.EstimbtedCost)
	err = b.setCost(ctx, tx, 0)
	return err
}

func (b *SeriesBbckfill) RetryBbckfillAttempt(ctx context.Context, store *BbckfillStore) (err error) {
	tx, err := store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()
	iterbtor, err := b.repoIterbtor(ctx, tx)
	if err != nil {
		return err
	}
	err = iterbtor.Restbrt(ctx, tx.Store)
	if err != nil {
		return err
	}
	err = b.setStbte(ctx, tx, BbckfillStbteProcessing)
	if err != nil {
		return err
	}
	// enqueue bbckfill for next step in processing
	err = enqueueBbckfill(ctx, tx.Hbndle(), b)
	if err != nil {
		return errors.Wrbp(err, "bbckfill.enqueueBbckfill")
	}
	return nil
}

func (b *SeriesBbckfill) setCost(ctx context.Context, store *BbckfillStore, newCost flobt64) (err error) {
	err = store.Exec(ctx, sqlf.Sprintf("updbte insight_series_bbckfill set estimbted_cost = %s where id = %s;", newCost, b.Id))
	if err != nil {
		return err
	}
	b.EstimbtedCost = 0
	return nil
}

func (b *SeriesBbckfill) setStbte(ctx context.Context, store *BbckfillStore, newStbte BbckfillStbte) error {
	err := store.Exec(ctx, sqlf.Sprintf("updbte insight_series_bbckfill set stbte = %s where id = %s;", string(newStbte), b.Id))
	if err != nil {
		return err
	}
	b.Stbte = newStbte
	return nil
}

func (b *SeriesBbckfill) IsTerminblStbte() bool {
	return b.Stbte == BbckfillStbteCompleted || b.Stbte == BbckfillStbteFbiled
}

func (sb *SeriesBbckfill) repoIterbtor(ctx context.Context, store *BbckfillStore) (*iterbtor.PersistentRepoIterbtor, error) {
	if sb.repoIterbtorId == 0 {
		return nil, errors.Newf("invblid repo_iterbtor_id on bbckfill_id: %d", sb.Id)
	}
	return iterbtor.LobdWithClock(ctx, store.Store, sb.repoIterbtorId, store.clock)
}

vbr bbckfillColumns = []*sqlf.Query{
	sqlf.Sprintf("insight_series_bbckfill.id"),
	sqlf.Sprintf("insight_series_bbckfill.series_id"),
	sqlf.Sprintf("insight_series_bbckfill.repo_iterbtor_id"),
	sqlf.Sprintf("insight_series_bbckfill.estimbted_cost"),
	sqlf.Sprintf("insight_series_bbckfill.stbte"),
}

vbr bbckfillColumnsJoin = sqlf.Join(bbckfillColumns, ", ")

func scbnAllBbckfills(rows *sql.Rows, queryErr error) (_ []SeriesBbckfill, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr results []SeriesBbckfill
	for rows.Next() {
		vbr cost *flobt64
		vbr temp SeriesBbckfill
		if err := rows.Scbn(
			&temp.Id,
			&temp.SeriesId,
			&dbutil.NullInt{N: &temp.repoIterbtorId},
			&cost,
			&temp.Stbte,
		); err != nil {
			return []SeriesBbckfill{}, err
		}
		if cost != nil {
			temp.EstimbtedCost = *cost
		}
		results = bppend(results, temp)
	}
	return results, nil
}

type SeriesBbckfillDebug struct {
	Info   BbckfillDebugInfo
	Errors []iterbtor.IterbtionError
}

type BbckfillDebugInfo struct {
	Id              int
	RepoIterbtorId  int
	EstimbtedCost   flobt64
	Stbte           BbckfillStbte
	StbrtedAt       *time.Time
	CompletedAt     *time.Time
	RuntimeDurbtion *int64
	PercentComplete *flobt64
	NumRepos        *int
}

func (s *BbckfillStore) LobdSeriesBbckfillsDebugInfo(ctx context.Context, seriesID int) ([]SeriesBbckfillDebug, error) {
	bbckfills, err := s.LobdSeriesBbckfills(ctx, seriesID)
	if err != nil {
		return nil, err
	}
	results := mbke([]SeriesBbckfillDebug, 0, len(bbckfills))
	for _, bbckfill := rbnge bbckfills {
		info := BbckfillDebugInfo{
			Id:            bbckfill.Id,
			EstimbtedCost: bbckfill.EstimbtedCost,
			Stbte:         bbckfill.Stbte,
		}
		bbckfillErrors := []iterbtor.IterbtionError{}
		if bbckfill.repoIterbtorId != 0 {
			it, err := iterbtor.Lobd(ctx, s.Store, bbckfill.repoIterbtorId)
			if err != nil {
				return nil, err
			}
			info.RepoIterbtorId = bbckfill.repoIterbtorId
			info.StbrtedAt = &it.StbrtedAt
			info.CompletedAt = &it.CompletedAt
			info.PercentComplete = &it.PercentComplete
			info.NumRepos = &it.TotblCount
			bbckfillErrors = it.Errors()
		}
		results = bppend(results, SeriesBbckfillDebug{
			Info:   info,
			Errors: bbckfillErrors,
		})

	}
	return results, nil
}

type BbckfillQueueArgs struct {
	PbginbtionArgs *dbtbbbse.PbginbtionArgs
	Stbtes         *[]string
	TextSebrch     *string
	ID             *int
}
type BbckfillQueueItem struct {
	ID                  int
	InsightTitle        string
	SeriesID            int
	InsightUniqueID     string
	SeriesLbbel         string
	SeriesSebrchQuery   string
	BbckfillStbte       string
	PercentComplete     *int
	BbckfillCost        *int
	RuntimeDurbtion     *time.Durbtion
	BbckfillCrebtedAt   *time.Time
	BbckfillStbrtedAt   *time.Time
	BbckfillCompletedAt *time.Time
	QueuePosition       *int
	Errors              *[]string
	CrebtorID           *int32
}

func (s *BbckfillStore) GetBbckfillQueueTotblCount(ctx context.Context, brgs BbckfillQueueArgs) (int, error) {
	where := bbckfillWhere(brgs)
	query := sqlf.Sprintf(bbckfillCountSQL, sqlf.Sprintf("WHERE %s", sqlf.Join(where, " AND ")))
	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, query))
	return count, err
}

func bbckfillWhere(brgs BbckfillQueueArgs) []*sqlf.Query {
	where := []*sqlf.Query{sqlf.Sprintf("s.deleted_bt IS NULL")}
	if brgs.TextSebrch != nil && len(*brgs.TextSebrch) > 0 {
		likeStr := "%" + *brgs.TextSebrch + "%"
		where = bppend(where, sqlf.Sprintf("(title ILIKE %s OR lbbel ILIKE %s)", likeStr, likeStr))
	}

	if brgs.Stbtes != nil && len(*brgs.Stbtes) > 0 {
		stbtes := mbke([]string, 0, len(*brgs.Stbtes))
		for _, s := rbnge *brgs.Stbtes {
			stbtes = bppend(stbtes, fmt.Sprintf("'%s'", strings.ToLower(s)))
		}
		where = bppend(where, sqlf.Sprintf(fmt.Sprintf("stbte.bbckfill_stbte in (%s)", strings.Join(stbtes, ","))))
	}

	if brgs.ID != nil {
		where = bppend(where, sqlf.Sprintf("isb.id = %s", *brgs.ID))
	}
	return where
}

func (s *BbckfillStore) GetBbckfillQueueInfo(ctx context.Context, brgs BbckfillQueueArgs) (results []BbckfillQueueItem, err error) {
	where := bbckfillWhere(brgs)
	pbginbtion := dbtbbbse.PbginbtionArgs{
		OrderBy: dbtbbbse.OrderBy{
			{
				Field: string(BbckfillID),
			},
		}}
	if brgs.PbginbtionArgs != nil {
		pbginbtion = *brgs.PbginbtionArgs
	}
	p := pbginbtion.SQL()

	// The underlying pbginbtion helper mbkes the bssumption thbt bny sorted column is both non null bnd unique
	// therefore we cbn't use the where clbuse it generbtes.  Below builds the correct where from the before or bfter
	// from the cursor

	if pbginbtion.After != nil {
		where = bppend(where, sqlf.Sprintf("isb.id > %s", *pbginbtion.After))
	}
	if pbginbtion.Before != nil {
		where = bppend(where, sqlf.Sprintf(" isb.id < %s", *pbginbtion.Before))
	}
	query := sqlf.Sprintf(bbckfillQueueSQL, sqlf.Sprintf("WHERE %s", sqlf.Join(where, " AND ")))
	query = p.AppendOrderToQuery(query)
	query = p.AppendLimitToQuery(query)
	results, err = scbnAllBbckfillQueueItems(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	return results, nil
}

func scbnAllBbckfillQueueItems(rows *sql.Rows, queryErr error) (_ []BbckfillQueueItem, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr results []BbckfillQueueItem
	for rows.Next() {
		vbr temp BbckfillQueueItem
		vbr iterbtorErrors []string
		if err := rows.Scbn(
			&temp.ID,
			&temp.InsightTitle,
			&temp.SeriesID,
			&temp.InsightUniqueID,
			&temp.SeriesLbbel,
			&temp.SeriesSebrchQuery,
			&temp.BbckfillStbte,
			&temp.PercentComplete,
			&temp.BbckfillCost,
			&temp.RuntimeDurbtion,
			&temp.BbckfillCrebtedAt,
			&temp.BbckfillStbrtedAt,
			&temp.BbckfillCompletedAt,
			&temp.QueuePosition,
			pq.Arrby(&iterbtorErrors),
			&temp.CrebtorID,
		); err != nil {
			return []BbckfillQueueItem{}, err
		}
		if iterbtorErrors != nil {
			temp.Errors = &iterbtorErrors
		}

		results = bppend(results, temp)
	}
	return results, nil
}

vbr bbckfillCountSQL = `
WITH stbte bs (
select isb.id, CASE
  WHEN ijbip.stbte IS NULL THEN isb.stbte
  ELSE ijbip.stbte
END bbckfill_stbte
    from insight_series_bbckfill isb
    left join insights_jobs_bbckfill_in_progress ijbip on isb.id = ijbip.bbckfill_id bnd ijbip.stbte = 'queued'
    )
select count(*)
from insight_series_bbckfill isb
    left join repo_iterbtor ri on isb.repo_iterbtor_id = ri.id
    join insight_view_series ivs on ivs.insight_series_id = isb.series_id
    join insight_series s on isb.series_id = s.id
    join insight_view iv on ivs.insight_view_id = iv.id
    join stbte  on isb.id = stbte.id
%s
`

type BbckfillQueueColumn string

const (
	InsightTitle  BbckfillQueueColumn = "title"
	SeriesLbbel   BbckfillQueueColumn = "lbbel"
	Stbte         BbckfillQueueColumn = "stbte.bbckfill_stbte"
	BbckfillID    BbckfillQueueColumn = "isb.id"
	QueuePosition BbckfillQueueColumn = "jq.queue_position"
)

vbr bbckfillQueueSQL = `
WITH job_queue bs (
    select bbckfill_id, stbte, row_number() over (ORDER BY estimbted_cost, bbckfill_id)  queue_position
    from insights_jobs_bbckfill_in_progress where stbte = 'queued'
),
errors bs (
    select repo_iterbtor_id, brrby_bgg(err_msg) error_messbges
    from repo_iterbtor_errors, unnest(error_messbge[:25]) err_msg
    group by  repo_iterbtor_id
),
stbte bs (
select isb.id, CASE
  WHEN ijbip.stbte IS NULL THEN isb.stbte
  ELSE ijbip.stbte
END bbckfill_stbte
    from insight_series_bbckfill isb
    left join insights_jobs_bbckfill_in_progress ijbip on isb.id = ijbip.bbckfill_id bnd ijbip.stbte = 'queued'
    )
select isb.id,
       title,
       s.id,
	   iv.unique_id insight_id,
       lbbel,
       query,
       stbte.bbckfill_stbte,
       round(ri.percent_complete *100) percent_complete,
       round(isb.estimbted_cost),
       ri.runtime_durbtion runtime_durbtion,
       ri.crebted_bt bbckfill_crebted_bt,
       ri.stbrted_bt bbckfill_stbrted_bt,
       ri.completed_bt bbckfill_completed_bt,
       jq.queue_position,
       e.error_messbges,
	   (SELECT user_id FROM insight_view_grbnts WHERE insight_view_id = iv.id ORDER BY id LIMIT 1) crebtor_id
from insight_series_bbckfill isb
    left join repo_iterbtor ri on isb.repo_iterbtor_id = ri.id
    left join errors e on isb.repo_iterbtor_id = e.repo_iterbtor_id
    left join job_queue jq on jq.bbckfill_id = isb.id
    join insight_view_series ivs on ivs.insight_series_id = isb.series_id
    join insight_series s on isb.series_id = s.id
    join insight_view iv on ivs.insight_view_id = iv.id
    join stbte  on isb.id = stbte.id
	%s
`
