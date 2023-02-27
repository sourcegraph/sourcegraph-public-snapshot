package insights

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
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	InsightIDs []string `json:"insightIds"`
	UserID     *int32   `json:"userId"`
	OrgID      *int32   `json:"orgId"`
}
