package sourcegraph

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/srclib/graph"
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

// ServiceType indicates which service is running on an origin. A
// repo whose origin service is GitHub, for example, should be
// accessed using a GitHub API client.
//
// If there are multiple API versions for a service, separate
// entries may be added per API version. In that case, the
// APIBaseURL may need to differ as well. The API client code is
// responsible for handling these cases.
type Origin_ServiceType int32

const (
	// GitHub indicates that the origin is GitHub.com or a GitHub
	// Enterprise server. If the latter, the origin base URL indicates
	// the URL to the GitHub Enterprise server's API.
	Origin_GitHub Origin_ServiceType = 0
)

var Origin_ServiceType_name = map[int32]string{
	0: "GitHub",
}
var Origin_ServiceType_value = map[string]int32{
	"GitHub": 0,
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

// Origin represents the origin of a resource that canonically lives
// on an external service (e.g., a repo hosted on GitHub).
type Origin struct {
	// ID is an identifier for the resource on its origin
	// service. Although numeric IDs are used on many services (GitHub
	// and Bitbucket, for example), this field is a string so that it
	// supports non-numeric IDs (which are used on Google Cloud
	// Platform and probably other services that Sourcegraph might
	// support in the future).
	//
	// If the ID is numeric, this string is the base-10 string
	// representation of the numeric ID (e.g., "1234"), with no
	// leading 0s.
	ID string `json:"ID,omitempty"`
	// Service is the type service that the resource canonically lives
	// on. It is used to determine which API client should be used to
	// access it on the origin service (e.g., GitHub vs. Bitbucket).
	Service Origin_ServiceType `json:""`
	// APIBaseURL is the base URL to the API of the origin service for
	// the resource. (E.g., "https://api.github.com" for
	// GitHub.com-hosted repos.)
	APIBaseURL string `json:"APIBaseURL,omitempty"`
}

// CombinedStatus is the combined status (i.e., incorporating statuses from all
// contexts) of the repository at a specific rev.
type CombinedStatus struct {
	// Rev is the revision that this status describes. It is set mutually exclusive with CommitID.
	Rev string `json:"Rev,omitempty"`
	// CommitID is the full commit ID of the commit this status describes. It is set mutually exclusively with Rev.
	CommitID string `json:"CommitID,omitempty"`
	// State is the combined status of the repository. Possible values are: failure,
	// pending, or success.
	State string `json:"State,omitempty"`
	// Statuses are the statuses for each context.
	Statuses []*RepoStatus `json:"Statuses,omitempty"`
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

// RepoConfig describes a repository's config. This config is
// Sourcegraph-specific and is persisted locally.
type RepoConfig struct {
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
	// HTTPCloneURL is the HTTPS clone URL of the repository (or the HTTP clone URL, if
	// no HTTPS clone URL is available).
	HTTPCloneURL string `json:"HTTPCloneURL,omitempty"`
	// SSHCloneURL is the SSH clone URL if the repository, if any.
	SSHCloneURL string `json:"SSHCloneURL,omitempty"`
	// HomepageURL is the URL to the repository's homepage, if any.
	HomepageURL string `json:"HomepageURL,omitempty"`
	// DefaultBranch is the default git branch used (typically "master").
	DefaultBranch string `json:"DefaultBranch,omitempty"`
	// Language is the primary programming language used in this repository.
	Language string `json:"Language,omitempty"`
	// Blocked is whether this repo has been blocked by an admin (and
	// will not be returned via the external API).
	Blocked bool `json:"Blocked,omitempty"`
	// Deprecated repositories are labeled as such and hidden from global search
	// results.
	Deprecated bool `json:"Deprecated,omitempty"`
	// Fork is whether this repository is a fork.
	Fork bool `json:"Fork,omitempty"`
	// Mirror indicates whether this repo's canonical location is on
	// another server. Mirror repos track their upstream. If this repo
	// canonically lives on a repo hosting that can supply additional
	// metadata (such as GitHub), the Origin field should be set.
	Mirror bool `json:"Mirror,omitempty"`
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
	// VCSSyncedAt is when this repository's VCS data was last synced
	// with the upstream. This field is only populated for mirror
	// repositories.
	VCSSyncedAt *time.Time `json:"VCSSyncedAt,omitempty"`
	// Origin describes the repo's canonical location. It is only
	// populated for mirror repos; for non-mirror repos, it is null.
	Origin *Origin `json:"Origin,omitempty"`
	// Permissions describes the actions that the current user (who
	// retrieved this repository from the API) may perform on the
	// repository. For public repositories retrieved by
	// non-contributors, the Permissions field may be null and the
	// Pull permission is implied (because it's a public repository).
	//
	// TODO(sqs): Currently these map directly to the user's GitHub
	// permissions for GitHub repositories; all other repositories
	// report that all users have the Pull permission only.
	Permissions *RepoPermissions `json:"Permissions,omitempty"`
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
	// Name filters the repository list by name.
	Name string `json:"Name,omitempty" url:",omitempty"`
	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string `json:"Query,omitempty" url:",omitempty"`
	// URIs specifies a set of repository URIs to list.
	URIs []string `json:"URIs,omitempty" url:",comma,omitempty"`
	// Sort determines how the returned list of repositories is sorted.
	Sort string `json:"Sort,omitempty" url:",omitempty"`
	// Direction determines the sort direction.
	Direction string `json:"Direction,omitempty" url:",omitempty"`
	// NoFork excludes forks from the list of returned repositories.
	NoFork bool `json:"NoFork,omitempty" url:",omitempty"`
	// Type of repositories to list. Possible values are currently
	// ones supported by the GitHub API, including: all, owner,
	// public, private, member. Default is "all".
	Type string `json:"Type,omitempty" url:",omitempty"`
	// Owner filters the list of repositories to those with the
	// specified owner.
	Owner string `json:"Owner,omitempty" url:",omitempty"`
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

// RepoStatus is the status of the repository at a specific rev (in a single
// context).
type RepoStatus struct {
	// State is the current status of the repository. Possible values are: pending,
	// success, error, or failure.
	State string `json:"State,omitempty"`
	// TargetURL is the URL of the page representing this status. It will be linked
	// from the UI to allow users to see the source of the status.
	TargetURL string `json:"TargetURL,omitempty"`
	// Description is a short, high-level summary of the status.
	Description string `json:"Description,omitempty"`
	// A string label to differentiate this status from the statuses of other systems.
	Context   string    `json:"Context,omitempty"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

type RepoStatusList struct {
	RepoStatuses []*RepoStatus `json:"RepoStatuses,omitempty"`
}

type RepoStatusesCreateOp struct {
	Repo   RepoRevSpec `json:"Repo"`
	Status RepoStatus  `json:"Status"`
}

type RepoList struct {
	Repos []*Repo `json:"Repos,omitempty"`
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

type RepoResolveOp struct {
	// Path is some repo path, such as "github.com/user/repo".
	Path string `json:"path,omitempty"`
	// Remote controls the behavior when Resolve locates a remote
	// repository that is not (yet) associated with an existing local
	// repository. If Remote is false (the default), then a NotFound
	// error is returned in that case. If Remote is true, then no
	// error is returned; the RepoResolution's Repo field will be
	// empty, but some metadata about the remote repository may be
	// provided.
	Remote bool `json:"remote,omitempty"`
}

// RepoResolution is the result of resolving a repo using
// Repos.Resolve.
type RepoResolution struct {
	// ID is the ID of the local repo (either a locally hosted repo,
	// or a locally added mirror).
	Repo int32 `json:"Repo,omitempty"`
	// CanonicalPath is the canonical repo path of the local repo
	// (with the canonical casing, etc.). Clients should generally
	// redirect the user to the canonical repo path if users access a
	// repo by a non-canonical path.
	CanonicalPath string `json:"CanonicalPath,omitempty"`
	// RemoteRepo holds metadata about the repo that exists on a
	// remote service (such as GitHub).
	RemoteRepo *Repo `json:"RemoteRepo,omitempty"`
}

// SrclibDataVersion specifies a srclib store version.
type SrclibDataVersion struct {
	CommitID      string `json:"CommitID,omitempty"`
	CommitsBehind int32  `json:"CommitsBehind,omitempty"`
}

type ReposCreateOp struct {
	// Types that are valid to be assigned to Op:
	//	*ReposCreateOp_New
	//	*ReposCreateOp_FromGitHubID
	//	*ReposCreateOp_Origin
	Op isReposCreateOp_Op
}

type isReposCreateOp_Op interface {
	isReposCreateOp_Op()
}

type ReposCreateOp_New struct {
	New *ReposCreateOp_NewRepo
}
type ReposCreateOp_FromGitHubID struct {
	FromGitHubID int32
}
type ReposCreateOp_Origin struct {
	Origin *Origin
}

func (*ReposCreateOp_New) isReposCreateOp_Op()          {}
func (*ReposCreateOp_FromGitHubID) isReposCreateOp_Op() {}
func (*ReposCreateOp_Origin) isReposCreateOp_Op()       {}

func (m *ReposCreateOp) GetOp() isReposCreateOp_Op {
	if m != nil {
		return m.Op
	}
	return nil
}

func (m *ReposCreateOp) GetNew() *ReposCreateOp_NewRepo {
	if x, ok := m.GetOp().(*ReposCreateOp_New); ok {
		return x.New
	}
	return nil
}

func (m *ReposCreateOp) GetFromGitHubID() int32 {
	if x, ok := m.GetOp().(*ReposCreateOp_FromGitHubID); ok {
		return x.FromGitHubID
	}
	return 0
}

func (m *ReposCreateOp) GetOrigin() *Origin {
	if x, ok := m.GetOp().(*ReposCreateOp_Origin); ok {
		return x.Origin
	}
	return nil
}

type ReposCreateOp_NewRepo struct {
	// URI is the desired URI of the new repository.
	URI string `json:"URI,omitempty"`
	// CloneURL is the clone URL of the repository for mirrored
	// repositories. If blank, a new hosted repository is created
	// (i.e., a repo whose origin is on the server). If Mirror is
	// true, a clone URL must be provided.
	CloneURL string `json:"CloneURL,omitempty"`
	// DefaultBranch is the repository's default Git branch.
	DefaultBranch string `json:"DefaultBranch,omitempty"`
	// Mirror is a boolean value indicating whether the newly created
	// repository should be a mirror. Mirror repositories are
	// periodically updated to track their upstream (which is
	// specified using the CloneURL field of this message).
	Mirror bool `json:"Mirror,omitempty"`
	// Description is the description of the repository.
	Description string `json:"Description,omitempty"`
	// Language is the primary programming language of the repository.
	Language string `json:"Language,omitempty"`
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
	// HTTPCloneURL, if non-empty, is the updated value of the HTTP clone URL.
	HTTPCloneURL string `json:"HTTPCloneURL,omitempty"`
	// SSHCloneURL, if non-empty, is the updated value of the SSH clone URL.
	SSHCloneURL string `json:"SSHCloneURL,omitempty"`
	// HomepageURL, if non-empty, is the updated value of the homepage URL.
	HomepageURL string `json:"HomepageURL,omitempty"`
	// DefaultBranch, if non-empty, is the updated value of the default branch.
	DefaultBranch string `json:"DefaultBranch,omitempty"`
	// Language, if non-empty, is the updated value of the language.
	Language string `json:"Language,omitempty"`
	// Origin is data about the repository origin (e.g., GitHub).
	Origin *Origin `json:"Origin,omitempty"`
	// Blocked, if non-empty, updates whether this repository is blocked.
	Blocked ReposUpdateOp_BoolType `json:"Blocked,omitempty"`
	// Deprecated, if non-empty, updates whether this repository is deprecated.
	Deprecated ReposUpdateOp_BoolType `json:"Deprecated,omitempty"`
	// Fork, if non-empty, updates whether this repository is a fork.
	Fork ReposUpdateOp_BoolType `json:"Fork,omitempty"`
	// Mirror, if non-empty, updates whether this repository is a mirror.
	Mirror ReposUpdateOp_BoolType `json:"Mirror,omitempty"`
	// Private, if non-empty, updates whether this repository is private.
	Private ReposUpdateOp_BoolType `json:"Private,omitempty"`
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

type ReposListBranchesOp struct {
	Repo int32                    `json:"Repo,omitempty"`
	Opt  *RepoListBranchesOptions `json:"Opt,omitempty"`
}

type RepoListBranchesOptions struct {
	IncludeCommit     bool   `json:"IncludeCommit,omitempty"`
	BehindAheadBranch string `json:"BehindAheadBranch,omitempty"`
	ContainsCommit    string `json:"ContainsCommit,omitempty"`
	ListOptions       `json:""`
}

type BranchList struct {
	Branches       []*vcs.Branch `json:"Branches,omitempty"`
	StreamResponse `json:""`
}

type ReposListTagsOp struct {
	Repo int32                `json:"Repo,omitempty"`
	Opt  *RepoListTagsOptions `json:"Opt,omitempty"`
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

type RepoListTagsOptions struct {
	ListOptions `json:""`
}

type TagList struct {
	Tags           []*vcs.Tag `json:"Tags,omitempty"`
	StreamResponse `json:""`
}

type MirrorReposRefreshVCSOp struct {
	Repo int32 `json:"Repo,omitempty"`
	// AsUser is the user whose auth token will be used for refreshing this
	// mirror repo. This can be used when refreshing a private repo mirror.
	AsUser *UserSpec `json:"AsUser,omitempty"`
}

// RemoteRepo is a repo canonically stored on an external host, and
// possibly mirrored on the local instance.
type RemoteRepo struct {
	// GitHubID is the repo's GitHub repository ID.
	GitHubID int32 `json:"GitHubID,omitempty"`
	// Owner is the login or org name of the repo's owner ("foo" in
	// github.com/foo/bar).
	Owner string `json:"Owner,omitempty"`
	// OwnerIsOrg is true if the repo's owner is an org (not a user).
	OwnerIsOrg bool `json:"OwnerIsOrg,omitempty"`
	// Name is the repo's name ("bar" in github.com/foo/bar).
	Name string `json:"Name,omitempty"`
	// VCS is "git".
	VCS string `json:"VCS,omitempty"`
	// HTTPCloneURL is the repo's HTTP (preferably HTTPS) clone URL.
	HTTPCloneURL string `json:"HTTPCloneURL,omitempty"`
	// DefaultBranch is the default Git branch for the repo.
	DefaultBranch string `json:"DefaultBranch,omitempty"`
	// Description is the repo's description from GitHub.
	Description string `json:"Description,omitempty"`
	// Language is the repo's primary programming language, as
	// reported by GitHub.
	Language string `json:"Language,omitempty"`
	// UpdatedAt is the date of the most recent update (push or
	// metadata edit) to the repo on GitHub.
	UpdatedAt *time.Time `json:"UpdatedAt,omitempty"`
	// PushedAt is the date of the most recent git push to the repo.
	PushedAt *time.Time `json:"PushedAt,omitempty"`
	// Private is true for private repos.
	Private bool `json:"Private,omitempty"`
	// Fork is true for repos that were forked from another repo using
	// GitHub's "fork" operation.
	Fork bool `json:"Fork,omitempty"`
	// Mirror is true for mirror repos (e.g., Apache Foundation
	// open-source repo mirrors on GitHub.com).
	Mirror bool `json:"Mirror,omitempty"`
	// Stars is the number of stargazers of the GitHub repo.
	Stars int32 `json:"Stars,omitempty"`
	// Permissions describes the actions that the current user may
	// perform on this remote repository on the remote service it came
	// from. These permissions currently map directly to GitHub
	// permissions but will be generalized/modified in the future to
	// support more repo hosting services.
	Permissions *RepoPermissions `json:"Permissions,omitempty"`
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

// BetaRegistration represents the user information needed to register for the
// beta program.
type BetaRegistration struct {
	// Email is the primary email address for the user. It is only used if the
	// user does not already have an email address set in their account, else
	// this field is no-op.
	Email string `json:"Email,omitempty"`
	// FirstName is the first name of the user.
	FirstName string `json:"FirstName,omitempty"`
	// LastName is the last name of the user.
	LastName string `json:"LastName,omitempty"`
	// Languages is a list of programming languages the user is interested in.
	Languages []string `json:"Languages,omitempty"`
	// Editors is a list of editors the user is interested in.
	Editors []string `json:"Editors,omitempty"`
	// Message contains any additional comments the user may have.
	Message string `json:"Message,omitempty"`
}

// BetaResponse is a response to a user registering for beta access.
type BetaResponse struct {
	// EmailAddress is the email address that was registered and will be
	// contacted once approved by an admin.
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

// Def is a code def returned by the Sourcegraph API.
type Def struct {
	graph.Def  `json:""`
	DocHTML    *htmlutil.HTML          `json:"DocHTML,omitempty"`
	FmtStrings *graph.DefFormatStrings `json:"FmtStrings,omitempty"`
	// StartLine and EndLine are populated if
	// DefGetOptions.ComputeLineRange is true.
	StartLine uint32 `json:"StartLine,omitempty"`
	EndLine   uint32 `json:"EndLine,omitempty"`
}

// DefGetOptions specifies options for DefsService.Get.
type DefGetOptions struct {
	Doc bool `json:"Doc,omitempty" url:",omitempty"`
	// ComputeLineRange is whether the server should compute the start
	// and end line numbers (1-indexed). This incurs additional cost,
	// so it's not always desired.
	ComputeLineRange bool `json:"ComputeLineRange,omitempty" url:",omitempty"`
}

// DefListOptions specifies options for DefsService.List.
type DefListOptions struct {
	Name string `json:"Name,omitempty" url:",omitempty"`
	// Specifies a search query for defs. If specified, then the Sort and Direction
	// options are ignored
	Query string `json:"Query,omitempty" url:",omitempty"`
	// ByteStart and ByteEnd will restrict the results to only definitions that overlap
	// with the specified start and end byte offsets. This filter is only applied if
	// both values are set.
	ByteStart uint32 `json:"ByteStart,omitempty"`
	// ByteStart and ByteEnd will restrict the results to only definitions that overlap
	// with the specified start and end byte offsets. This filter is only applied if
	// both values are set.
	ByteEnd uint32 `json:"ByteEnd,omitempty"`
	// DefKeys, if set, will return the definitions that match the given DefKey
	DefKeys []*graph.DefKey `json:"DefKeys,omitempty"`
	// RepoRevs constrains the results to a set of repository revisions (given by their
	// URIs plus an optional "@" and a revision specifier). For example,
	// "repo.com/foo@revspec".
	//
	// TODO(repo-key): Make this use repo IDs, not URIs.
	RepoRevs []string `json:"RepoRevs,omitempty" url:",omitempty,comma"`
	UnitType string   `json:"UnitType,omitempty" url:",omitempty"`
	Unit     string   `json:"Unit,omitempty" url:",omitempty"`
	Path     string   `json:"Path,omitempty" url:",omitempty"`
	// Files, if specified, will restrict the results to only defs defined in the
	// specified file.
	Files []string `json:"Files,omitempty" url:",omitempty"`
	// FilePathPrefix, if specified, will restrict the results to only defs defined in
	// files whose path is underneath the specified prefix.
	FilePathPrefix string   `json:"FilePathPrefix,omitempty" url:",omitempty"`
	Kinds          []string `json:"Kinds,omitempty" url:",omitempty,comma"`
	Exported       bool     `json:"Exported,omitempty" url:",omitempty"`
	Nonlocal       bool     `json:"Nonlocal,omitempty" url:",omitempty"`
	// IncludeTest is whether the results should include definitions in test files.
	IncludeTest bool `json:"IncludeTest,omitempty" url:",omitempty"`
	// Enhancements
	Doc   bool `json:"Doc,omitempty" url:",omitempty"`
	Fuzzy bool `json:"Fuzzy,omitempty" url:",omitempty"`
	// Sorting
	Sort      string `json:"Sort,omitempty" url:",omitempty"`
	Direction string `json:"Direction,omitempty" url:",omitempty"`
	// Paging
	ListOptions `json:""`
}

// DeprecatedDefListRefsOptions configures the scope of ref search for a def.
type DeprecatedDefListRefsOptions struct {
	Repo        int32    `json:"Repo,omitempty" url:",omitempty"`
	CommitID    string   `json:"CommitID,omitempty" url:",omitempty"`
	Files       []string `json:"Files,omitempty" url:",omitempty"`
	ListOptions `json:""`
}

// DefSpec specifies a def.
type DefSpec struct {
	Repo     int32  `json:"Repo,omitempty"`
	CommitID string `json:"CommitID,omitempty"`
	UnitType string `json:"UnitType,omitempty"`
	Unit     string `json:"Unit,omitempty"`
	Path     string `json:"Path,omitempty"`
}

type DefsGetOp struct {
	Def DefSpec        `json:"Def"`
	Opt *DefGetOptions `json:"Opt,omitempty"`
}

type DefList struct {
	Defs         []*Def `json:"Defs,omitempty"`
	ListResponse `json:""`
}

type DeprecatedDefsListRefsOp struct {
	Def DefSpec                       `json:"Def"`
	Opt *DeprecatedDefListRefsOptions `json:"Opt,omitempty"`
}

type RefList struct {
	Refs           []*graph.Ref `json:"Refs,omitempty"`
	StreamResponse `json:""`
}

// DeprecatedDefListRefLocationsOptions holds the options for fetching
// all locations referencing a def.
type DeprecatedDefListRefLocationsOptions struct {
	// Repos is the list of repos to restrict the results to.
	// If empty, all repos are searched for references.
	Repos []string `json:"Repos,omitempty" url:",omitempty"`
	// ListOptions specifies options for paginating
	// the result.
	ListOptions `json:""`
}

// DeprecatedDefListRefLocationsOptions holds the options for fetching
// all locations referencing the specified def.
type DeprecatedDefsListRefLocationsOp struct {
	// Def identifies the definition whose locations are requested.
	Def DefSpec `json:"Def"`
	// Opt controls the scope of the search for ref locations of this def.
	Opt *DeprecatedDefListRefLocationsOptions `json:"Opt,omitempty"`
}

// DeprecatedRefLocationsList lists the repos and files that reference a def.
type DeprecatedRefLocationsList struct {
	// RepoRefs holds the repos and files referencing the def.
	RepoRefs []*DeprecatedDefRepoRef `json:"RepoRefs,omitempty"`
	// StreamResponse specifies if more results are available.
	StreamResponse `json:""`
	// TotalRepos is the total number of repos which reference the def.
	TotalRepos int32 `json:"TotalRepos,omitempty"`
}

// DeprecatedDefRepoRef identifies a repo and its files that reference a def.
type DeprecatedDefRepoRef struct {
	// Repo is the name of repo that references the def.
	Repo string `json:"Repo,omitempty"`
	// Count is the number of references to the def in the repo.
	Count int32 `json:"Count,omitempty"`
	// Score is the importance score of this repo for the def.
	Score float32 `json:"Score,omitempty"`
	// Files is the list of files in this repo referencing the def.
	Files []*DeprecatedDefFileRef `json:"Files,omitempty"`
}

// DeprecatedFilePosition represents a line:column in a file.
type DeprecatedFilePosition struct {
	Line   int32 `json:"Line,omitempty"`
	Column int32 `json:"Column,omitempty"`
}

// DeprecatedDefFileRef identifies a file that references a def.
type DeprecatedDefFileRef struct {
	// Path is the path of this file.
	Path string `json:"Path,omitempty"`
	// Count is the number of references to the def in this file.
	Count int32 `json:"Count,omitempty"`
	// Positions is the locations in the file that the def is referenced.
	Positions []*DeprecatedFilePosition `json:"Positions,omitempty"`
	// Score is the importance score of this file for the def.
	Score float32 `json:"Score,omitempty"`
}

// RepoTreeGetOptions specifies options for (RepoTreeService).Get.
type RepoTreeGetOptions struct {
	ContentsAsString bool `json:"ContentsAsString,omitempty" url:",omitempty"`
	GetFileOptions   `json:""`
	// NoSrclibAnns indicates whether or not srclib annotations should be
	// excluded in the (optional) returned list of annotations.
	NoSrclibAnns bool `json:"NoSrclibAnns,omitempty"`
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

type DefsRefreshIndexOp struct {
	// Repo is the repo whose graph data is to be re-indexed
	// for global ref locations.
	Repo int32 `json:"Repo,omitempty"`
	// Force ensures we reindex, even if we have already indexed the latest
	// commit for repo
	Force bool `json:"Force,omitempty"`
}

type AsyncRefreshIndexesOp struct {
	// Repo will have all of its indexes refreshed.
	Repo int32 `json:"Repo,omitempty"`
	// Source helps tie back async jobs to their source.
	Source string `json:"Source,omitempty"`
	// Force will ensure all indexes are refreshed, even if the index
	// already includes the latest commit.
	Force bool `json:"Force,omitempty"`
}

// ServerConfig describes the server's configuration.
//
// DEV NOTE: There is some overlap with Go CLI flag structs. In the
// future we may consolidate these.
type ServerConfig struct {
	// Version is the version of Sourcegraph that this server is
	// running.
	Version string `json:"Version,omitempty"`
	// AppURL is the base URL of the user-facing web application
	// (e.g., "https://sourcegraph.com").
	AppURL string `json:"AppURL,omitempty"`
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

// An Annotation is metadata (such as a srclib ref) attached to a
// portion of a file.
type Annotation struct {
	// URL is the location where more information about the
	// annotation's topic may be found (e.g., for a srclib ref, it's
	// the def's URL).
	URL string `json:"URL,omitempty"`
	// StartByte is the start of the byte range.
	StartByte uint32 `json:"StartByte"`
	// EndByte is the end of the byte range.
	EndByte uint32 `json:"EndByte"`
	// Class is the HTML class name that should be applied to this
	// region.
	Class string `json:"Class,omitempty"`
	// Def is whether this annotation marks the definition of the
	// item described in URL or URLs. For example, "Foo" in "func Foo() {}"
	// would have its annotation with Def=true.
	//
	// Marking whether this annotation is a def lets us
	// jump-to-definition here from other refs in the same file
	// without needing to load the defs for those refs.
	Def bool `json:"Def,omitempty"`
	// WantInner indicates that this annotation, when being applied to
	// the underlying text, should be inner (i.e., more deeply
	// nested). Larger numbers mean greater precedence for being
	// nested more deeply.
	WantInner int32 `json:"WantInner,omitempty"`
	// URLs can be set instead of URL if there are multiple URLs on an
	// annotation.
	URLs []string `json:"URLs,omitempty"`
}

// AnnotationList is a list of annotations.
type AnnotationList struct {
	Annotations    []*Annotation `json:"Annotations,omitempty"`
	LineStartBytes []uint32      `json:"LineStartBytes,omitempty"`
}

// AnnotationsListOptions specifies options for Annotations.List.
type AnnotationsListOptions struct {
	// Entry specifies the file in a specific repository at a specific
	// version.
	Entry TreeEntrySpec `json:"Entry"`
	// Range specifies the range within the file that annotations
	// should be fetched for. If it is not set, then all of the file's
	// annotations are returned.
	Range *FileRange `json:"Range,omitempty"`
	// NoSrclibAnns indicates whether or not srclib annotations should be
	// excluded in the returned list of annotations.
	NoSrclibAnns bool `json:"NoSrclibAnns,omitempty"`
}

// AnnotationsGetDefAtPosOptions specifies options for Annotations.GetDefAtPos
type AnnotationsGetDefAtPosOptions struct {
	// Entry specifies the file in a specific repository at a specific
	// version.
	Entry TreeEntrySpec `json:"Entry"`
	// Line specifies the line of the ref (zero based).
	Line uint32 `json:"Line,omitempty"`
	// Character specifies the character of the ref in the line (zero based).
	Character uint32 `json:"Character,omitempty"`
}

// SearchOptions configures the scope of a global search.
type SearchOptions struct {
	// Repos is the list of repos to restrict the results to.
	// If empty, all repos are searched.
	Repos []int32 `json:"Repos,omitempty" url:",omitempty"`
	// NotRepos is the list of repos from which to exclude results.
	// If empty, then no repositories are excluded.
	NotRepos []int32 `json:"NotRepos,omitempty" url:",omitempty"`
	// Languages, if specified, limits the returned results to just the
	// specified languages.
	//
	// The values are case-insensitive, e.g. "java", "Java", and "jAvA" will
	// all match the Java programming language.
	Languages []string `json:"Languages,omitempty"`
	// NotLanguages, if specified, excludes the specified languages from the
	// returned results.
	//
	// The values are case-insensitive, e.g. "java", "Java", and "jAvA" will
	// all exclude the Java programming language.
	NotLanguages []string `json:"NotLanguages,omitempty"`
	// Kinds, if specified, limits the returned results to just the specified
	// kinds of definitions (func, var, etc).
	Kinds []string `json:"Kinds,omitempty"`
	// NotKinds, if specified, excludes the specified kinds of definitions
	// (func, var, etc) from the returned results.
	NotKinds []string `json:"NotKinds,omitempty"`
	// ListOptions specifies options for paginating
	// the result.
	ListOptions `json:""`
	// IncludeRepos indicates whether to include repository search results.
	IncludeRepos bool `json:"IncludeRepos,omitempty"`
	// Fast specifies that the client would like results returned more quickly at
	// the cost of missing lower ranked results. If there are no high-ranking results
	// and Fast is set to true, there may be no results returned at all.
	Fast bool `json:"Fast,omitempty"`
	// AllowEmpty may be used to signal that a query with undefined search parameters is still valid.
	// It was introduced with the sole purpose of running offline queries against all defs.
	// Do not use in the product, as it will likely cause a significant performance hit.
	AllowEmpty bool `json:"AllowEmpty,omitempty"`
}

// SearchOp holds the options for global search for a given query.
type SearchOp struct {
	// Query is the text string being searched for.
	Query string `json:"Query,omitempty"`
	// Opt controls the scope of the search.
	Opt *SearchOptions `json:"Opt,omitempty"`
}

// RepoSearchResult holds a result of a global repo search.
type RepoSearchResult struct {
	// Repo represents a source code repository.
	Repo *Repo `json:"Repo,omitempty"`
}

// SearchReposResultList holds a result of a repo search.
type SearchReposResultList struct {
	// Repos is the list of repo search results.
	Repos []*Repo `json:"Repos,omitempty"`
}

// DefSearchResult holds a result of a global def search.
type DefSearchResult struct {
	// Def is the def matching the search query.
	Def `json:""`
	// Score is the computed relevance of this result to the search query.
	Score float32 `json:"Score,omitempty"`
	// RefCount is global ref count for this def.
	RefCount int32 `json:"RefCount,omitempty"`
}

// SearchResultsList is a list of results to a global search.
type SearchResultsList struct {
	// RepoResults is the list of repo search results
	RepoResults []*RepoSearchResult `json:"RepoResults,omitempty"`
	// DefResults is the list of def search results
	DefResults []*DefSearchResult `json:"DefResults,omitempty"`
	// SearchOptions is a list of options to a global search query.
	SearchQueryOptions []*SearchOptions `json:"SearchQueryOptions,omitempty"`
}

// OrgListOptions holds the options for listing organization details
type OrgListOptions struct {
	OrgName  string `json:"OrgName,omitempty"`
	Username string `json:"Username,omitempty"`
	OrgID    string `json:"OrgID,omitempty"`
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
