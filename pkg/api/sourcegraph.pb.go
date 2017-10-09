package sourcegraph

import (
	"fmt"
	"time"

	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type TreeEntryType int32

const (
	FileEntry      TreeEntryType = 0
	DirEntry       TreeEntryType = 1
	SymlinkEntry   TreeEntryType = 2
	SubmoduleEntry TreeEntryType = 3
)

var TreeEntryType_name = map[int32]string{
	0: "FileEntry",
	1: "DirEntry",
	2: "SymlinkEntry",
	3: "SubmoduleEntry",
}
var TreeEntryType_value = map[string]int32{
	"FileEntry":      0,
	"DirEntry":       1,
	"SymlinkEntry":   2,
	"SubmoduleEntry": 3,
}

// ListOptions specifies general pagination options for fetching a list of results.
type ListOptions struct {
	PerPage int32 `json:"PerPage,omitempty" url:",omitempty"`
	Page    int32 `json:"Page,omitempty" url:",omitempty"`
}

// ListResponse specifies a general paginated response when fetching a list of results.
type ListResponse struct {
	// Total is the total number of results in the list.
	Total int32 `json:"Total,omitempty" url:",omitempty"`
}

// StreamResponse specifies a paginated response where the total number of results
// that can be returned is too expensive to compute, unbounded, or unknown.
type StreamResponse struct {
	// HasMore is true if there are more results available after the returned page.
	HasMore bool `json:"HasMore,omitempty" url:",omitempty"`
}

// Repo represents a source code repository.
type Repo struct {
	// ID is the unique numeric ID for this repository.
	ID int32 `json:"ID,omitempty"`
	// URI is a normalized identifier for this repository based on its primary clone
	// URL. E.g., "github.com/user/repo".
	URI string `json:"URI,omitempty"`
	// Description is a brief description of the repository.
	Description string `json:"Description,omitempty"`
	// HomepageURL is the URL to the repository's homepage, if any.
	HomepageURL string `json:"HomepageURL,omitempty"`
	// DefaultBranch is the default git branch used (typically "master").
	DefaultBranch string `json:"DefaultBranch,omitempty"`
	// Language is the primary programming language used in this repository.
	Language string `json:"Language,omitempty"`
	// Blocked is whether this repo has been blocked by an admin (and
	// will not be returned via the external API).
	Blocked bool `json:"Blocked,omitempty"`
	// Fork is whether this repository is a fork.
	Fork bool `json:"Fork,omitempty"`
	// StarsCount is the number of users who have starred this repository.
	// Not persisted in DB!
	StarsCount *uint `json:"Stars,omitempty"`
	// ForksCount is the number of forks of this repository that exist.
	// Not persisted in DB!
	ForksCount *uint `json:"Forks,omitempty"`
	// Private is whether this repository is private. Note: this field
	// is currently only used when the repository is hosted on GitHub.
	// All locally hosted repositories should be public. If Private is
	// true for a locally hosted repository, the repository might never
	// be returned.
	Private bool `json:"Private,omitempty"`
	// CreatedAt is when this repository was created. If it represents an externally
	// hosted (e.g., GitHub) repository, the creation date is when it was created at
	// that origin.
	CreatedAt *time.Time `json:"CreatedAt,omitempty"`
	// UpdatedAt is when this repository's metadata was last updated (on its origin if
	// it's an externally hosted repository).
	UpdatedAt *time.Time `json:"UpdatedAt,omitempty"`
	// PushedAt is when this repository's was last (VCS-)pushed to.
	PushedAt *time.Time `json:"PushedAt,omitempty"`
	// IndexedRevision is the revision that the global index is currently based on. It is only used
	// by the indexer to determine if reindexing is necessary. Setting it to nil/null will cause
	// the indexer to reindex the next time it gets triggered for this repository.
	IndexedRevision *string `json:"IndexedRevision,omitempty"`
	// FreezeIndexedRevision, when true, tells the indexer not to
	// update the indexed revision if it is already set. This is a
	// kludge that lets us freeze the indexed repository revision for
	// specific deployments
	FreezeIndexedRevision bool `json:"FreezeIndexedRevision,omitempty"`
}

// GitHubRepoWithDetails represents a GitHub source code repository with additional context
// These types are used for data logging/capturing when a GitHub user signs in to Sourcegraph
type GitHubRepoWithDetails struct {
	URI         string                `json:"URI,omitempty"`
	Fork        bool                  `json:"Fork,omitempty"`
	Private     bool                  `json:"Private,omitempty"`
	CreatedAt   *time.Time            `json:"CreatedAt,omitempty"`
	PushedAt    *time.Time            `json:"PushedAt,omitempty"`
	Languages   []*GitHubRepoLanguage `json:"Languages,omitempty"`
	CommitTimes []*time.Time          `json:"Commits,omitempty"`
	// ErrorFetchingDetails is provided if tracker code receives error
	// responses from GitHub while fetching language or commit details from
	// https://api.github.com/repos/org/name/[languages|commits] URLs
	ErrorFetchingDetails bool `json:"error_fetching_details,omitempty"`
	Skipped              bool `json:"skipped,omitempty"`
}

type GitHubRepoLanguage struct {
	Language string `json:"Language,omitempty"`
	Count    int    `json:"Count,omitempty"`
}

type Contributor struct {
	Login         string `json:"Login,omitempty"`
	AvatarURL     string `json:"AvatarURL,omitempty"`
	Contributions int    `json:"Contributions,omitempty"`
}

// RepoPermissions describes the actions that a user may perform on a
// repo. Currently, the definition of these permissions directly maps
// to GitHub permissions, except for "Pull", which means read access.
type RepoPermissions struct {
	Pull  bool `json:"Pull,omitempty"`
	Push  bool `json:"Push,omitempty"`
	Admin bool `json:"Admin,omitempty"`
}

type RepoListOptions struct {
	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string `json:"Query,omitempty" url:",omitempty"`
	// ListOptions controls pagination.
	ListOptions `json:""`
}

// RepoWebhookOptions is used for enable repository webhook.
type RepoWebhookOptions struct {
	URI string `json:"URI,omitempty"`
}

// RepoRevSpec specifies a repository at a specific commit.
type RepoRevSpec struct {
	Repo int32 `json:"Repo,omitempty"`
	// CommitID is the 40-character SHA-1 of the Git commit ID.
	//
	// Revision specifiers are not allowed here. To resolve a revision
	// specifier (such as a branch name or "master~7"), call
	// Repos.GetCommit.
	CommitID string `json:"CommitID,omitempty"`
}

// RepoSpec specifies a repository.
type RepoSpec struct {
	ID int32 `json:"ID,omitempty"`
}

type RepoList struct {
	Repos []*Repo `json:"Repos,omitempty"`
}

type GitHubReposWithDetailsList struct {
	ReposWithDetails []*GitHubRepoWithDetails `json:"ReposWithDetails,omitempty"`
}

// ReposResolveRevOp specifies a Repos.ResolveRev operation.
type ReposResolveRevOp struct {
	Repo int32 `json:"repo,omitempty"`
	// Rev is a VCS revision specifier, such as a branch or
	// "master~7".
	Rev string `json:"rev,omitempty"`
}

// ResolvedRev is the result of resolving a VCS revision specifier to
// an absolute commit ID.
type ResolvedRev struct {
	// CommitID is the 40-character absolute SHA-1 hex digest of the
	// commit's Git oid.
	CommitID string `json:"CommitID,omitempty"`
}

type URIList struct {
	URIs []string `json:"URIs,omitempty"`
}

type ReposListCommitsOp struct {
	Repo int32                   `json:"Repo,omitempty"`
	Opt  *RepoListCommitsOptions `json:"Opt,omitempty"`
}

type RepoListCommitsOptions struct {
	Head        string `json:"Head,omitempty" url:",omitempty"`
	Base        string `json:"Base,omitempty" url:",omitempty"`
	ListOptions `json:""`
	Path        string `json:"Path,omitempty" url:",omitempty"`
}

type CommitList struct {
	Commits        []*vcs.Commit `json:"Commits,omitempty"`
	StreamResponse `json:""`
}

type ReposListCommittersOp struct {
	Repo int32                      `json:"Repo,omitempty"`
	Opt  *RepoListCommittersOptions `json:"Opt,omitempty"`
}

type RepoListCommittersOptions struct {
	Rev         string `json:"Rev,omitempty"`
	ListOptions `json:""`
}

type CommitterList struct {
	Committers     []*vcs.Committer `json:"Committers,omitempty"`
	StreamResponse `json:""`
}

// UserSpec specifies a user. At least one of Login and UID must be
// nonempty.
type UserSpec struct {
	// UID is a user's UID.
	UID string `json:"UID,omitempty"`
}

// SubmitFormResponse is a response to a user submitting a form (such
// as, e.g., a beta signup form).
type SubmitFormResponse struct {
	// EmailAddress is the email address of the user that submitted the
	// form
	EmailAddress string `json:"EmailAddress,omitempty"`
}

// AuthInfo describes the currently authenticated client and/or user
// (if any).
type AuthInfo struct {
	// UID is the UID of the currently authenticated user (if any).
	UID string `json:"UID,omitempty"`
	// Login is the login of the currently authenticated user (if any).
	Login string `json:"Login,omitempty"`
	// Write is set if the user (if any) has write access on this server.
	Write bool `json:"Write,omitempty"`
	// Admin is set if the user (if any) has admin access on this server.
	Admin bool `json:"Admin,omitempty"`
}

// ExternalToken specifies an auth token of a user for an external host.
type ExternalToken struct {
	// UID is the UID of the user authorized by the token.
	UID string `json:"uid,omitempty"`
	// Host is the external service which granted the token.
	Host string `json:"host,omitempty"`
	// Token is the auth token authorizing a user on an external service.
	Token string `json:"token,omitempty"`
	// Scope lists the permissions the token is entitled to.
	Scope string `json:"scope,omitempty"`
}

// DependencyReferencesOptions specifies options for querying dependency references.
type DependencyReferencesOptions struct {
	Language        string // e.g. "go"
	RepoID          int32  // repository whose file:line:character describe the symbol of interest
	CommitID        string
	File            string
	Line, Character int

	// Limit specifies the number of dependency references to return.
	Limit int // e.g. 20
}

type DependencyReferences struct {
	References []*DependencyReference
	Location   lspext.SymbolLocationInformation
}

// DependencyReference effectively says that RepoID has made a reference to a
// dependency.
type DependencyReference struct {
	DepData map[string]interface{} // includes additional information about the dependency, e.g. whether or not it is vendored for Go
	RepoID  int32                  // the repository who made the reference to the dependency.
	Hints   map[string]interface{} // hints which should be passed to workspace/xreferences in order to more quickly find the definition.
}

func (d *DependencyReference) String() string {
	return fmt.Sprintf("DependencyReference{DepData: %v, RepoID: %v, Hints: %v}", d.DepData, d.RepoID, d.Hints)
}

// UserEvent encodes any user initiated event on the local instance.
type UserEvent struct {
	Type    string `json:"Type,omitempty"`
	UID     string `json:"UID,omitempty"`
	Service string `json:"Service,omitempty"`
	Method  string `json:"Method,omitempty"`
	Result  string `json:"Result,omitempty"`
	// CreatedAt holds the time when this event was logged.
	CreatedAt *time.Time `json:"CreatedAt,omitempty"`
	Message   string     `json:"Message,omitempty"`
	// Version holds the release version of the Sourcegraph binary.
	Version string `json:"Version,omitempty"`
	// URL holds the http request url.
	URL string `json:"URL,omitempty"`
}

// Event is any action logged on a Sourcegraph instance.
type Event struct {
	// Type specifies the action type, eg. "AccountCreate" or "AddRepo".
	Type string `json:"Type,omitempty"`
	// UserID is the unique identifier of a user on a Sourcegraph instance.
	// It is constructed as "login@short-client-id", where short-client-id
	// is the first 6 characters of this sourcegraph instance's public key
	// fingerprint (i.e. it's ClientID).
	UserID string `json:"UserID,omitempty"`
	// DeviceID is the unique identifier of an anonymous user on a Sourcegraph
	// instance.
	DeviceID string `json:"DeviceID,omitempty"`
	// Timestamp records the instant when this event was logged.
	Timestamp *time.Time `json:"Timestamp,omitempty"`
	// UserProperties holds metadata relating to user who performed this
	// action, eg. "Email".
	UserProperties map[string]string `json:"UserProperties,omitempty"`
	// EventProperties holds metadata relating to the action logged by
	// this event, eg. for "AddRepo" event, a property is "Source" which
	// specifies if the repo is local or mirrored.
	EventProperties map[string]string `json:"EventProperties,omitempty"`
}

// EventList is a list of logged Sourcegraph events.
type EventList struct {
	// Events holds the list of events.
	Events []*Event `json:"Events,omitempty"`
	// Version holds the release version of the Sourcegraph binary.
	Version string `json:"Version,omitempty"`
	// AppURL holds the base URL of the Sourcegraph app.
	AppURL string `json:"AppURL,omitempty"`
}

// OrgsList is a list of GitHub organizations for a given user
type OrgsList struct {
	Orgs []*GitHubOrg `json:"Orgs,omitempty"`
}

// GitHubOrg holds the result of an org for Orgs.ListOrgs
type GitHubOrg struct {
	Login         string `json:"Login"`
	ID            int32  `json:"ID"`
	AvatarURL     string `json:"AvatarURL,omitempty"`
	Name          string `json:"Name,omitempty"`
	Blog          string `json:"Blog,omitempty"`
	Location      string `json:"Location,omitempty"`
	Email         string `json:"Email,omitempty"`
	Description   string `json:"Description,omitempty"`
	Collaborators int32  `json:"Collaborators,omitempty"`
}

// UserInvite holds the result of an invite for Orgs.InviteUser
type UserInvite struct {
	UserLogin string `json:"UserID,omitempty"`
	UserEmail string `json:"UserEmail,omitempty"`
	// OrgID is a string representation of the organiztion's unique GitHub ID (e.g., for Sourcegraph: "3979584")
	OrgID    string     `json:"OrgID,omitempty"`
	OrgLogin string     `json:"OrgName,omitempty"`
	SentAt   *time.Time `json:"SentAt,omitempty"`
	URI      string     `json:"URI,omitempty"`
}
type UserInviteResponse int

const (
	InviteSuccess UserInviteResponse = iota
	InviteMissingEmail
	InviteError
)

// User represents a registered user.
type User struct {
	ID          int32     `json:"ID,omitempty"`
	Auth0ID     string    `json:"auth0ID,omitempty"`
	Email       string    `json:"email,omitempty"`
	Username    string    `json:"username,omitempty"`
	DisplayName string    `json:"displayName,omitempty"`
	AvatarURL   *string   `json:"avatarURL,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// OrgRepo represents a repo that exists on a native client's filesystem, but
// does not necessarily have its contents cloned to a remote Sourcegraph server.
type OrgRepo struct {
	ID        int32
	RemoteURI string
	OrgID     int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ThreadLines struct {
	HTMLBefore               string `json:"HTMLBefore,omitempty"`
	HTML                     string `json:"HTML,omitempty"`
	HTMLAfter                string `json:"HTMLAfter,omitempty"`
	TextBefore               string `json:"TextBefore,omitempty"`
	Text                     string `json:"Text,omitempty"`
	TextAfter                string `json:"TextAfter,omitempty"`
	TextSelectionRangeStart  int32  `json:"TextSelectionRangeStart,omitempty"`
	TextSelectionRangeLength int32  `json:"TextSelectionRangeLength,omitempty"`
}

type Thread struct {
	ID             int32        `json:"ID,omitempty"`
	OrgRepoID      int32        `json:"OrgRepoID,omitempty"`
	File           string       `json:"File,omitempty"`
	Revision       string       `json:"Revision,omitempty"`
	Branch         *string      `json:"Branch,omitempty"`
	StartLine      int32        `json:"StartLine,omitempty"`
	EndLine        int32        `json:"EndLine,omitempty"`
	StartCharacter int32        `json:"StartCharacter,omitempty"`
	EndCharacter   int32        `json:"EndCharacter,omitempty"`
	RangeLength    int32        `json:"RangeLength,omitempty"`
	CreatedAt      time.Time    `json:"CreatedAt,omitempty"`
	UpdatedAt      time.Time    `json:"UpdatedAt,omitempty"`
	ArchivedAt     *time.Time   `json:"ArchivedAt,omitempty"`
	AuthorUserID   string       `json:"AuthorUserID,omitempty"`
	Lines          *ThreadLines `json:"ThreadLines,omitempty"`
}

type Comment struct {
	ID        int32     `json:"ID,omitempty"`
	ThreadID  int32     `json:"ThreadID,omitempty"`
	Contents  string    `json:"Contents,omitempty"`
	CreatedAt time.Time `json:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"UpdatedAt,omitempty"`
	// TODO(nick): remove author fields and resolve data using author id instead
	AuthorName   string `json:"AuthorName,omitempty"`
	AuthorEmail  string `json:"AuthorEmail,omitempty"`
	AuthorUserID string `json:"AuthorUserID,omitempty"`
}

// SharedItem represents a shared thread or comment. Note that a code snippet
// is also just a thread.
type SharedItem struct {
	ULID         string `json:"ULID"`
	AuthorUserID string `json:"AuthorUserID"`

	// Only one of the below is present.
	ThreadID  *int32 `json:"ThreadID,omitempty"`
	CommentID *int32 `json:"CommentID,omitempty"`
}

type Org struct {
	ID          int32     `json:"ID"`
	Name        string    `json:"Name,omitempty"`
	DisplayName *string   `json:"DisplayName,omitempty"`
	CreatedAt   time.Time `json:"CreatedAt,omitempty"`
	UpdatedAt   time.Time `json:"UpdatedAt,omitempty"`
}

type OrgMember struct {
	ID          int32     `json:"ID"`
	OrgID       int32     `json:"OrgID"`
	UserID      string    `json:"UserID"`
	Username    string    `json:"Username,omitempty"`
	Email       string    `json:"Email,omitempty"`
	DisplayName string    `json:"DisplayName,omitempty"`
	AvatarURL   *string   `json:"AvatarURL,omitempty"`
	CreatedAt   time.Time `json:"CreatedAt,omitempty"`
	UpdatedAt   time.Time `json:"UpdatedAt,omitempty"`
}

type UserTag struct {
	ID     int32  `json:"ID"`
	UserID int32  `json:"UserID"`
	Name   string `json:"Name,omitempty"`
}

type OrgTag struct {
	ID    int32  `json:"ID"`
	OrgID int32  `json:"OrgID"`
	Name  string `json:"Name,omitempty"`
}
