package sourcegraph

import (
	"time"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
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
	// IndexedRevision is the revision that the global index is currently based on.
	IndexedRevision *string `json:"IndexedRevision,omitempty"`
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
}

// SrclibDataVersion specifies a srclib store version.
type SrclibDataVersion struct {
	CommitID      string `json:"CommitID,omitempty"`
	CommitsBehind int32  `json:"CommitsBehind,omitempty"`
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

func (d *DeprecatedRefLocationsList) Convert() *DeprecatedRefLocations {
	sourceRefs := make([]*DeprecatedSourceRef, len(d.RepoRefs))
	for i, r := range d.RepoRefs {
		sourceRefs[i] = r.Convert()
	}
	return &DeprecatedRefLocations{
		SourceRefs:     sourceRefs,
		StreamResponse: d.StreamResponse,
		TotalSources:   int(d.TotalRepos),
	}
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

func (d *DeprecatedDefRepoRef) Convert() *DeprecatedSourceRef {
	files := make([]*DeprecatedFileRef, len(d.Files))
	for i, f := range d.Files {
		files[i] = f.Convert()
	}
	return &DeprecatedSourceRef{Source: d.Repo, Refs: int(d.Count), Score: int16(d.Score), FileRefs: files}
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

func (d *DeprecatedDefFileRef) Convert() *DeprecatedFileRef {
	// Use d.Count since d.Positions is not actually populated today. This at
	// least gives us valid "N times in file X" counts.
	positions := make([]lsp.Range, d.Count)
	return &DeprecatedFileRef{File: d.Path, Positions: positions, Score: int16(d.Score)}
}

// DeprecatedRefLocations lists the sources and files that reference a def.
type DeprecatedRefLocations struct {
	// SourceRefs holds the sources and files referencing the def.
	SourceRefs []*DeprecatedSourceRef
	// StreamResponse specifies if more results are available.
	StreamResponse
	// TotalSources is the total number of sources which reference the def.
	TotalSources int
}

// DeprecatedSourceRef identifies a source (e.g. a repo) and its files that reference a
// def.
type DeprecatedSourceRef struct {
	// Scheme is the URI scheme for the source, e.g. "git"
	Scheme string

	// Source is the source that references the def (e.g. a repo URI).
	Source string

	// Version is the version of the source that references the def.
	Version string

	// Files is the number of files in the source that reference the def.
	Files int

	// Refs is the total number of references to the def in the source.
	Refs int

	// Score is the importance score of this source for the def.
	Score int16

	// FileRefs is the list of files in this source referencing the def.
	FileRefs []*DeprecatedFileRef
}

// DeprecatedFileRef identifies a file that references a def.
type DeprecatedFileRef struct {
	// Scheme is the URI scheme for the source, e.g. "git"
	Scheme string

	// Source is the source that references the def (e.g. a repo URI).
	Source string

	// Version is the version of the source that references the def.
	Version string

	// File is the filepath that references the def.
	File string

	// Positions is the locations in the file that the def is referenced.
	Positions []lsp.Range

	// Score is the importance score of this file for the def.
	Score int16
}

// RefLocationsOptions specifies options for querying locations that reference
// a definition.
type RefLocationsOptions struct {
	Language        string
	RepoID          int32
	File            string
	Line, Character int
}

// RefLocations is a lists of reference locations to a definition.
type RefLocations struct {
	// Locations is the actual locations that reference a definition.
	Locations []*RefLocation

	// StreamResponse specifies if more results are available.
	StreamResponse
}

// RefLocation represents the location of a reference to a definition.
type RefLocation struct {
	// Scheme, Host, Path, Version, and File combined compose a URI at which a
	// reference to a definition has been made. For example:
	//
	//  <scheme>://<host><path>?<version>#<file>
	// 	git://github.com/gorilla/mux?master#sub/pkg/main.go
	//
	// The scheme is limited to VCS protocol schemes (e.g. `git` or `svn`)
	// which are viewable directly by Sourcegraph itself. No language-specific
	// schemes like `npm` etc are applicable here.
	Scheme, Host, Path, Version, File string

	// Start and end line/character positions at which the reference begins and
	// ends, respectively.
	StartLine, StartChar, EndLine, EndChar int
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

type DefsRefreshIndexOp struct {
	// Repo is the repo whose graph data is to be re-indexed
	// for global ref locations.
	Repo int32 `json:"Repo,omitempty"`
	// Force ensures we reindex, even if we have already indexed the latest
	// commit for repo
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
