// Package api contains an API client and types for cross-service communication.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

// RepoID is the unique identifier for a repository.
type RepoID int32

// RepoName is the name of a repository, consisting of one or more "/"-separated path components.
//
// Previously, this was called RepoURI.
type RepoName string

// RepoHashedName is the hashed name of a repo
type RepoHashedName string

func (r RepoName) Equal(o RepoName) bool {
	return strings.EqualFold(string(r), string(o))
}

var deletedRegex = lazyregexp.New("DELETED-[0-9]+\\.[0-9]+-")

// UndeletedRepoName will "undelete" a repo name. When we soft-delete a repo we
// change its name in the database, this function extracts the original repo
// name.
func UndeletedRepoName(name RepoName) RepoName {
	return RepoName(deletedRegex.ReplaceAllString(string(name), ""))
}

// CommitID is the 40-character SHA-1 hash for a Git commit.
type CommitID string

// Short returns the SHA-1 commit hash truncated to 7 characters
func (c CommitID) Short() string {
	if len(c) >= 7 {
		return string(c)[:7]
	}
	return string(c)
}

// RevSpec is a revision range specifier suitable for passing to git. See
// the manpage gitrevisions(7).
type RevSpec string

// RepoCommit scopes a commit to a repository.
type RepoCommit struct {
	Repo     RepoName
	CommitID CommitID
}

func (rc RepoCommit) LogFields() []log.Field {
	return []log.Field{
		log.String("repo", string(rc.Repo)),
		log.String("commitID", string(rc.CommitID)),
	}
}

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository on Sourcegraph.
	ID RepoID

	// ExternalRepo identifies this repository by its ID on the external service where it resides (and the external
	// service itself).
	ExternalRepo *ExternalRepoSpec

	// Name is the name of the repository (such as "github.com/user/repo").
	Name RepoName
	// Enabled is whether the repository is enabled. Disabled repositories are
	// not accessible by users (except site admins).
	Enabled bool
}

// ExternalRepoSpec specifies a repository on an external service (such as GitHub or GitLab).
type ExternalRepoSpec struct {
	// ID is the repository's ID on the external service. Its value is opaque except to the repo-updater.
	//
	// For GitHub, this is the GitHub GraphQL API's node ID for the repository.
	ID string

	// ServiceType is the type of external service. Its value is opaque except to the repo-updater.
	//
	// Example: "github", "gitlab", etc.
	ServiceType string

	// ServiceID is the particular instance of the external service where this repository resides. Its value is
	// opaque but typically consists of the canonical base URL to the service.
	//
	// Implementations must take care to normalize this URL. For example, if different GitHub.com repository code
	// paths used slightly different values here (such as "https://github.com/" and "https://github.com", note the
	// lack of trailing slash), then the same logical repository would be incorrectly treated as multiple distinct
	// repositories depending on the code path that provided its ServiceID value.
	//
	// Example: "https://github.com/", "https://github-enterprise.example.com/"
	ServiceID string
}

// Equal returns true if r is equal to s.
func (r ExternalRepoSpec) Equal(s *ExternalRepoSpec) bool {
	return r.ID == s.ID && r.ServiceType == s.ServiceType && r.ServiceID == s.ServiceID
}

// Compare returns -1 if r < s, 0 if r == s or 1 if r > s
func (r ExternalRepoSpec) Compare(s ExternalRepoSpec) int {
	if r.ServiceType != s.ServiceType {
		return cmp(r.ServiceType, s.ServiceType)
	}
	if r.ServiceID != s.ServiceID {
		return cmp(r.ServiceID, s.ServiceID)
	}
	return cmp(r.ID, s.ID)
}

func (r ExternalRepoSpec) String() string {
	return fmt.Sprintf("ExternalRepoSpec{%s %s %s}", r.ServiceID, r.ServiceType, r.ID)
}

// A SettingsSubject is something that can have settings. Exactly 1 field must be nonzero.
type SettingsSubject struct {
	Default bool   // whether this is for default settings
	Site    bool   // whether this is for global settings
	Org     *int32 // the org's ID
	User    *int32 // the user's ID
}

func (s SettingsSubject) String() string {
	switch {
	case s.Default:
		return "DefaultSettings"
	case s.Site:
		return "site"
	case s.Org != nil:
		return fmt.Sprintf("org %d", *s.Org)
	case s.User != nil:
		return fmt.Sprintf("user %d", *s.User)
	default:
		return "unknown settings subject"
	}
}

// Settings contains settings for a subject.
type Settings struct {
	ID           int32           // the unique ID of this settings value
	Subject      SettingsSubject // the subject of these settings
	AuthorUserID *int32          // the ID of the user who authored this settings value
	Contents     string          // the raw JSON (with comments and trailing commas allowed)
	CreatedAt    time.Time       // the date when this settings value was created
}

func cmp(a, b string) int {
	switch {
	case a < b:
		return -1
	case b < a:
		return 1
	default:
		return 0
	}
}

// SavedQueryInfo represents information about a saved query that was executed.
type SavedQueryInfo struct {
	// Query is the search query in question.
	Query string

	// LastExecuted is the timestamp of the last time that the search query was
	// executed.
	LastExecuted time.Time

	// LatestResult is the timestamp of the latest-known result for the search
	// query. Therefore, searching `after:<LatestResult>` will return the new
	// search results not yet seen.
	LatestResult time.Time

	// ExecDuration is the amount of time it took for the query to execute.
	ExecDuration time.Duration
}

type SavedQueryIDSpec struct {
	Subject SettingsSubject
	Key     string
}

// ConfigSavedQuery is the JSON shape of a saved query entry in the JSON configuration
// (i.e., an entry in the {"search.savedQueries": [...]} array).
type ConfigSavedQuery struct {
	Key             string  `json:"key,omitempty"`
	Description     string  `json:"description"`
	Query           string  `json:"query"`
	Notify          bool    `json:"notify,omitempty"`
	NotifySlack     bool    `json:"notifySlack,omitempty"`
	UserID          *int32  `json:"userID"`
	OrgID           *int32  `json:"orgID"`
	SlackWebhookURL *string `json:"slackWebhookURL"`
}

func (sq ConfigSavedQuery) Equals(other ConfigSavedQuery) bool {
	a, _ := json.Marshal(sq)
	b, _ := json.Marshal(other)
	return bytes.Equal(a, b)
}

// PartialConfigSavedQueries is the JSON configuration shape, including only the
// search.savedQueries section.
type PartialConfigSavedQueries struct {
	SavedQueries []ConfigSavedQuery `json:"search.savedQueries"`
}

// SavedQuerySpecAndConfig represents a saved query configuration its unique ID.
type SavedQuerySpecAndConfig struct {
	Spec   SavedQueryIDSpec
	Config ConfigSavedQuery
}
