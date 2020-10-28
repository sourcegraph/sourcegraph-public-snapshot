// Package types defines types used by the frontend.
package types

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// RepoFields are lazy loaded data fields on a Repo (from the DB).
type RepoFields struct {
	// URI is the full name for this repository (e.g.,
	// "github.com/user/repo"). See the documentation for the Name field.
	URI string

	// Description is a brief description of the repository.
	Description string

	// Fork is whether this repository is a fork of another repository.
	Fork bool

	// Archived is whether this repository has been archived.
	Archived bool

	// Cloned is whether this repository is cloned.
	Cloned bool

	// CreatedAt indicates when the repository record was created.
	CreatedAt time.Time

	// UpdatedAt is when this repository's metadata was last updated on Sourcegraph.
	UpdatedAt time.Time

	// DeletedAt is when this repository was soft-deleted from Sourcegraph.
	DeletedAt time.Time

	// Metadata contains the raw source code host JSON metadata.
	Metadata interface{}

	// Sources identifies all the repo sources this Repo belongs to.
	// The key is a URN created by extsvc.URN
	Sources map[string]*SourceInfo
}

// A SourceInfo represents a source a Repo belongs to (such as an external service).
type SourceInfo struct {
	ID       string
	CloneURL string
}

// ExternalServiceID returns the ID of the external service this
// SourceInfo refers to.
func (i SourceInfo) ExternalServiceID() int64 {
	ps := strings.SplitN(i.ID, ":", 3)
	if len(ps) != 3 {
		return -1
	}

	id, err := strconv.ParseInt(ps[2], 10, 64)
	if err != nil {
		return -1
	}

	return id
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

// CloneURLs returns all the clone URLs this repo is clonable from.
func (r *Repo) CloneURLs() []string {
	urls := make([]string, 0, len(r.Sources))
	for _, src := range r.Sources {
		if src != nil && src.CloneURL != "" {
			urls = append(urls, src.CloneURL)
		}
	}
	return urls
}

// IsDeleted returns true if the repo is deleted.
func (r *Repo) IsDeleted() bool { return !r.DeletedAt.IsZero() }

// ExternalServiceIDs returns the IDs of the external services this
// repo belongs to.
func (r *Repo) ExternalServiceIDs() []int64 {
	ids := make([]int64, 0, len(r.Sources))
	for _, src := range r.Sources {
		ids = append(ids, src.ExternalServiceID())
	}
	return ids
}

// Update updates Repo r with the fields from the given newer Repo n,
// returning true if modified.
func (r *Repo) Update(n *Repo) (modified bool) {
	if r.Name != n.Name {
		r.Name, modified = n.Name, true
	}

	if r.URI != n.URI {
		r.URI, modified = n.URI, true
	}

	if r.Description != n.Description {
		r.Description, modified = n.Description, true
	}

	if n.ExternalRepo != (api.ExternalRepoSpec{}) &&
		!r.ExternalRepo.Equal(&n.ExternalRepo) {
		r.ExternalRepo, modified = n.ExternalRepo, true
	}

	if r.Archived != n.Archived {
		r.Archived, modified = n.Archived, true
	}

	if r.Fork != n.Fork {
		r.Fork, modified = n.Fork, true
	}

	if r.Private != n.Private {
		r.Private, modified = n.Private, true
	}

	if !reflect.DeepEqual(r.Sources, n.Sources) {
		r.Sources, modified = n.Sources, true
	}

	// As a special case, we clear out the value of ViewerPermission for GitHub repos as
	// the value is dependent on the token used to fetch it. We don't want to store this in the DB as it will
	// flip flop as we fetch the same repo from different external services.
	switch x := n.Metadata.(type) {
	case *github.Repository:
		cp := *x
		cp.ViewerPermission = ""
		n = n.With(func(clone *Repo) {
			// Repo.Clone does not currently clone metadata for any types as they could contain hard to clone
			// items such as maps. However, we know that copying github.Repository is safe as it only contains values.
			clone.Metadata = &cp
		})
	}

	if !reflect.DeepEqual(r.Metadata, n.Metadata) {
		r.Metadata, modified = n.Metadata, true
	}

	return modified
}

// Clone returns a clone of the given repo.
func (r *Repo) Clone() *Repo {
	if r == nil {
		return nil
	}
	clone := *r
	if r.Sources != nil {
		clone.Sources = make(map[string]*SourceInfo, len(r.Sources))
		for k, v := range r.Sources {
			clone.Sources[k] = v
		}
	}
	return &clone
}

// Apply applies the given functional options to the Repo.
func (r *Repo) Apply(opts ...func(*Repo)) {
	if r == nil {
		return
	}

	for _, opt := range opts {
		opt(r)
	}
}

// With returns a clone of the given repo with the given functional options applied.
func (r *Repo) With(opts ...func(*Repo)) *Repo {
	clone := r.Clone()
	clone.Apply(opts...)
	return clone
}

// Less compares Repos by the important fields (fields with constraints in our
// DB). Additionally it will compare on Sources to give a deterministic order
// on repos returned from a sourcer.
//
// NewDiff relies on Less to deterministically decide on the order to merge
// repositories, as well as which repository to keep on conflicts.
//
// Context on using other fields such as timestamps to order/resolve
// conflicts: We only want to rely on values that have constraints in our
// database. Timestamps have the following downsides:
//
//   - We need to assume the upstream codehost has reasonable values for them
//   - Not all codehosts set them to relevant values (eg gitolite or other)
//   - They could change often for codehosts that do set them.
func (r *Repo) Less(s *Repo) bool {
	if r.ID != s.ID {
		return r.ID < s.ID
	}
	if r.Name != s.Name {
		return r.Name < s.Name
	}
	if cmp := r.ExternalRepo.Compare(s.ExternalRepo); cmp != 0 {
		return cmp == -1
	}

	return sortedSliceLess(sourcesKeys(r.Sources), sourcesKeys(s.Sources))
}

func (r *Repo) String() string {
	eid := fmt.Sprintf("{%s %s %s}", r.ExternalRepo.ServiceID, r.ExternalRepo.ServiceType, r.ExternalRepo.ID)
	if r.IsDeleted() {
		return fmt.Sprintf("Repo{ID: %d, Name: %q, EID: %s, IsDeleted: true}", r.ID, r.Name, eid)
	}
	return fmt.Sprintf("Repo{ID: %d, Name: %q, EID: %s}", r.ID, r.Name, eid)
}

func sourcesKeys(m map[string]*SourceInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedSliceLess returns true if a < b
func sortedSliceLess(a, b []string) bool {
	for i, v := range a {
		if i == len(b) {
			return false
		}
		if v != b[i] {
			return v < b[i]
		}
	}
	return len(a) != len(b)
}

// Repos is an utility type with convenience methods for operating on lists of Repos.
type Repos []*Repo

func (rs Repos) Len() int           { return len(rs) }
func (rs Repos) Less(i, j int) bool { return rs[i].ID < rs[j].ID }
func (rs Repos) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// IDs returns the list of ids from all Repos.
func (rs Repos) IDs() []api.RepoID {
	ids := make([]api.RepoID, len(rs))
	for i := range rs {
		ids[i] = rs[i].ID
	}
	return ids
}

// Names returns the list of names from all Repos.
func (rs Repos) Names() []string {
	names := make([]string, len(rs))
	for i := range rs {
		names[i] = string(rs[i].Name)
	}
	return names
}

// NamesSummary caps the number of repos to 20 when composing a space-separated list string.
// Used in logging statements.
func (rs Repos) NamesSummary() string {
	if len(rs) > 20 {
		return strings.Join(rs[:20].Names(), " ") + "..."
	}
	return strings.Join(rs.Names(), " ")
}

// Kinds returns the unique set of kinds from all Repos.
func (rs Repos) Kinds() (kinds []string) {
	set := map[string]bool{}
	for _, r := range rs {
		kind := strings.ToUpper(r.ExternalRepo.ServiceType)
		if !set[kind] {
			kinds = append(kinds, kind)
			set[kind] = true
		}
	}
	return kinds
}

// ExternalRepos returns the list of set ExternalRepoSpecs from all Repos.
func (rs Repos) ExternalRepos() []api.ExternalRepoSpec {
	specs := make([]api.ExternalRepoSpec, 0, len(rs))
	for _, r := range rs {
		specs = append(specs, r.ExternalRepo)
	}
	return specs
}

// Sources returns a map of all the sources per repo id.
func (rs Repos) Sources() map[api.RepoID][]SourceInfo {
	sources := make(map[api.RepoID][]SourceInfo)
	for i := range rs {
		for _, info := range rs[i].Sources {
			sources[rs[i].ID] = append(sources[rs[i].ID], *info)
		}
	}

	return sources
}

// Concat adds the given Repos to the end of rs.
func (rs *Repos) Concat(others ...Repos) {
	for _, o := range others {
		*rs = append(*rs, o...)
	}
}

// Clone returns a clone of Repos.
func (rs Repos) Clone() Repos {
	o := make(Repos, 0, len(rs))
	for _, r := range rs {
		o = append(o, r.Clone())
	}
	return o
}

// Apply applies the given functional options to the Repo.
func (rs Repos) Apply(opts ...func(*Repo)) {
	for _, r := range rs {
		r.Apply(opts...)
	}
}

// With returns a clone of the given repos with the given functional options applied.
func (rs Repos) With(opts ...func(*Repo)) Repos {
	clone := rs.Clone()
	clone.Apply(opts...)
	return clone
}

// Filter returns all the Repos that match the given predicate.
func (rs Repos) Filter(pred func(*Repo) bool) (fs Repos) {
	for _, r := range rs {
		if pred(r) {
			fs = append(fs, r)
		}
	}
	return fs
}

// ExternalService is a connection to an external service.
type ExternalService struct {
	ID              int64
	Kind            string
	DisplayName     string
	Config          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	LastSyncAt      *time.Time
	NextSyncAt      *time.Time
	NamespaceUserID *int32
}

// URN returns a unique resource identifier of this external service.
func (e *ExternalService) URN() string {
	return extsvc.URN(e.Kind, e.ID)
}

type GlobalState struct {
	SiteID      string
	Initialized bool // whether the initial site admin account has been created
}

// User represents a registered user.
type User struct {
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

// UserUsageCounts captures the usage numbers of a user in a single day.
type UserUsageCounts struct {
	Date           time.Time
	UserID         uint32
	SearchCount    int32
	CodeIntelCount int32
}

// UserDates captures the created and deleted dates of a single user.
type UserDates struct {
	UserID    int32
	CreatedAt time.Time
	DeletedAt time.Time
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
}

// NOTE: DO NOT alter this struct without making a symmetric change
// to the updatecheck handler. This struct is marshalled and sent to
// BigQuery, which requires the input match its schema exactly.
type CampaignsUsageStatistics struct {
	CampaignsCount              int32
	ActionChangesetsCount       int32
	ActionChangesetsMergedCount int32
	ManualChangesetsCount       int32
	ManualChangesetsMergedCount int32
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
	UsersCount  int32
	EventsCount *int32
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

// SiteUsageSummary is an alternate view of SiteUsageStatistics which is
// calculated in the database layer.
type SiteUsageSummary struct {
	Month                   time.Time
	Week                    time.Time
	Day                     time.Time
	UniquesMonth            int32
	UniquesWeek             int32
	UniquesDay              int32
	RegisteredUniquesMonth  int32
	RegisteredUniquesWeek   int32
	RegisteredUniquesDay    int32
	IntegrationUniquesMonth int32
	IntegrationUniquesWeek  int32
	IntegrationUniquesDay   int32
	ManageUniquesMonth      int32
	CodeUniquesMonth        int32
	VerifyUniquesMonth      int32
	MonitorUniquesMonth     int32
	ManageUniquesWeek       int32
	CodeUniquesWeek         int32
	VerifyUniquesWeek       int32
	MonitorUniquesWeek      int32
}

// AggregatedEvent represents the total events, unique users, and
// latencies over the current month, week, and day for a single event.
type AggregatedEvent struct {
	Name           string
	Month          time.Time
	Week           time.Time
	Day            time.Time
	TotalMonth     int32
	TotalWeek      int32
	TotalDay       int32
	UniquesMonth   int32
	UniquesWeek    int32
	UniquesDay     int32
	LatenciesMonth []float64
	LatenciesWeek  []float64
	LatenciesDay   []float64
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

// GrowthStatistics represents the total users that were created,
// deleted, resurrected, churned and retained over the current month.
type GrowthStatistics struct {
	DeletedUsers     int32
	CreatedUsers     int32
	ResurrectedUsers int32
	ChurnedUsers     int32
	RetainedUsers    int32
}

// SavedSearches represents the total number of saved searches, users
// using saved searches, and usage of saved searches.
type SavedSearches struct {
	TotalSavedSearches   int32
	UniqueUsers          int32
	NotificationsSent    int32
	NotificationsClicked int32
	UniqueUserPageViews  int32
	OrgSavedSearches     int32
}

// Panel homepage represents interaction data on the
// enterprise homepage panels.
type HomepagePanels struct {
	RecentFilesClickedPercentage           float64
	RecentSearchClickedPercentage          float64
	RecentRepositoriesClickedPercentage    float64
	SavedSearchesClickedPercentage         float64
	NewSavedSearchesClickedPercentage      float64
	TotalPanelViews                        float64
	UsersFilesClickedPercentage            float64
	UsersSearchClickedPercentage           float64
	UsersRepositoriesClickedPercentage     float64
	UsersSavedSearchesClickedPercentage    float64
	UsersNewSavedSearchesClickedPercentage float64
	PercentUsersShown                      float64
}

// Secret represents the secrets table
type Secret struct {
	ID int32

	// The table containing an object whose token is being encrypted.
	SourceType sql.NullString

	// The ID of the object in the SourceType table.
	SourceID sql.NullInt32

	// KeyName represents a unique key for the case where we're storing key-value pairs.
	KeyName sql.NullString

	// Value contains the encrypted string
	Value string
}
