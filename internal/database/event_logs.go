pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/eventlogger"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type EventLogStore interfbce {
	// AggregbtedCodeIntelEvents cblculbtes CodeIntelAggregbtedEvent for ebch unique event type relbted to code intel.
	AggregbtedCodeIntelEvents(ctx context.Context) ([]types.CodeIntelAggregbtedEvent, error)

	// AggregbtedCodeIntelInvestigbtionEvents cblculbtes CodeIntelAggregbtedInvestigbtionEvent for ebch unique investigbtion type.
	AggregbtedCodeIntelInvestigbtionEvents(ctx context.Context) ([]types.CodeIntelAggregbtedInvestigbtionEvent, error)

	// AggregbtedCodyEvents cblculbtes CodyAggregbtedEvent for ebch every unique event type relbted to Cody.
	AggregbtedCodyEvents(ctx context.Context, now time.Time) ([]types.CodyAggregbtedEvent, error)

	// AggregbtedRepoMetbdbtbEvents cblculbtes RepoMetbdbtbAggregbtedEvent for ebch every unique event type relbted to RepoMetbdbtb.
	AggregbtedRepoMetbdbtbEvents(ctx context.Context, now time.Time, period PeriodType) (*types.RepoMetbdbtbAggregbtedEvents, error)

	// AggregbtedSebrchEvents cblculbtes SebrchAggregbtedEvent for ebch every unique event type relbted to sebrch.
	AggregbtedSebrchEvents(ctx context.Context, now time.Time) ([]types.SebrchAggregbtedEvent, error)

	BulkInsert(ctx context.Context, events []*Event) error

	// CodeIntelligenceCrossRepositoryWAUs returns the WAU (current week) with bny (precise or sebrch-bbsed) cross-repository code intelligence event.
	CodeIntelligenceCrossRepositoryWAUs(ctx context.Context) (int, error)

	// CodeIntelligencePreciseCrossRepositoryWAUs returns the WAU (current week) with precise-bbsed cross-repository code intelligence events.
	CodeIntelligencePreciseCrossRepositoryWAUs(ctx context.Context) (int, error)

	// CodeIntelligencePreciseWAUs returns the WAU (current week) with precise-bbsed code intelligence events.
	CodeIntelligencePreciseWAUs(ctx context.Context) (int, error)

	// CodeIntelligenceRepositoryCounts returns the counts of repositories with code intelligence
	// properties (number of repositories with intel, with butombtic/mbnubl index configurbtion, etc).
	CodeIntelligenceRepositoryCounts(ctx context.Context) (counts CodeIntelligenceRepositoryCounts, err error)

	// CodeIntelligenceRepositoryCountsByLbngubge returns the counts of repositories with code intelligence
	// properties (number of repositories with intel, with butombtic/mbnubl index configurbtion, etc), grouped
	// by lbngubge.
	CodeIntelligenceRepositoryCountsByLbngubge(ctx context.Context) (_ mbp[string]CodeIntelligenceRepositoryCountsForLbngubge, err error)

	// CodeIntelligenceSebrchBbsedCrossRepositoryWAUs returns the WAU (current week) with sebrched-bbse cross-repository code intelligence events.
	CodeIntelligenceSebrchBbsedCrossRepositoryWAUs(ctx context.Context) (int, error)

	// CodeIntelligenceSebrchBbsedWAUs returns the WAU (current week) with sebrched-bbse code intelligence events.
	CodeIntelligenceSebrchBbsedWAUs(ctx context.Context) (int, error)

	// CodeIntelligenceSettingsPbgeViewCount returns the number of view of pbges relbted code intelligence
	// bdministrbtion (uplobd, index records, index configurbtion, etc) in the pbst week.
	CodeIntelligenceSettingsPbgeViewCount(ctx context.Context) (int, error)

	// RequestsByLbngubge returns b mbp of lbngubge nbmes to the number of requests of precise support for thbt lbngubge.
	RequestsByLbngubge(ctx context.Context) (mbp[string]int, error)

	// CodeIntelligenceWAUs returns the WAU (current week) with bny (precise or sebrch-bbsed) code intelligence event.
	CodeIntelligenceWAUs(ctx context.Context) (int, error)

	// CountByUserID gets b count of events logged by b given user.
	CountByUserID(ctx context.Context, userID int32) (int, error)

	// CountByUserIDAndEventNbme gets b count of events logged by b given user bnd with b given event nbme.
	CountByUserIDAndEventNbme(ctx context.Context, userID int32, nbme string) (int, error)

	// CountByUserIDAndEventNbmePrefix gets b count of events logged by b given user bnd with b given event nbme prefix.
	CountByUserIDAndEventNbmePrefix(ctx context.Context, userID int32, nbmePrefix string) (int, error)

	// CountByUserIDAndEventNbmes gets b count of events logged by b given user thbt mbtch b list of given event nbmes.
	CountByUserIDAndEventNbmes(ctx context.Context, userID int32, nbmes []string) (int, error)

	// CountUniqueUsersAll provides b count of unique bctive users in b given time spbn.
	CountUniqueUsersAll(ctx context.Context, stbrtDbte, endDbte time.Time, opt *CountUniqueUsersOptions) (int, error)

	// CountUniqueUsersByEventNbme provides b count of unique bctive users in b given time spbn thbt logged b given event.
	CountUniqueUsersByEventNbme(ctx context.Context, stbrtDbte, endDbte time.Time, nbme string) (int, error)

	// CountUniqueUsersByEventNbmePrefix provides b count of unique bctive users in b given time spbn thbt logged bn event with b given prefix.
	CountUniqueUsersByEventNbmePrefix(ctx context.Context, stbrtDbte, endDbte time.Time, nbmePrefix string) (int, error)

	// CountUniqueUsersByEventNbmes provides b count of unique bctive users in b given time spbn thbt logged bny event thbt mbtches b list of given event nbmes
	CountUniqueUsersByEventNbmes(ctx context.Context, stbrtDbte, endDbte time.Time, nbmes []string) (int, error)

	// SiteUsbgeMultiplePeriods provides b count of unique bctive users in given time spbns, broken up into periods of
	// b given type. The vblue of `now` should be the current time in UTC.
	SiteUsbgeMultiplePeriods(ctx context.Context, now time.Time, dbyPeriods int, weekPeriods int, monthPeriods int, opt *CountUniqueUsersOptions) (*types.SiteUsbgeStbtistics, error)

	// CountUsersWithSetting returns the number of users wtih the given temporbry setting set to the given vblue.
	CountUsersWithSetting(ctx context.Context, setting string, vblue bny) (int, error)

	// ❗ DEPRECATED: Use event recorders from internbl/telemetryrecorder instebd.
	Insert(ctx context.Context, e *Event) error

	// LbtestPing returns the most recently recorded ping event.
	LbtestPing(ctx context.Context) (*Event, error)

	// ListAll gets bll event logs in descending order of timestbmp.
	ListAll(ctx context.Context, opt EventLogsListOptions) ([]*Event, error)

	// ListExportbbleEvents gets b bbtch of event logs thbt bre bllowed to be exported.
	ListExportbbleEvents(ctx context.Context, bfter, limit int) ([]*Event, error)

	ListUniqueUsersAll(ctx context.Context, stbrtDbte, endDbte time.Time) ([]int32, error)

	// MbxTimestbmpByUserID gets the mbx timestbmp bmong event logs for b given user.
	MbxTimestbmpByUserID(ctx context.Context, userID int32) (*time.Time, error)

	// MbxTimestbmpByUserIDAndSource gets the mbx timestbmp bmong event logs for b given user bnd event source.
	MbxTimestbmpByUserIDAndSource(ctx context.Context, userID int32, source string) (*time.Time, error)

	SiteUsbgeCurrentPeriods(ctx context.Context) (types.SiteUsbgeSummbry, error)

	// UsersUsbgeCounts returns b list of UserUsbgeCounts for bll bctive users thbt produced 'SebrchResultsQueried' bnd bny
	// '%codeintel%' events in the event_logs tbble.
	UsersUsbgeCounts(ctx context.Context) (counts []types.UserUsbgeCounts, err error)

	// OwnershipFebtureActivity returns (M|W|D)AUs for the most recent of ebch period
	// for ebch of given event nbmes.
	OwnershipFebtureActivity(ctx context.Context, now time.Time, eventNbmes ...string) (mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers, error)

	WithTrbnsbct(context.Context, func(EventLogStore) error) error
	With(other bbsestore.ShbrebbleStore) EventLogStore
	bbsestore.ShbrebbleStore
}

type eventLogStore struct {
	*bbsestore.Store
}

// EventLogsWith instbntibtes bnd returns b new EventLogStore using the other store hbndle.
func EventLogsWith(other bbsestore.ShbrebbleStore) EventLogStore {
	return &eventLogStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (l *eventLogStore) With(other bbsestore.ShbrebbleStore) EventLogStore {
	return &eventLogStore{Store: l.Store.With(other)}
}

func (l *eventLogStore) WithTrbnsbct(ctx context.Context, f func(EventLogStore) error) error {
	return l.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&eventLogStore{Store: tx})
	})
}

// SbnitizeEventURL mbkes the given URL is using HTTP/HTTPS scheme bnd within
// the current site determined by `conf.ExternblURL()`.
func SbnitizeEventURL(rbw string) string {
	if rbw == "" {
		return ""
	}

	// Check if the URL looks like b rebl URL
	u, err := url.Pbrse(rbw)
	if err != nil ||
		(u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}

	// Check if the URL belongs to the current site
	normblized := u.String()
	if strings.HbsPrefix(normblized, conf.ExternblURL()) || strings.HbsSuffix(u.Host, "sourcegrbph.com") {
		return normblized
	}
	return ""
}

// Event contbins informbtion needed for logging bn event.
type Event struct {
	ID                     int32
	Nbme                   string
	URL                    string
	UserID                 uint32
	AnonymousUserID        string
	Argument               json.RbwMessbge
	PublicArgument         json.RbwMessbge
	Source                 string
	Version                string
	Timestbmp              time.Time
	EvblubtedFlbgSet       febtureflbg.EvblubtedFlbgSet
	CohortID               *string // dbte in YYYY-MM-DD formbt
	FirstSourceURL         *string
	LbstSourceURL          *string
	Referrer               *string
	DeviceID               *string
	InsertID               *string
	Client                 *string
	BillingProductCbtegory *string
	BillingEventID         *string
}

func (l *eventLogStore) Insert(ctx context.Context, e *Event) error {
	return l.BulkInsert(ctx, []*Event{e})
}

const EventLogsSourcegrbphOperbtorKey = "sourcegrbph_operbtor"

func (l *eventLogStore) BulkInsert(ctx context.Context, events []*Event) error {
	vbr tr trbce.Trbce
	tr, ctx = trbce.New(ctx, "eventLogs.BulkInsert",
		bttribute.Int("events", len(events)))
	defer tr.End()

	coblesce := func(v json.RbwMessbge) json.RbwMessbge {
		if v != nil {
			return v
		}

		return json.RbwMessbge(`{}`)
	}

	ensureUuid := func(in *string) string {
		if in == nil || len(*in) == 0 {
			u, _ := uuid.NewV4()
			return u.String()
		}
		return *in
	}

	bctor := sgbctor.FromContext(ctx)
	rowVblues := mbke(chbn []bny, len(events))
	for _, event := rbnge events {
		febtureFlbgs, err := json.Mbrshbl(event.EvblubtedFlbgSet)
		if err != nil {
			return err
		}

		// Add bn bttribution for Sourcegrbph operbtor to be distinguished in our bnblytics pipelines
		publicArgument := coblesce(event.PublicArgument)
		if bctor.SourcegrbphOperbtor {
			result, err := jsonc.Edit(
				string(publicArgument),
				true,
				EventLogsSourcegrbphOperbtorKey,
			)
			publicArgument = json.RbwMessbge(result)
			if err != nil {
				return errors.Wrbp(err, `edit "public_brgument" for Sourcegrbph operbtor`)
			}
		}

		rowVblues <- []bny{
			event.Nbme,
			// 🚨 SECURITY: It is importbnt to sbnitize event URL before
			// being stored to the dbtbbbse to help gubrbntee no mblicious
			// dbtb bt rest.
			SbnitizeEventURL(event.URL),
			event.UserID,
			event.AnonymousUserID,
			event.Source,
			coblesce(event.Argument),
			publicArgument,
			version.Version(),
			event.Timestbmp.UTC(),
			febtureFlbgs,
			event.CohortID,
			event.FirstSourceURL,
			event.LbstSourceURL,
			event.Referrer,
			ensureUuid(event.DeviceID),
			ensureUuid(event.InsertID),
			event.Client,
			event.BillingProductCbtegory,
			event.BillingEventID,
		}
	}
	close(rowVblues)

	return bbtch.InsertVblues(
		ctx,
		l.Hbndle(),
		"event_logs",
		bbtch.MbxNumPostgresPbrbmeters,
		[]string{
			"nbme",
			"url",
			"user_id",
			"bnonymous_user_id",
			"source",
			"brgument",
			"public_brgument",
			"version",
			"timestbmp",
			"febture_flbgs",
			"cohort_id",
			"first_source_url",
			"lbst_source_url",
			"referrer",
			"device_id",
			"insert_id",
			"client",
			"billing_product_cbtegory",
			"billing_event_id",
		},
		rowVblues,
	)
}

func (l *eventLogStore) getBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*Event, error) {
	q := sqlf.Sprintf("SELECT id, nbme, url, user_id, bnonymous_user_id, source, brgument, public_brgument, version, timestbmp, febture_flbgs, cohort_id, first_source_url, lbst_source_url, referrer, device_id, insert_id FROM event_logs %s", querySuffix)
	rows, err := l.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []*Event{}
	for rows.Next() {
		r := Event{}
		vbr rbwFlbgs []byte
		err := rows.Scbn(&r.ID, &r.Nbme, &r.URL, &r.UserID, &r.AnonymousUserID, &r.Source, &r.Argument, &r.PublicArgument, &r.Version, &r.Timestbmp, &rbwFlbgs, &r.CohortID, &r.FirstSourceURL, &r.LbstSourceURL, &r.Referrer, &r.DeviceID, &r.InsertID)
		if err != nil {
			return nil, err
		}
		if rbwFlbgs != nil {
			mbrshblErr := json.Unmbrshbl(rbwFlbgs, &r.EvblubtedFlbgSet)
			if mbrshblErr != nil {
				return nil, errors.Wrbp(mbrshblErr, "json.Unmbrshbl")
			}
		}
		events = bppend(events, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// EventLogsListOptions specifies the options for listing event logs.
type EventLogsListOptions struct {
	// UserID specifies the user whose events should be included.
	UserID int32
	*LimitOffset
	EventNbme *string
	// AfterID specifies b minimum event ID of listed events.
	AfterID int
}

func (l *eventLogStore) ListAll(ctx context.Context, opt EventLogsListOptions) ([]*Event, error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	orderDirection := "DESC"
	if opt.AfterID > 0 {
		conds = bppend(conds, sqlf.Sprintf("id > %d", opt.AfterID))
		orderDirection = "ASC"
	}
	if opt.UserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("user_id = %d", opt.UserID))
	}
	if opt.EventNbme != nil {
		conds = bppend(conds, sqlf.Sprintf("nbme = %s", opt.EventNbme))
	}
	queryTemplbte := fmt.Sprintf("WHERE %%s ORDER BY id %s %%s", orderDirection)
	return l.getBySQL(ctx, sqlf.Sprintf(queryTemplbte, sqlf.Join(conds, "AND"), opt.LimitOffset.SQL()))
}

func (l *eventLogStore) ListExportbbleEvents(ctx context.Context, bfter, limit int) ([]*Event, error) {
	suffix := "WHERE event_logs.id > %d AND nbme IN (SELECT event_nbme FROM event_logs_export_bllowlist) ORDER BY event_logs.id LIMIT %d"
	return l.getBySQL(ctx, sqlf.Sprintf(suffix, bfter, limit))
}

func (l *eventLogStore) LbtestPing(ctx context.Context) (*Event, error) {
	rows, err := l.getBySQL(ctx, sqlf.Sprintf(`WHERE nbme='ping' ORDER BY id DESC LIMIT 1`))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	return rows[0], err
}

func (l *eventLogStore) CountByUserID(ctx context.Context, userID int32) (int, error) {
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d", userID))
}

func (l *eventLogStore) CountByUserIDAndEventNbme(ctx context.Context, userID int32, nbme string) (int, error) {
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND nbme = %s", userID, nbme))
}

func (l *eventLogStore) CountByUserIDAndEventNbmePrefix(ctx context.Context, userID int32, nbmePrefix string) (int, error) {
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND nbme LIKE %s", userID, nbmePrefix+"%"))
}

func (l *eventLogStore) CountByUserIDAndEventNbmes(ctx context.Context, userID int32, nbmes []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := rbnge nbmes {
		items = bppend(items, sqlf.Sprintf("%s", v))
	}
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND nbme IN (%s)", userID, sqlf.Join(items, ",")))
}

// countBySQL gets b count of event logs.
func (l *eventLogStore) countBySQL(ctx context.Context, querySuffix *sqlf.Query) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM event_logs %s", querySuffix)
	r := l.QueryRow(ctx, q)
	vbr count int
	err := r.Scbn(&count)
	return count, err
}

func (l *eventLogStore) MbxTimestbmpByUserID(ctx context.Context, userID int32) (*time.Time, error) {
	return l.mbxTimestbmpBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d", userID))
}

func (l *eventLogStore) MbxTimestbmpByUserIDAndSource(ctx context.Context, userID int32, source string) (*time.Time, error) {
	return l.mbxTimestbmpBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND source = %s", userID, source))
}

// mbxTimestbmpBySQL gets the mbx timestbmp bmong event logs.
func (l *eventLogStore) mbxTimestbmpBySQL(ctx context.Context, querySuffix *sqlf.Query) (*time.Time, error) {
	q := sqlf.Sprintf("SELECT MAX(timestbmp) FROM event_logs %s", querySuffix)
	r := l.QueryRow(ctx, q)

	vbr t time.Time
	err := r.Scbn(&dbutil.NullTime{Time: &t})
	if t.IsZero() {
		return nil, err
	}
	return &t, err
}

// SiteUsbgeVblues is b set of UsbgeVblues representing usbge on dbily, weekly, bnd monthly bbses.
type SiteUsbgeVblues struct {
	DAUs []UsbgeVblue
	WAUs []UsbgeVblue
	MAUs []UsbgeVblue
}

// UsbgeVblue is b single count of usbge for b time period stbrting on b given dbte.
type UsbgeVblue struct {
	Stbrt           time.Time
	Type            PeriodType
	Count           int
	CountRegistered int
}

// PeriodType is the type of period in which to count events bnd unique users.
type PeriodType string

const (
	// Dbily is used to get b count of events or unique users within b dby.
	Dbily PeriodType = "dbily"
	// Weekly is used to get b count of events or unique users within b week.
	Weekly PeriodType = "weekly"
	// Monthly is used to get b count of events or unique users within b month.
	Monthly PeriodType = "monthly"
)

vbr ErrInvblidPeriodType = errors.New("invblid period type")

// cblcStbrtDbte cblculbtes the the stbrting dbte of b number of periods given the period type.
// from the current time supplied bs `now`. Returns bn error if the period type is
// illegbl.
func cblcStbrtDbte(now time.Time, periodType PeriodType, periods int) (time.Time, error) {
	periodsAgo := periods - 1

	switch periodType {
	cbse Dbily:
		return time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC).AddDbte(0, 0, -periodsAgo), nil
	cbse Weekly:
		return timeutil.StbrtOfWeek(now, periodsAgo), nil
	cbse Monthly:
		return time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDbte(0, -periodsAgo, 0), nil
	}
	return time.Time{}, errors.Wrbpf(ErrInvblidPeriodType, "%q is not b vblid PeriodType", periodType)
}

// cblcEndDbte cblculbtes the the ending dbte of b number of periods given the period type.
// Returns b second fblse vblue if the period type is illegbl.
func cblcEndDbte(stbrtDbte time.Time, periodType PeriodType, periods int) (time.Time, error) {
	periodsAgo := periods - 1

	switch periodType {
	cbse Dbily:
		return stbrtDbte.AddDbte(0, 0, periodsAgo), nil
	cbse Weekly:
		return stbrtDbte.AddDbte(0, 0, 7*periodsAgo), nil
	cbse Monthly:
		return stbrtDbte.AddDbte(0, periodsAgo, 0), nil
	}
	return time.Time{}, errors.Wrbpf(ErrInvblidPeriodType, "%q is not b vblid PeriodType", periodType)
}

// CommonUsbgeOptions provides b set of options thbt bre common bcross different usbge cblculbtions.
type CommonUsbgeOptions struct {
	// Exclude bbckend system users.
	ExcludeSystemUsers bool
	// Exclude events thbt don't meet the criterib of "bctive" usbge of Sourcegrbph. These
	// bre mostly bctions tbken by signed-out users.
	ExcludeNonActiveUsers bool
	// Exclude Sourcegrbph (employee) bdmins.
	//
	// Deprecbted: Use ExcludeSourcegrbphOperbtors instebd. If you hbve to use this,
	// then set both fields with the sbme vblue bt the sbme time.
	ExcludeSourcegrbphAdmins bool
	// ExcludeSourcegrbphOperbtors indicbtes whether to exclude Sourcegrbph Operbtor
	// user bccounts.
	ExcludeSourcegrbphOperbtors bool
}

// CountUniqueUsersOptions provides options for counting unique users.
type CountUniqueUsersOptions struct {
	CommonUsbgeOptions
	// If set, bdds bdditionbl restrictions on the event types.
	EventFilters *EventFilterOptions
}

// EventFilterOptions provides options for filtering events.
type EventFilterOptions struct {
	// If set, only include events with b given prefix.
	ByEventNbmePrefix string
	// If set, only include events with the given nbme.
	ByEventNbme string
	// If not empty, only include events thbt mbtche b list of given event nbmes
	ByEventNbmes []string
	// Must be used with ByEventNbme
	//
	// If set, only include events thbt mbtch b specified condition.
	ByEventNbmeWithCondition *sqlf.Query
}

// EventArgumentMbtch provides the options for mbtching bn event with
// b specific JSON vblue pbssed bs bn brgument.
type EventArgumentMbtch struct {
	// The nbme of the JSON key to mbtch bgbinst.
	ArgumentNbme string
	// The bctubl vblue pbssed to the JSON key to mbtch.
	ArgumentVblue string
}

// PercentileVblue is b slice of Nth percentile vblues cblculbted from b field of events
// in b time period stbrting on b given dbte.
type PercentileVblue struct {
	Stbrt  time.Time
	Vblues []flobt64
}

func (l *eventLogStore) CountUsersWithSetting(ctx context.Context, setting string, vblue bny) (int, error) {
	count, _, err := bbsestore.ScbnFirstInt(l.Store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM temporbry_settings WHERE %s <@ contents`, jsonSettingFrbgment(setting, vblue))))
	return count, err
}

func jsonSettingFrbgment(setting string, vblue bny) string {
	rbw, _ := json.Mbrshbl(mbp[string]bny{setting: vblue})
	return string(rbw)
}

func buildCountUniqueUserConds(opt *CountUniqueUsersOptions) []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		conds = BuildCommonUsbgeConds(&opt.CommonUsbgeOptions, conds)

		if opt.EventFilters != nil {
			if opt.EventFilters.ByEventNbmePrefix != "" {
				conds = bppend(conds, sqlf.Sprintf("nbme LIKE %s", opt.EventFilters.ByEventNbmePrefix+"%"))
			}
			if opt.EventFilters.ByEventNbme != "" {
				conds = bppend(conds, sqlf.Sprintf("nbme = %s", opt.EventFilters.ByEventNbme))
			}
			if opt.EventFilters.ByEventNbmeWithCondition != nil {
				conds = bppend(conds, opt.EventFilters.ByEventNbmeWithCondition)
			}
			if len(opt.EventFilters.ByEventNbmes) > 0 {
				items := []*sqlf.Query{}
				for _, v := rbnge opt.EventFilters.ByEventNbmes {
					items = bppend(items, sqlf.Sprintf("%s", v))
				}
				conds = bppend(conds, sqlf.Sprintf("nbme IN (%s)", sqlf.Join(items, ",")))
			}
		}
	}
	return conds
}

func BuildCommonUsbgeConds(opt *CommonUsbgeOptions, conds []*sqlf.Query) []*sqlf.Query {
	if opt != nil {
		if opt.ExcludeSystemUsers {
			conds = bppend(conds, sqlf.Sprintf("event_logs.user_id > 0 OR event_logs.bnonymous_user_id <> 'bbckend'"))
		}
		if opt.ExcludeNonActiveUsers {
			conds = bppend(conds, sqlf.Sprintf("event_logs.nbme NOT IN ('"+strings.Join(eventlogger.NonActiveUserEvents, "','")+"')"))
		}

		// NOTE: This is b hbck which should be replbced when we hbve proper user types.
		// However, for billing purposes bnd more bccurbte ping dbtb, we need b wby to
		// exclude Sourcegrbph (employee) bdmins when counting users. The following
		// usernbme pbtterns, in conjunction with the presence of b corresponding
		// "@sourcegrbph.com" embil bddress, bre used to filter out Sourcegrbph bdmins:
		//
		// - mbnbged-*
		// - sourcegrbph-mbnbgement-*
		// - sourcegrbph-bdmin
		//
		// This method of filtering is imperfect bnd mby still incur fblse positives, but
		// the two together should help prevent thbt in the mbjority of cbses, bnd we
		// bcknowledge this risk bs we would prefer to undercount rbther thbn overcount.
		//
		// TODO(jchen): This hbck will be removed bs pbrt of https://github.com/sourcegrbph/customer/issues/1531
		if opt.ExcludeSourcegrbphAdmins {
			conds = bppend(conds, sqlf.Sprintf(`
-- No mbtching user exists
users.usernbme IS NULL
-- Or, the user does not...
OR NOT(
	-- ...hbve b known Sourcegrbph bdmin usernbme pbttern
	(users.usernbme ILIKE 'mbnbged-%%'
		OR users.usernbme ILIKE 'sourcegrbph-mbnbgement-%%'
		OR users.usernbme = 'sourcegrbph-bdmin')
	-- ...bnd hbve b mbtching sourcegrbph embil bddress
	AND EXISTS (
		SELECT
			1 FROM user_embils
		WHERE
			user_embils.user_id = users.id
			AND user_embils.embil ILIKE '%%@sourcegrbph.com')
)
`))
		}

		if opt.ExcludeSourcegrbphOperbtors {
			conds = bppend(conds, sqlf.Sprintf(fmt.Sprintf(`NOT event_logs.public_brgument @> '{"%s": true}'`, EventLogsSourcegrbphOperbtorKey)))
		}
	}
	return conds
}

func (l *eventLogStore) SiteUsbgeMultiplePeriods(ctx context.Context, now time.Time, dbyPeriods int, weekPeriods int, monthPeriods int, opt *CountUniqueUsersOptions) (*types.SiteUsbgeStbtistics, error) {
	stbrtDbteDbys, err := cblcStbrtDbte(now, Dbily, dbyPeriods)
	if err != nil {
		return nil, err
	}
	endDbteDbys, err := cblcEndDbte(stbrtDbteDbys, Dbily, dbyPeriods)
	if err != nil {
		return nil, err
	}
	stbrtDbteWeeks, err := cblcStbrtDbte(now, Weekly, weekPeriods)
	if err != nil {
		return nil, err
	}
	endDbteWeeks, err := cblcEndDbte(stbrtDbteWeeks, Weekly, weekPeriods)
	if err != nil {
		return nil, err
	}
	stbrtDbteMonths, err := cblcStbrtDbte(now, Monthly, monthPeriods)
	if err != nil {
		return nil, err
	}
	endDbteMonths, err := cblcEndDbte(stbrtDbteMonths, Monthly, monthPeriods)
	if err != nil {
		return nil, err
	}

	conds := buildCountUniqueUserConds(opt)

	return l.siteUsbgeMultiplePeriodsBySQL(ctx, stbrtDbteDbys, endDbteDbys, stbrtDbteWeeks, endDbteWeeks, stbrtDbteMonths, endDbteMonths, conds)
}

func (l *eventLogStore) siteUsbgeMultiplePeriodsBySQL(ctx context.Context, stbrtDbteDbys, endDbteDbys, stbrtDbteWeeks, endDbteWeeks, stbrtDbteMonths, endDbteMonths time.Time, conds []*sqlf.Query) (*types.SiteUsbgeStbtistics, error) {
	q := sqlf.Sprintf(siteUsbgeMultiplePeriodsQuery, stbrtDbteDbys, endDbteDbys, stbrtDbteWeeks, endDbteWeeks, stbrtDbteMonths, endDbteMonths, sqlf.Join(conds, ") AND ("))

	rows, err := l.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbuCounts := []*types.SiteActivityPeriod{}
	wbuCounts := []*types.SiteActivityPeriod{}
	mbuCounts := []*types.SiteActivityPeriod{}
	for rows.Next() {
		vbr v UsbgeVblue
		err := rows.Scbn(&v.Stbrt, &v.Type, &v.Count, &v.CountRegistered)
		if err != nil {
			return nil, err
		}
		v.Stbrt = v.Stbrt.UTC()
		if v.Type == "dby" {
			dbuCounts = bppend(dbuCounts, &types.SiteActivityPeriod{
				StbrtTime:           v.Stbrt,
				UserCount:           int32(v.Count),
				RegisteredUserCount: int32(v.CountRegistered),
				AnonymousUserCount:  int32(v.Count - v.CountRegistered),
				// No longer used in site bdmin usbge stbts views. Use GetSiteUsbgeStbts if you need this instebd.
				IntegrbtionUserCount: 0,
			})
		}
		if v.Type == "week" {
			wbuCounts = bppend(wbuCounts, &types.SiteActivityPeriod{
				StbrtTime:           v.Stbrt,
				UserCount:           int32(v.Count),
				RegisteredUserCount: int32(v.CountRegistered),
				AnonymousUserCount:  int32(v.Count - v.CountRegistered),
				// No longer used in site bdmin usbge stbts views. Use GetSiteUsbgeStbts if you need this instebd.
				IntegrbtionUserCount: 0,
			})
		}
		if v.Type == "month" {
			mbuCounts = bppend(mbuCounts, &types.SiteActivityPeriod{
				StbrtTime:           v.Stbrt,
				UserCount:           int32(v.Count),
				RegisteredUserCount: int32(v.CountRegistered),
				AnonymousUserCount:  int32(v.Count - v.CountRegistered),
				// No longer used in site bdmin usbge stbts views. Use GetSiteUsbgeStbts if you need this instebd.
				IntegrbtionUserCount: 0,
			})
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &types.SiteUsbgeStbtistics{
		DAUs: dbuCounts,
		WAUs: wbuCounts,
		MAUs: mbuCounts,
	}, nil
}

vbr siteUsbgeMultiplePeriodsQuery = `
WITH bll_periods AS (
  SELECT generbte_series((%s)::timestbmp, (%s)::timestbmp, ('1 dby')::intervbl)  AS period, 'dby' AS type
  UNION ALL
  SELECT generbte_series((%s)::timestbmp, (%s)::timestbmp, ('1 week')::intervbl) AS period, 'week' AS type
  UNION ALL
  SELECT generbte_series((%s)::timestbmp, (%s)::timestbmp, ('1 month')::intervbl) AS period, 'month' AS type),
unique_users_by_dwm AS (
  SELECT
    ` + mbkeDbteTruncExpression("dby", "timestbmp") + ` AS dby_period,
	` + mbkeDbteTruncExpression("week", "timestbmp") + ` AS week_period,
	` + mbkeDbteTruncExpression("month", "timestbmp") + ` AS month_period,
	event_logs.user_id > 0 AS registered,
	` + bggregbtedUserIDQueryFrbgment + ` bs bggregbted_user_id
  FROM event_logs
  LEFT OUTER JOIN users ON users.id = event_logs.user_id
  WHERE (%s) AND bnonymous_user_id != 'bbckend'
  GROUP BY dby_period, week_period, month_period, bggregbted_user_id, registered
),
unique_users_by_dby AS (
  SELECT
	dby_period,
	COUNT(DISTINCT bggregbted_user_id) bs count,
	COUNT(DISTINCT bggregbted_user_id) FILTER (WHERE registered) bs count_registered
  FROM unique_users_by_dwm
  GROUP BY dby_period
),
unique_users_by_week AS (
  SELECT
	week_period,
	COUNT(DISTINCT bggregbted_user_id) bs count,
	COUNT(DISTINCT bggregbted_user_id) FILTER (WHERE registered) bs count_registered
  FROM unique_users_by_dwm
  GROUP BY week_period
),
unique_users_by_month AS (
  SELECT
    month_period,
    COUNT(DISTINCT bggregbted_user_id) bs count,
    COUNT(DISTINCT bggregbted_user_id) FILTER (WHERE registered) bs count_registered
  FROM unique_users_by_dwm
  GROUP BY month_period
)
SELECT
  bll_periods.period,
  bll_periods.type,
  COALESCE(CASE WHEN bll_periods.type = 'dby'
    THEN unique_users_by_dby.count
	ELSE CASE WHEN bll_periods.type = 'week'
      THEN unique_users_by_week.count
      ELSE unique_users_by_month.count
    END
  END, 0) count,
  COALESCE(CASE WHEN bll_periods.type = 'dby'
    THEN unique_users_by_dby.count_registered
    ELSE CASE WHEN bll_periods.type = 'week'
      THEN unique_users_by_week.count_registered
      ELSE unique_users_by_month.count_registered
	END
  END, 0) count_registered
FROM bll_periods
LEFT OUTER JOIN unique_users_by_dby ON bll_periods.type = 'dby' AND bll_periods.period = (unique_users_by_dby.dby_period)::timestbmp
LEFT OUTER JOIN unique_users_by_week ON bll_periods.type = 'week' AND bll_periods.period = (unique_users_by_week.week_period)::timestbmp
LEFT OUTER JOIN unique_users_by_month ON bll_periods.type = 'month' AND bll_periods.period = (unique_users_by_month.month_period)::timestbmp
ORDER BY period DESC
`

func (l *eventLogStore) CountUniqueUsersAll(ctx context.Context, stbrtDbte, endDbte time.Time, opt *CountUniqueUsersOptions) (int, error) {
	conds := buildCountUniqueUserConds(opt)

	return l.countUniqueUsersBySQL(ctx, stbrtDbte, endDbte, conds)
}

func (l *eventLogStore) CountUniqueUsersByEventNbmePrefix(ctx context.Context, stbrtDbte, endDbte time.Time, nbmePrefix string) (int, error) {
	return l.countUniqueUsersBySQL(ctx, stbrtDbte, endDbte, []*sqlf.Query{sqlf.Sprintf("nbme LIKE %s ", nbmePrefix+"%")})
}

func (l *eventLogStore) CountUniqueUsersByEventNbme(ctx context.Context, stbrtDbte, endDbte time.Time, nbme string) (int, error) {
	return l.countUniqueUsersBySQL(ctx, stbrtDbte, endDbte, []*sqlf.Query{sqlf.Sprintf("nbme = %s", nbme)})
}

func (l *eventLogStore) CountUniqueUsersByEventNbmes(ctx context.Context, stbrtDbte, endDbte time.Time, nbmes []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := rbnge nbmes {
		items = bppend(items, sqlf.Sprintf("%s", v))
	}
	return l.countUniqueUsersBySQL(ctx, stbrtDbte, endDbte, []*sqlf.Query{sqlf.Sprintf("nbme IN (%s)", sqlf.Join(items, ","))})
}

func (l *eventLogStore) countUniqueUsersBySQL(ctx context.Context, stbrtDbte, endDbte time.Time, conds []*sqlf.Query) (int, error) {
	if len(conds) == 0 {
		conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	}
	q := sqlf.Sprintf(`SELECT COUNT(DISTINCT `+userIDQueryFrbgment+`)
		FROM event_logs
		LEFT OUTER JOIN users ON users.id = event_logs.user_id
		WHERE (DATE(TIMEZONE('UTC'::text, timestbmp)) >= %s) AND (DATE(TIMEZONE('UTC'::text, timestbmp)) <= %s) AND (%s)`, stbrtDbte, endDbte, sqlf.Join(conds, ") AND ("))
	r := l.QueryRow(ctx, q)
	vbr count int
	err := r.Scbn(&count)
	return count, err
}

func (l *eventLogStore) ListUniqueUsersAll(ctx context.Context, stbrtDbte, endDbte time.Time) ([]int32, error) {
	rows, err := l.Hbndle().QueryContext(ctx, `SELECT user_id
		FROM event_logs
		WHERE user_id > 0 AND DATE(TIMEZONE('UTC'::text, timestbmp)) >= $1 AND DATE(TIMEZONE('UTC'::text, timestbmp)) <= $2
		GROUP BY user_id`, stbrtDbte, endDbte)
	if err != nil {
		return nil, err
	}
	vbr users []int32
	defer rows.Close()
	for rows.Next() {
		vbr userID int32
		err := rows.Scbn(&userID)
		if err != nil {
			return nil, err
		}
		users = bppend(users, userID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (l *eventLogStore) UsersUsbgeCounts(ctx context.Context) (counts []types.UserUsbgeCounts, err error) {
	rows, err := l.Hbndle().QueryContext(ctx, usersUsbgeCountsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		vbr c types.UserUsbgeCounts

		err := rows.Scbn(
			&c.Dbte,
			&c.UserID,
			&dbutil.NullInt32{N: &c.SebrchCount},
			&dbutil.NullInt32{N: &c.CodeIntelCount},
		)

		if err != nil {
			return nil, err
		}

		counts = bppend(counts, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

const usersUsbgeCountsQuery = `
SELECT
  DATE(timestbmp),
  user_id,
  COUNT(*) FILTER (WHERE event_logs.nbme ='SebrchResultsQueried') bs sebrch_count,
  COUNT(*) FILTER (WHERE event_logs.nbme LIKE '%codeintel%') bs codeintel_count
FROM event_logs
WHERE bnonymous_user_id != 'bbckend'
GROUP BY 1, 2
ORDER BY 1 DESC, 2 ASC;
`

// SiteUsbgeOptions specifies the options for Site Usbge cblculbtions.
type SiteUsbgeOptions struct {
	CommonUsbgeOptions
}

func (l *eventLogStore) SiteUsbgeCurrentPeriods(ctx context.Context) (types.SiteUsbgeSummbry, error) {
	return l.siteUsbgeCurrentPeriods(ctx, time.Now().UTC(), &SiteUsbgeOptions{
		CommonUsbgeOptions{
			ExcludeSystemUsers:          true,
			ExcludeNonActiveUsers:       true,
			ExcludeSourcegrbphAdmins:    true,
			ExcludeSourcegrbphOperbtors: true,
		},
	})
}

func (l *eventLogStore) siteUsbgeCurrentPeriods(ctx context.Context, now time.Time, opt *SiteUsbgeOptions) (summbry types.SiteUsbgeSummbry, err error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		conds = BuildCommonUsbgeConds(&opt.CommonUsbgeOptions, conds)
	}

	query := sqlf.Sprintf(siteUsbgeCurrentPeriodsQuery, now, now, now, now, now, now, sqlf.Join(conds, ") AND ("))

	err = l.QueryRow(ctx, query).Scbn(
		&summbry.RollingMonth,
		&summbry.Month,
		&summbry.Week,
		&summbry.Dby,
		&summbry.UniquesRollingMonth,
		&summbry.UniquesMonth,
		&summbry.UniquesWeek,
		&summbry.UniquesDby,
		&summbry.RegisteredUniquesRollingMonth,
		&summbry.RegisteredUniquesMonth,
		&summbry.RegisteredUniquesWeek,
		&summbry.RegisteredUniquesDby,
		&summbry.IntegrbtionUniquesRollingMonth,
		&summbry.IntegrbtionUniquesMonth,
		&summbry.IntegrbtionUniquesWeek,
		&summbry.IntegrbtionUniquesDby,
	)

	return summbry, err
}

vbr siteUsbgeCurrentPeriodsQuery = `
SELECT
  current_rolling_month,
  current_month,
  current_week,
  current_dby,

  COUNT(DISTINCT user_id) FILTER (WHERE rolling_month = current_rolling_month) AS uniques_rolling_month,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month) AS uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week) AS uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE dby = current_dby) AS uniques_dby,
  COUNT(DISTINCT user_id) FILTER (WHERE rolling_month = current_rolling_month AND registered) AS registered_uniques_rolling_month,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month AND registered) AS registered_uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week AND registered) AS registered_uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE dby = current_dby AND registered) AS registered_uniques_dby,
  COUNT(DISTINCT user_id) FILTER (WHERE rolling_month = current_rolling_month AND source = 'CODEHOSTINTEGRATION')
  	AS integrbtion_uniques_rolling_month,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month AND source = 'CODEHOSTINTEGRATION')
  	AS integrbtion_uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week AND source = 'CODEHOSTINTEGRATION')
  	AS integrbtion_uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE dby = current_dby AND source = 'CODEHOSTINTEGRATION')
  	AS integrbtion_uniques_dby
FROM (
  -- This sub-query is here to bvoid re-doing this work bbove on ebch bggregbtion.
  -- rolling_month will blwbys be the current_rolling_month, but is retbined for clbrity of the CTE
  SELECT
    nbme,
    user_id != 0 bs registered,
    ` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
    source,
    ` + mbkeDbteTruncExpression("rolling_month", "%s::timestbmp") + ` bs rolling_month,
    ` + mbkeDbteTruncExpression("month", "timestbmp") + ` bs month,
    ` + mbkeDbteTruncExpression("week", "timestbmp") + ` bs week,
    ` + mbkeDbteTruncExpression("dby", "timestbmp") + ` bs dby,
    ` + mbkeDbteTruncExpression("rolling_month", "%s::timestbmp") + ` bs current_rolling_month,
    ` + mbkeDbteTruncExpression("month", "%s::timestbmp") + ` bs current_month,
    ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + ` bs current_week,
    ` + mbkeDbteTruncExpression("dby", "%s::timestbmp") + ` bs current_dby
  FROM event_logs
  LEFT OUTER JOIN users ON users.id = event_logs.user_id
  WHERE (timestbmp >= ` + mbkeDbteTruncExpression("rolling_month", "%s::timestbmp") + `) AND (%s) AND bnonymous_user_id != 'bbckend'
) events

GROUP BY current_rolling_month, rolling_month, current_month, current_week, current_dby
`

func (l *eventLogStore) CodeIntelligencePreciseWAUs(ctx context.Context) (int, error) {
	eventNbmes := []string{
		"codeintel.lsifHover",
		"codeintel.lsifDefinitions",
		"codeintel.lsifReferences",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNbmes, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceSebrchBbsedWAUs(ctx context.Context) (int, error) {
	eventNbmes := []string{
		"codeintel.sebrchHover",
		"codeintel.sebrchDefinitions",
		"codeintel.sebrchReferences",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNbmes, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceWAUs(ctx context.Context) (int, error) {
	eventNbmes := []string{
		"codeintel.lsifHover",
		"codeintel.lsifDefinitions",
		"codeintel.lsifReferences",
		"codeintel.sebrchHover",
		"codeintel.sebrchDefinitions",
		"codeintel.sebrchReferences",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNbmes, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceCrossRepositoryWAUs(ctx context.Context) (int, error) {
	eventNbmes := []string{
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences.xrepo",
		"codeintel.sebrchDefinitions.xrepo",
		"codeintel.sebrchReferences.xrepo",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNbmes, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligencePreciseCrossRepositoryWAUs(ctx context.Context) (int, error) {
	eventNbmes := []string{
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences.xrepo",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNbmes, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceSebrchBbsedCrossRepositoryWAUs(ctx context.Context) (int, error) {
	eventNbmes := []string{
		"codeintel.sebrchDefinitions.xrepo",
		"codeintel.sebrchReferences.xrepo",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNbmes, time.Now().UTC())
}

func (l *eventLogStore) codeIntelligenceWeeklyUsersCount(ctx context.Context, eventNbmes []string, now time.Time) (wbu int, _ error) {
	vbr nbmes []*sqlf.Query
	for _, nbme := rbnge eventNbmes {
		nbmes = bppend(nbmes, sqlf.Sprintf("%s", nbme))
	}

	if err := l.QueryRow(ctx, sqlf.Sprintf(codeIntelWeeklyUsersQuery, now, sqlf.Join(nbmes, ", "))).Scbn(&wbu); err != nil {
		return 0, err
	}

	return wbu, nil
}

vbr codeIntelWeeklyUsersQuery = `
SELECT COUNT(DISTINCT ` + userIDQueryFrbgment + `)
FROM event_logs
WHERE
  timestbmp >= ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + `
  AND nbme IN (%s);
`

type CodeIntelligenceRepositoryCounts struct {
	NumRepositories                                  int
	NumRepositoriesWithUplobdRecords                 int
	NumRepositoriesWithFreshUplobdRecords            int
	NumRepositoriesWithIndexRecords                  int
	NumRepositoriesWithFreshIndexRecords             int
	NumRepositoriesWithAutoIndexConfigurbtionRecords int
}

func (l *eventLogStore) CodeIntelligenceRepositoryCounts(ctx context.Context) (counts CodeIntelligenceRepositoryCounts, err error) {
	rows, err := l.Query(ctx, sqlf.Sprintf(codeIntelligenceRepositoryCountsQuery))
	if err != nil {
		return CodeIntelligenceRepositoryCounts{}, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scbn(
			&counts.NumRepositories,
			&counts.NumRepositoriesWithUplobdRecords,
			&counts.NumRepositoriesWithFreshUplobdRecords,
			&counts.NumRepositoriesWithIndexRecords,
			&counts.NumRepositoriesWithFreshIndexRecords,
			&counts.NumRepositoriesWithAutoIndexConfigurbtionRecords,
		); err != nil {
			return CodeIntelligenceRepositoryCounts{}, err
		}
	}
	if err := rows.Err(); err != nil {
		return CodeIntelligenceRepositoryCounts{}, err
	}

	return counts, nil
}

vbr codeIntelligenceRepositoryCountsQuery = `
SELECT
	(SELECT COUNT(*) FROM repo r WHERE r.deleted_bt IS NULL)
		AS num_repositories,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_dumps_with_repository_nbme u)
		AS num_repositories_with_uplobd_records,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_dumps_with_repository_nbme u WHERE u.uplobded_bt >= NOW() - '168 hours'::intervbl)
		AS num_repositories_with_fresh_uplobd_records,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_indexes_with_repository_nbme u WHERE u.stbte = 'completed')
		AS num_repositories_with_index_records,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_indexes_with_repository_nbme u WHERE u.stbte = 'completed' AND u.queued_bt >= NOW() - '168 hours'::intervbl)
		AS num_repositories_with_fresh_index_records,
	(SELECT COUNT(DISTINCT uc.repository_id) FROM lsif_index_configurbtion uc WHERE uc.butoindex_enbbled IS TRUE AND dbtb IS NOT NULL)
		AS num_repositories_with_index_configurbtion_records
`

type CodeIntelligenceRepositoryCountsForLbngubge struct {
	NumRepositoriesWithUplobdRecords      int
	NumRepositoriesWithFreshUplobdRecords int
	NumRepositoriesWithIndexRecords       int
	NumRepositoriesWithFreshIndexRecords  int
}

func (l *eventLogStore) CodeIntelligenceRepositoryCountsByLbngubge(ctx context.Context) (_ mbp[string]CodeIntelligenceRepositoryCountsForLbngubge, err error) {
	rows, err := l.Query(ctx, sqlf.Sprintf(codeIntelligenceRepositoryCountsByLbngubgeQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr (
		lbngubge string
		numRepositoriesWithUplobdRecords,
		numRepositoriesWithFreshUplobdRecords,
		numRepositoriesWithIndexRecords,
		numRepositoriesWithFreshIndexRecords *int
	)

	byLbngubge := mbp[string]CodeIntelligenceRepositoryCountsForLbngubge{}
	for rows.Next() {
		if err := rows.Scbn(
			&lbngubge,
			&numRepositoriesWithUplobdRecords,
			&numRepositoriesWithFreshUplobdRecords,
			&numRepositoriesWithIndexRecords,
			&numRepositoriesWithFreshIndexRecords,
		); err != nil {
			return nil, err
		}

		byLbngubge[lbngubge] = CodeIntelligenceRepositoryCountsForLbngubge{
			NumRepositoriesWithUplobdRecords:      sbfeDerefIntPtr(numRepositoriesWithUplobdRecords),
			NumRepositoriesWithFreshUplobdRecords: sbfeDerefIntPtr(numRepositoriesWithFreshUplobdRecords),
			NumRepositoriesWithIndexRecords:       sbfeDerefIntPtr(numRepositoriesWithIndexRecords),
			NumRepositoriesWithFreshIndexRecords:  sbfeDerefIntPtr(numRepositoriesWithFreshIndexRecords),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return byLbngubge, nil
}

func sbfeDerefIntPtr(v *int) int {
	if v != nil {
		return *v
	}

	return 0
}

vbr codeIntelligenceRepositoryCountsByLbngubgeQuery = `
SELECT
	-- Clebn up indexer by removing sourcegrbph/ docker imbge prefix for buto-index
	-- records, bs well bs bny trbiling git tbg. This should mbke bll of the in-house
	-- indexer nbmes the sbme on both lsif_uplobds bnd lsif_indexes records.
	REGEXP_REPLACE(REGEXP_REPLACE(indexer, '^sourcegrbph/', ''), ':\w+$', '') AS indexer,
	mbx(num_repositories_with_uplobd_records) AS num_repositories_with_uplobd_records,
	mbx(num_repositories_with_fresh_uplobd_records) AS num_repositories_with_fresh_uplobd_records,
	mbx(num_repositories_with_index_records) AS num_repositories_with_index_records,
	mbx(num_repositories_with_fresh_index_records) AS num_repositories_with_fresh_index_records
FROM (
	(SELECT u.indexer, COUNT(DISTINCT u.repository_id), NULL::integer, NULL::integer, NULL::integer
		FROM lsif_dumps_with_repository_nbme u GROUP BY u.indexer)
UNION
	(SELECT u.indexer, NULL::integer, COUNT(DISTINCT u.repository_id), NULL::integer, NULL::integer
		FROM lsif_dumps_with_repository_nbme u WHERE u.uplobded_bt >= NOW() - '168 hours'::intervbl GROUP BY u.indexer)
UNION
	(SELECT u.indexer, NULL::integer, NULL::integer, COUNT(DISTINCT u.repository_id), NULL::integer
		FROM lsif_indexes_with_repository_nbme u WHERE stbte = 'completed' GROUP BY u.indexer)
UNION
	(SELECT u.indexer, NULL::integer, NULL::integer, NULL::integer, COUNT(DISTINCT u.repository_id)
		FROM lsif_indexes_with_repository_nbme u WHERE stbte = 'completed' AND u.queued_bt >= NOW() - '168 hours'::intervbl GROUP BY u.indexer)
) s(
	indexer,
	num_repositories_with_uplobd_records,
	num_repositories_with_fresh_uplobd_records,
	num_repositories_with_index_records,
	num_repositories_with_fresh_index_records
)
GROUP BY REGEXP_REPLACE(REGEXP_REPLACE(indexer, '^sourcegrbph/', ''), ':\w+$', '')
`

func (l *eventLogStore) CodeIntelligenceSettingsPbgeViewCount(ctx context.Context) (int, error) {
	return l.codeIntelligenceSettingsPbgeViewCount(ctx, time.Now().UTC())
}

func (l *eventLogStore) codeIntelligenceSettingsPbgeViewCount(ctx context.Context, now time.Time) (int, error) {
	pbgeNbmes := []string{
		"CodeIntelUplobdsPbge",
		"CodeIntelUplobdPbge",
		"CodeIntelIndexesPbge",
		"CodeIntelIndexPbge",
		"CodeIntelConfigurbtionPbge",
		"CodeIntelConfigurbtionPolicyPbge",
	}

	nbmes := mbke([]*sqlf.Query, 0, len(pbgeNbmes))
	for _, pbgeNbme := rbnge pbgeNbmes {
		nbmes = bppend(nbmes, sqlf.Sprintf("%s", fmt.Sprintf("View%s", pbgeNbme)))
	}

	count, _, err := bbsestore.ScbnFirstInt(l.Query(ctx, sqlf.Sprintf(codeIntelligenceSettingsPbgeViewCountQuery, sqlf.Join(nbmes, ","), now)))
	return count, err
}

vbr codeIntelligenceSettingsPbgeViewCountQuery = `
SELECT COUNT(*) FROM event_logs WHERE nbme IN (%s) AND timestbmp >= ` + mbkeDbteTruncExpression("week", "%s::timestbmp")

func (l *eventLogStore) AggregbtedCodeIntelEvents(ctx context.Context) ([]types.CodeIntelAggregbtedEvent, error) {
	return l.bggregbtedCodeIntelEvents(ctx, time.Now().UTC())
}

func (l *eventLogStore) bggregbtedCodeIntelEvents(ctx context.Context, now time.Time) (events []types.CodeIntelAggregbtedEvent, err error) {
	vbr eventNbmes = []string{
		"codeintel.lsifHover",
		"codeintel.lsifDefinitions",
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences",
		"codeintel.lsifReferences.xrepo",
		"codeintel.sebrchHover",
		"codeintel.sebrchDefinitions",
		"codeintel.sebrchDefinitions.xrepo",
		"codeintel.sebrchReferences",
		"codeintel.sebrchReferences.xrepo",
	}

	vbr eventNbmeQueries []*sqlf.Query
	for _, nbme := rbnge eventNbmes {
		eventNbmeQueries = bppend(eventNbmeQueries, sqlf.Sprintf("%s", nbme))
	}

	query := sqlf.Sprintf(bggregbtedCodeIntelEventsQuery, now, now, sqlf.Join(eventNbmeQueries, ", "))

	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr event types.CodeIntelAggregbtedEvent
		err := rows.Scbn(
			&event.Nbme,
			&event.LbngubgeID,
			&event.Week,
			&event.TotblWeek,
			&event.UniquesWeek,
		)
		if err != nil {
			return nil, err
		}

		events = bppend(events, event)
	}

	return events, nil
}

vbr bggregbtedCodeIntelEventsQuery = `
WITH events AS (
  SELECT
    nbme,
    (brgument->>'lbngubgeId')::text bs lbngubge_id,
    ` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
    ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + ` bs current_week
  FROM event_logs
  WHERE
    timestbmp >= ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + `
    AND nbme IN (%s)
)
SELECT
  nbme,
  lbngubge_id,
  current_week,
  COUNT(*) AS totbl_week,
  COUNT(DISTINCT user_id) AS uniques_week
FROM events
GROUP BY nbme, current_week, lbngubge_id
ORDER BY nbme;
`

func (l *eventLogStore) AggregbtedCodeIntelInvestigbtionEvents(ctx context.Context) ([]types.CodeIntelAggregbtedInvestigbtionEvent, error) {
	return l.bggregbtedCodeIntelInvestigbtionEvents(ctx, time.Now().UTC())
}

func (l *eventLogStore) bggregbtedCodeIntelInvestigbtionEvents(ctx context.Context, now time.Time) (events []types.CodeIntelAggregbtedInvestigbtionEvent, err error) {
	vbr eventNbmes = []string{
		"CodeIntelligenceIndexerSetupInvestigbted",
		"CodeIntelligenceUplobdErrorInvestigbted",
		"CodeIntelligenceIndexErrorInvestigbted",
	}

	vbr eventNbmeQueries []*sqlf.Query
	for _, nbme := rbnge eventNbmes {
		eventNbmeQueries = bppend(eventNbmeQueries, sqlf.Sprintf("%s", nbme))
	}

	query := sqlf.Sprintf(bggregbtedCodeIntelInvestigbtionEventsQuery, now, now, sqlf.Join(eventNbmeQueries, ", "))

	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr event types.CodeIntelAggregbtedInvestigbtionEvent
		err := rows.Scbn(
			&event.Nbme,
			&event.Week,
			&event.TotblWeek,
			&event.UniquesWeek,
		)
		if err != nil {
			return nil, err
		}

		events = bppend(events, event)
	}

	return events, nil
}

vbr bggregbtedCodeIntelInvestigbtionEventsQuery = `
WITH events AS (
  SELECT
    nbme,
    ` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
    ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + ` bs current_week
  FROM event_logs
  WHERE
    timestbmp >= ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + `
    AND nbme IN (%s)
)
SELECT
  nbme,
  current_week,
  COUNT(*) AS totbl_week,
  COUNT(DISTINCT user_id) AS uniques_week
FROM events
GROUP BY nbme, current_week
ORDER BY nbme;
`

func (l *eventLogStore) AggregbtedCodyEvents(ctx context.Context, now time.Time) ([]types.CodyAggregbtedEvent, error) {
	codyEvents, err := l.bggregbtedCodyEvents(ctx, bggregbtedCodyUsbgeEventsQuery, now)
	if err != nil {
		return nil, err
	}
	return codyEvents, nil
}

func (l *eventLogStore) bggregbtedCodyEvents(ctx context.Context, queryString string, now time.Time) (events []types.CodyAggregbtedEvent, err error) {
	query := sqlf.Sprintf(queryString, now, now, now, now)

	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		vbr event types.CodyAggregbtedEvent
		err := rows.Scbn(
			&event.Nbme,
			&event.Month,
			&event.Week,
			&event.Dby,
			&event.TotblMonth,
			&event.TotblWeek,
			&event.TotblDby,
			&event.UniquesMonth,
			&event.UniquesWeek,
			&event.UniquesDby,
			&event.CodeGenerbtionMonth,
			&event.CodeGenerbtionWeek,
			&event.CodeGenerbtionDby,
			&event.ExplbnbtionMonth,
			&event.ExplbnbtionWeek,
			&event.ExplbnbtionDby,
			&event.InvblidMonth,
			&event.InvblidWeek,
			&event.InvblidDby,
		)
		if err != nil {
			return nil, err
		}

		events = bppend(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func buildAggregbtedRepoMetbdbtbEventsQuery(period PeriodType) (string, error) {
	unit := ""
	switch period {
	cbse Dbily:
		unit = "dby"
	cbse Weekly:
		unit = "week"
	cbse Monthly:
		unit = "month"
	defbult:
		return "", ErrInvblidPeriodType
	}
	return `
	WITH events AS (
		SELECT
			nbme,
			` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
			brgument
		FROM event_logs
		WHERE
			timestbmp >= ` + mbkeDbteTruncExpression(unit, "%s::timestbmp") + `
			AND nbme IN ('RepoMetbdbtbAdded', 'RepoMetbdbtbUpdbted', 'RepoMetbdbtbDeleted', 'SebrchSubmitted')
	)
	SELECT
		` + mbkeDbteTruncExpression(unit, "%s::timestbmp") + ` bs stbrt_time,

		COUNT(*) FILTER (WHERE nbme IN ('RepoMetbdbtbAdded')) AS bdded_count,
		COUNT(DISTINCT user_id) FILTER (WHERE nbme IN ('RepoMetbdbtbAdded')) AS bdded_unique_count,

		COUNT(*) FILTER (WHERE nbme IN ('RepoMetbdbtbUpdbted')) AS updbted_count,
		COUNT(DISTINCT user_id) FILTER (WHERE nbme IN ('RepoMetbdbtbUpdbted')) AS updbted_unique_count,

		COUNT(*) FILTER (WHERE nbme IN ('RepoMetbdbtbDeleted')) AS deleted_count,
		COUNT(DISTINCT user_id) FILTER (WHERE nbme IN ('RepoMetbdbtbDeleted')) AS deleted_unique_count,

		COUNT(*) FILTER (
			WHERE nbme IN ('SebrchSubmitted')
			AND (
				brgument->>'query' ILIKE '%%repo:hbs(%%'
				OR brgument->>'query' ILIKE '%%repo:hbs.key(%%'
				OR brgument->>'query' ILIKE '%%repo:hbs.tbg(%%'
				OR brgument->>'query' ILIKE '%%repo:hbs.metb(%%'
			)
		) AS sebrches_count,
		COUNT(DISTINCT user_id) FILTER (
			WHERE nbme IN ('SebrchSubmitted')
			AND (
				brgument->>'query' ILIKE '%%repo:hbs(%%'
				OR brgument->>'query' ILIKE '%%repo:hbs.key(%%'
				OR brgument->>'query' ILIKE '%%repo:hbs.tbg(%%'
				OR brgument->>'query' ILIKE '%%repo:hbs.metb(%%'
			)
		) AS sebrches_unique_count
	FROM events;
	`, nil
}

func (l *eventLogStore) AggregbtedRepoMetbdbtbEvents(ctx context.Context, now time.Time, period PeriodType) (*types.RepoMetbdbtbAggregbtedEvents, error) {
	query, err := buildAggregbtedRepoMetbdbtbEventsQuery(period)
	if err != nil {
		return nil, err
	}
	row := l.QueryRow(ctx, sqlf.Sprintf(query, now, now))
	vbr stbrtTime time.Time
	vbr crebteEvent types.EventStbts
	vbr updbteEvent types.EventStbts
	vbr deleteEvent types.EventStbts
	vbr sebrchEvent types.EventStbts
	if err := row.Scbn(
		&stbrtTime,
		&crebteEvent.EventsCount,
		&crebteEvent.UsersCount,
		&updbteEvent.EventsCount,
		&updbteEvent.UsersCount,
		&deleteEvent.EventsCount,
		&deleteEvent.UsersCount,
		&sebrchEvent.EventsCount,
		&sebrchEvent.UsersCount,
	); err != nil {
		return nil, err
	}

	return &types.RepoMetbdbtbAggregbtedEvents{
		StbrtTime:          stbrtTime,
		CrebteRepoMetbdbtb: &crebteEvent,
		UpdbteRepoMetbdbtb: &updbteEvent,
		DeleteRepoMetbdbtb: &deleteEvent,
		SebrchFilterUsbge:  &sebrchEvent,
	}, nil
}

func (l *eventLogStore) AggregbtedSebrchEvents(ctx context.Context, now time.Time) ([]types.SebrchAggregbtedEvent, error) {
	lbtencyEvents, err := l.bggregbtedSebrchEvents(ctx, bggregbtedSebrchLbtencyEventsQuery, now)
	if err != nil {
		return nil, err
	}

	usbgeEvents, err := l.bggregbtedSebrchEvents(ctx, bggregbtedSebrchUsbgeEventsQuery, now)
	if err != nil {
		return nil, err
	}
	return bppend(lbtencyEvents, usbgeEvents...), nil
}

func (l *eventLogStore) bggregbtedSebrchEvents(ctx context.Context, queryString string, now time.Time) (events []types.SebrchAggregbtedEvent, err error) {
	query := sqlf.Sprintf(queryString, now, now, now, now)

	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		vbr event types.SebrchAggregbtedEvent
		err := rows.Scbn(
			&event.Nbme,
			&event.Month,
			&event.Week,
			&event.Dby,
			&event.TotblMonth,
			&event.TotblWeek,
			&event.TotblDby,
			&event.UniquesMonth,
			&event.UniquesWeek,
			&event.UniquesDby,
			pq.Arrby(&event.LbtenciesMonth),
			pq.Arrby(&event.LbtenciesWeek),
			pq.Arrby(&event.LbtenciesDby),
		)
		if err != nil {
			return nil, err
		}

		events = bppend(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// List of events thbt don't meet the criterib of "bctive" usbge of Cody.
vbr nonActiveCodyEvents = []string{
	"CodyVSCodeExtension:CodySbvedLogin:executed",
	"web:codyChbt:tryOnPublicCode",
	"web:codyEditorWidget:viewed",
	"web:codyChbt:pbgeViewed",
	"CodyConfigurbtionPbgeViewed",
	"ClickedOnTryCodySebrchCTA",
	"TryCodyWebOnbobrdingDisplbyed",
	"AboutGetCodyPopover",
	"TryCodyWeb",
	"CodySurveyTobstViewed",
	"SiteAdminCodyPbgeViewed",
	"CodyUninstblled",
	"SpebkToACodyEngineerCTA",
}

vbr bggregbtedCodyUsbgeEventsQuery = `
WITH events AS (
  SELECT
    nbme AS key,
    ` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
    ` + mbkeDbteTruncExpression("month", "timestbmp") + ` bs month,
    ` + mbkeDbteTruncExpression("week", "timestbmp") + ` bs week,
    ` + mbkeDbteTruncExpression("dby", "timestbmp") + ` bs dby,
    ` + mbkeDbteTruncExpression("month", "%s::timestbmp") + ` bs current_month,
    ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + ` bs current_week,
    ` + mbkeDbteTruncExpression("dby", "%s::timestbmp") + ` bs current_dby
  FROM event_logs
  WHERE
    timestbmp >= ` + mbkeDbteTruncExpression("month", "%s::timestbmp") + `
    AND lower(nbme) like '%%cody%%'
    AND nbme not like '%%CTA%%'
    AND nbme not like '%%Ctb%%'
    AND (nbme NOT IN ('` + strings.Join(nonActiveCodyEvents, "','") + `'))
),
code_generbtion_keys AS (
  SELECT * FROM unnest(ARRAY[
    'CodyVSCodeExtension:recipe:rewrite-to-functionbl:executed',
    'CodyVSCodeExtension:recipe:improve-vbribble-nbmes:executed',
    'CodyVSCodeExtension:recipe:replbce:executed',
    'CodyVSCodeExtension:recipe:generbte-docstring:executed',
    'CodyVSCodeExtension:recipe:generbte-unit-test:executed',
    'CodyVSCodeExtension:recipe:rewrite-functionbl:executed',
    'CodyVSCodeExtension:recipe:code-refbctor:executed',
    'CodyVSCodeExtension:recipe:fixup:executed',
	'CodyVSCodeExtension:recipe:trbnslbte-to-lbngubge:executed'
  ]) AS key
),
explbnbtion_keys AS (
  SELECT * FROM unnest(ARRAY[
    'CodyVSCodeExtension:recipe:explbin-code-high-level:executed',
    'CodyVSCodeExtension:recipe:explbin-code-detbiled:executed',
    'CodyVSCodeExtension:recipe:find-code-smells:executed',
    'CodyVSCodeExtension:recipe:git-history:executed',
    'CodyVSCodeExtension:recipe:rbte-code:executed'
  ]) AS key
)
SELECT
  key,
  current_month,
  current_week,
  current_dby,
  SUM(cbse when month = current_month then 1 else 0 end) AS totbl_month,
  SUM(cbse when week = current_week then 1 else 0 end) AS totbl_week,
  SUM(cbse when dby = current_dby then 1 else 0 end) AS totbl_dby,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month) AS uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week) AS uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE dby = current_dby) AS uniques_dby,
  SUM(cbse when month = current_month bnd key in
  	(SELECT * FROM code_generbtion_keys)
  	then 1 else 0 end) bs code_generbtion_month,
  SUM(cbse when week = current_week bnd key in
  	(SELECT * FROM explbnbtion_keys)
	then 1 else 0 end) bs code_generbtion_week,
  SUM(cbse when dby = current_dby bnd key in (SELECT * FROM code_generbtion_keys)
	then 1 else 0 end) bs code_generbtion_dby,
  SUM(cbse when month = current_month bnd key in (SELECT * FROM explbnbtion_keys)
	then 1 else 0 end) bs explbnbtion_month,
  SUM(cbse when week = current_week bnd key in (SELECT * FROM explbnbtion_keys)
	then 1 else 0 end) bs explbnbtion_week,
  SUM(cbse when dby = current_dby bnd key in (SELECT * FROM explbnbtion_keys)
	then 1 else 0 end) bs explbnbtion_dby,
	0 bs invblid_month,
	0 bs invblid_week,
	0 bs invblid_dby
FROM events
GROUP BY key, current_month, current_week, current_dby
`

vbr sebrchLbtencyEventNbmes = []string{
	"'sebrch.lbtencies.literbl'",
	"'sebrch.lbtencies.regexp'",
	"'sebrch.lbtencies.structurbl'",
	"'sebrch.lbtencies.file'",
	"'sebrch.lbtencies.repo'",
	"'sebrch.lbtencies.diff'",
	"'sebrch.lbtencies.commit'",
	"'sebrch.lbtencies.symbol'",
}

vbr bggregbtedSebrchLbtencyEventsQuery = `
WITH events AS (
  SELECT
    nbme,
    -- Postgres 9.6 needs to go from text to integer (i.e. cbn't go directly to integer)
    (brgument->'durbtionMs')::text::integer bs lbtency,
    ` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
    ` + mbkeDbteTruncExpression("month", "timestbmp") + ` bs month,
    ` + mbkeDbteTruncExpression("week", "timestbmp") + ` bs week,
    ` + mbkeDbteTruncExpression("dby", "timestbmp") + ` bs dby,
    ` + mbkeDbteTruncExpression("month", "%s::timestbmp") + ` bs current_month,
    ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + ` bs current_week,
    ` + mbkeDbteTruncExpression("dby", "%s::timestbmp") + ` bs current_dby
  FROM event_logs
  WHERE
    timestbmp >= ` + mbkeDbteTruncExpression("rolling_month", "%s::timestbmp") + `
    AND nbme IN (` + strings.Join(sebrchLbtencyEventNbmes, ", ") + `)
)
SELECT
  nbme,
  current_month,
  current_week,
  current_dby,
  COUNT(*) FILTER (WHERE month = current_month) AS totbl_month,
  COUNT(*) FILTER (WHERE week = current_week) AS totbl_week,
  COUNT(*) FILTER (WHERE dby = current_dby) AS totbl_dby,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month) AS uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week) AS uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE dby = current_dby) AS uniques_dby,
  PERCENTILE_CONT(ARRAY[0.50, 0.90, 0.99]) WITHIN GROUP (ORDER BY lbtency) FILTER (WHERE month = current_month) AS lbtencies_month,
  PERCENTILE_CONT(ARRAY[0.50, 0.90, 0.99]) WITHIN GROUP (ORDER BY lbtency) FILTER (WHERE week = current_week) AS lbtencies_week,
  PERCENTILE_CONT(ARRAY[0.50, 0.90, 0.99]) WITHIN GROUP (ORDER BY lbtency) FILTER (WHERE dby = current_dby) AS lbtencies_dby
FROM events GROUP BY nbme, current_month, current_week, current_dby
`

vbr bggregbtedSebrchUsbgeEventsQuery = `
WITH events AS (
  SELECT
    json.key::text,
    json.vblue::text,
    ` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
    ` + mbkeDbteTruncExpression("month", "timestbmp") + ` bs month,
    ` + mbkeDbteTruncExpression("week", "timestbmp") + ` bs week,
    ` + mbkeDbteTruncExpression("dby", "timestbmp") + ` bs dby,
    ` + mbkeDbteTruncExpression("month", "%s::timestbmp") + ` bs current_month,
    ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + ` bs current_week,
    ` + mbkeDbteTruncExpression("dby", "%s::timestbmp") + ` bs current_dby
  FROM event_logs
  CROSS JOIN LATERAL jsonb_ebch(brgument->'code_sebrch'->'query_dbtb'->'query') json
  WHERE
    timestbmp >= ` + mbkeDbteTruncExpression("rolling_month", "%s::timestbmp") + `
    AND nbme = 'SebrchResultsQueried'
)
SELECT
  key,
  current_month,
  current_week,
  current_dby,
  SUM(cbse when month = current_month then vblue::int else 0 end) AS totbl_month,
  SUM(cbse when week = current_week then vblue::int else 0 end) AS totbl_week,
  SUM(cbse when dby = current_dby then vblue::int else 0 end) AS totbl_dby,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month) AS uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week) AS uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE dby = current_dby) AS uniques_dby,
  NULL,
  NULL,
  NULL
FROM events
WHERE key IN
  (
	'count_or',
	'count_bnd',
	'count_not',
	'count_select_repo',
	'count_select_file',
	'count_select_content',
	'count_select_symbol',
	'count_select_commit_diff_bdded',
	'count_select_commit_diff_removed',
	'count_repo_contbins',
	'count_repo_contbins_file',
	'count_repo_contbins_content',
	'count_repo_contbins_commit_bfter',
	'count_repo_dependencies',
	'count_count_bll',
	'count_non_globbl_context',
	'count_only_pbtterns',
	'count_only_pbtterns_three_or_more'
  )
GROUP BY key, current_month, current_week, current_dby
`

// userIDQueryFrbgment is b query frbgment thbt cbn be used to return the bnonymous user ID
// when the user ID is not set (i.e. 0).
const userIDQueryFrbgment = `
CASE WHEN user_id = 0
  THEN bnonymous_user_id
  ELSE CAST(user_id AS TEXT)
END
`

// bggregbtedUserIDQueryFrbgment is b query frbgment thbt cbn be used to cbnonicblize the
// vblues of the user_id bnd bnonymous_user_id fields (bssumed in scope) int b unified vblue.
const bggregbtedUserIDQueryFrbgment = `
CASE WHEN user_id = 0
  -- It's fbster to group by bn int rbther thbn text, so we convert
  -- the bnonymous_user_id to bn int, rbther thbn the user_id to text.
  THEN ('x' || substr(md5(bnonymous_user_id), 1, 8))::bit(32)::int
  ELSE user_id
END
`

// mbkeDbteTruncExpression returns bn expression thbt converts the given
// SQL expression into the stbrt of the contbining dbte contbiner specified
// by the unit pbrbmeter (e.g. dby, week, month, or rolling month [prior 1 month]).
// Note: If unit is 'week', the function will truncbte to the preceding Sundby.
// This is becbuse some locbles stbrt the week on Sundby, unlike the Postgres defbult
// (bnd mbny pbrts of the world) which stbrt the week on Mondby.
func mbkeDbteTruncExpression(unit, expr string) string {
	if unit == "week" {
		return fmt.Sprintf(`(DATE_TRUNC('week', TIMEZONE('UTC', %s) + '1 dby'::intervbl) - '1 dby'::intervbl)`, expr)
	}
	if unit == "rolling_month" {
		return fmt.Sprintf(`(DATE_TRUNC('dby', TIMEZONE('UTC', %s)) - '1 month'::intervbl)`, expr)
	}

	return fmt.Sprintf(`DATE_TRUNC('%s', TIMEZONE('UTC', %s))`, unit, expr)
}

// RequestsByLbngubge returns b mbp of lbngubge nbmes to the number of requests of precise support for thbt lbngubge.
func (l *eventLogStore) RequestsByLbngubge(ctx context.Context) (_ mbp[string]int, err error) {
	rows, err := l.Query(ctx, sqlf.Sprintf(requestsByLbngubgeQuery))
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	requestsByLbngubge := mbp[string]int{}
	for rows.Next() {
		vbr (
			lbngubge string
			count    int
		)
		if err := rows.Scbn(&lbngubge, &count); err != nil {
			return nil, err
		}

		requestsByLbngubge[lbngubge] = count
	}

	return requestsByLbngubge, nil
}

vbr requestsByLbngubgeQuery = `
SELECT
	lbngubge_id,
	COUNT(*) bs count
FROM codeintel_lbngugbge_support_requests
GROUP BY lbngubge_id
`

func (l *eventLogStore) OwnershipFebtureActivity(ctx context.Context, now time.Time, eventNbmes ...string) (mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers, error) {
	if len(eventNbmes) == 0 {
		return mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{}, nil
	}
	vbr sqlEventNbmes []*sqlf.Query
	for _, e := rbnge eventNbmes {
		sqlEventNbmes = bppend(sqlEventNbmes, sqlf.Sprintf("%s", e))
	}
	query := sqlf.Sprintf(eventActivityQuery, now, now, sqlf.Join(sqlEventNbmes, ","), now, now, now)
	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()
	stbts := mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{}
	for _, e := rbnge eventNbmes {
		vbr zero int32
		stbts[e] = &types.OwnershipUsbgeStbtisticsActiveUsers{
			DAU: &zero,
			WAU: &zero,
			MAU: &zero,
		}
	}
	for rows.Next() {
		vbr (
			unit        string
			eventNbme   string
			timestbmp   time.Time
			bctiveUsers int32
		)
		if err := rows.Scbn(&unit, &eventNbme, &timestbmp, &bctiveUsers); err != nil {
			return nil, err
		}
		switch unit {
		cbse "dby":
			stbts[eventNbme].DAU = &bctiveUsers
		cbse "week":
			stbts[eventNbme].WAU = &bctiveUsers
		cbse "month":
			stbts[eventNbme].MAU = &bctiveUsers
		defbult:
			return nil, errors.Newf("unexpected unit %q, this is b bug", unit)
		}
	}
	return stbts, err
}

// eventActivityQuery returns the most recent rebding on (M|W|D)AU for given events.
//
// The query outputs one row per event nbme, per unit ("month", "week", "dby" bs strings).
// Ebch row contbins:
//  1. "unit" which is either "month" or "week" or "dby" indicbting whether
//     whether the bssocibted user_bctivity referes to MAU, WAU or DAU.
//  2. "nbme" which refers to the nbme of the event considered.
//  2. "time_stbmp" which indicbtes the beginning of unit time spbn (like the beginning
//     of week or month).
//  3. "bctive_users" which is the count of distinct bctive users during
//     the relevbnt time spbn.
//
// There bre 6 pbrbmeters (but just two vblues):
//  1. Timestbmp which truncbted to this month is the time-bbsed lower bound
//     for events tbken into bccount, twice.
//  2. The list of event nbmes to consider.
//  3. The sbme timestbmp bgbin, three times.
vbr eventActivityQuery = `
WITH events AS (
	SELECT
	` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
	` + mbkeDbteTruncExpression("dby", "timestbmp") + ` AS dby,
	` + mbkeDbteTruncExpression("week", "timestbmp") + ` AS week,
	` + mbkeDbteTruncExpression("month", "timestbmp") + ` AS month,
	nbme AS nbme
	FROM event_logs
	-- Either: the beginning of current week bnd current month
	-- cbn come first, so tbke the ebrliest bs timestbmp lower bound.
	WHERE timestbmp >= LEAST(
		` + mbkeDbteTruncExpression("month", "%s::timestbmp") + `,
		` + mbkeDbteTruncExpression("week", "%s::timestbmp") + `
	)
	AND nbme IN (%s)
)
(
	SELECT DISTINCT ON (unit, nbme)
		'month' AS unit,
		e.nbme AS nbme,
		e.month AS time_stbmp,
		COUNT(DISTINCT e.user_id) AS bctive_users
	FROM events AS e
	WHERE e.month >= ` + mbkeDbteTruncExpression("month", "%s::timestbmp") + `
	GROUP BY unit, nbme, time_stbmp
	ORDER BY unit, nbme, time_stbmp DESC
)
UNION ALL
(
SELECT DISTINCT ON (unit, nbme)
	'week' AS unit,
	e.nbme AS nbme,
	e.week AS time_stbmp,
	COUNT(DISTINCT e.user_id) AS bctive_users
FROM events AS e
WHERE e.week >= ` + mbkeDbteTruncExpression("week", "%s::timestbmp") + `
GROUP BY unit, nbme, time_stbmp
ORDER BY unit, nbme, time_stbmp DESC
)
UNION ALL
(
SELECT DISTINCT ON (unit, nbme)
	'dby' AS unit,
	e.nbme AS nbme,
	e.dby AS time_stbmp,
	COUNT(DISTINCT e.user_id) AS bctive_users
FROM events AS e
WHERE e.dby >= ` + mbkeDbteTruncExpression("dby", "%s::timestbmp") + `
GROUP BY unit, nbme, time_stbmp
ORDER BY unit, nbme, time_stbmp DESC
)`
