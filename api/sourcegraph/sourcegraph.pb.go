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

type ReposUpdateOp_BoolType int32

const (
	ReposUpdateOp_NONE  ReposUpdateOp_BoolType = 0
	ReposUpdateOp_TRUE  ReposUpdateOp_BoolType = 1
	ReposUpdateOp_FALSE ReposUpdateOp_BoolType = 2
)

var ReposUpdateOp_BoolType_name = map[int32]string{
	0: "NONE",
	1: "TRUE",
	2: "FALSE",
}
var ReposUpdateOp_BoolType_value = map[string]int32{
	"NONE":  0,
	"TRUE":  1,
	"FALSE": 2,
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
	// Owner is the repository owner (user or organizatin) of the repository. (For
	// example, for "github.com/user/repo", the owner is "user".)
	Owner string `json:"Owner,omitempty"`
	// Name is the base name (the final path component) of the repository, typically
	// the name of the directory that the repository would be cloned into. (For
	// example, for git://example.com/foo.git, the name is "foo".)
	Name string `json:"Name,omitempty"`
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
	// Canonical is true if a repository should be treated as the
	// canonical instance of source code regardless of the value of
	// Fork. Sometimes forks become the canonical repository while the
	// original repository is abandoned. If Canonical is false, then
	// the repository is canonical if and only if Fork is false.
	Canonical bool `json:"Canonical,omitempty"`
}

// GitHubRepoWithDetails represents a GitHub source code repository with additional context
// These types are used for data logging/capturing when a GitHub user signs in to Sourcegraph
type GitHubRepoWithDetails struct {
	URI         string                `json:"URI,omitempty"`
	Owner       string                `json:"Owner,omitempty"`
	Name        string                `json:"Name,omitempty"`
	Fork        bool                  `json:"Fork,omitempty"`
	Private     bool                  `json:"Private,omitempty"`
	CreatedAt   *time.Time            `json:"CreatedAt,omitempty"`
	Languages   []*GitHubRepoLanguage `json:"Languages,omitempty"`
	CommitTimes []*time.Time          `json:"Commits,omitempty"`
	// ErrorFetchingDetails is provided if tracker code receives error
	// responses from GitHub while fetching language or commit details from
	// https://api.github.com/repos/org/name/[languages|commits] URLs
	ErrorFetchingDetails bool `json:"error_fetching_details,omitempty"`
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
	// RemoteOnly makes the endpoint return repositories hosted on
	// GitHub that the currently authenticated user has access to.
	//
	// When RemoteOnly is true, the only other option field that is
	// obeyed is Type (pagination is ignored, too), and returned
	// repositories do not have valid IDs. If you want to get the ID of
	// a repository, fetch it individually.
	RemoteOnly bool `json:"RemoteOnly,omitempty" url:",omitempty"`
	// RemoteSearch should be used in conjunction with the Query field.
	// If true, it will issue an external query to remote code host
	// search APIs and augment the list of returned results with
	// repositories that exist on external hosts but not yet on
	// Sourcegraph. RemoteSearch is ignored if there is no authenticated
	// user.
	RemoteSearch bool `json:"RemoteSearch,omitempty" url:",omitempty"`
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

// ReposUpdateOp is an operation to update a repository's metadata.
type ReposUpdateOp struct {
	// Repo is the repository to update.
	Repo int32 `json:"Repo,omitempty"`
	// URI, if non-empty, is the updated value of the URI
	URI string `json:"URI,omitempty"`
	// Owner, if non-empty, is the updated value of the owner.
	Owner string `json:"Owner,omitempty"`
	// Name, if non-empty, is the updated value of the name.
	Name string `json:"Name,omitempty"`
	// Description, if non-empty, is the updated value of the description.
	Description string `json:"Description,omitempty"`
	// HomepageURL, if non-empty, is the updated value of the homepage URL.
	HomepageURL string `json:"HomepageURL,omitempty"`
	// DefaultBranch, if non-empty, is the updated value of the default branch.
	DefaultBranch string `json:"DefaultBranch,omitempty"`
	// Language, if non-empty, is the updated value of the language.
	Language string `json:"Language,omitempty"`
	// Blocked, if non-empty, updates whether this repository is blocked.
	Blocked ReposUpdateOp_BoolType `json:"Blocked,omitempty"`
	// Fork, if non-empty, updates whether this repository is a fork.
	Fork ReposUpdateOp_BoolType `json:"Fork,omitempty"`
	// Private, if non-empty, updates whether this repository is private.
	Private ReposUpdateOp_BoolType `json:"Private,omitempty"`
	// IndexedRevision is the revision that the global index is currently based on.
	IndexedRevision string `json:"IndexedRevision,omitempty"`
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

// EmailAddr is an email address associated with a user.
type EmailAddr struct {
	// the email address (case-insensitively compared in the DB and API)
	Email string `json:"Email,omitempty"`
	// whether this email address has been verified
	Verified bool `json:"Verified,omitempty"`
	// indicates this is the user's primary email (only 1 email can be primary per user)
	Primary bool `json:"Primary,omitempty"`
	// whether Sourcegraph inferred via public data that this is an email for the user
	Guessed bool `json:"Guessed,omitempty"`
	// indicates that this email should not be associated with the user (even if guessed in the future)
	Blacklisted bool `json:"Blacklisted,omitempty"`
}

type UserList struct {
	Users []*User `json:"Users,omitempty"`
}

// User represents a registered user.
type User struct {
	// UID is the numeric primary key for a user.
	UID string `json:"UID"`
	// Login is the user's username.
	Login string `json:"Login"`
	// Name is the (possibly empty) full name of the user.
	Name string `json:"Name,omitempty"`
	// IsOrganization is whether this user represents an organization.
	IsOrganization bool `json:"IsOrganization,omitempty"`
	// AvatarURL is the URL to an avatar image specified by the user.
	AvatarURL string `json:"AvatarURL,omitempty"`
	// Location is the user's physical location.
	Location string `json:"Location,omitempty"`
	// Company is the user's company.
	Company string `json:"Company,omitempty"`
	// HomepageURL is the user's homepage or blog URL.
	HomepageURL string `json:"HomepageURL,omitempty"`
	// Disabled is whether the user account is disabled.
	Disabled bool `json:"Disabled,omitempty"`
	// Admin is whether the user is a site admin for the site.
	Admin bool `json:"Admin,omitempty"`
	// Betas is a list of betas which the user is enrolled in. A user may be
	// granted access to any beta string listed in both:
	//
	//  pkg/betautil/betautil.go
	//  app/web_modules/sourcegraph/util/betautil.js
	//
	// Only admin users may set this field.
	Betas []string `json:"Betas,omitempty"`
	// Write is whether the user has write access for the site.
	Write bool `json:"Write,omitempty"`
	// RegisteredAt is the date that the user registered. If the user has not
	// registered (i.e., we have processed their repos but they haven't signed into
	// Sourcegraph), it is null.
	RegisteredAt *time.Time `json:"RegisteredAt,omitempty"`
}

// UserSpec specifies a user. At least one of Login and UID must be
// nonempty.
type UserSpec struct {
	// UID is a user's UID.
	UID string `json:"UID,omitempty"`
}

type EmailAddrList struct {
	EmailAddrs []*EmailAddr `json:"EmailAddrs,omitempty"`
}

// UpdateEmailsOp represents options for updating a user's email addresses.
type UpdateEmailsOp struct {
	// UserSpec is the user whose email addresses are to be updated.
	UserSpec UserSpec `json:"UserSpec"`
	// Add is a list of email addresses to add, if they do not already exist.
	Add *EmailAddrList `json:"Add,omitempty"`
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

// RepoTreeGetOptions specifies options for (RepoTreeService).Get.
type RepoTreeGetOptions struct {
	ContentsAsString bool `json:"ContentsAsString,omitempty" url:",omitempty"`
	GetFileOptions   `json:""`
}

// GetFileOptions specifies options for GetFileWithOptions.
type GetFileOptions struct {
	// line or byte range to fetch (can't set both line *and* byte range)
	FileRange `json:""`
	// EntireFile is whether the entire file contents should be returned. If true,
	// Start/EndLine and Start/EndBytes are ignored.
	EntireFile bool `json:"EntireFile,omitempty" url:",omitempty"`
	// ExpandContextLines is how many lines of additional output context to include (if
	// Start/EndLine and Start/EndBytes are specified). For example, specifying 2 will
	// expand the range to include 2 full lines before the beginning and 2 full lines
	// after the end of the range specified by Start/EndLine and Start/EndBytes.
	ExpandContextLines int32 `json:"ExpandContextLines,omitempty" url:",omitempty"`
	// FullLines is whether a range that includes partial lines should be extended to
	// the nearest line boundaries on both sides. It is only valid if StartByte and
	// EndByte are specified.
	FullLines bool `json:"FullLines,omitempty" url:",omitempty"`
	// Recursive only applies if the returned entry is a directory. It will
	// return the full file tree of the host repository, recursing into all
	// sub-directories.
	Recursive bool `json:"Recursive,omitempty" url:",omitempty"`
	// RecurseSingleSubfolderLimit only applies if the returned entry is a directory.
	// If nonzero, it will recursively find and include all singleton sub-directory chains,
	// up to a limit of RecurseSingleSubfolderLimit.
	RecurseSingleSubfolderLimit int32 `json:"RecurseSingleSubfolderLimit,omitempty" url:",omitempty"`
}

type RepoTreeGetOp struct {
	Entry TreeEntrySpec       `json:"Entry"`
	Opt   *RepoTreeGetOptions `json:"Opt,omitempty"`
}

type RepoTreeListOp struct {
	Rev RepoRevSpec `json:"Rev"`
}

type RepoTreeListResult struct {
	Files []string `json:"Files,omitempty"`
}

type VCSSearchResultList struct {
	SearchResults []*vcs.SearchResult `json:"SearchResults,omitempty"`
	ListResponse  `json:""`
}

// TreeEntry is a file or directory in a repository.
type TreeEntry struct {
	*BasicTreeEntry `json:""`
	*FileRange      `json:""`
	ContentsString  string `json:"ContentsString,omitempty"`
}

type BasicTreeEntry struct {
	Name     string            `json:"Name,omitempty"`
	Type     TreeEntryType     `json:"Type,omitempty"`
	CommitID string            `json:"CommitID,omitempty"`
	Contents []byte            `json:"Contents,omitempty"`
	Entries  []*BasicTreeEntry `json:"Entries,omitempty"`
}

type TreeEntrySpec struct {
	RepoRev RepoRevSpec `json:"RepoRev"`
	Path    string      `json:"Path,omitempty"`
}

// FileRange is a line and byte range in a file.
type FileRange struct {
	// start of line range
	StartLine int64 `json:"StartLine,omitempty" url:",omitempty"`
	// end of line range
	EndLine int64 `json:"EndLine,omitempty" url:",omitempty"`
	// start of byte range
	StartByte int64 `json:"StartByte,omitempty" url:",omitempty"`
	// end of byte range
	EndByte int64 `json:"EndByte,omitempty" url:",omitempty"`
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

// OrgListOptions holds the options for listing organization details
type OrgListOptions struct {
	OrgName  string `json:"OrgName,omitempty"`
	Username string `json:"Username,omitempty"`
	OrgID    string `json:"OrgID,omitempty"`
	// ListOptions controls pagination.
	ListOptions `json:""`
}

// OrgsList is a list of GitHub organizations for a given user
type OrgsList struct {
	Orgs []*Org `json:"Orgs,omitempty"`
}

// Org holds the result of an org for Orgs.ListOrgs
type Org struct {
	Login       string `json:"Login"`
	ID          int32  `json:"ID"`
	AvatarURL   string `json:"AvatarURL,omitempty"`
	Name        string `json:"Name,omitempty"`
	Blog        string `json:"Blog,omitempty"`
	Location    string `json:"Location,omitempty"`
	Email       string `json:"Email,omitempty"`
	Description string `json:"Description,omitempty"`
}

// OrgMembersList is a list of GitHub organization members for an organization
type OrgMembersList struct {
	OrgMembers []*OrgMember `json:"OrgMembers,omitempty"`
}

// OrgMember holds the result of an org member for Orgs.ListOrgMembers
type OrgMember struct {
	Login           string      `json:"Login"`
	ID              int32       `json:"ID"`
	AvatarURL       string      `json:"AvatarURL,omitempty"`
	Email           string      `json:"Email,omitempty"`
	SourcegraphUser bool        `json:"SourcegraphUser,omitempty"`
	CanInvite       bool        `json:"CanInvite,omitempty"`
	Invite          *UserInvite `json:"Invite,omitempty"`
}

// UserInvite holds the result of an invite for Orgs.InviteUser
type UserInvite struct {
	UserID    string     `json:"UserID,omitempty"`
	UserEmail string     `json:"UserEmail,omitempty"`
	OrgID     string     `json:"OrgID,omitempty"`
	OrgName   string     `json:"OrgName,omitempty"`
	SentAt    *time.Time `json:"SentAt,omitempty"`
	URI       string     `json:"URI,omitempty"`
}

type UserInviteResponse struct {
	OrgName string `json:"OrgName,omitempty"`
	OrgID   string `json:"OrgID,omitempty"`
}
