package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	integrationSource = "CODEHOSTINTEGRATION"
)

type EventLogStore interface {
	// AggregatedCodeIntelEvents calculates CodeIntelAggregatedEvent for each unique event type related to code intel.
	AggregatedCodeIntelEvents(ctx context.Context) ([]types.CodeIntelAggregatedEvent, error)

	// AggregatedCodeIntelInvestigationEvents calculates CodeIntelAggregatedInvestigationEvent for each unique investigation type.
	AggregatedCodeIntelInvestigationEvents(ctx context.Context) ([]types.CodeIntelAggregatedInvestigationEvent, error)

	// AggregatedSearchEvents calculates SearchAggregatedEvent for each every unique event type related to search.
	AggregatedSearchEvents(ctx context.Context, now time.Time) ([]types.SearchAggregatedEvent, error)

	BulkInsert(ctx context.Context, events []*Event) error

	// CodeIntelligenceCrossRepositoryWAUs returns the WAU (current week) with any (precise or search-based) cross-repository code intelligence event.
	CodeIntelligenceCrossRepositoryWAUs(ctx context.Context) (int, error)

	// CodeIntelligencePreciseCrossRepositoryWAUs returns the WAU (current week) with precise-based cross-repository code intelligence events.
	CodeIntelligencePreciseCrossRepositoryWAUs(ctx context.Context) (int, error)

	// CodeIntelligencePreciseWAUs returns the WAU (current week) with precise-based code intelligence events.
	CodeIntelligencePreciseWAUs(ctx context.Context) (int, error)

	// CodeIntelligenceRepositoryCounts returns the counts of repositories with code intelligence
	// properties (number of repositories with intel, with automatic/manual index configuration, etc).
	CodeIntelligenceRepositoryCounts(ctx context.Context) (counts CodeIntelligenceRepositoryCounts, err error)

	// CodeIntelligenceRepositoryCountsByLanguage returns the counts of repositories with code intelligence
	// properties (number of repositories with intel, with automatic/manual index configuration, etc), grouped
	// by language.
	CodeIntelligenceRepositoryCountsByLanguage(ctx context.Context) (_ map[string]CodeIntelligenceRepositoryCountsForLanguage, err error)

	// CodeIntelligenceSearchBasedCrossRepositoryWAUs returns the WAU (current week) with searched-base cross-repository code intelligence events.
	CodeIntelligenceSearchBasedCrossRepositoryWAUs(ctx context.Context) (int, error)

	// CodeIntelligenceSearchBasedWAUs returns the WAU (current week) with searched-base code intelligence events.
	CodeIntelligenceSearchBasedWAUs(ctx context.Context) (int, error)

	// CodeIntelligenceSettingsPageViewCount returns the number of view of pages related code intelligence
	// administration (upload, index records, index configuration, etc) in the past week.
	CodeIntelligenceSettingsPageViewCount(ctx context.Context) (int, error)

	// RequestsByLanguage returns a map of language names to the number of requests of precise support for that language.
	RequestsByLanguage(ctx context.Context) (map[string]int, error)

	// CodeIntelligenceWAUs returns the WAU (current week) with any (precise or search-based) code intelligence event.
	CodeIntelligenceWAUs(ctx context.Context) (int, error)

	// CountByUserID gets a count of events logged by a given user.
	CountByUserID(ctx context.Context, userID int32) (int, error)

	// CountByUserIDAndEventName gets a count of events logged by a given user and with a given event name.
	CountByUserIDAndEventName(ctx context.Context, userID int32, name string) (int, error)

	// CountByUserIDAndEventNamePrefix gets a count of events logged by a given user and with a given event name prefix.
	CountByUserIDAndEventNamePrefix(ctx context.Context, userID int32, namePrefix string) (int, error)

	// CountByUserIDAndEventNames gets a count of events logged by a given user that match a list of given event names.
	CountByUserIDAndEventNames(ctx context.Context, userID int32, names []string) (int, error)

	// CountUniqueUsersAll provides a count of unique active users in a given time span.
	CountUniqueUsersAll(ctx context.Context, startDate, endDate time.Time, opt *CountUniqueUsersOptions) (int, error)

	// CountUniqueUsersByEventName provides a count of unique active users in a given time span that logged a given event.
	CountUniqueUsersByEventName(ctx context.Context, startDate, endDate time.Time, name string) (int, error)

	// CountUniqueUsersByEventNamePrefix provides a count of unique active users in a given time span that logged an event with a given prefix.
	CountUniqueUsersByEventNamePrefix(ctx context.Context, startDate, endDate time.Time, namePrefix string) (int, error)

	// CountUniqueUsersByEventNames provides a count of unique active users in a given time span that logged any event that matches a list of given event names
	CountUniqueUsersByEventNames(ctx context.Context, startDate, endDate time.Time, names []string) (int, error)

	// SiteUsageMultiplePeriods provides a count of unique active users in given time spans, broken up into periods of
	// a given type. The value of `now` should be the current time in UTC.
	SiteUsageMultiplePeriods(ctx context.Context, now time.Time, dayPeriods int, weekPeriods int, monthPeriods int, opt *CountUniqueUsersOptions) (*types.SiteUsageStatistics, error)

	// CountUsersWithSetting returns the number of users wtih the given temporary setting set to the given value.
	CountUsersWithSetting(ctx context.Context, setting string, value any) (int, error)

	Insert(ctx context.Context, e *Event) error

	// LatestPing returns the most recently recorded ping event.
	LatestPing(ctx context.Context) (*Event, error)

	// ListAll gets all event logs in descending order of timestamp.
	ListAll(ctx context.Context, opt EventLogsListOptions) ([]*Event, error)

	// ListExportableEvents gets all event logs that are allowed to be exported.
	ListExportableEvents(ctx context.Context, after, limit int) ([]*Event, error)

	ListUniqueUsersAll(ctx context.Context, startDate, endDate time.Time) ([]int32, error)

	// MaxTimestampByUserID gets the max timestamp among event logs for a given user.
	MaxTimestampByUserID(ctx context.Context, userID int32) (*time.Time, error)

	// MaxTimestampByUserIDAndSource gets the max timestamp among event logs for a given user and event source.
	MaxTimestampByUserIDAndSource(ctx context.Context, userID int32, source string) (*time.Time, error)

	SiteUsageCurrentPeriods(ctx context.Context) (types.SiteUsageSummary, error)

	// UsersUsageCounts returns a list of UserUsageCounts for all active users that produced 'SearchResultsQueried' and any
	// '%codeintel%' events in the event_logs table.
	UsersUsageCounts(ctx context.Context) (counts []types.UserUsageCounts, err error)

	Transact(ctx context.Context) (EventLogStore, error)
	Done(error) error
	With(other basestore.ShareableStore) EventLogStore
	basestore.ShareableStore
}

type eventLogStore struct {
	*basestore.Store
}

// EventLogsWith instantiates and returns a new EventLogStore using the other store handle.
func EventLogsWith(other basestore.ShareableStore) EventLogStore {
	return &eventLogStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (l *eventLogStore) With(other basestore.ShareableStore) EventLogStore {
	return &eventLogStore{Store: l.Store.With(other)}
}

func (l *eventLogStore) Transact(ctx context.Context) (EventLogStore, error) {
	txBase, err := l.Store.Transact(ctx)
	return &eventLogStore{Store: txBase}, err
}

// SanitizeEventURL makes the given URL is using HTTP/HTTPS scheme and within
// the current site determined by `conf.ExternalURL()`.
func SanitizeEventURL(raw string) string {
	if raw == "" {
		return ""
	}

	// Check if the URL looks like a real URL
	u, err := url.Parse(raw)
	if err != nil ||
		(u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}

	// Check if the URL belongs to the current site
	normalized := u.String()
	if strings.HasPrefix(normalized, conf.ExternalURL()) || strings.HasSuffix(u.Host, "sourcegraph.com") {
		return normalized
	}
	return ""
}

// Event contains information needed for logging an event.
type Event struct {
	ID               int32
	Name             string
	URL              string
	UserID           uint32
	AnonymousUserID  string
	Argument         json.RawMessage
	PublicArgument   json.RawMessage
	Source           string
	Version          string
	Timestamp        time.Time
	EvaluatedFlagSet featureflag.EvaluatedFlagSet
	CohortID         *string // date in YYYY-MM-DD format
	FirstSourceURL   *string
	LastSourceURL    *string
	Referrer         *string
	DeviceID         *string
	InsertID         *string
}

func (l *eventLogStore) Insert(ctx context.Context, e *Event) error {
	return l.BulkInsert(ctx, []*Event{e})
}

func (l *eventLogStore) BulkInsert(ctx context.Context, events []*Event) error {
	coalesce := func(v json.RawMessage) json.RawMessage {
		if v != nil {
			return v
		}

		return json.RawMessage(`{}`)
	}

	ensureUuid := func(in *string) string {
		if in == nil || len(*in) == 0 {
			u, _ := uuid.NewV4()
			return u.String()
		}
		return *in
	}

	rowValues := make(chan []any, len(events))
	for _, event := range events {
		featureFlags, err := json.Marshal(event.EvaluatedFlagSet)
		if err != nil {
			return err
		}

		rowValues <- []any{
			event.Name,
			// 🚨 SECURITY: It is important to sanitize event URL before
			// being stored to the database to help guarantee no malicious
			// data at rest.
			SanitizeEventURL(event.URL),
			event.UserID,
			event.AnonymousUserID,
			event.Source,
			coalesce(event.Argument),
			coalesce(event.PublicArgument),
			version.Version(),
			event.Timestamp.UTC(),
			featureFlags,
			event.CohortID,
			event.FirstSourceURL,
			event.LastSourceURL,
			event.Referrer,
			ensureUuid(event.DeviceID),
			ensureUuid(event.InsertID),
		}
	}
	close(rowValues)

	return batch.InsertValues(
		ctx,
		l.Handle(),
		"event_logs",
		batch.MaxNumPostgresParameters,
		[]string{
			"name",
			"url",
			"user_id",
			"anonymous_user_id",
			"source",
			"argument",
			"public_argument",
			"version",
			"timestamp",
			"feature_flags",
			"cohort_id",
			"first_source_url",
			"last_source_url",
			"referrer",
			"device_id",
			"insert_id",
		},
		rowValues,
	)
}

func (l *eventLogStore) getBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*Event, error) {
	q := sqlf.Sprintf("SELECT id, name, url, user_id, anonymous_user_id, source, argument, version, timestamp, feature_flags, cohort_id, first_source_url, last_source_url, referrer, device_id, insert_id FROM event_logs %s", querySuffix)
	rows, err := l.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []*Event{}
	for rows.Next() {
		r := Event{}
		var rawFlags []byte
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.UserID, &r.AnonymousUserID, &r.Source, &r.Argument, &r.Version, &r.Timestamp, &rawFlags, &r.CohortID, &r.FirstSourceURL, &r.LastSourceURL, &r.Referrer, &r.DeviceID, &r.InsertID)
		if err != nil {
			return nil, err
		}
		if rawFlags != nil {
			marshalErr := json.Unmarshal(rawFlags, &r.EvaluatedFlagSet)
			if marshalErr != nil {
				return nil, errors.Wrap(marshalErr, "json.Unmarshal")
			}
		}
		events = append(events, &r)
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

	EventName *string
}

func (l *eventLogStore) ListAll(ctx context.Context, opt EventLogsListOptions) ([]*Event, error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_id = %d", opt.UserID))
	}
	if opt.EventName != nil {
		conds = append(conds, sqlf.Sprintf("name = %s", opt.EventName))
	}
	return l.getBySQL(ctx, sqlf.Sprintf("WHERE %s ORDER BY timestamp DESC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL()))
}

func (l *eventLogStore) ListExportableEvents(ctx context.Context, after, limit int) ([]*Event, error) {
	suffix := "WHERE event_logs.id > %d AND name IN (SELECT event_name FROM event_logs_export_allowlist) ORDER BY event_logs.id LIMIT %d"
	return l.getBySQL(ctx, sqlf.Sprintf(suffix, after, limit))
}

func (l *eventLogStore) LatestPing(ctx context.Context) (*Event, error) {
	rows, err := l.getBySQL(ctx, sqlf.Sprintf(`WHERE name='ping' ORDER BY id DESC LIMIT 1`))
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

func (l *eventLogStore) CountByUserIDAndEventName(ctx context.Context, userID int32, name string) (int, error) {
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND name = %s", userID, name))
}

func (l *eventLogStore) CountByUserIDAndEventNamePrefix(ctx context.Context, userID int32, namePrefix string) (int, error) {
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND name LIKE %s", userID, namePrefix+"%"))
}

func (l *eventLogStore) CountByUserIDAndEventNames(ctx context.Context, userID int32, names []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := range names {
		items = append(items, sqlf.Sprintf("%s", v))
	}
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND name IN (%s)", userID, sqlf.Join(items, ",")))
}

// countBySQL gets a count of event logs.
func (l *eventLogStore) countBySQL(ctx context.Context, querySuffix *sqlf.Query) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM event_logs %s", querySuffix)
	r := l.QueryRow(ctx, q)
	var count int
	err := r.Scan(&count)
	return count, err
}

func (l *eventLogStore) MaxTimestampByUserID(ctx context.Context, userID int32) (*time.Time, error) {
	return l.maxTimestampBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d", userID))
}

func (l *eventLogStore) MaxTimestampByUserIDAndSource(ctx context.Context, userID int32, source string) (*time.Time, error) {
	return l.maxTimestampBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND source = %s", userID, source))
}

// maxTimestampBySQL gets the max timestamp among event logs.
func (l *eventLogStore) maxTimestampBySQL(ctx context.Context, querySuffix *sqlf.Query) (*time.Time, error) {
	q := sqlf.Sprintf("SELECT MAX(timestamp) FROM event_logs %s", querySuffix)
	r := l.QueryRow(ctx, q)

	var t time.Time
	err := r.Scan(&dbutil.NullTime{Time: &t})
	if t.IsZero() {
		return nil, err
	}
	return &t, err
}

// SiteUsageValues is a set of UsageValues representing usage on daily, weekly, and monthly bases.
type SiteUsageValues struct {
	DAUs []UsageValue
	WAUs []UsageValue
	MAUs []UsageValue
}

// UsageValue is a single count of usage for a time period starting on a given date.
type UsageValue struct {
	Start           time.Time
	Type            PeriodType
	Count           int
	CountRegistered int
}

// PeriodType is the type of period in which to count events and unique users.
type PeriodType string

const (
	// Daily is used to get a count of events or unique users within a day.
	Daily PeriodType = "daily"
	// Weekly is used to get a count of events or unique users within a week.
	Weekly PeriodType = "weekly"
	// Monthly is used to get a count of events or unique users within a month.
	Monthly PeriodType = "monthly"
)

// intervalByPeriodType is a map of generate_series step values by period type.
var intervalByPeriodType = map[PeriodType]*sqlf.Query{
	Daily:   sqlf.Sprintf("'1 day'"),
	Weekly:  sqlf.Sprintf("'1 week'"),
	Monthly: sqlf.Sprintf("'1 month'"),
}

// periodByPeriodType is a map of SQL fragments that produce a timestamp bucket by period
// type. This assumes the existence of a  field named `timestamp` in the enclosing query.
var periodByPeriodType = map[PeriodType]*sqlf.Query{
	Daily:   sqlf.Sprintf(makeDateTruncExpression("day", "timestamp")),
	Weekly:  sqlf.Sprintf(makeDateTruncExpression("week", "timestamp")),
	Monthly: sqlf.Sprintf(makeDateTruncExpression("month", "timestamp")),
}

// calcStartDate calculates the the starting date of a number of periods given the period type.
// from the current time supplied as `now`. Returns a second false value if the period type is
// illegal.
func calcStartDate(now time.Time, periodType PeriodType, periods int) (time.Time, bool) {
	periodsAgo := periods - 1

	switch periodType {
	case Daily:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -periodsAgo), true
	case Weekly:
		return timeutil.StartOfWeek(now, periodsAgo), true
	case Monthly:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -periodsAgo, 0), true
	}
	return time.Time{}, false
}

// calcEndDate calculates the the ending date of a number of periods given the period type.
// Returns a second false value if the period type is illegal.
func calcEndDate(startDate time.Time, periodType PeriodType, periods int) (time.Time, bool) {
	periodsAgo := periods - 1

	switch periodType {
	case Daily:
		return startDate.AddDate(0, 0, periodsAgo), true
	case Weekly:
		return startDate.AddDate(0, 0, 7*periodsAgo), true
	case Monthly:
		return startDate.AddDate(0, periodsAgo, 0), true
	}

	return time.Time{}, false
}

// CountUniqueUsersOptions provides options for counting unique users.
type CountUniqueUsersOptions struct {
	// If set, adds additional restrictions on the event types.
	EventFilters *EventFilterOptions
	// If set, excludes backend system users
	ExcludeSystemUsers bool
	// If set, excludes events that don't meet the criteria of "active" usage of Sourcegraph.
	// These are mostly actions taken by signed-out users.
	ExcludeNonActiveUsers bool
}

// EventFilterOptions provides options for filtering events.
type EventFilterOptions struct {
	// If set, only include events with a given prefix.
	ByEventNamePrefix string
	// If set, only include events with the given name.
	ByEventName string
	// If not empty, only include events that matche a list of given event names
	ByEventNames []string
	// Must be used with ByEventName
	//
	// If set, only include events that match a specified condition.
	ByEventNameWithCondition *sqlf.Query
}

// EventArgumentMatch provides the options for matching an event with
// a specific JSON value passed as an argument.
type EventArgumentMatch struct {
	// The name of the JSON key to match against.
	ArgumentName string
	// The actual value passed to the JSON key to match.
	ArgumentValue string
}

// PercentileValue is a slice of Nth percentile values calculated from a field of events
// in a time period starting on a given date.
type PercentileValue struct {
	Start  time.Time
	Values []float64
}

func (l *eventLogStore) CountUsersWithSetting(ctx context.Context, setting string, value any) (int, error) {
	count, _, err := basestore.ScanFirstInt(l.Store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM temporary_settings WHERE %s <@ contents`, jsonSettingFragment(setting, value))))
	return count, err
}

func jsonSettingFragment(setting string, value any) string {
	raw, _ := json.Marshal(map[string]any{setting: value})
	return string(raw)
}

func createCountUniqueUserConds(opt *CountUniqueUsersOptions) []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		if opt.ExcludeSystemUsers {
			conds = append(conds, sqlf.Sprintf("user_id > 0 OR anonymous_user_id <> 'backend'"))
		}
		if opt.ExcludeNonActiveUsers {
			conds = append(conds, sqlf.Sprintf("name NOT IN ('"+strings.Join(eventlogger.NonActiveUserEvents, "','")+"')"))
		}
		if opt.EventFilters != nil {
			if opt.EventFilters.ByEventNamePrefix != "" {
				conds = append(conds, sqlf.Sprintf("name LIKE %s", opt.EventFilters.ByEventNamePrefix+"%"))
			}
			if opt.EventFilters.ByEventName != "" {
				conds = append(conds, sqlf.Sprintf("name = %s", opt.EventFilters.ByEventName))
			}
			if opt.EventFilters.ByEventNameWithCondition != nil {
				conds = append(conds, opt.EventFilters.ByEventNameWithCondition)
			}
			if len(opt.EventFilters.ByEventNames) > 0 {
				items := []*sqlf.Query{}
				for _, v := range opt.EventFilters.ByEventNames {
					items = append(items, sqlf.Sprintf("%s", v))
				}
				conds = append(conds, sqlf.Sprintf("name IN (%s)", sqlf.Join(items, ",")))
			}
		}
	}
	return conds
}

func (l *eventLogStore) SiteUsageMultiplePeriods(ctx context.Context, now time.Time, dayPeriods int, weekPeriods int, monthPeriods int, opt *CountUniqueUsersOptions) (*types.SiteUsageStatistics, error) {
	startDateDays, _ := calcStartDate(now, Daily, dayPeriods)
	endDateDays, _ := calcEndDate(startDateDays, Daily, dayPeriods)
	startDateWeeks, _ := calcStartDate(now, Weekly, weekPeriods)
	endDateWeeks, _ := calcEndDate(startDateWeeks, Weekly, weekPeriods)
	startDateMonths, _ := calcStartDate(now, Monthly, monthPeriods)
	endDateMonths, _ := calcEndDate(startDateMonths, Monthly, monthPeriods)

	conds := createCountUniqueUserConds(opt)

	return l.siteUsageMultiplePeriodsBySQL(ctx, startDateDays, endDateDays, startDateWeeks, endDateWeeks, startDateMonths, endDateMonths, conds)
}

func (l *eventLogStore) siteUsageMultiplePeriodsBySQL(ctx context.Context, startDateDays, endDateDays, startDateWeeks, endDateWeeks, startDateMonths, endDateMonths time.Time, conds []*sqlf.Query) (*types.SiteUsageStatistics, error) {
	q := sqlf.Sprintf(siteUsageMultiplePeriodsQuery, startDateDays, endDateDays, startDateWeeks, endDateWeeks, startDateMonths, endDateMonths, sqlf.Join(conds, ") AND ("))

	rows, err := l.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dauCounts := []*types.SiteActivityPeriod{}
	wauCounts := []*types.SiteActivityPeriod{}
	mauCounts := []*types.SiteActivityPeriod{}
	for rows.Next() {
		var v UsageValue
		err := rows.Scan(&v.Start, &v.Type, &v.Count, &v.CountRegistered)
		if err != nil {
			return nil, err
		}
		v.Start = v.Start.UTC()
		if v.Type == "day" {
			dauCounts = append(dauCounts, &types.SiteActivityPeriod{
				StartTime:           v.Start,
				UserCount:           int32(v.Count),
				RegisteredUserCount: int32(v.CountRegistered),
				AnonymousUserCount:  int32(v.Count - v.CountRegistered),
				// No longer used in site admin usage stats views. Use GetSiteUsageStats if you need this instead.
				IntegrationUserCount: 0,
			})
		}
		if v.Type == "week" {
			wauCounts = append(wauCounts, &types.SiteActivityPeriod{
				StartTime:           v.Start,
				UserCount:           int32(v.Count),
				RegisteredUserCount: int32(v.CountRegistered),
				AnonymousUserCount:  int32(v.Count - v.CountRegistered),
				// No longer used in site admin usage stats views. Use GetSiteUsageStats if you need this instead.
				IntegrationUserCount: 0,
			})
		}
		if v.Type == "month" {
			mauCounts = append(mauCounts, &types.SiteActivityPeriod{
				StartTime:           v.Start,
				UserCount:           int32(v.Count),
				RegisteredUserCount: int32(v.CountRegistered),
				AnonymousUserCount:  int32(v.Count - v.CountRegistered),
				// No longer used in site admin usage stats views. Use GetSiteUsageStats if you need this instead.
				IntegrationUserCount: 0,
			})
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &types.SiteUsageStatistics{
		DAUs: dauCounts,
		WAUs: wauCounts,
		MAUs: mauCounts,
	}, nil
}

var siteUsageMultiplePeriodsQuery = `
WITH all_periods AS (
  SELECT generate_series((%s)::timestamp, (%s)::timestamp, ('1 day')::interval)  AS period, 'day' AS type
  UNION ALL
  SELECT generate_series((%s)::timestamp, (%s)::timestamp, ('1 week')::interval) AS period, 'week' AS type
  UNION ALL
  SELECT generate_series((%s)::timestamp, (%s)::timestamp, ('1 month')::interval) AS period, 'month' AS type),
unique_users_by_dwm AS (
  SELECT
    ` + makeDateTruncExpression("day", "timestamp") + ` AS day_period,
	` + makeDateTruncExpression("week", "timestamp") + ` AS week_period,
	` + makeDateTruncExpression("month", "timestamp") + ` AS month_period,
	user_id > 0 AS registered,
	` + aggregatedUserIDQueryFragment + ` as aggregated_user_id
  FROM event_logs
  WHERE (%s)
  GROUP BY day_period, week_period, month_period, aggregated_user_id, registered
),
unique_users_by_day AS (
  SELECT
	day_period,
	COUNT(DISTINCT aggregated_user_id) as count,
	COUNT(DISTINCT aggregated_user_id) FILTER (WHERE registered) as count_registered
  FROM unique_users_by_dwm
  GROUP BY day_period
),
unique_users_by_week AS (
  SELECT
	week_period,
	COUNT(DISTINCT aggregated_user_id) as count,
	COUNT(DISTINCT aggregated_user_id) FILTER (WHERE registered) as count_registered
  FROM unique_users_by_dwm
  GROUP BY week_period
),
unique_users_by_month AS (
  SELECT
    month_period,
    COUNT(DISTINCT aggregated_user_id) as count,
    COUNT(DISTINCT aggregated_user_id) FILTER (WHERE registered) as count_registered
  FROM unique_users_by_dwm
  GROUP BY month_period
)
SELECT
  all_periods.period,
  all_periods.type,
  COALESCE(CASE WHEN all_periods.type = 'day'
    THEN unique_users_by_day.count
	ELSE CASE WHEN all_periods.type = 'week'
      THEN unique_users_by_week.count
      ELSE unique_users_by_month.count
    END
  END, 0) count,
  COALESCE(CASE WHEN all_periods.type = 'day'
    THEN unique_users_by_day.count_registered
    ELSE CASE WHEN all_periods.type = 'week'
      THEN unique_users_by_week.count_registered
      ELSE unique_users_by_month.count_registered
	END
  END, 0) count_registered
FROM all_periods
LEFT OUTER JOIN unique_users_by_day ON all_periods.type = 'day' AND all_periods.period = (unique_users_by_day.day_period)::timestamp
LEFT OUTER JOIN unique_users_by_week ON all_periods.type = 'week' AND all_periods.period = (unique_users_by_week.week_period)::timestamp
LEFT OUTER JOIN unique_users_by_month ON all_periods.type = 'month' AND all_periods.period = (unique_users_by_month.month_period)::timestamp
ORDER BY period DESC
`

func (l *eventLogStore) CountUniqueUsersAll(ctx context.Context, startDate, endDate time.Time, opt *CountUniqueUsersOptions) (int, error) {
	conds := createCountUniqueUserConds(opt)

	return l.countUniqueUsersBySQL(ctx, startDate, endDate, conds)
}

func (l *eventLogStore) CountUniqueUsersByEventNamePrefix(ctx context.Context, startDate, endDate time.Time, namePrefix string) (int, error) {
	return l.countUniqueUsersBySQL(ctx, startDate, endDate, []*sqlf.Query{sqlf.Sprintf("name LIKE %s ", namePrefix+"%")})
}

func (l *eventLogStore) CountUniqueUsersByEventName(ctx context.Context, startDate, endDate time.Time, name string) (int, error) {
	return l.countUniqueUsersBySQL(ctx, startDate, endDate, []*sqlf.Query{sqlf.Sprintf("name = %s", name)})
}

func (l *eventLogStore) CountUniqueUsersByEventNames(ctx context.Context, startDate, endDate time.Time, names []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := range names {
		items = append(items, sqlf.Sprintf("%s", v))
	}
	return l.countUniqueUsersBySQL(ctx, startDate, endDate, []*sqlf.Query{sqlf.Sprintf("name IN (%s)", sqlf.Join(items, ","))})
}

func (l *eventLogStore) countUniqueUsersBySQL(ctx context.Context, startDate, endDate time.Time, conds []*sqlf.Query) (int, error) {
	if len(conds) == 0 {
		conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	}
	q := sqlf.Sprintf(`SELECT COUNT(DISTINCT `+userIDQueryFragment+`)
		FROM event_logs
		WHERE (DATE(TIMEZONE('UTC'::text, timestamp)) >= %s) AND (DATE(TIMEZONE('UTC'::text, timestamp)) <= %s) AND (%s)`, startDate, endDate, sqlf.Join(conds, ") AND ("))
	r := l.QueryRow(ctx, q)
	var count int
	err := r.Scan(&count)
	return count, err
}

func (l *eventLogStore) ListUniqueUsersAll(ctx context.Context, startDate, endDate time.Time) ([]int32, error) {
	rows, err := l.Handle().QueryContext(ctx, `SELECT user_id
		FROM event_logs
		WHERE user_id > 0 AND DATE(TIMEZONE('UTC'::text, timestamp)) >= $1 AND DATE(TIMEZONE('UTC'::text, timestamp)) <= $2
		GROUP BY user_id`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	var users []int32
	defer rows.Close()
	for rows.Next() {
		var userID int32
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (l *eventLogStore) UsersUsageCounts(ctx context.Context) (counts []types.UserUsageCounts, err error) {
	rows, err := l.Handle().QueryContext(ctx, usersUsageCountsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c types.UserUsageCounts

		err := rows.Scan(
			&c.Date,
			&c.UserID,
			&dbutil.NullInt32{N: &c.SearchCount},
			&dbutil.NullInt32{N: &c.CodeIntelCount},
		)

		if err != nil {
			return nil, err
		}

		counts = append(counts, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

const usersUsageCountsQuery = `
-- source: internal/database/event_logs.go:UsersUsageCounts
SELECT
  DATE(timestamp),
  user_id,
  COUNT(*) FILTER (WHERE event_logs.name ='SearchResultsQueried') as search_count,
  COUNT(*) FILTER (WHERE event_logs.name LIKE '%codeintel%') as codeintel_count
FROM event_logs
GROUP BY 1, 2
ORDER BY 1 DESC, 2 ASC;
`

// SiteUsageOptions specifies the options for Site Usage calculations.
type SiteUsageOptions struct {
	// Exclude backend system users.
	ExcludeSystemUsers bool
	// Exclude events that don't meet the criteria of "active" usage of Sourcegraph. These are mostly actions taken by signed-out users.
	ExcludeNonActiveUsers bool
}

func (l *eventLogStore) SiteUsageCurrentPeriods(ctx context.Context) (types.SiteUsageSummary, error) {
	return l.siteUsageCurrentPeriods(ctx, time.Now().UTC(), &SiteUsageOptions{
		ExcludeSystemUsers:    true,
		ExcludeNonActiveUsers: true,
	})
}

func (l *eventLogStore) siteUsageCurrentPeriods(ctx context.Context, now time.Time, opt *SiteUsageOptions) (summary types.SiteUsageSummary, err error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		if opt.ExcludeSystemUsers {
			conds = append(conds, sqlf.Sprintf("user_id > 0 OR anonymous_user_id <> 'backend'"))
		}
		if opt.ExcludeNonActiveUsers {
			conds = append(conds, sqlf.Sprintf("name NOT IN ('"+strings.Join(eventlogger.NonActiveUserEvents, "','")+"')"))
		}
	}

	query := sqlf.Sprintf(siteUsageCurrentPeriodsQuery, now, now, now, now, sqlf.Join(conds, ") AND ("))

	err = l.QueryRow(ctx, query).Scan(
		&summary.Month,
		&summary.Week,
		&summary.Day,
		&summary.UniquesMonth,
		&summary.UniquesWeek,
		&summary.UniquesDay,
		&summary.RegisteredUniquesMonth,
		&summary.RegisteredUniquesWeek,
		&summary.RegisteredUniquesDay,
		&summary.IntegrationUniquesMonth,
		&summary.IntegrationUniquesWeek,
		&summary.IntegrationUniquesDay,
	)

	return summary, err
}

var siteUsageCurrentPeriodsQuery = `
SELECT
  current_month,
  current_week,
  current_day,

  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month) AS uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week) AS uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE day = current_day) AS uniques_day,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month AND registered) AS registered_uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week AND registered) AS registered_uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE day = current_day AND registered) AS registered_uniques_day,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month AND source = 'CODEHOSTINTEGRATION')
  	AS integration_uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week AND source = 'CODEHOSTINTEGRATION')
  	AS integration_uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE day = current_day AND source = 'CODEHOSTINTEGRATION')
  	AS integration_uniques_day
FROM (
  -- This sub-query is here to avoid re-doing this work above on each aggregation.
  SELECT
    name,
    user_id != 0 as registered,
    ` + aggregatedUserIDQueryFragment + ` AS user_id,
    source,
    ` + makeDateTruncExpression("month", "timestamp") + ` as month,
    ` + makeDateTruncExpression("week", "timestamp") + ` as week,
    ` + makeDateTruncExpression("day", "timestamp") + ` as day,
    ` + makeDateTruncExpression("month", "%s::timestamp") + ` as current_month,
    ` + makeDateTruncExpression("week", "%s::timestamp") + ` as current_week,
    ` + makeDateTruncExpression("day", "%s::timestamp") + ` as current_day
  FROM event_logs
  WHERE (timestamp >= ` + makeDateTruncExpression("month", "%s::timestamp") + `) AND (%s)
) events

GROUP BY current_month, current_week, current_day
`

func (l *eventLogStore) CodeIntelligencePreciseWAUs(ctx context.Context) (int, error) {
	eventNames := []string{
		"codeintel.lsifHover",
		"codeintel.lsifDefinitions",
		"codeintel.lsifReferences",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNames, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceSearchBasedWAUs(ctx context.Context) (int, error) {
	eventNames := []string{
		"codeintel.searchHover",
		"codeintel.searchDefinitions",
		"codeintel.searchReferences",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNames, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceWAUs(ctx context.Context) (int, error) {
	eventNames := []string{
		"codeintel.lsifHover",
		"codeintel.lsifDefinitions",
		"codeintel.lsifReferences",
		"codeintel.searchHover",
		"codeintel.searchDefinitions",
		"codeintel.searchReferences",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNames, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceCrossRepositoryWAUs(ctx context.Context) (int, error) {
	eventNames := []string{
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences.xrepo",
		"codeintel.searchDefinitions.xrepo",
		"codeintel.searchReferences.xrepo",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNames, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligencePreciseCrossRepositoryWAUs(ctx context.Context) (int, error) {
	eventNames := []string{
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences.xrepo",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNames, time.Now().UTC())
}

func (l *eventLogStore) CodeIntelligenceSearchBasedCrossRepositoryWAUs(ctx context.Context) (int, error) {
	eventNames := []string{
		"codeintel.searchDefinitions.xrepo",
		"codeintel.searchReferences.xrepo",
	}

	return l.codeIntelligenceWeeklyUsersCount(ctx, eventNames, time.Now().UTC())
}

func (l *eventLogStore) codeIntelligenceWeeklyUsersCount(ctx context.Context, eventNames []string, now time.Time) (wau int, _ error) {
	var names []*sqlf.Query
	for _, name := range eventNames {
		names = append(names, sqlf.Sprintf("%s", name))
	}

	if err := l.QueryRow(ctx, sqlf.Sprintf(codeIntelWeeklyUsersQuery, now, sqlf.Join(names, ", "))).Scan(&wau); err != nil {
		return 0, err
	}

	return wau, nil
}

var codeIntelWeeklyUsersQuery = `
-- source: internal/database/event_logs.go:codeIntelligenceWeeklyUsersCount
SELECT COUNT(DISTINCT ` + userIDQueryFragment + `)
FROM event_logs
WHERE
  timestamp >= ` + makeDateTruncExpression("week", "%s::timestamp") + `
  AND name IN (%s);
`

type CodeIntelligenceRepositoryCounts struct {
	NumRepositories                                  int
	NumRepositoriesWithUploadRecords                 int
	NumRepositoriesWithFreshUploadRecords            int
	NumRepositoriesWithIndexRecords                  int
	NumRepositoriesWithFreshIndexRecords             int
	NumRepositoriesWithAutoIndexConfigurationRecords int
}

func (l *eventLogStore) CodeIntelligenceRepositoryCounts(ctx context.Context) (counts CodeIntelligenceRepositoryCounts, err error) {
	rows, err := l.Query(ctx, sqlf.Sprintf(codeIntelligenceRepositoryCountsQuery))
	if err != nil {
		return CodeIntelligenceRepositoryCounts{}, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&counts.NumRepositories,
			&counts.NumRepositoriesWithUploadRecords,
			&counts.NumRepositoriesWithFreshUploadRecords,
			&counts.NumRepositoriesWithIndexRecords,
			&counts.NumRepositoriesWithFreshIndexRecords,
			&counts.NumRepositoriesWithAutoIndexConfigurationRecords,
		); err != nil {
			return CodeIntelligenceRepositoryCounts{}, err
		}
	}
	if err := rows.Err(); err != nil {
		return CodeIntelligenceRepositoryCounts{}, err
	}

	return counts, nil
}

var codeIntelligenceRepositoryCountsQuery = `
-- source: internal/database/event_logs.go:CodeIntelligenceRepositoryCounts
SELECT
	(SELECT COUNT(*) FROM repo r WHERE r.deleted_at IS NULL)
		AS num_repositories,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_dumps_with_repository_name u)
		AS num_repositories_with_upload_records,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_dumps_with_repository_name u WHERE u.uploaded_at >= NOW() - '168 hours'::interval)
		AS num_repositories_with_fresh_upload_records,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_indexes_with_repository_name u WHERE u.state = 'completed')
		AS num_repositories_with_index_records,
	(SELECT COUNT(DISTINCT u.repository_id) FROM lsif_indexes_with_repository_name u WHERE u.state = 'completed' AND u.queued_at >= NOW() - '168 hours'::interval)
		AS num_repositories_with_fresh_index_records,
	(SELECT COUNT(DISTINCT uc.repository_id) FROM lsif_index_configuration uc WHERE uc.autoindex_enabled IS TRUE AND data IS NOT NULL)
		AS num_repositories_with_index_configuration_records
`

type CodeIntelligenceRepositoryCountsForLanguage struct {
	NumRepositoriesWithUploadRecords      int
	NumRepositoriesWithFreshUploadRecords int
	NumRepositoriesWithIndexRecords       int
	NumRepositoriesWithFreshIndexRecords  int
}

func (l *eventLogStore) CodeIntelligenceRepositoryCountsByLanguage(ctx context.Context) (_ map[string]CodeIntelligenceRepositoryCountsForLanguage, err error) {
	rows, err := l.Query(ctx, sqlf.Sprintf(codeIntelligenceRepositoryCountsByLanguageQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		language string
		numRepositoriesWithUploadRecords,
		numRepositoriesWithFreshUploadRecords,
		numRepositoriesWithIndexRecords,
		numRepositoriesWithFreshIndexRecords *int
	)

	byLanguage := map[string]CodeIntelligenceRepositoryCountsForLanguage{}
	for rows.Next() {
		if err := rows.Scan(
			&language,
			&numRepositoriesWithUploadRecords,
			&numRepositoriesWithFreshUploadRecords,
			&numRepositoriesWithIndexRecords,
			&numRepositoriesWithFreshIndexRecords,
		); err != nil {
			return nil, err
		}

		byLanguage[language] = CodeIntelligenceRepositoryCountsForLanguage{
			NumRepositoriesWithUploadRecords:      safeDerefIntPtr(numRepositoriesWithUploadRecords),
			NumRepositoriesWithFreshUploadRecords: safeDerefIntPtr(numRepositoriesWithFreshUploadRecords),
			NumRepositoriesWithIndexRecords:       safeDerefIntPtr(numRepositoriesWithIndexRecords),
			NumRepositoriesWithFreshIndexRecords:  safeDerefIntPtr(numRepositoriesWithFreshIndexRecords),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return byLanguage, nil
}

func safeDerefIntPtr(v *int) int {
	if v != nil {
		return *v
	}

	return 0
}

var codeIntelligenceRepositoryCountsByLanguageQuery = `
-- source: internal/database/event_logs.go:CodeIntelligenceRepositoryCountsByLanguage
SELECT
	-- Clean up indexer by removing sourcegraph/ docker image prefix for auto-index
	-- records, as well as any trailing git tag. This should make all of the in-house
	-- indexer names the same on both lsif_uploads and lsif_indexes records.
	REGEXP_REPLACE(REGEXP_REPLACE(indexer, '^sourcegraph/', ''), ':\w+$', '') AS indexer,
	max(num_repositories_with_upload_records) AS num_repositories_with_upload_records,
	max(num_repositories_with_fresh_upload_records) AS num_repositories_with_fresh_upload_records,
	max(num_repositories_with_index_records) AS num_repositories_with_index_records,
	max(num_repositories_with_fresh_index_records) AS num_repositories_with_fresh_index_records
FROM (
	(SELECT u.indexer, COUNT(DISTINCT u.repository_id), NULL::integer, NULL::integer, NULL::integer
		FROM lsif_dumps_with_repository_name u GROUP BY u.indexer)
UNION
	(SELECT u.indexer, NULL::integer, COUNT(DISTINCT u.repository_id), NULL::integer, NULL::integer
		FROM lsif_dumps_with_repository_name u WHERE u.uploaded_at >= NOW() - '168 hours'::interval GROUP BY u.indexer)
UNION
	(SELECT u.indexer, NULL::integer, NULL::integer, COUNT(DISTINCT u.repository_id), NULL::integer
		FROM lsif_indexes_with_repository_name u WHERE state = 'completed' GROUP BY u.indexer)
UNION
	(SELECT u.indexer, NULL::integer, NULL::integer, NULL::integer, COUNT(DISTINCT u.repository_id)
		FROM lsif_indexes_with_repository_name u WHERE state = 'completed' AND u.queued_at >= NOW() - '168 hours'::interval GROUP BY u.indexer)
) s(
	indexer,
	num_repositories_with_upload_records,
	num_repositories_with_fresh_upload_records,
	num_repositories_with_index_records,
	num_repositories_with_fresh_index_records
)
GROUP BY REGEXP_REPLACE(REGEXP_REPLACE(indexer, '^sourcegraph/', ''), ':\w+$', '')
`

func (l *eventLogStore) CodeIntelligenceSettingsPageViewCount(ctx context.Context) (int, error) {
	return l.codeIntelligenceSettingsPageViewCount(ctx, time.Now().UTC())
}

func (l *eventLogStore) codeIntelligenceSettingsPageViewCount(ctx context.Context, now time.Time) (int, error) {
	pageNames := []string{
		"CodeIntelUploadsPage",
		"CodeIntelUploadPage",
		"CodeIntelIndexesPage",
		"CodeIntelIndexPage",
		"CodeIntelConfigurationPage",
		"CodeIntelConfigurationPolicyPage",
	}

	names := make([]*sqlf.Query, 0, len(pageNames))
	for _, pageName := range pageNames {
		names = append(names, sqlf.Sprintf("%s", fmt.Sprintf("View%s", pageName)))
	}

	count, _, err := basestore.ScanFirstInt(l.Query(ctx, sqlf.Sprintf(codeIntelligenceSettingsPageViewCountQuery, sqlf.Join(names, ","), now)))
	return count, err
}

var codeIntelligenceSettingsPageViewCountQuery = `
-- source: internal/database/event_logs.go:CodeIntelligenceSettingsPageViewCount
SELECT COUNT(*) FROM event_logs WHERE name IN (%s) AND timestamp >= ` + makeDateTruncExpression("week", "%s::timestamp")

func (l *eventLogStore) AggregatedCodeIntelEvents(ctx context.Context) ([]types.CodeIntelAggregatedEvent, error) {
	return l.aggregatedCodeIntelEvents(ctx, time.Now().UTC())
}

func (l *eventLogStore) aggregatedCodeIntelEvents(ctx context.Context, now time.Time) (events []types.CodeIntelAggregatedEvent, err error) {
	var eventNames = []string{
		"codeintel.lsifHover",
		"codeintel.lsifDefinitions",
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences",
		"codeintel.lsifReferences.xrepo",
		"codeintel.searchHover",
		"codeintel.searchDefinitions",
		"codeintel.searchDefinitions.xrepo",
		"codeintel.searchReferences",
		"codeintel.searchReferences.xrepo",
	}

	var eventNameQueries []*sqlf.Query
	for _, name := range eventNames {
		eventNameQueries = append(eventNameQueries, sqlf.Sprintf("%s", name))
	}

	query := sqlf.Sprintf(aggregatedCodeIntelEventsQuery, now, now, sqlf.Join(eventNameQueries, ", "))

	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var event types.CodeIntelAggregatedEvent
		err := rows.Scan(
			&event.Name,
			&event.LanguageID,
			&event.Week,
			&event.TotalWeek,
			&event.UniquesWeek,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

var aggregatedCodeIntelEventsQuery = `
-- source: internal/database/event_logs.go:aggregatedCodeIntelEvents
WITH events AS (
  SELECT
    name,
    (argument->>'languageId')::text as language_id,
    ` + aggregatedUserIDQueryFragment + ` AS user_id,
    ` + makeDateTruncExpression("week", "%s::timestamp") + ` as current_week
  FROM event_logs
  WHERE
    timestamp >= ` + makeDateTruncExpression("week", "%s::timestamp") + `
    AND name IN (%s)
)
SELECT
  name,
  language_id,
  current_week,
  COUNT(*) AS total_week,
  COUNT(DISTINCT user_id) AS uniques_week
FROM events
GROUP BY name, current_week, language_id
ORDER BY name;
`

func (l *eventLogStore) AggregatedCodeIntelInvestigationEvents(ctx context.Context) ([]types.CodeIntelAggregatedInvestigationEvent, error) {
	return l.aggregatedCodeIntelInvestigationEvents(ctx, time.Now().UTC())
}

func (l *eventLogStore) aggregatedCodeIntelInvestigationEvents(ctx context.Context, now time.Time) (events []types.CodeIntelAggregatedInvestigationEvent, err error) {
	var eventNames = []string{
		"CodeIntelligenceIndexerSetupInvestigated",
		"CodeIntelligenceUploadErrorInvestigated",
		"CodeIntelligenceIndexErrorInvestigated",
	}

	var eventNameQueries []*sqlf.Query
	for _, name := range eventNames {
		eventNameQueries = append(eventNameQueries, sqlf.Sprintf("%s", name))
	}

	query := sqlf.Sprintf(aggregatedCodeIntelInvestigationEventsQuery, now, now, sqlf.Join(eventNameQueries, ", "))

	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var event types.CodeIntelAggregatedInvestigationEvent
		err := rows.Scan(
			&event.Name,
			&event.Week,
			&event.TotalWeek,
			&event.UniquesWeek,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

var aggregatedCodeIntelInvestigationEventsQuery = `
-- source: internal/database/event_logs.go:aggregatedCodeIntelInvestigationEvents
WITH events AS (
  SELECT
    name,
    ` + aggregatedUserIDQueryFragment + ` AS user_id,
    ` + makeDateTruncExpression("week", "%s::timestamp") + ` as current_week
  FROM event_logs
  WHERE
    timestamp >= ` + makeDateTruncExpression("week", "%s::timestamp") + `
    AND name IN (%s)
)
SELECT
  name,
  current_week,
  COUNT(*) AS total_week,
  COUNT(DISTINCT user_id) AS uniques_week
FROM events
GROUP BY name, current_week
ORDER BY name;
`

func (l *eventLogStore) AggregatedSearchEvents(ctx context.Context, now time.Time) ([]types.SearchAggregatedEvent, error) {
	latencyEvents, err := l.aggregatedSearchEvents(ctx, aggregatedSearchLatencyEventsQuery, now)
	if err != nil {
		return nil, err
	}

	usageEvents, err := l.aggregatedSearchEvents(ctx, aggregatedSearchUsageEventsQuery, now)
	if err != nil {
		return nil, err
	}
	return append(latencyEvents, usageEvents...), nil
}

func (l *eventLogStore) aggregatedSearchEvents(ctx context.Context, queryString string, now time.Time) (events []types.SearchAggregatedEvent, err error) {
	query := sqlf.Sprintf(queryString, now, now, now, now)

	rows, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event types.SearchAggregatedEvent
		err := rows.Scan(
			&event.Name,
			&event.Month,
			&event.Week,
			&event.Day,
			&event.TotalMonth,
			&event.TotalWeek,
			&event.TotalDay,
			&event.UniquesMonth,
			&event.UniquesWeek,
			&event.UniquesDay,
			pq.Array(&event.LatenciesMonth),
			pq.Array(&event.LatenciesWeek),
			pq.Array(&event.LatenciesDay),
		)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

var searchLatencyEventNames = []string{
	"'search.latencies.literal'",
	"'search.latencies.regexp'",
	"'search.latencies.structural'",
	"'search.latencies.file'",
	"'search.latencies.repo'",
	"'search.latencies.diff'",
	"'search.latencies.commit'",
	"'search.latencies.symbol'",
}

var aggregatedSearchLatencyEventsQuery = `
-- source: internal/database/event_logs.go:aggregatedSearchLatencyEvents
WITH events AS (
  SELECT
    name,
    -- Postgres 9.6 needs to go from text to integer (i.e. can't go directly to integer)
    (argument->'durationMs')::text::integer as latency,
    ` + aggregatedUserIDQueryFragment + ` AS user_id,
    ` + makeDateTruncExpression("month", "timestamp") + ` as month,
    ` + makeDateTruncExpression("week", "timestamp") + ` as week,
    ` + makeDateTruncExpression("day", "timestamp") + ` as day,
    ` + makeDateTruncExpression("month", "%s::timestamp") + ` as current_month,
    ` + makeDateTruncExpression("week", "%s::timestamp") + ` as current_week,
    ` + makeDateTruncExpression("day", "%s::timestamp") + ` as current_day
  FROM event_logs
  WHERE
    timestamp >= ` + makeDateTruncExpression("month", "%s::timestamp") + `
    AND name IN (` + strings.Join(searchLatencyEventNames, ", ") + `)
)
SELECT
  name,
  current_month,
  current_week,
  current_day,
  COUNT(*) FILTER (WHERE month = current_month) AS total_month,
  COUNT(*) FILTER (WHERE week = current_week) AS total_week,
  COUNT(*) FILTER (WHERE day = current_day) AS total_day,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month) AS uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week) AS uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE day = current_day) AS uniques_day,
  PERCENTILE_CONT(ARRAY[0.50, 0.90, 0.99]) WITHIN GROUP (ORDER BY latency) FILTER (WHERE month = current_month) AS latencies_month,
  PERCENTILE_CONT(ARRAY[0.50, 0.90, 0.99]) WITHIN GROUP (ORDER BY latency) FILTER (WHERE week = current_week) AS latencies_week,
  PERCENTILE_CONT(ARRAY[0.50, 0.90, 0.99]) WITHIN GROUP (ORDER BY latency) FILTER (WHERE day = current_day) AS latencies_day
FROM events GROUP BY name, current_month, current_week, current_day
`

var aggregatedSearchUsageEventsQuery = `
-- source: internal/database/event_logs.go:aggregatedSearchUsageEvents
WITH events AS (
  SELECT
    json.key::text,
    json.value::text,
    ` + aggregatedUserIDQueryFragment + ` AS user_id,
    ` + makeDateTruncExpression("month", "timestamp") + ` as month,
    ` + makeDateTruncExpression("week", "timestamp") + ` as week,
    ` + makeDateTruncExpression("day", "timestamp") + ` as day,
    ` + makeDateTruncExpression("month", "%s::timestamp") + ` as current_month,
    ` + makeDateTruncExpression("week", "%s::timestamp") + ` as current_week,
    ` + makeDateTruncExpression("day", "%s::timestamp") + ` as current_day
  FROM event_logs
  CROSS JOIN LATERAL jsonb_each(argument->'code_search'->'query_data'->'query') json
  WHERE
    timestamp >= ` + makeDateTruncExpression("month", "%s::timestamp") + `
    AND name = 'SearchResultsQueried'
)
SELECT
  key,
  current_month,
  current_week,
  current_day,
  SUM(case when month = current_month then value::int else 0 end) AS total_month,
  SUM(case when week = current_week then value::int else 0 end) AS total_week,
  SUM(case when day = current_day then value::int else 0 end) AS total_day,
  COUNT(DISTINCT user_id) FILTER (WHERE month = current_month) AS uniques_month,
  COUNT(DISTINCT user_id) FILTER (WHERE week = current_week) AS uniques_week,
  COUNT(DISTINCT user_id) FILTER (WHERE day = current_day) AS uniques_day,
  NULL,
  NULL,
  NULL
FROM events
WHERE key IN
  (
	'count_or',
	'count_and',
	'count_not',
	'count_select_repo',
	'count_select_file',
	'count_select_content',
	'count_select_symbol',
	'count_select_commit_diff_added',
	'count_select_commit_diff_removed',
	'count_repo_contains',
	'count_repo_contains_file',
	'count_repo_contains_content',
	'count_repo_contains_commit_after',
	'count_repo_dependencies',
	'count_count_all',
	'count_non_global_context',
	'count_only_patterns',
	'count_only_patterns_three_or_more'
  )
GROUP BY key, current_month, current_week, current_day
`

// userIDQueryFragment is a query fragment that can be used to return the anonymous user ID
// when the user ID is not set (i.e. 0).
const userIDQueryFragment = `
CASE WHEN user_id = 0
  THEN anonymous_user_id
  ELSE CAST(user_id AS TEXT)
END
`

// aggregatedUserIDQueryFragment is a query fragment that can be used to canonicalize the
// values of the user_id and anonymous_user_id fields (assumed in scope) int a unified value.
const aggregatedUserIDQueryFragment = `
CASE WHEN user_id = 0
  -- It's faster to group by an int rather than text, so we convert
  -- the anonymous_user_id to an int, rather than the user_id to text.
  THEN ('x' || substr(md5(anonymous_user_id), 1, 8))::bit(32)::int
  ELSE user_id
END
`

// makeDateTruncExpression returns an expresson that converts the given
// SQL expression into the start of the containing date container specified
// by the unit parameter (e.g. day, week, month, or).
func makeDateTruncExpression(unit, expr string) string {
	if unit == "week" {
		return fmt.Sprintf(`DATE_TRUNC('%s', TIMEZONE('UTC', %s) + '1 day'::interval) - '1 day'::interval`, unit, expr)
	}

	return fmt.Sprintf(`DATE_TRUNC('%s', TIMEZONE('UTC', %s))`, unit, expr)
}

// RequestsByLanguage returns a map of language names to the number of requests of precise support for that language.
func (l *eventLogStore) RequestsByLanguage(ctx context.Context) (_ map[string]int, err error) {
	rows, err := l.Query(ctx, sqlf.Sprintf(requestsByLanguageQuery))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	requestsByLanguage := map[string]int{}
	for rows.Next() {
		var (
			language string
			count    int
		)
		if err := rows.Scan(&language, &count); err != nil {
			return nil, err
		}

		requestsByLanguage[language] = count
	}

	return requestsByLanguage, nil
}

var requestsByLanguageQuery = `
-- source: internal/database/event_logs.go:RequestsByLanguage
SELECT
	language_id,
	COUNT(*) as count
FROM codeintel_langugage_support_requests
GROUP BY language_id
`
