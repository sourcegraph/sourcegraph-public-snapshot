package insights

import (
	"time"
)

type permissionAssociations struct {
	userID *int32
	orgID  *int32
}

type insightViewSeriesMetadata struct {
	Label  string
	Stroke string
}

type insightView struct {
	ID               int
	Title            string
	Description      string
	UniqueID         string
	Filters          insightViewFilters
	OtherThreshold   *float64
	PresentationType string
}

type insightViewFilters struct {
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
	SearchContexts   []string
}

type timeInterval struct {
	unit  string
	value int
}

type settingsMigrationJob struct {
	UserId             *int
	OrgId              *int
	Global             bool
	MigratedInsights   int
	MigratedDashboards int
	Runs               int
}

type userOrOrg struct {
	ID          int32
	Name        string
	DisplayName *string
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
	Repositories               []string
	SampleIntervalUnit         string
	SampleIntervalValue        int
	GeneratedFromCaptureGroups bool
	JustInTime                 bool
	GenerationMethod           string
	GroupBy                    *string
}

func (t timeInterval) StepForwards(start time.Time) time.Time {
	switch t.unit {
	case "YEAR":
		return start.AddDate(t.value, 0, 0)
	case "MONTH":
		return start.AddDate(0, t.value, 0)
	case "WEEK":
		return start.AddDate(0, 0, 7*t.value)
	case "DAY":
		return start.AddDate(0, 0, t.value)
	case "HOUR":
		return start.Add(time.Hour * time.Duration(t.value))
	default:
		// this doesn't really make sense, so return something?
		return start.AddDate(t.value, 0, 0)
	}
}

//
// JSON UNMARSHALLED
//

type settings struct {
	ID       int32           // the unique ID of this settings value
	Subject  settingsSubject // the subject of these settings
	Contents string          // the raw JSON (with comments and trailing commas allowed)
}

type settingsSubject struct {
	Org  *int32 // the org's ID
	User *int32 // the user's ID
}

type langStatsInsight struct {
	ID             string
	Title          string
	Repository     string
	OtherThreshold float64
	OrgID          *int32
	UserID         *int32
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

type defaultFilters struct {
	IncludeRepoRegexp *string
	ExcludeRepoRegexp *string
}

type settingDashboard struct {
	ID         string   `json:"id,omitempty"`
	Title      string   `json:"title,omitempty"`
	InsightIds []string `json:"insightIds,omitempty"`
	UserID     *int32
	OrgID      *int32
}
