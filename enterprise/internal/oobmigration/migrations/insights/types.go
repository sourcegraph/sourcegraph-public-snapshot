package insights

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type timeSeries struct {
	Name   string
	Stroke string
	Query  string
}

type interval struct {
	Years  *int
	Months *int
	Weeks  *int
	Days   *int
	Hours  *int
}

type searchInsight struct {
	ID           string
	Title        string
	Description  string
	Repositories []string
	Series       []timeSeries
	Step         interval
	Visibility   string
	OrgID        *int32
	UserID       *int32
	Filters      *defaultFilters
}

type defaultFilters struct {
	IncludeRepoRegexp *string
	ExcludeRepoRegexp *string
}

type permissionAssociations struct {
	userID *int32
	orgID  *int32
}

type langStatsInsight struct {
	ID             string
	Title          string
	Repository     string
	OtherThreshold float64
	OrgID          *int32
	UserID         *int32
}

type insightViewSeriesMetadata struct {
	Label  string
	Stroke string
}

type insightView struct {
	ID                  int
	Title               string
	Description         string
	UniqueID            string
	Filters             insightViewFilters
	OtherThreshold      *float64
	PresentationType    presentationType
	IsFrozen            bool
	SeriesSortMode      *seriesSortMode
	SeriesSortDirection *seriesSortDirection
	SeriesLimit         *int32
}

type insightViewFilters struct {
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
	SearchContexts   []string
}

type presentationType string

const (
	Line presentationType = "LINE"
	Pie  presentationType = "PIE"
)

type seriesSortMode string

const (
	ResultCount     seriesSortMode = "RESULT_COUNT"    // Sorts by the number of results for the most recent datapoint of a series.
	DateAdded       seriesSortMode = "DATE_ADDED"      // Sorts by the date of the earliest datapoint in the series.
	Lexicographical seriesSortMode = "LEXICOGRAPHICAL" // Sorts by label: first by semantic version and then alphabetically.
)

type seriesSortDirection string

const (
	Asc  seriesSortDirection = "ASC"
	Desc seriesSortDirection = "DESC"
)

type insightViewGrant struct {
	UserID *int
	OrgID  *int
	Global *bool
}

func userGrant(userID int) insightViewGrant {
	return insightViewGrant{UserID: &userID}
}

func orgGrant(orgID int) insightViewGrant {
	return insightViewGrant{OrgID: &orgID}
}

func globalGrant() insightViewGrant {
	b := true
	return insightViewGrant{Global: &b}
}

type timeInterval struct {
	unit  intervalUnit
	value int
}

type migrationBatch string

const (
	backend  migrationBatch = "backend"
	frontend migrationBatch = "frontend"
)

type migrator struct {
	frontendStore *basestore.Store
	insightsStore *basestore.Store
}

type settingsMigrationJobType string

const (
	UserJob   settingsMigrationJobType = "USER"
	OrgJob    settingsMigrationJobType = "ORG"
	GlobalJob settingsMigrationJobType = "GLOBAL"
)

type dashboardType string

const (
	standard dashboardType = "standard"
	// This is a singleton dashboard that facilitates users having global access to their insights in Limited Access Mode.
	limitedAccessMode dashboardType = "limited_access_mode"
)

type settingsMigrationJob struct {
	UserId             *int
	OrgId              *int
	Global             bool
	MigratedInsights   int
	MigratedDashboards int
	Runs               int
}

type organization struct {
	ID          int32
	Name        string
	DisplayName *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type user struct {
	ID                    int32
	Username              string
	DisplayName           string
	AvatarURL             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	SiteAdmin             bool
	BuiltinAuth           bool
	Tags                  []string
	InvalidatedSessionsAt time.Time
	TosAccepted           bool
	Searchable            bool
}

type settingsSubject struct {
	Default bool   // whether this is for default settings
	Site    bool   // whether this is for global settings
	Org     *int32 // the org's ID
	User    *int32 // the user's ID
}
type settings struct {
	ID           int32           // the unique ID of this settings value
	Subject      settingsSubject // the subject of these settings
	AuthorUserID *int32          // the ID of the user who authored this settings value
	Contents     string          // the raw JSON (with comments and trailing commas allowed)
	CreatedAt    time.Time       // the date when this settings value was created
}

type insightSeries struct {
	ID                         int
	SeriesID                   string
	Query                      string
	CreatedAt                  time.Time
	OldestHistoricalAt         time.Time
	LastRecordedAt             time.Time
	NextRecordingAfter         time.Time
	LastSnapshotAt             time.Time
	NextSnapshotAfter          time.Time
	BackfillQueuedAt           time.Time
	Enabled                    bool
	Repositories               []string
	SampleIntervalUnit         string
	SampleIntervalValue        int
	GeneratedFromCaptureGroups bool
	JustInTime                 bool
	GenerationMethod           generationMethod
	GroupBy                    *string
	BackfillAttempts           int32
}

type generationMethod string

const (
	Search         generationMethod = "search"
	SearchCompute  generationMethod = "search-compute"
	LanguageStats  generationMethod = "language-stats"
	MappingCompute generationMethod = "mapping-compute"
)

type TimeInterval struct {
	Unit  intervalUnit
	Value int
}

func (t TimeInterval) StepForwards(start time.Time) time.Time {
	return t.step(start, forward)
}

type stepDirection int

const forward stepDirection = 1
const backward stepDirection = -1

func (t TimeInterval) step(start time.Time, direction stepDirection) time.Time {
	switch t.Unit {
	case Year:
		return start.AddDate(int(direction)*t.Value, 0, 0)
	case Month:
		return start.AddDate(0, int(direction)*t.Value, 0)
	case Week:
		return start.AddDate(0, 0, int(direction)*7*t.Value)
	case Day:
		return start.AddDate(0, 0, int(direction)*t.Value)
	case Hour:
		return start.Add(time.Hour * time.Duration(t.Value) * time.Duration(direction))
	default:
		// this doesn't really make sense, so return something?
		return start.AddDate(int(direction)*t.Value, 0, 0)
	}
}

type intervalUnit string

const (
	Month intervalUnit = "MONTH"
	Day   intervalUnit = "DAY"
	Week  intervalUnit = "WEEK"
	Year  intervalUnit = "YEAR"
	Hour  intervalUnit = "HOUR"
)

type settingDashboard struct {
	ID         string   `json:"id,omitempty"`
	Title      string   `json:"title,omitempty"`
	InsightIds []string `json:"insightIds,omitempty"`
	UserID     *int32
	OrgID      *int32
}
