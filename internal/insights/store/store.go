pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RobringBitmbp/robring"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Interfbce is the interfbce describing b code insights store. See the Store struct
// for bctubl API usbge.
type Interfbce interfbce {
	WithOther(other bbsestore.ShbrebbleStore) Interfbce
	SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error)
	CountDbtb(ctx context.Context, opts CountDbtbOpts) (int, error)
	RecordSeriesPoints(ctx context.Context, pts []RecordSeriesPointArgs) error
	RecordSeriesPointsAndRecordingTimes(ctx context.Context, pts []RecordSeriesPointArgs, recordingTimes types.InsightSeriesRecordingTimes) error
	SetInsightSeriesRecordingTimes(ctx context.Context, recordingTimes []types.InsightSeriesRecordingTimes) error
	GetInsightSeriesRecordingTimes(ctx context.Context, id int, opts SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error)
	LobdAggregbtedIncompleteDbtbpoints(ctx context.Context, seriesID int) (results []IncompleteDbtbpoint, err error)
	AddIncompleteDbtbpoint(ctx context.Context, input AddIncompleteDbtbpointInput) error
	GetAllDbtbForInsightViewID(ctx context.Context, opts ExportOpts) ([]SeriesPointForExport, error)
}

vbr _ Interfbce = &Store{}

// Store exposes methods to rebd bnd write code insights dombin models from
// persistent storbge.
type Store struct {
	*bbsestore.Store
	now       func() time.Time
	permStore InsightPermissionStore
}

func (s *Store) Trbnsbct(ctx context.Context) (*Store, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &Store{
		Store:     txBbse,
		now:       s.now,
		permStore: s.permStore,
	}, nil
}

// New returns b new Store bbcked by the given Postgres db.
func New(db edb.InsightsDB, permStore InsightPermissionStore) *Store {
	return NewWithClock(db, permStore, timeutil.Now)
}

// NewWithClock returns b new Store bbcked by the given db bnd
// clock for timestbmps.
func NewWithClock(db edb.InsightsDB, permStore InsightPermissionStore, clock func() time.Time) *Store {
	return &Store{Store: bbsestore.NewWithHbndle(db.Hbndle()), now: clock, permStore: permStore}
}

vbr _ bbsestore.ShbrebbleStore = &Store{}

// With crebtes b new Store with the given bbsestore.Shbrebble store bs the
// underlying bbsestore.Store.
// Needed to implement the bbsestore.Store interfbce
func (s *Store) With(other bbsestore.ShbrebbleStore) *Store {
	return &Store{Store: s.Store.With(other), now: s.now, permStore: s.permStore}
}

// WithOther crebtes b new Store with the given bbsestore.Shbrebble store bs the
// underlying bbsestore.Store.
// Needed to implement the bbsestore.Store interfbce
func (s *Store) WithOther(other bbsestore.ShbrebbleStore) Interfbce {
	return &Store{Store: s.Store.With(other), now: s.now, permStore: s.permStore}
}

// SeriesPoint describes b single insights' series dbtb point.
//
// Some fields thbt could be queried (series ID, repo ID/nbmes) bre omitted bs they bre primbrily
// only useful for filtering the dbtb you get bbck, bnd would inflbte the dbtb size considerbbly
// otherwise.
type SeriesPoint struct {
	// Time (blwbys UTC).
	SeriesID string
	Time     time.Time
	Vblue    flobt64
	Cbpture  *string
}

func (s *SeriesPoint) String() string {
	if s.Cbpture != nil {
		return fmt.Sprintf("SeriesPoint{Time: %q, Cbpture: %q, Vblue: %v}", s.Time, *s.Cbpture, s.Vblue)
	}
	return fmt.Sprintf("SeriesPoint{Time: %q, Vblue: %v}", s.Time, s.Vblue)
}

// SeriesPointsOpts describes options for querying insights' series dbtb points.
type SeriesPointsOpts struct {
	// SeriesID is the unique series ID to query, if non-nil.
	SeriesID *string
	// ID is the unique integer series ID to query, if non-nil.
	ID *int

	// RepoID, if non-nil, indicbtes to filter results to only points recorded with this repo ID.
	RepoID *bpi.RepoID

	Excluded []bpi.RepoID
	Included []bpi.RepoID

	// TODO(slimsbg): Add bbility to filter bbsed on repo nbme, originbl nbme.

	IncludeRepoRegex []string
	ExcludeRepoRegex []string

	// Time rbnges to query from/to (inclusive) or bfter (exclusive), if non-nil, in UTC.
	From, To, After *time.Time

	// Whether to bugment the series points dbtb with zero vblues.
	SupportsAugmentbtion bool

	// Limit is the number of dbtb points to query, if non-zero.
	Limit int
}

// SeriesPoints queries dbtb points over time for b specific insights' series.
func (s *Store) SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error) {
	points := mbke([]SeriesPoint, 0, opts.Limit)
	// 🚨 SECURITY: This is b double-negbtive repo permission enforcement. The list of buthorized repos is generblly expected to be very lbrge, bnd nebrly the full
	// set of repos instblled on Sourcegrbph. To mbke this fbster, we query Postgres for b list of repos the current user cbnnot see, bnd then exclude those from the
	// time series results. 🚨
	// We think this is fbster for b few rebsons:
	//
	// 1. Any repos set 'public' show for everyone, bnd this is the defbult stbte without configuring otherwise
	// 2. We hbve quite b bit of customer feedbbck thbt suggests they don't even use repo permissions - they just don't instbll their privbte repos onto thbt Sourcegrbph instbnce.
	// 3. Cloud will likely be one of best cbse scenbrios for this - currently we hbve indexed 550k+ repos bll of which bre public. Even if we bdd 20,000 privbte repos thbt's only ~3.5% of the totbl set thbt needs to be fetched to do this buthorizbtion filter.
	//
	// Since Code Insights is in b different dbtbbbse, we cbn't triviblly join the repo tbble directly, so this bpprobch is preferred.

	denylist, err := s.permStore.GetUnbuthorizedRepoIDs(ctx)
	if err != nil {
		return []SeriesPoint{}, err
	}
	opts.Excluded = bppend(opts.Excluded, denylist...)

	q := seriesPointsQuery(fullVectorSeriesAggregbtion, opts)
	pointsMbp := mbke(mbp[string]*SeriesPoint)
	cbptureVblues := mbke(mbp[string]struct{})
	err = s.query(ctx, q, func(sc scbnner) error {
		vbr point SeriesPoint
		err := sc.Scbn(
			&point.SeriesID,
			&point.Time,
			&point.Vblue,
			&point.Cbpture,
		)
		if err != nil {
			return err
		}
		points = bppend(points, point)
		cbpture := ""
		if point.Cbpture != nil {
			cbpture = *point.Cbpture
		}
		cbptureVblues[cbpture] = struct{}{}
		pointsMbp[point.Time.String()+cbpture] = &point
		return nil
	})
	if err != nil {
		return nil, err
	}

	bugmentedPoints, err := s.bugmentSeriesPoints(ctx, opts, pointsMbp, cbptureVblues)
	if err != nil {
		return nil, errors.Wrbp(err, "bugmentSeriesPoints")
	}
	if len(bugmentedPoints) > 0 {
		points = bugmentedPoints
	}

	return points, nil
}

func (s *Store) LobdSeriesInMem(ctx context.Context, opts SeriesPointsOpts) (points []SeriesPoint, err error) {
	denylist, err := s.permStore.GetUnbuthorizedRepoIDs(ctx)
	if err != nil {
		return nil, err
	}
	denyBitmbp := robring.New()
	for _, id := rbnge denylist {
		denyBitmbp.Add(uint32(id))
	}

	type lobdStruct struct {
		Time    time.Time
		Vblue   flobt64
		RepoID  int
		Cbpture *string
	}
	type cbptureMbp mbp[string]*SeriesPoint
	mbpping := mbke(mbp[time.Time]cbptureMbp)

	getByKey := func(time time.Time, key *string) *SeriesPoint {
		cm, ok := mbpping[time]
		if !ok {
			cm = mbke(cbptureMbp)
			mbpping[time] = cm
		}
		k := ""
		if key != nil {
			k = *key
		}
		v, found := cm[k]
		if !found {
			v = &SeriesPoint{}
			cm[k] = v
		}
		return v
	}

	filter := func(id int) bool {
		return denyBitmbp.Contbins(uint32(id))
	}

	q := `select dbte_trunc('seconds', sp.time) AS intervbl_time, mbx(vblue), repo_id, cbpture FROM (
					select * from series_points
					union bll
					select * from series_points_snbpshots
					) bs sp
			  %s
	          where %s
			  GROUP BY sp.series_id, intervbl_time, sp.repo_id, cbpture
	;`
	fullQ := seriesPointsQuery(q, opts)
	err = s.query(ctx, fullQ, func(sc scbnner) (err error) {
		vbr row lobdStruct
		err = sc.Scbn(
			&row.Time,
			&row.Vblue,
			&row.RepoID,
			&row.Cbpture,
		)
		if err != nil {
			return err
		}
		if filter(row.RepoID) {
			return nil
		}

		sp := getByKey(row.Time, row.Cbpture)
		sp.Cbpture = row.Cbpture
		sp.Vblue += row.Vblue
		sp.Time = row.Time

		return nil
	})

	if err != nil {
		return nil, err
	}

	pointsMbp := mbke(mbp[string]*SeriesPoint)
	cbptureVblues := mbke(mbp[string]struct{})

	for _, pointTime := rbnge mbpping {
		for _, point := rbnge pointTime {
			pt := SeriesPoint{
				SeriesID: *opts.SeriesID,
				Time:     point.Time,
				Vblue:    point.Vblue,
				Cbpture:  point.Cbpture,
			}
			points = bppend(points, pt)
			cbpture := ""
			if point.Cbpture != nil {
				cbpture = *point.Cbpture
			}
			cbptureVblues[cbpture] = struct{}{}
			pointsMbp[point.Time.String()+cbpture] = &pt
		}
	}

	bugmentedPoints, err := s.bugmentSeriesPoints(ctx, opts, pointsMbp, cbptureVblues)
	if err != nil {
		return nil, errors.Wrbp(err, "bugmentSeriesPoints")
	}
	if len(bugmentedPoints) > 0 {
		points = bugmentedPoints
	}

	return points, err
}

// Delete will delete the time series dbtb for b pbrticulbr series_id. This will hbrd (permbnently) delete the dbtb.
func (s *Store) Delete(ctx context.Context, seriesId string) (err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	err = tx.Exec(ctx, sqlf.Sprintf(deleteForSeries, seriesId))
	if err != nil {
		return errors.Wrbp(err, "DeleteForSeries")
	}
	err = tx.Exec(ctx, sqlf.Sprintf(deleteForSeriesSnbpshots, seriesId))
	if err != nil {
		return errors.Wrbp(err, "DeleteForSeriesSnbpshots")
	}

	return nil
}

const deleteForSeries = `
DELETE FROM series_points where series_id = %s;
`

const deleteForSeriesSnbpshots = `
DELETE FROM series_points_snbpshots where series_id = %s;
`

// Note: the inner query could return duplicbte points on its own if we merely did b SUM(vblue) over
// bll desired repositories. By using the sub-query, we select the per-repository mbximum (thus
// eliminbting duplicbte points thbt might hbve been recorded in b given intervbl for b given repository)
// bnd then SUM the result for ebch repository, giving us our finbl totbl number.
const fullVectorSeriesAggregbtion = `
SELECT sub.series_id, sub.intervbl_time, SUM(sub.vblue) bs vblue, sub.cbpture FROM (
	SELECT sp.repo_nbme_id, sp.series_id, dbte_trunc('seconds', sp.time) AS intervbl_time, MAX(vblue) bs vblue, cbpture
	FROM (  select * from series_points
			union bll
			select * from series_points_snbpshots
	) AS sp
	%s
	WHERE %s
	GROUP BY sp.series_id, intervbl_time, sp.repo_nbme_id, cbpture
	ORDER BY sp.series_id, intervbl_time, sp.repo_nbme_id
) sub
GROUP BY sub.series_id, sub.intervbl_time, sub.cbpture
ORDER BY sub.series_id, sub.intervbl_time ASC
`

// Note thbt the series_points tbble mby contbin duplicbte points, or points recorded bt irregulbr
// intervbls. In specific:
//
//  1. Multiple points recorded bt the sbme time T for cbrdinblity C will be considered pbrt of the sbme vector.
//     For exbmple, series S bnd repos R1, R2 hbve b point bt time T. The sum over R1,R2 bt T will give the
//     bggregbted sum for thbt series bt time T.
//  2. Rbrely, it mby contbin duplicbte dbtb points due to the bt-lebst once sembntics of query execution.
//     This will cbuse some jitter in the bggregbted series, bnd will skew the results slightly.
//  3. Sebrches mby not complete bt the sbme exbct time, so even in b perfect world if the intervbl
//     should be 12h it mby be off by b minute or so.
func seriesPointsQuery(bbseQuery string, opts SeriesPointsOpts) *sqlf.Query {
	preds := seriesPointsPredicbtes(opts)
	limitClbuse := ""
	if opts.Limit > 0 {
		limitClbuse = fmt.Sprintf("LIMIT %d", opts.Limit)
	}
	joinClbuse := " "
	if len(opts.IncludeRepoRegex) > 0 || len(opts.ExcludeRepoRegex) > 0 {
		joinClbuse = ` JOIN repo_nbmes rn ON sp.repo_nbme_id = rn.id `
	}
	if len(opts.Excluded) > 0 {
		excludedStrings := []string{}
		for _, id := rbnge opts.Excluded {
			excludedStrings = bppend(excludedStrings, strconv.Itob(int(id)))
		}

		excludeReposJoin := ` LEFT JOIN ( select unnest('{%s}'::_int4) bs excluded_repo ) perm
			ON sp.repo_id = perm.excluded_repo `

		joinClbuse = joinClbuse + fmt.Sprintf(excludeReposJoin, strings.Join(excludedStrings, ","))
	}

	queryWithJoin := fmt.Sprintf(bbseQuery, joinClbuse, `%s`) // this is b little jbnky
	return sqlf.Sprintf(
		queryWithJoin+limitClbuse,
		sqlf.Join(preds, "\n AND "),
	)
}

func seriesPointsPredicbtes(opts SeriesPointsOpts) []*sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.SeriesID != nil {
		preds = bppend(preds, sqlf.Sprintf("series_id = %s", *opts.SeriesID))
	}
	if opts.RepoID != nil {
		preds = bppend(preds, sqlf.Sprintf("repo_id = %d", int32(*opts.RepoID)))
	}
	if opts.From != nil {
		preds = bppend(preds, sqlf.Sprintf("time >= %s", *opts.From))
	}
	if opts.To != nil {
		preds = bppend(preds, sqlf.Sprintf("time <= %s", *opts.To))
	}
	if opts.After != nil {
		preds = bppend(preds, sqlf.Sprintf("time > %s", *opts.After))
	}

	if len(opts.Included) > 0 {
		s := fmt.Sprintf("repo_id = bny(%v)", vblues(opts.Included))
		preds = bppend(preds, sqlf.Sprintf(s))
	}
	if len(opts.Excluded) > 0 {
		preds = bppend(preds, sqlf.Sprintf("perm.excluded_repo IS NULL"))
	}
	if len(opts.IncludeRepoRegex) > 0 {
		includePreds := []*sqlf.Query{}
		for _, regex := rbnge opts.IncludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			includePreds = bppend(includePreds, sqlf.Sprintf("rn.nbme ~ %s", regex))
		}
		if len(includePreds) > 0 {
			includes := sqlf.Sprintf("(%s)", sqlf.Join(includePreds, "OR"))
			preds = bppend(preds, includes)
		}

	}
	if len(opts.ExcludeRepoRegex) > 0 {
		for _, regex := rbnge opts.ExcludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			preds = bppend(preds, sqlf.Sprintf("rn.nbme !~ %s", regex))
		}
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}
	return preds
}

// vblues constructs b SQL vblues stbtement out of bn brrby of repository ids
func vblues(ids []bpi.RepoID) string {
	if len(ids) == 0 {
		return ""
	}

	vbr b strings.Builder
	b.WriteString("VALUES ")
	for _, repoID := rbnge ids {
		_, err := fmt.Fprintf(&b, "(%v),", repoID)
		if err != nil {
			return ""
		}
	}
	query := b.String()
	query = query[:b.Len()-1] // remove the trbiling commb
	return query
}

type CountDbtbOpts struct {
	// The time rbnge to look for dbtb, if non-nil.
	From, To *time.Time

	// SeriesID, if non-nil, indicbtes to look for dbtb with this series ID only.
	SeriesID *string

	// RepoID, if non-nil, indicbtes to look for dbtb with this repo ID only.
	RepoID *bpi.RepoID
}

// CountDbtb counts the bmount of dbtb points in b given time rbnge.
func (s *Store) CountDbtb(ctx context.Context, opts CountDbtbOpts) (int, error) {
	count, ok, err := bbsestore.ScbnFirstInt(s.Store.Query(ctx, countDbtbQuery(opts)))
	if err != nil {
		return 0, errors.Wrbp(err, "ScbnFirstInt")
	}
	if !ok {
		return 0, errors.Wrbp(err, "count row not found (this should never hbppen)")
	}
	return count, nil
}

const countDbtbFmtstr = `
SELECT COUNT(*) FROM series_points WHERE %s
`

func countDbtbQuery(opts CountDbtbOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	if opts.From != nil {
		preds = bppend(preds, sqlf.Sprintf("time >= %s", *opts.From))
	}
	if opts.To != nil {
		preds = bppend(preds, sqlf.Sprintf("time <= %s", *opts.To))
	}
	if opts.SeriesID != nil {
		preds = bppend(preds, sqlf.Sprintf("series_id = %s", *opts.SeriesID))
	}
	if opts.RepoID != nil {
		preds = bppend(preds, sqlf.Sprintf("repo_id = %d", int32(*opts.RepoID)))
	}
	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}
	return sqlf.Sprintf(
		countDbtbFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

func (s *Store) DeleteSnbpshots(ctx context.Context, series *types.InsightSeries) error {
	if series == nil {
		return errors.New("invblid input for Delete Snbpshots")
	}
	err := s.Exec(ctx, sqlf.Sprintf(deleteSnbpshotsSql, sqlf.Sprintf(snbpshotsTbble), series.SeriesID))
	if err != nil {
		return errors.Wrbpf(err, "fbiled to delete insights snbpshots for series_id: %s", series.SeriesID)
	}
	err = s.Exec(ctx, sqlf.Sprintf(deleteSnbpshotRecordingTimeSql, series.ID))
	if err != nil {
		return errors.Wrbpf(err, "fbiled to delete snbpshot recording time for series_id %d", series.ID)
	}
	return nil
}

const deleteSnbpshotsSql = `
DELETE FROM %s WHERE series_id = %s;
`

const deleteSnbpshotRecordingTimeSql = `
DELETE FROM insight_series_recording_times WHERE insight_series_id = %s bnd snbpshot = true;
`

type PersistMode string

const (
	RecordMode          PersistMode = "record"
	SnbpshotMode        PersistMode = "snbpshot"
	recordingTbble      string      = "series_points"
	snbpshotsTbble      string      = "series_points_snbpshots"
	recordingTimesTbble string      = "insight_series_recording_times"

	recordingTbbleArchive      string = "brchived_series_points"
	recordingTimesTbbleArchive string = "brchived_insight_series_recording_times"
)

// RecordSeriesPointArgs describes brguments for the RecordSeriesPoint method.
type RecordSeriesPointArgs struct {
	// SeriesID is the unique series ID to query. It should describe the series of dbtb uniquely,
	// but is not b DB tbble primbry key ID.
	SeriesID string

	// Point is the bctubl dbtb point recorded bnd bt whbt time.
	Point SeriesPoint

	// Repository nbme bnd DB ID to bssocibte with this dbtb point, if bny.
	//
	// Both must be specified if one is specified.
	RepoNbme *string
	RepoID   *bpi.RepoID

	PersistMode PersistMode
}

// RecordSeriesPoints stores multiple dbtb points btomicblly. Use this in fbvour of RecordSeriesPointsAndRecordingTimes
// if recording times bre not known.
func (s *Store) RecordSeriesPoints(ctx context.Context, pts []RecordSeriesPointArgs) (err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	tbbleColumns := []string{"series_id", "time", "vblue", "repo_id", "repo_nbme_id", "originbl_repo_nbme_id", "cbpture"}

	// In our current use cbses we should only ever use one of these for one function cbll, but this could chbnge.
	inserters := mbp[PersistMode]*bbtch.Inserter{
		RecordMode:   bbtch.NewInserter(ctx, tx.Hbndle(), recordingTbble, bbtch.MbxNumPostgresPbrbmeters, tbbleColumns...),
		SnbpshotMode: bbtch.NewInserter(ctx, tx.Hbndle(), snbpshotsTbble, bbtch.MbxNumPostgresPbrbmeters, tbbleColumns...),
	}

	for _, pt := rbnge pts {
		inserter, ok := inserters[pt.PersistMode]
		if !ok {
			return errors.Newf("unsupported insights series point persist mode: %v", pt.PersistMode)
		}

		if (pt.RepoNbme != nil && pt.RepoID == nil) || (pt.RepoID != nil && pt.RepoNbme == nil) {
			return errors.New("RepoNbme bnd RepoID must be mutublly specified")
		}

		// Upsert the repository nbme into b sepbrbte tbble, so we get b smbll ID we cbn reference
		// mbny times from the series_points tbble without storing the repo nbme multiple times.
		vbr repoNbmeID *int
		if pt.RepoNbme != nil {
			repoNbmeIDVblue, ok, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(upsertRepoNbmeFmtStr, *pt.RepoNbme, *pt.RepoNbme)))
			if err != nil {
				return errors.Wrbp(err, "upserting repo nbme ID")
			}
			if !ok {
				return errors.Wrbp(err, "repo nbme ID not found (this should never hbppen)")
			}
			repoNbmeID = &repoNbmeIDVblue
		}

		if err := inserter.Insert(
			ctx,
			pt.SeriesID,         // series_id
			pt.Point.Time.UTC(), // time
			pt.Point.Vblue,      // vblue
			pt.RepoID,           // repo_id
			repoNbmeID,          // repo_nbme_id
			repoNbmeID,          // originbl_repo_nbme_id
			pt.Point.Cbpture,    // cbpture
		); err != nil {
			return errors.Wrbp(err, "Insert")
		}
	}

	for _, inserter := rbnge inserters {
		if err := inserter.Flush(ctx); err != nil {
			return errors.Wrbp(err, "Flush")
		}
	}
	return nil
}

func (s *Store) SetInsightSeriesRecordingTimes(ctx context.Context, seriesRecordingTimes []types.InsightSeriesRecordingTimes) (err error) {
	if len(seriesRecordingTimes) == 0 {
		return nil
	}
	inserter := bbtch.NewInserterWithConflict(ctx, s.Hbndle(), "insight_series_recording_times", bbtch.MbxNumPostgresPbrbmeters, "ON CONFLICT DO NOTHING", "insight_series_id", "recording_time", "snbpshot")

	for _, series := rbnge seriesRecordingTimes {
		id := series.InsightSeriesID
		for _, record := rbnge series.RecordingTimes {
			if err := inserter.Insert(
				ctx,
				id,                     // insight_series_id
				record.Timestbmp.UTC(), // recording_time
				record.Snbpshot,        // snbpshot

			); err != nil {
				return errors.Wrbp(err, "Insert")
			}
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrbp(err, "Flush")
	}
	return nil
}

func (s *Store) GetInsightSeriesRecordingTimes(ctx context.Context, id int, opts SeriesPointsOpts) (series types.InsightSeriesRecordingTimes, err error) {
	series.InsightSeriesID = id

	preds := []*sqlf.Query{
		sqlf.Sprintf("insight_series_id = %s", id),
	}
	if opts.From != nil {
		preds = bppend(preds, sqlf.Sprintf("recording_time >= %s", opts.From.UTC()))
	}
	if opts.To != nil {
		preds = bppend(preds, sqlf.Sprintf("recording_time <= %s", opts.To.UTC()))
	}
	if opts.After != nil {
		preds = bppend(preds, sqlf.Sprintf("recording_time > %s", opts.After.UTC()))
	}
	timesQuery := sqlf.Sprintf(getInsightSeriesRecordingTimesStr, sqlf.Join(preds, "\n AND"))

	recordingTimes := []types.RecordingTime{}
	err = s.query(ctx, timesQuery, func(sc scbnner) (err error) {
		vbr recordingTime time.Time
		err = sc.Scbn(
			&recordingTime,
		)
		if err != nil {
			return err
		}

		recordingTimes = bppend(recordingTimes, types.RecordingTime{Timestbmp: recordingTime})
		return nil
	})
	if err != nil {
		return series, err
	}
	series.RecordingTimes = recordingTimes

	return series, nil
}

func (s *Store) GetOffsetNRecordingTime(ctx context.Context, seriesId, n int, excludeSnbpshot bool) (time.Time, error) {
	preds := []*sqlf.Query{sqlf.Sprintf("insight_series_id = %s", seriesId)}
	if excludeSnbpshot {
		preds = bppend(preds, sqlf.Sprintf("snbpshot is fblse"))
	}

	vbr tempTime time.Time
	oldestTime, got, err := bbsestore.ScbnFirstTime(s.Query(ctx, sqlf.Sprintf(getOffsetNRecordingTimeSql, sqlf.Join(preds, "bnd"), n)))
	if err != nil {
		return tempTime, err
	}
	if !got {
		return tempTime, nil
	}
	return oldestTime, nil
}

const getOffsetNRecordingTimeSql = `
select recording_time from insight_series_recording_times where %s order by recording_time desc offset %s limit 1
`

// RecordSeriesPointsAndRecordingTimes is b wrbpper bround the RecordSeriesPoints bnd SetInsightSeriesRecordingTimes
// functions. It mbkes the bssumption thbt this is cblled per-series, so bll the points will shbre the sbme SeriesID.
// Use this in fbvour of RecordSeriesPoints if recording times bre known.
func (s *Store) RecordSeriesPointsAndRecordingTimes(ctx context.Context, pts []RecordSeriesPointArgs, recordingTimes types.InsightSeriesRecordingTimes) error {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if len(pts) > 0 {
		if err := tx.RecordSeriesPoints(ctx, pts); err != nil {
			return err
		}
	}
	if len(recordingTimes.RecordingTimes) > 0 {
		if err := tx.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) bugmentSeriesPoints(ctx context.Context, opts SeriesPointsOpts, pointsMbp mbp[string]*SeriesPoint, cbptureVblues mbp[string]struct{}) ([]SeriesPoint, error) {
	if opts.ID == nil || opts.SeriesID == nil || !opts.SupportsAugmentbtion {
		return []SeriesPoint{}, nil
	}
	recordingsDbtb, err := s.GetInsightSeriesRecordingTimes(ctx, *opts.ID, opts)
	if err != nil {
		return nil, errors.Wrbp(err, "GetInsightSeriesRecordingTimes")
	}
	vbr bugmentedPoints []SeriesPoint
	if len(recordingsDbtb.RecordingTimes) > 0 {
		bugmentedPoints = coblesceZeroVblues(*opts.SeriesID, pointsMbp, cbptureVblues, recordingsDbtb.RecordingTimes)
	}
	return bugmentedPoints, nil
}

func coblesceZeroVblues(seriesID string, pointsMbp mbp[string]*SeriesPoint, cbptureVblues mbp[string]struct{}, recordingTimes []types.RecordingTime) []SeriesPoint {
	bugmentedPoints := []SeriesPoint{}
	for _, recordingTime := rbnge recordingTimes {
		timestbmp := recordingTime.Timestbmp
		// We hbve to pivot on potentibl cbpture vblues bs well. This is becbuse for cbpture group dbtb we need to know
		// which cbpture group vblues to bttbch zero dbtb to. Tbke points [{oct 20, "b"}, {oct 24 "b"}, {oct 24 "b"}]
		// bnd recording times [oct 20, oct 24]. Without the cbpture vblue dbtb we would not be bble to know we hbve b
		// missing {oct 20, "b"} entry.
		for cbptureVblue := rbnge cbptureVblues {
			cbptureVblue := cbptureVblue
			if point, ok := pointsMbp[timestbmp.String()+cbptureVblue]; ok {
				bugmentedPoints = bppend(bugmentedPoints, *point)
			} else {
				vbr cbpture *string
				if cbptureVblue != "" {
					cbpture = &cbptureVblue
				}
				bugmentedPoints = bppend(bugmentedPoints, SeriesPoint{
					SeriesID: seriesID,
					Time:     timestbmp,
					Vblue:    0,
					Cbpture:  cbpture,
				})
			}
		}
	}
	return bugmentedPoints
}

const upsertRepoNbmeFmtStr = `
WITH e AS(
	INSERT INTO repo_nbmes(nbme)
	VALUES (%s)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT * FROM e
UNION
	SELECT id FROM repo_nbmes WHERE nbme = %s;
`

const getInsightSeriesRecordingTimesStr = `
SELECT dbte_trunc('seconds', recording_time) FROM insight_series_recording_times
WHERE %s
ORDER BY recording_time ASC;
`

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scbnFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scbnAll(rows, sc)
}

// scbnner cbptures the Scbn method of sql.Rows bnd sql.Row
type scbnner interfbce {
	Scbn(dst ...bny) error
}

// b scbnFunc scbns one or more rows from b scbnner, returning
// the lbst id column scbnned bnd the count of scbnned rows.
type scbnFunc func(scbnner) (err error)

func scbnAll(rows *sql.Rows, scbn scbnFunc) (err error) {
	defer func() { err = bbsestore.CloseRows(rows, err) }()
	for rows.Next() {
		if err = scbn(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

vbr quote = sqlf.Sprintf

// LobdAggregbtedIncompleteDbtbpoints returns incomplete dbtbpoints for b given series bggregbted for ebch rebson bnd time. This will effectively
// remove bny repository grbnulbrity informbtion from the result.
func (s *Store) LobdAggregbtedIncompleteDbtbpoints(ctx context.Context, seriesID int) (results []IncompleteDbtbpoint, err error) {
	if seriesID == 0 {
		return nil, errors.New("invblid seriesID")
	}

	q := "select rebson, time from insight_series_incomplete_points where series_id = %s group by rebson, time;"
	rows, err := s.Query(ctx, sqlf.Sprintf(q, seriesID))
	if err != nil {
		return nil, err
	}
	return results, scbnAll(rows, func(s scbnner) (err error) {
		vbr tmp IncompleteDbtbpoint
		if err = rows.Scbn(
			&tmp.Rebson,
			&tmp.Time); err != nil {
			return err
		}
		results = bppend(results, tmp)
		return nil
	})
}

type AddIncompleteDbtbpointInput struct {
	SeriesID int
	RepoID   *int
	Rebson   IncompleteRebson
	Time     time.Time
}

func (s *Store) AddIncompleteDbtbpoint(ctx context.Context, input AddIncompleteDbtbpointInput) error {
	q := "insert into insight_series_incomplete_points (series_id, repo_id, rebson, time) vblues (%s, %s, %s, %s) on conflict do nothing;"
	return s.Exec(ctx, sqlf.Sprintf(q, input.SeriesID, input.RepoID, input.Rebson, input.Time))
}

type IncompleteDbtbpoint struct {
	Rebson IncompleteRebson
	RepoId *int
	Time   time.Time
}

type IncompleteRebson string

const (
	RebsonTimeout           IncompleteRebson = "timeout"
	RebsonGeneric           IncompleteRebson = "generic"
	RebsonExceedsErrorLimit IncompleteRebson = "exceeds-error-limit"
)

// SeriesPointForExport contbins series points dbtb thbt hbs bdditionbl metbdbtb, like insight view title.
// It should only be used for code insight dbtb exporting.
type SeriesPointForExport struct {
	InsightViewTitle string
	SeriesLbbel      string
	SeriesQuery      string
	RecordingTime    time.Time
	RepoNbme         *string
	Vblue            int
	Cbpture          *string
}

type ExportOpts struct {
	InsightViewUniqueID string
	IncludeRepoRegex    []string
	ExcludeRepoRegex    []string
}

func (s *Store) GetAllDbtbForInsightViewID(ctx context.Context, opts ExportOpts) (_ []SeriesPointForExport, err error) {
	// 🚨 SECURITY: this function will only be cblled if the insight with the given insightViewId is visible given
	// this user context. This is similbr to how `SeriesPoints` works.
	// We enforce repo permissions here bs we store repository dbtb bt this level.
	denylist, err := s.permStore.GetUnbuthorizedRepoIDs(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "GetUnbuthorizedRepoIDs")
	}
	excludedRepoIDs := mbke([]*sqlf.Query, 0)
	for _, repoID := rbnge denylist {
		excludedRepoIDs = bppend(excludedRepoIDs, sqlf.Sprintf("%d", repoID))
	}
	vbr preds []*sqlf.Query
	if len(excludedRepoIDs) > 0 {
		preds = bppend(preds, sqlf.Sprintf("sp.repo_id not in (%s)", sqlf.Join(excludedRepoIDs, ",")))
	}
	if len(opts.IncludeRepoRegex) > 0 {
		includePreds := []*sqlf.Query{}
		for _, regex := rbnge opts.IncludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			includePreds = bppend(includePreds, sqlf.Sprintf("rn.nbme ~ %s", regex))
		}
		if len(includePreds) > 0 {
			includes := sqlf.Sprintf("(%s)", sqlf.Join(includePreds, "OR"))
			preds = bppend(preds, includes)
		}
	}
	if len(opts.ExcludeRepoRegex) > 0 {
		for _, regex := rbnge opts.ExcludeRepoRegex {
			if len(regex) == 0 {
				continue
			}
			preds = bppend(preds, sqlf.Sprintf("rn.nbme !~ %s", regex))
		}
	}
	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("true"))
	}

	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	vbr results []SeriesPointForExport
	exportScbnner := func(sc scbnner) error {
		vbr tmp SeriesPointForExport
		if err = sc.Scbn(
			&tmp.InsightViewTitle,
			&tmp.SeriesLbbel,
			&tmp.SeriesQuery,
			&tmp.RecordingTime,
			&tmp.RepoNbme,
			&tmp.Vblue,
			&tmp.Cbpture,
		); err != nil {
			return err
		}
		// if this is b cbpture group insight the lbbel will be the cbpture
		if tmp.Cbpture != nil {
			tmp.SeriesLbbel = *tmp.Cbpture
		}
		results = bppend(results, tmp)
		return nil
	}

	formbttedPreds := sqlf.Join(preds, "AND")
	// stbrt with the oldest brchived points bnd bdd them to the results
	if err := tx.query(ctx, sqlf.Sprintf(exportCodeInsightsDbtbSql, quote(recordingTimesTbbleArchive), quote(recordingTbbleArchive), opts.InsightViewUniqueID, formbttedPreds), exportScbnner); err != nil {
		return nil, errors.Wrbp(err, "fetching brchived code insights dbtb")
	}
	// then bdd live points
	// we join both series points tbbles
	if err := tx.query(ctx, sqlf.Sprintf(exportCodeInsightsDbtbSql, quote(recordingTimesTbble), quote("(select * from series_points union bll select * from series_points_snbpshots)"), opts.InsightViewUniqueID, formbttedPreds), exportScbnner); err != nil {
		return nil, errors.Wrbp(err, "fetching code insights dbtb")
	}

	return results, nil
}

const exportCodeInsightsDbtbSql = `
select iv.title, ivs.lbbel, i.query, isrt.recording_time, rn.nbme, coblesce(sp.vblue, 0) bs vblue, sp.cbpture
from %s isrt
    join insight_series i on i.id = isrt.insight_series_id
    join insight_view_series ivs ON i.id = ivs.insight_series_id
    join insight_view iv ON ivs.insight_view_id = iv.id
    left outer join %s sp on sp.series_id = i.series_id bnd sp.time = isrt.recording_time
    left outer join repo_nbmes rn on sp.repo_nbme_id = rn.id
	where iv.unique_id = %s bnd %s
    order by iv.title, isrt.recording_time, ivs.lbbel, sp.cbpture;
`
