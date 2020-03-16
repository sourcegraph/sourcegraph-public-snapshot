// Package types defines types used by the frontend.
package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// RepoFields are lazy loaded data fields on a Repo (from the DB).
type RepoFields struct {
	// URI is the full name for this repository (e.g.,
	// "github.com/user/repo"). See the documentation for the Name field.
	URI string

	// Description is a brief description of the repository.
	Description string

	// DEPRECATED: this field is always empty for new repositories as of
	// https://github.com/sourcegraph/sourcegraph/issues/2586. Do not use it.
	//
	// Language is the primary programming language used in this repository.
	Language string

	// Fork is whether this repository is a fork of another repository.
	Fork bool
}

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository.
	ID api.RepoID
	// ExternalRepo identifies this repository by its ID on the external service where it resides (and the external
	// service itself).
	ExternalRepo api.ExternalRepoSpec
	// Name is the name for this repository (e.g., "github.com/user/repo"). It
	// is the same as URI, unless the user configures a non-default
	// repositoryPathPattern.
	//
	// Previously, this was called RepoURI.
	Name api.RepoName

	// Private is whether the repository is private on the code host.
	Private bool

	// RepoFields contains fields that are loaded from the DB only when necessary.
	// This is to reduce memory usage when loading thousands of repos.
	*RepoFields
}

// Repos is an utility type of a list of repos.
type Repos []*Repo

func (rs Repos) Len() int           { return len(rs) }
func (rs Repos) Less(i, j int) bool { return rs[i].ID < rs[j].ID }
func (rs Repos) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// ExternalService is a connection to an external service.
type ExternalService struct {
	ID          int64
	Kind        string
	DisplayName string
	Config      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type GlobalState struct {
	SiteID      string
	Initialized bool // whether the initial site admin account has been created
}

// User represents a registered user.
type User struct {
	ID          int32
	Username    string
	DisplayName string
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SiteAdmin   bool
	BuiltinAuth bool
	Tags        []string
}

type Org struct {
	ID          int32
	Name        string
	DisplayName *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrgMembership struct {
	ID        int32
	OrgID     int32
	UserID    int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PhabricatorRepo struct {
	ID       int32
	Name     api.RepoName
	URL      string
	Callsign string
}

type UserUsageStatistics struct {
	UserID                      int32
	PageViews                   int32
	SearchQueries               int32
	CodeIntelligenceActions     int32
	FindReferencesActions       int32
	LastActiveTime              *time.Time
	LastCodeHostIntegrationTime *time.Time
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type SiteUsageStatistics struct {
	DAUs []*SiteActivityPeriod
	WAUs []*SiteActivityPeriod
	MAUs []*SiteActivityPeriod
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type SiteActivityPeriod struct {
	StartTime            time.Time
	UserCount            int32
	RegisteredUserCount  int32
	AnonymousUserCount   int32
	IntegrationUserCount int32
	Stages               *Stages
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type Stages struct {
	Manage    int32 `json:"mng"`
	Plan      int32 `json:"plan"`
	Code      int32 `json:"code"`
	Review    int32 `json:"rev"`
	Verify    int32 `json:"ver"`
	Package   int32 `json:"pkg"`
	Deploy    int32 `json:"depl"`
	Configure int32 `json:"conf"`
	Monitor   int32 `json:"mtr"`
	Secure    int32 `json:"sec"`
	Automate  int32 `json:"auto"`
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type CampaignsUsageStatistics struct {
	CampaignsCount int32
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type CodeIntelUsageStatistics struct {
	Daily   []*CodeIntelUsagePeriod
	Weekly  []*CodeIntelUsagePeriod
	Monthly []*CodeIntelUsagePeriod
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type CodeIntelUsagePeriod struct {
	StartTime   time.Time
	Hover       *CodeIntelEventCategoryStatistics
	Definitions *CodeIntelEventCategoryStatistics
	References  *CodeIntelEventCategoryStatistics
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type CodeIntelEventCategoryStatistics struct {
	LSIF   *CodeIntelEventStatistics
	LSP    *CodeIntelEventStatistics
	Search *CodeIntelEventStatistics
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type CodeIntelEventStatistics struct {
	UsersCount     int32
	EventsCount    *int32
	EventLatencies *CodeIntelEventLatencies
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type CodeIntelEventLatencies struct {
	P50 float64
	P90 float64
	P99 float64
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type SearchUsageStatistics struct {
	Daily   []*SearchUsagePeriod
	Weekly  []*SearchUsagePeriod
	Monthly []*SearchUsagePeriod
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type SearchUsagePeriod struct {
	StartTime          time.Time
	TotalUsers         int32
	Literal            *SearchEventStatistics
	Regexp             *SearchEventStatistics
	After              *SearchCountStatistics
	Archived           *SearchCountStatistics
	Author             *SearchCountStatistics
	Before             *SearchCountStatistics
	Case               *SearchCountStatistics
	Commit             *SearchEventStatistics
	Committer          *SearchCountStatistics
	Content            *SearchCountStatistics
	Count              *SearchCountStatistics
	Diff               *SearchEventStatistics
	File               *SearchEventStatistics
	Fork               *SearchCountStatistics
	Index              *SearchCountStatistics
	Lang               *SearchCountStatistics
	Message            *SearchCountStatistics
	PatternType        *SearchCountStatistics
	Repo               *SearchEventStatistics
	Repohascommitafter *SearchCountStatistics
	Repohasfile        *SearchCountStatistics
	Repogroup          *SearchCountStatistics
	Structural         *SearchEventStatistics
	Symbol             *SearchEventStatistics
	Timeout            *SearchCountStatistics
	Type               *SearchCountStatistics
	SearchModes        *SearchModeUsageStatistics
}

type SearchModeUsageStatistics struct {
	Interactive *SearchCountStatistics
	PlainText   *SearchCountStatistics
}

type SearchCountStatistics struct {
	UserCount   *int32
	EventsCount *int32
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type SearchEventStatistics struct {
	UserCount      *int32
	EventsCount    *int32
	EventLatencies *SearchEventLatencies
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type SearchEventLatencies struct {
	P50 float64
	P90 float64
	P99 float64
}

type SurveyResponse struct {
	ID        int32
	UserID    *int32
	Email     *string
	Score     int32
	Reason    *string
	Better    *string
	CreatedAt time.Time
}

type Event struct {
	ID              int32
	Name            string
	URL             string
	UserID          *int32
	AnonymousUserID string
	Argument        string
	Source          string
	Version         string
	Timestamp       time.Time
}
