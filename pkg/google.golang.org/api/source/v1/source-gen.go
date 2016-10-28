// Package source provides access to the Cloud Source Repositories API.
//
// See https://cloud.google.com/eap/cloud-repositories/cloud-source-api
//
// Usage example:
//
//   import "google.golang.org/api/source/v1"
//   ...
//   sourceService, err := source.New(oauthHttpClient)
package source // import "sourcegraph.com/sourcegraph/sourcegraph/pkg/google.golang.org/api/source/v1"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	context "golang.org/x/net/context"
	ctxhttp "golang.org/x/net/context/ctxhttp"
	gensupport "google.golang.org/api/gensupport"
	googleapi "google.golang.org/api/googleapi"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = gensupport.MarshalJSON
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "source:v1"
const apiName = "source"
const apiVersion = "v1"
const basePath = "https://source.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Projects = NewProjectsService(s)
	s.V1 = NewV1Service(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Projects *ProjectsService

	V1 *V1Service
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewProjectsService(s *Service) *ProjectsService {
	rs := &ProjectsService{s: s}
	rs.Repos = NewProjectsReposService(s)
	return rs
}

type ProjectsService struct {
	s *Service

	Repos *ProjectsReposService
}

func NewProjectsReposService(s *Service) *ProjectsReposService {
	rs := &ProjectsReposService{s: s}
	rs.Aliases = NewProjectsReposAliasesService(s)
	rs.Files = NewProjectsReposFilesService(s)
	rs.Revisions = NewProjectsReposRevisionsService(s)
	rs.Workspaces = NewProjectsReposWorkspacesService(s)
	return rs
}

type ProjectsReposService struct {
	s *Service

	Aliases *ProjectsReposAliasesService

	Files *ProjectsReposFilesService

	Revisions *ProjectsReposRevisionsService

	Workspaces *ProjectsReposWorkspacesService
}

func NewProjectsReposAliasesService(s *Service) *ProjectsReposAliasesService {
	rs := &ProjectsReposAliasesService{s: s}
	rs.Files = NewProjectsReposAliasesFilesService(s)
	return rs
}

type ProjectsReposAliasesService struct {
	s *Service

	Files *ProjectsReposAliasesFilesService
}

func NewProjectsReposAliasesFilesService(s *Service) *ProjectsReposAliasesFilesService {
	rs := &ProjectsReposAliasesFilesService{s: s}
	return rs
}

type ProjectsReposAliasesFilesService struct {
	s *Service
}

func NewProjectsReposFilesService(s *Service) *ProjectsReposFilesService {
	rs := &ProjectsReposFilesService{s: s}
	return rs
}

type ProjectsReposFilesService struct {
	s *Service
}

func NewProjectsReposRevisionsService(s *Service) *ProjectsReposRevisionsService {
	rs := &ProjectsReposRevisionsService{s: s}
	rs.Files = NewProjectsReposRevisionsFilesService(s)
	return rs
}

type ProjectsReposRevisionsService struct {
	s *Service

	Files *ProjectsReposRevisionsFilesService
}

func NewProjectsReposRevisionsFilesService(s *Service) *ProjectsReposRevisionsFilesService {
	rs := &ProjectsReposRevisionsFilesService{s: s}
	return rs
}

type ProjectsReposRevisionsFilesService struct {
	s *Service
}

func NewProjectsReposWorkspacesService(s *Service) *ProjectsReposWorkspacesService {
	rs := &ProjectsReposWorkspacesService{s: s}
	rs.Files = NewProjectsReposWorkspacesFilesService(s)
	rs.Snapshots = NewProjectsReposWorkspacesSnapshotsService(s)
	return rs
}

type ProjectsReposWorkspacesService struct {
	s *Service

	Files *ProjectsReposWorkspacesFilesService

	Snapshots *ProjectsReposWorkspacesSnapshotsService
}

func NewProjectsReposWorkspacesFilesService(s *Service) *ProjectsReposWorkspacesFilesService {
	rs := &ProjectsReposWorkspacesFilesService{s: s}
	return rs
}

type ProjectsReposWorkspacesFilesService struct {
	s *Service
}

func NewProjectsReposWorkspacesSnapshotsService(s *Service) *ProjectsReposWorkspacesSnapshotsService {
	rs := &ProjectsReposWorkspacesSnapshotsService{s: s}
	rs.Files = NewProjectsReposWorkspacesSnapshotsFilesService(s)
	return rs
}

type ProjectsReposWorkspacesSnapshotsService struct {
	s *Service

	Files *ProjectsReposWorkspacesSnapshotsFilesService
}

func NewProjectsReposWorkspacesSnapshotsFilesService(s *Service) *ProjectsReposWorkspacesSnapshotsFilesService {
	rs := &ProjectsReposWorkspacesSnapshotsFilesService{s: s}
	return rs
}

type ProjectsReposWorkspacesSnapshotsFilesService struct {
	s *Service
}

func NewV1Service(s *Service) *V1Service {
	rs := &V1Service{s: s}
	return rs
}

type V1Service struct {
	s *Service
}

// Action: An action to perform on a path in a workspace.
type Action struct {
	// CopyAction: Copy the contents of one path to another.
	CopyAction *CopyAction `json:"copyAction,omitempty"`

	// DeleteAction: Delete a file or directory.
	DeleteAction *DeleteAction `json:"deleteAction,omitempty"`

	// WriteAction: Create or modify a file.
	WriteAction *WriteAction `json:"writeAction,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CopyAction") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CopyAction") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Action) MarshalJSON() ([]byte, error) {
	type noMethod Action
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Alias: An alias is a named reference to a revision. Examples include
// git
// branches and tags.
type Alias struct {
	// Kind: The alias kind.
	//
	// Possible values:
	//   "ANY" - ANY is used to indicate to ListAliases to return aliases of
	// all kinds,
	// and when used with GetAlias, the GetAlias function will return a
	// FIXED,
	// or MOVABLE, in that priority order. Using ANY
	// with CreateAlias or DeleteAlias will result in an error.
	//   "FIXED" - Git tag
	//   "MOVABLE" - Git branch
	//   "MERCURIAL_BRANCH_DEPRECATED"
	//   "OTHER" - OTHER is used to fetch non-standard aliases, which are
	// none
	// of the kinds above or below. For example, if a git repo
	// has a ref named "refs/foo/bar", it is considered to be OTHER.
	//   "SPECIAL_DEPRECATED" - DO NOT USE.
	Kind string `json:"kind,omitempty"`

	// Name: The alias name.
	Name string `json:"name,omitempty"`

	// RevisionId: The revision referred to by this alias.
	// For git tags and branches, this is the corresponding hash.
	RevisionId string `json:"revisionId,omitempty"`

	// WorkspaceNames: The list of workspace names whose alias is this
	// one.
	// NOT YET IMPLEMENTED (b/16943429).
	WorkspaceNames []string `json:"workspaceNames,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Kind") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Kind") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Alias) MarshalJSON() ([]byte, error) {
	type noMethod Alias
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// AliasContext: An alias to a repo revision.
type AliasContext struct {
	// Kind: The alias kind.
	//
	// Possible values:
	//   "ANY" - Do not use.
	//   "FIXED" - Git tag
	//   "MOVABLE" - Git branch
	//   "OTHER" - OTHER is used to specify non-standard aliases, those not
	// of the kinds
	// above. For example, if a Git repo has a ref named "refs/foo/bar",
	// it
	// is considered to be of kind OTHER.
	Kind string `json:"kind,omitempty"`

	// Name: The alias name.
	Name string `json:"name,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Kind") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Kind") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *AliasContext) MarshalJSON() ([]byte, error) {
	type noMethod AliasContext
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ChangedFileInfo: Represents file information.
type ChangedFileInfo struct {
	// FromPath: Related file path for copies or renames.
	//
	// For copies, the type will be ADDED and the from_path will point to
	// the
	// source of the copy. For renames, the type will be ADDED, the
	// from_path
	// will point to the source of the rename, and another ChangedFileInfo
	// record
	// with that path will appear with type DELETED. In other words, a
	// rename is
	// represented as a copy plus a delete of the old path.
	FromPath string `json:"fromPath,omitempty"`

	// Hash: A hex-encoded hash for the file.
	// Not necessarily a hash of the file's contents. Two paths in the
	// same
	// revision with the same hash have the same contents with high
	// probability.
	// Empty if the operation is CONFLICTED.
	Hash string `json:"hash,omitempty"`

	// Operation: The operation type for the file.
	//
	// Possible values:
	//   "OPERATION_UNSPECIFIED" - No operation was specified.
	//   "ADDED" - The file was added.
	//   "DELETED" - The file was deleted.
	//   "MODIFIED" - The file was modified.
	//   "CONFLICTED" - The result of merging the file is a conflict.
	// The CONFLICTED type only appears in Workspace.changed_files
	// or
	// Snapshot.changed_files when the workspace is in a merge state.
	Operation string `json:"operation,omitempty"`

	// Path: The path of the file.
	Path string `json:"path,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FromPath") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "FromPath") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ChangedFileInfo) MarshalJSON() ([]byte, error) {
	type noMethod ChangedFileInfo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CloudRepoSourceContext: A CloudRepoSourceContext denotes a particular
// revision in a cloud
// repo (a repo hosted by the Google Cloud Platform).
type CloudRepoSourceContext struct {
	// AliasContext: An alias, which may be a branch or tag.
	AliasContext *AliasContext `json:"aliasContext,omitempty"`

	// AliasName: The name of an alias (branch, tag, etc.).
	AliasName string `json:"aliasName,omitempty"`

	// RepoId: The ID of the repo.
	RepoId *RepoId `json:"repoId,omitempty"`

	// RevisionId: A revision ID.
	RevisionId string `json:"revisionId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AliasContext") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "AliasContext") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CloudRepoSourceContext) MarshalJSON() ([]byte, error) {
	type noMethod CloudRepoSourceContext
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CloudWorkspaceId: A CloudWorkspaceId is a unique identifier for a
// cloud workspace.
// A cloud workspace is a place associated with a repo where modified
// files
// can be stored before they are committed.
type CloudWorkspaceId struct {
	// Name: The unique name of the workspace within the repo.  This is the
	// name
	// chosen by the client in the Source API's CreateWorkspace method.
	Name string `json:"name,omitempty"`

	// RepoId: The ID of the repo containing the workspace.
	RepoId *RepoId `json:"repoId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Name") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Name") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CloudWorkspaceId) MarshalJSON() ([]byte, error) {
	type noMethod CloudWorkspaceId
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CloudWorkspaceSourceContext: A CloudWorkspaceSourceContext denotes a
// workspace at a particular snapshot.
type CloudWorkspaceSourceContext struct {
	// SnapshotId: The ID of the snapshot.
	// An empty snapshot_id refers to the most recent snapshot.
	SnapshotId string `json:"snapshotId,omitempty"`

	// WorkspaceId: The ID of the workspace.
	WorkspaceId *CloudWorkspaceId `json:"workspaceId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "SnapshotId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "SnapshotId") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CloudWorkspaceSourceContext) MarshalJSON() ([]byte, error) {
	type noMethod CloudWorkspaceSourceContext
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CommitWorkspaceRequest: Request for CommitWorkspace.
type CommitWorkspaceRequest struct {
	// Author: Author of the commit in the format: "Author Name
	// <author@example.com>"
	// required
	Author string `json:"author,omitempty"`

	// CurrentSnapshotId: If non-empty, current_snapshot_id must refer to
	// the most recent update to
	// the workspace, or ABORTED is returned.
	CurrentSnapshotId string `json:"currentSnapshotId,omitempty"`

	// Message: The commit message.
	// required
	Message string `json:"message,omitempty"`

	// Paths: The subset of modified paths to commit. If empty, then commit
	// all
	// modified paths.
	Paths []string `json:"paths,omitempty"`

	// WorkspaceId: The ID of the workspace.
	WorkspaceId *CloudWorkspaceId `json:"workspaceId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Author") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Author") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CommitWorkspaceRequest) MarshalJSON() ([]byte, error) {
	type noMethod CommitWorkspaceRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CopyAction: Copy the contents of a file or directory at from_path in
// the specified
// revision or snapshot to to_path.
//
// To rename a file, copy it to the new path and delete the old.
type CopyAction struct {
	// FromPath: The path to copy from.
	FromPath string `json:"fromPath,omitempty"`

	// FromRevisionId: The revision ID from which to copy the file.
	FromRevisionId string `json:"fromRevisionId,omitempty"`

	// FromSnapshotId: The snapshot ID from which to copy the file.
	FromSnapshotId string `json:"fromSnapshotId,omitempty"`

	// ToPath: The path to copy to.
	ToPath string `json:"toPath,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FromPath") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "FromPath") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CopyAction) MarshalJSON() ([]byte, error) {
	type noMethod CopyAction
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// CreateWorkspaceRequest: Request for CreateWorkspace.
type CreateWorkspaceRequest struct {
	// Actions: An ordered sequence of actions to perform in the workspace.
	// Can be empty.
	// Specifying actions here instead of using ModifyWorkspace saves one
	// RPC.
	Actions []*Action `json:"actions,omitempty"`

	// RepoId: The repo within which to create the workspace.
	RepoId *RepoId `json:"repoId,omitempty"`

	// Workspace: The following fields of workspace, with the allowable
	// exception of
	// baseline, must be set. No other fields of workspace should be
	// set.
	//
	// id.name
	// Provides the name of the workspace and must be unique within the
	// repo.
	// Note: Do not set field id.repo_id.  The repo_id is provided above as
	// a
	// CreateWorkspaceRequest field.
	//
	// alias:
	// If alias names an existing movable alias, the workspace's baseline
	// is set to the alias's revision.
	//
	// If alias does not name an existing movable alias, then the workspace
	// is
	// created with no baseline. When the workspace is committed, a new
	// root
	// revision is created with no parents. The new revision becomes
	// the
	// workspace's baseline and the alias name is used to create a movable
	// alias
	// referring to the revision.
	//
	// baseline:
	// A revision ID (hexadecimal string) for sequencing. If non-empty,
	// alias
	// must name an existing movable alias and baseline must match the
	// alias's
	// revision ID.
	Workspace *Workspace `json:"workspace,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Actions") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Actions") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *CreateWorkspaceRequest) MarshalJSON() ([]byte, error) {
	type noMethod CreateWorkspaceRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// DeleteAction: Delete a file or directory.
type DeleteAction struct {
	// Path: The path of the file or directory. If path refers to
	// a
	// directory, the directory and its contents are deleted.
	Path string `json:"path,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Path") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Path") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *DeleteAction) MarshalJSON() ([]byte, error) {
	type noMethod DeleteAction
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// DirectoryEntry: Information about a directory.
type DirectoryEntry struct {
	// Info: Information about the entry.
	Info *FileInfo `json:"info,omitempty"`

	// IsDir: Whether the entry is a file or directory.
	IsDir bool `json:"isDir,omitempty"`

	// LastModifiedRevisionId: ID of the revision that most recently
	// modified this file.
	LastModifiedRevisionId string `json:"lastModifiedRevisionId,omitempty"`

	// Name: Name of the entry relative to the directory.
	Name string `json:"name,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Info") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Info") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *DirectoryEntry) MarshalJSON() ([]byte, error) {
	type noMethod DirectoryEntry
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Empty: A generic empty message that you can re-use to avoid defining
// duplicated
// empty messages in your APIs. A typical example is to use it as the
// request
// or the response type of an API method. For instance:
//
//     service Foo {
//       rpc Bar(google.protobuf.Empty) returns
// (google.protobuf.Empty);
//     }
//
// The JSON representation for `Empty` is empty JSON object `{}`.
type Empty struct {
	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`
}

// ExternalReference: A submodule or subrepository.
type ExternalReference struct {
}

// File: A file, with contents and metadata.
//
// Pagination can be used to limit the size of the file. Otherwise,
// there is a
// default max size for the contents. Whether the file has been
// truncated can
// be determined by comparing len(contents) to info.Size.
type File struct {
	// Contents: The contents of the file.
	Contents string `json:"contents,omitempty"`

	// Info: Information about the file.
	Info *FileInfo `json:"info,omitempty"`

	// Path: The path to the file starting from the root of the revision.
	Path string `json:"path,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Contents") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Contents") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *File) MarshalJSON() ([]byte, error) {
	type noMethod File
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// FileInfo: File metadata, including a hash of the file contents.
type FileInfo struct {
	// Hash: A hex-encoded cryptographic hash of the file's contents,
	// possibly with other data.
	Hash string `json:"hash,omitempty"`

	// IsText: An educated guess as to whether the file is human-readable
	// text, or
	// binary. Typically available only when file contents are retrieved
	// (since
	// the guess depends on examining a prefix of the contents), but some
	// systems
	// might store this metadata for every file.
	IsText bool `json:"isText,omitempty"`

	// Mode: The mode of the file: an executable, a symbolic link, or
	// neither.
	//
	// Possible values:
	//   "FILE_MODE_UNSPECIFIED" - No file mode was specified.
	//   "NORMAL" - Neither a symbolic link nor executable.
	//   "SYMLINK" - A symbolic link.
	//   "EXECUTABLE" - An executable.
	Mode string `json:"mode,omitempty"`

	// Size: The size of the file in bytes.
	Size int64 `json:"size,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "Hash") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Hash") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *FileInfo) MarshalJSON() ([]byte, error) {
	type noMethod FileInfo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GerritSourceContext: A SourceContext referring to a Gerrit project.
type GerritSourceContext struct {
	// AliasContext: An alias, which may be a branch or tag.
	AliasContext *AliasContext `json:"aliasContext,omitempty"`

	// AliasName: The name of an alias (branch, tag, etc.).
	AliasName string `json:"aliasName,omitempty"`

	// GerritProject: The full project name within the host. Projects may be
	// nested, so
	// "project/subproject" is a valid project name.
	// The "repo name" is hostURI/project.
	GerritProject string `json:"gerritProject,omitempty"`

	// HostUri: The URI of a running Gerrit instance.
	HostUri string `json:"hostUri,omitempty"`

	// RevisionId: A revision (commit) ID.
	RevisionId string `json:"revisionId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AliasContext") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "AliasContext") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GerritSourceContext) MarshalJSON() ([]byte, error) {
	type noMethod GerritSourceContext
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GetRevisionsResponse: Response for GetRevisions.
type GetRevisionsResponse struct {
	// Revisions: The revisions.
	Revisions []*Revision `json:"revisions,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Revisions") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Revisions") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GetRevisionsResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetRevisionsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// GitSourceContext: A GitSourceContext denotes a particular revision in
// a third party Git
// repository (e.g. GitHub).
type GitSourceContext struct {
	// RevisionId: Git commit hash.
	// required.
	RevisionId string `json:"revisionId,omitempty"`

	// Url: Git repository URL.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "RevisionId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "RevisionId") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *GitSourceContext) MarshalJSON() ([]byte, error) {
	type noMethod GitSourceContext
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListAliasesResponse: Response for ListAliases.
type ListAliasesResponse struct {
	// Aliases: The list of aliases.
	Aliases []*Alias `json:"aliases,omitempty"`

	// NextPageToken: Use as the value of page_token in the next
	// call to obtain the next page of results.
	// If empty, there are no more results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// TotalAliases: The total number of aliases in the repo of the kind
	// specified in the
	// request.
	TotalAliases int64 `json:"totalAliases,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Aliases") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Aliases") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListAliasesResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListAliasesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListChangedFilesRequest: Request for ListChangedFiles.
type ListChangedFilesRequest struct {
	// PageSize: The maximum number of ChangedFileInfo values to return.
	PageSize int64 `json:"pageSize,omitempty"`

	// PageToken: The value of next_page_token from the previous call.
	// Omit for the first page.
	PageToken string `json:"pageToken,omitempty"`

	// SourceContext1: The starting source context to compare.
	SourceContext1 *SourceContext `json:"sourceContext1,omitempty"`

	// SourceContext2: The ending source context to compare.
	SourceContext2 *SourceContext `json:"sourceContext2,omitempty"`

	// ForceSendFields is a list of field names (e.g. "PageSize") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "PageSize") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListChangedFilesRequest) MarshalJSON() ([]byte, error) {
	type noMethod ListChangedFilesRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListChangedFilesResponse: Response for ListChangedFiles.
type ListChangedFilesResponse struct {
	// ChangedFiles: Note: ChangedFileInfo.from_path is not set here.
	// ListChangedFiles does not
	// perform rename/copy detection.
	//
	// The ChangedFileInfo.Type describes the changes from source_context1
	// to
	// source_context2. Thus ADDED would mean a file is not present
	// in
	// source_context1 but is present in source_context2.
	ChangedFiles []*ChangedFileInfo `json:"changedFiles,omitempty"`

	// NextPageToken: Use as the value of page_token in the next
	// call to obtain the next page of results.
	// If empty, there are no more results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ChangedFiles") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ChangedFiles") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListChangedFilesResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListChangedFilesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListFilesResponse: Response for ListFiles.
type ListFilesResponse struct {
	// Files: The contents field is empty.
	Files []*File `json:"files,omitempty"`

	// NextPageToken: Use as the value of page_token in the next
	// call to obtain the next page of results.
	// If empty, there are no more results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Files") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Files") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListFilesResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListFilesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListReposResponse: Response for ListRepos.
type ListReposResponse struct {
	// Repos: The listed repos.
	Repos []*Repo `json:"repos,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Repos") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Repos") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListReposResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListReposResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListRevisionsResponse: Response for ListRevisions.
type ListRevisionsResponse struct {
	// NextPageToken: Use as the value of page_token in the next
	// call to obtain the next page of results.
	// If empty, there are no more results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// Revisions: The list of revisions.
	Revisions []*Revision `json:"revisions,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NextPageToken") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListRevisionsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListRevisionsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListSnapshotsResponse: Response for ListSnapshots.
type ListSnapshotsResponse struct {
	// NextPageToken: Use as the value of page_token in the next
	// call to obtain the next page of results.
	// If empty, there are no more results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// Snapshots: The list of snapshots.
	Snapshots []*Snapshot `json:"snapshots,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "NextPageToken") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListSnapshotsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListSnapshotsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ListWorkspacesResponse: Response for ListWorkspaces.
type ListWorkspacesResponse struct {
	// Workspaces: The listed workspaces.
	Workspaces []*Workspace `json:"workspaces,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Workspaces") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Workspaces") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ListWorkspacesResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListWorkspacesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// MergeInfo: MergeInfo holds information needed while resolving
// merges, and
// refreshes that
// involve conflicts.
type MergeInfo struct {
	// CommonAncestorRevisionId: Revision ID of the closest common ancestor
	// of the file trees that are
	// participating in a refresh or merge.  During a refresh, the
	// common
	// ancestor is the baseline of the workspace.  During a merge of
	// two
	// branches, the common ancestor is derived from the workspace baseline
	// and
	// the alias of the branch being merged in.  The repository state at
	// the
	// common ancestor provides the base version for a three-way merge.
	CommonAncestorRevisionId string `json:"commonAncestorRevisionId,omitempty"`

	// IsRefresh: If true, a refresh operation is in progress.  If false, a
	// merge is in
	// progress.
	IsRefresh bool `json:"isRefresh,omitempty"`

	// OtherRevisionId: During a refresh, the ID of the revision with which
	// the workspace is being
	// refreshed. This is the revision ID to which the workspace's alias
	// refers
	// at the time of the RefreshWorkspace call. During a merge, the ID of
	// the
	// revision that's being merged into the workspace's alias. This is
	// the
	// revision_id field of the MergeRequest.
	OtherRevisionId string `json:"otherRevisionId,omitempty"`

	// WorkspaceAfterSnapshotId: The workspace snapshot immediately after
	// the refresh or merge RPC
	// completes.  If a file has conflicts, this snapshot contains
	// the
	// version of the file with conflict markers.
	WorkspaceAfterSnapshotId string `json:"workspaceAfterSnapshotId,omitempty"`

	// WorkspaceBeforeSnapshotId: During a refresh, the snapshot ID of the
	// latest change to the workspace
	// before the refresh.  During a merge, the workspace's baseline, which
	// is
	// identical to the commit hash of the workspace's alias before
	// initiating
	// the merge.
	WorkspaceBeforeSnapshotId string `json:"workspaceBeforeSnapshotId,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "CommonAncestorRevisionId") to unconditionally include in API
	// requests. By default, fields with empty values are omitted from API
	// requests. However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CommonAncestorRevisionId")
	// to include in API requests with the JSON null value. By default,
	// fields with empty values are omitted from API requests. However, any
	// field with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *MergeInfo) MarshalJSON() ([]byte, error) {
	type noMethod MergeInfo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// MergeRequest: Request for Merge.
type MergeRequest struct {
	// RevisionId: The other revision to be merged.
	RevisionId string `json:"revisionId,omitempty"`

	// WorkspaceId: The workspace to use for the merge. The revision
	// referred to
	// by the workspace's alias will be one of the revisions merged.
	WorkspaceId *CloudWorkspaceId `json:"workspaceId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "RevisionId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "RevisionId") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *MergeRequest) MarshalJSON() ([]byte, error) {
	type noMethod MergeRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ModifyWorkspaceRequest: Request for ModifyWorkspace.
type ModifyWorkspaceRequest struct {
	// Actions: An ordered sequence of actions to perform in the workspace.
	// May not be
	// empty.
	Actions []*Action `json:"actions,omitempty"`

	// CurrentSnapshotId: If non-empty, current_snapshot_id must refer to
	// the most recent update to
	// the workspace, or ABORTED is returned.
	CurrentSnapshotId string `json:"currentSnapshotId,omitempty"`

	// WorkspaceId: The ID of the workspace.
	WorkspaceId *CloudWorkspaceId `json:"workspaceId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Actions") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Actions") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ModifyWorkspaceRequest) MarshalJSON() ([]byte, error) {
	type noMethod ModifyWorkspaceRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ProjectRepoId: Selects a repo using a Google Cloud Platform project
// ID
// (e.g. winged-cargo-31) and a repo name within that project.
type ProjectRepoId struct {
}

// ReadResponse: Response to read request. Exactly one of entries, file
// or external_reference
// will be populated, depending on what the path in the request denotes.
type ReadResponse struct {
	// Entries: Contains the directory entries if the request specifies a
	// directory.
	Entries []*DirectoryEntry `json:"entries,omitempty"`

	// ExternalReference: The read path denotes a Git submodule.
	ExternalReference *ExternalReference `json:"externalReference,omitempty"`

	// File: Contains file metadata and contents if the request specifies a
	// file.
	File *File `json:"file,omitempty"`

	// NextPageToken: Use as the value of page_token in the next
	// call to obtain the next page of results.
	// If empty, there are no more results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// SourceContext: Returns the SourceContext actually used, resolving any
	// alias in the input
	// SourceContext into its revision ID and returning the actual
	// current
	// snapshot ID if the read was from a workspace with an unspecified
	// snapshot
	// ID.
	SourceContext *SourceContext `json:"sourceContext,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Entries") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Entries") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ReadResponse) MarshalJSON() ([]byte, error) {
	type noMethod ReadResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// RefreshWorkspaceRequest: Request for RefreshWorkspace.
type RefreshWorkspaceRequest struct {
	// WorkspaceId: The ID of the workspace.
	WorkspaceId *CloudWorkspaceId `json:"workspaceId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "WorkspaceId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "WorkspaceId") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *RefreshWorkspaceRequest) MarshalJSON() ([]byte, error) {
	type noMethod RefreshWorkspaceRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Repo: A repository (or repo) stores files for a version-control
// system.
type Repo struct {
	// CreateTime: Timestamp when the repo was created.
	CreateTime string `json:"createTime,omitempty"`

	// Id: Randomly generated ID that uniquely identifies a repo.
	Id string `json:"id,omitempty"`

	// Name: Human-readable, user-defined name of the repository. Names must
	// be
	// alphanumeric, lowercase, begin with a letter, and be between 3 and
	// 63
	// characters long. The - character can appear in the middle
	// positions.
	// (Names must satisfy the regular expression
	// a-z{1,61}[a-z0-9].)
	Name string `json:"name,omitempty"`

	// ProjectId: Immutable, globally unique, DNS-compatible textual
	// identifier.
	// Examples: user-chosen-project-id, yellow-banana-33.
	ProjectId string `json:"projectId,omitempty"`

	// RepoSyncConfig: How RepoSync is configured for this repo. If missing,
	// this
	// repo is not set up for RepoSync.
	RepoSyncConfig *RepoSyncConfig `json:"repoSyncConfig,omitempty"`

	// State: The state the repo is in.
	//
	// Possible values:
	//   "STATE_UNSPECIFIED" - No state was specified.
	//   "LIVE" - The repo is live and available for use.
	//   "DELETED" - The repo has been deleted.
	State string `json:"state,omitempty"`

	// Vcs: The version control system of the repo.
	//
	// Possible values:
	//   "VCS_UNSPECIFIED" - No version control system was specified.
	//   "GIT" - The Git version control system.
	Vcs string `json:"vcs,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "CreateTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CreateTime") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Repo) MarshalJSON() ([]byte, error) {
	type noMethod Repo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// RepoId: A unique identifier for a cloud repo.
type RepoId struct {
	// ProjectRepoId: A combination of a project ID and a repo name.
	ProjectRepoId *ProjectRepoId `json:"projectRepoId,omitempty"`

	// Uid: A server-assigned, globally unique identifier.
	Uid string `json:"uid,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ProjectRepoId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ProjectRepoId") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *RepoId) MarshalJSON() ([]byte, error) {
	type noMethod RepoId
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// RepoSyncConfig: RepoSync configuration information.
type RepoSyncConfig struct {
	// ExternalRepoUrl: If this repo is enabled for RepoSync, this will be
	// the URL of the
	// external repo that this repo should sync with.
	ExternalRepoUrl string `json:"externalRepoUrl,omitempty"`

	// Status: The status of RepoSync.
	//
	// Possible values:
	//   "REPO_SYNC_STATUS_UNSPECIFIED" - No RepoSync status was specified.
	//   "OK" - RepoSync is working.
	//   "FAILED_AUTH" - RepoSync failed because of
	// authorization/authentication.
	//   "FAILED_OTHER" - RepoSync failed for a reason other than auth.
	//   "FAILED_NOT_FOUND" - RepoSync failed because the repository was not
	// found.
	Status string `json:"status,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ExternalRepoUrl") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ExternalRepoUrl") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

func (s *RepoSyncConfig) MarshalJSON() ([]byte, error) {
	type noMethod RepoSyncConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// ResolveFilesRequest: Request for ResolveFiles.
type ResolveFilesRequest struct {
	// ResolvedPaths: Files that should be marked as resolved in the
	// workspace.  All files in
	// resolved_paths must currently be in the CONFLICTED state
	// in
	// Workspace.changed_files.  NOTE: Changing a file's contents to match
	// the
	// contents in the workspace baseline, then calling ResolveFiles on it,
	// will
	// cause the file to be removed from the changed_files list entirely.
	// If resolved_paths is empty, INVALID_ARGUMENT is returned.
	// If resolved_paths contains duplicates, INVALID_ARGUMENT is
	// returned.
	// If resolved_paths contains a path that was never unresolved,
	// or has already been resolved, FAILED_PRECONDITION is returned.
	ResolvedPaths []string `json:"resolvedPaths,omitempty"`

	// WorkspaceId: The ID of the workspace.
	WorkspaceId *CloudWorkspaceId `json:"workspaceId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ResolvedPaths") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ResolvedPaths") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *ResolveFilesRequest) MarshalJSON() ([]byte, error) {
	type noMethod ResolveFilesRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// RevertRefreshRequest: Request for RevertRefresh.
type RevertRefreshRequest struct {
	// WorkspaceId: The ID of the workspace.
	WorkspaceId *CloudWorkspaceId `json:"workspaceId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "WorkspaceId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "WorkspaceId") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *RevertRefreshRequest) MarshalJSON() ([]byte, error) {
	type noMethod RevertRefreshRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Revision: A revision is a snapshot of a file tree, with associated
// metadata. This
// message contains metadata only. Use the Read or
// ReadFromWorkspaceOrAlias
// rpcs to read the contents of the revision's file tree.
type Revision struct {
	// Author: The name of the user who wrote the revision. (In Git, this
	// can
	// differ from committer.)
	Author string `json:"author,omitempty"`

	// BranchName: Mercurial branch name.
	BranchName string `json:"branchName,omitempty"`

	// ChangedFiles: Files changed in this revision.
	ChangedFiles []*ChangedFileInfo `json:"changedFiles,omitempty"`

	// ChangedFilesUnknown: In some cases changed-file
	// information is generated asynchronously. So there is a period
	// of time when it is not available. This field encodes that fact.
	// (An empty changed_files field is not sufficient, since it is
	// possible for a revision to have no changed files.)
	ChangedFilesUnknown bool `json:"changedFilesUnknown,omitempty"`

	// CommitMessage: The message added by the committer.
	CommitMessage string `json:"commitMessage,omitempty"`

	// CommitTime: When the revision was committed.
	CommitTime string `json:"commitTime,omitempty"`

	// Committer: The name of the user who committed the revision.
	Committer string `json:"committer,omitempty"`

	// CreateTime: When the revision was made. This may or may not be
	// reliable, depending on
	// the version control system being used.
	CreateTime string `json:"createTime,omitempty"`

	// Id: The unique ID of the revision. For many version control systems,
	// this
	// will be string of hex digits representing a hash value.
	Id string `json:"id,omitempty"`

	// ParentIds: The revision IDs of this revision's parents.
	ParentIds []string `json:"parentIds,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Author") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Author") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Revision) MarshalJSON() ([]byte, error) {
	type noMethod Revision
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Snapshot: A snapshot is a version of a workspace. Each change to a
// workspace's files
// creates a new snapshot. A workspace consists of a sequence of
// snapshots.
type Snapshot struct {
	// ChangedFiles: The set of files modified in this snapshot, relative to
	// the workspace
	// baseline. ChangedFileInfo.from_path is not set.
	ChangedFiles []*ChangedFileInfo `json:"changedFiles,omitempty"`

	// CreateTime: Timestamp when the snapshot was created.
	CreateTime string `json:"createTime,omitempty"`

	// SnapshotId: The ID of the snapshot.
	SnapshotId string `json:"snapshotId,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ChangedFiles") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "ChangedFiles") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Snapshot) MarshalJSON() ([]byte, error) {
	type noMethod Snapshot
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// SourceContext: A SourceContext is a reference to a tree of files. A
// SourceContext together
// with a path point to a unique revision of a single file or directory.
type SourceContext struct {
	// CloudRepo: A SourceContext referring to a revision in a cloud repo.
	CloudRepo *CloudRepoSourceContext `json:"cloudRepo,omitempty"`

	// CloudWorkspace: A SourceContext referring to a snapshot in a cloud
	// workspace.
	CloudWorkspace *CloudWorkspaceSourceContext `json:"cloudWorkspace,omitempty"`

	// Gerrit: A SourceContext referring to a Gerrit project.
	Gerrit *GerritSourceContext `json:"gerrit,omitempty"`

	// Git: A SourceContext referring to any third party Git repo (e.g.
	// GitHub).
	Git *GitSourceContext `json:"git,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CloudRepo") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "CloudRepo") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *SourceContext) MarshalJSON() ([]byte, error) {
	type noMethod SourceContext
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// UpdateRepoRequest: Request for UpdateRepo.
type UpdateRepoRequest struct {
	// RepoId: The ID of the repo to be updated.
	RepoId *RepoId `json:"repoId,omitempty"`

	// RepoName: Renames the repo. repo_name cannot already be in use by a
	// LIVE repo
	// within the project. This field is ignored if left blank or set to the
	// empty
	// string. If you want to rename a repo to "default," you need to
	// explicitly
	// set that value here.
	RepoName string `json:"repoName,omitempty"`

	// RepoSyncConfig: Sets or updates the RepoSync config. When the
	// repo_sync_config field is not
	// set it actually clears the repo sync config.
	RepoSyncConfig *RepoSyncConfig `json:"repoSyncConfig,omitempty"`

	// ForceSendFields is a list of field names (e.g. "RepoId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "RepoId") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *UpdateRepoRequest) MarshalJSON() ([]byte, error) {
	type noMethod UpdateRepoRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// Workspace: A Cloud Workspace stores modified files before they are
// committed to
// a repo. This message contains metadata. Use the Read
// or
// ReadFromWorkspaceOrAlias methods to read files from the
// workspace,
// and use ModifyWorkspace to change files.
type Workspace struct {
	// Alias: The alias associated with the workspace. When the workspace is
	// committed,
	// this alias will be moved to point to the new revision.
	Alias string `json:"alias,omitempty"`

	// Baseline: The revision of the workspace's alias when the workspace
	// was
	// created.
	Baseline string `json:"baseline,omitempty"`

	// ChangedFiles: The set of files modified in this workspace.
	ChangedFiles []*ChangedFileInfo `json:"changedFiles,omitempty"`

	// CurrentSnapshotId: If non-empty, current_snapshot_id refers to the
	// most recent update to the
	// workspace.
	CurrentSnapshotId string `json:"currentSnapshotId,omitempty"`

	// Id: The ID of the workspace.
	Id *CloudWorkspaceId `json:"id,omitempty"`

	// MergeInfo: Information needed to manage a refresh or merge operation.
	// Present only
	// during a merge (i.e. after a call to Merge) or a call
	// to
	// RefreshWorkspace which results in conflicts.
	MergeInfo *MergeInfo `json:"mergeInfo,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Alias") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Alias") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *Workspace) MarshalJSON() ([]byte, error) {
	type noMethod Workspace
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// WriteAction: Create or modify a file.
type WriteAction struct {
	// Contents: The new contents of the file.
	Contents string `json:"contents,omitempty"`

	// Mode: The new mode of the file.
	//
	// Possible values:
	//   "FILE_MODE_UNSPECIFIED" - No file mode was specified.
	//   "NORMAL" - Neither a symbolic link nor executable.
	//   "SYMLINK" - A symbolic link.
	//   "EXECUTABLE" - An executable.
	Mode string `json:"mode,omitempty"`

	// Path: The path of the file to write.
	Path string `json:"path,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Contents") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Contents") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

func (s *WriteAction) MarshalJSON() ([]byte, error) {
	type noMethod WriteAction
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields, s.NullFields)
}

// method id "source.projects.repos.create":

type ProjectsReposCreateCall struct {
	s          *Service
	projectId  string
	repo       *Repo
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Create: Creates a repo in the given project. The provided repo
// message should have
// its name field set to the desired repo name. No other repo fields
// should
// be set. Omitting the name is the same as specifying "default"
//
// Repo names must satisfy the regular expression
// `a-z{1,61}[a-z0-9]`. (Note that repo names must contain at
// least three characters and may not contain underscores.) The special
// name
// "default" is the default repo for the project; this is the repo shown
// when
// visiting the Cloud Developers Console, and can be accessed via git's
// HTTP
// protocol at `https://source.developers.google.com/p/PROJECT_ID`. You
// may
// create other repos with this API and access them
// at
// `https://source.developers.google.com/p/PROJECT_ID/r/NAME`.
func (r *ProjectsReposService) Create(projectId string, repo *Repo) *ProjectsReposCreateCall {
	c := &ProjectsReposCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repo = repo
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposCreateCall) Fields(s ...googleapi.Field) *ProjectsReposCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposCreateCall) Context(ctx context.Context) *ProjectsReposCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.repo)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.create" call.
// Exactly one of *Repo or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Repo.ServerResponse.Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ProjectsReposCreateCall) Do(opts ...googleapi.CallOption) (*Repo, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Repo{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates a repo in the given project. The provided repo message should have\nits name field set to the desired repo name. No other repo fields should\nbe set. Omitting the name is the same as specifying \"default\"\n\nRepo names must satisfy the regular expression\n`a-z{1,61}[a-z0-9]`. (Note that repo names must contain at\nleast three characters and may not contain underscores.) The special name\n\"default\" is the default repo for the project; this is the repo shown when\nvisiting the Cloud Developers Console, and can be accessed via git's HTTP\nprotocol at `https://source.developers.google.com/p/PROJECT_ID`. You may\ncreate other repos with this API and access them at\n`https://source.developers.google.com/p/PROJECT_ID/r/NAME`.",
	//   "flatPath": "v1/projects/{projectId}/repos",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.create",
	//   "parameterOrder": [
	//     "projectId"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The project in which to create the repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos",
	//   "request": {
	//     "$ref": "Repo"
	//   },
	//   "response": {
	//     "$ref": "Repo"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.delete":

type ProjectsReposDeleteCall struct {
	s          *Service
	projectId  string
	repoName   string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes a repo.
func (r *ProjectsReposService) Delete(projectId string, repoName string) *ProjectsReposDeleteCall {
	c := &ProjectsReposDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposDeleteCall) RepoIdUid(repoIdUid string) *ProjectsReposDeleteCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposDeleteCall) Fields(s ...googleapi.Field) *ProjectsReposDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposDeleteCall) Context(ctx context.Context) *ProjectsReposDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposDeleteCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsReposDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Deletes a repo.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}",
	//   "httpMethod": "DELETE",
	//   "id": "source.projects.repos.delete",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.get":

type ProjectsReposGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Returns information about a repo.
func (r *ProjectsReposService) Get(projectId string, repoName string) *ProjectsReposGetCall {
	c := &ProjectsReposGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposGetCall) RepoIdUid(repoIdUid string) *ProjectsReposGetCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposGetCall) Fields(s ...googleapi.Field) *ProjectsReposGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposGetCall) IfNoneMatch(entityTag string) *ProjectsReposGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposGetCall) Context(ctx context.Context) *ProjectsReposGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.get" call.
// Exactly one of *Repo or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Repo.ServerResponse.Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ProjectsReposGetCall) Do(opts ...googleapi.CallOption) (*Repo, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Repo{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns information about a repo.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}",
	//   "response": {
	//     "$ref": "Repo"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.list":

type ProjectsReposListCall struct {
	s            *Service
	projectId    string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Returns all repos belonging to a project, specified by its
// project ID. The
// response list is sorted by name with the default repo listed first.
func (r *ProjectsReposService) List(projectId string) *ProjectsReposListCall {
	c := &ProjectsReposListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposListCall) Fields(s ...googleapi.Field) *ProjectsReposListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposListCall) IfNoneMatch(entityTag string) *ProjectsReposListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposListCall) Context(ctx context.Context) *ProjectsReposListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.list" call.
// Exactly one of *ListReposResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListReposResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposListCall) Do(opts ...googleapi.CallOption) (*ListReposResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListReposResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns all repos belonging to a project, specified by its project ID. The\nresponse list is sorted by name with the default repo listed first.",
	//   "flatPath": "v1/projects/{projectId}/repos",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.list",
	//   "parameterOrder": [
	//     "projectId"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The project ID whose repos should be listed.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos",
	//   "response": {
	//     "$ref": "ListReposResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.merge":

type ProjectsReposMergeCall struct {
	s            *Service
	projectId    string
	repoName     string
	mergerequest *MergeRequest
	urlParams_   gensupport.URLParams
	ctx_         context.Context
}

// Merge: Merges a revision into a movable alias, using a workspace
// associated with
// that alias to store modified files. The workspace must not have
// any
// modified files. Note that Merge neither creates the workspace nor
// commits
// it; those actions must be done separately. Returns ABORTED when
// the
// workspace is simultaneously modified by another client.
func (r *ProjectsReposService) Merge(projectId string, repoName string, mergerequest *MergeRequest) *ProjectsReposMergeCall {
	c := &ProjectsReposMergeCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.mergerequest = mergerequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposMergeCall) Fields(s ...googleapi.Field) *ProjectsReposMergeCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposMergeCall) Context(ctx context.Context) *ProjectsReposMergeCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposMergeCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.mergerequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}:merge")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.merge" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposMergeCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Merges a revision into a movable alias, using a workspace associated with\nthat alias to store modified files. The workspace must not have any\nmodified files. Note that Merge neither creates the workspace nor commits\nit; those actions must be done separately. Returns ABORTED when the\nworkspace is simultaneously modified by another client.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}:merge",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.merge",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}:merge",
	//   "request": {
	//     "$ref": "MergeRequest"
	//   },
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.update":

type ProjectsReposUpdateCall struct {
	s                 *Service
	projectId         string
	repoName          string
	updatereporequest *UpdateRepoRequest
	urlParams_        gensupport.URLParams
	ctx_              context.Context
}

// Update: Updates an existing repo. The only things you can change
// about a repo are:
//   1) its repo_sync_config (and then only to add one that is not
// present);
//   2) its last-updated time; and
//   3) its name.
func (r *ProjectsReposService) Update(projectId string, repoName string, updatereporequest *UpdateRepoRequest) *ProjectsReposUpdateCall {
	c := &ProjectsReposUpdateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.updatereporequest = updatereporequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposUpdateCall) Fields(s ...googleapi.Field) *ProjectsReposUpdateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposUpdateCall) Context(ctx context.Context) *ProjectsReposUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposUpdateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.updatereporequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.update" call.
// Exactly one of *Repo or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Repo.ServerResponse.Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ProjectsReposUpdateCall) Do(opts ...googleapi.CallOption) (*Repo, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Repo{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing repo. The only things you can change about a repo are:\n  1) its repo_sync_config (and then only to add one that is not present);\n  2) its last-updated time; and\n  3) its name.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}",
	//   "httpMethod": "PUT",
	//   "id": "source.projects.repos.update",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}",
	//   "request": {
	//     "$ref": "UpdateRepoRequest"
	//   },
	//   "response": {
	//     "$ref": "Repo"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.aliases.create":

type ProjectsReposAliasesCreateCall struct {
	s          *Service
	projectId  string
	repoName   string
	alias      *Alias
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Create: Creates a new alias. It is an ALREADY_EXISTS error if an
// alias with that
// name and kind already exists.
func (r *ProjectsReposAliasesService) Create(projectId string, repoName string, alias *Alias) *ProjectsReposAliasesCreateCall {
	c := &ProjectsReposAliasesCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.alias = alias
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposAliasesCreateCall) RepoIdUid(repoIdUid string) *ProjectsReposAliasesCreateCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposAliasesCreateCall) Fields(s ...googleapi.Field) *ProjectsReposAliasesCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposAliasesCreateCall) Context(ctx context.Context) *ProjectsReposAliasesCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposAliasesCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.alias)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/aliases")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.aliases.create" call.
// Exactly one of *Alias or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Alias.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsReposAliasesCreateCall) Do(opts ...googleapi.CallOption) (*Alias, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Alias{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates a new alias. It is an ALREADY_EXISTS error if an alias with that\nname and kind already exists.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/aliases",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.aliases.create",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/aliases",
	//   "request": {
	//     "$ref": "Alias"
	//   },
	//   "response": {
	//     "$ref": "Alias"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.aliases.delete":

type ProjectsReposAliasesDeleteCall struct {
	s          *Service
	projectId  string
	repoName   string
	kind       string
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes the alias with the given name and kind. Kind cannot
// be ANY.  If
// the alias does not exist, NOT_FOUND is returned.  If the request
// provides
// a revision ID and the alias does not refer to that revision, ABORTED
// is
// returned.
func (r *ProjectsReposAliasesService) Delete(projectId string, repoName string, kind string, name string) *ProjectsReposAliasesDeleteCall {
	c := &ProjectsReposAliasesDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.kind = kind
	c.name = name
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposAliasesDeleteCall) RepoIdUid(repoIdUid string) *ProjectsReposAliasesDeleteCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// RevisionId sets the optional parameter "revisionId": If non-empty,
// must match the revision that the alias refers to.
func (c *ProjectsReposAliasesDeleteCall) RevisionId(revisionId string) *ProjectsReposAliasesDeleteCall {
	c.urlParams_.Set("revisionId", revisionId)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposAliasesDeleteCall) Fields(s ...googleapi.Field) *ProjectsReposAliasesDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposAliasesDeleteCall) Context(ctx context.Context) *ProjectsReposAliasesDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposAliasesDeleteCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"kind":      c.kind,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.aliases.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsReposAliasesDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Deletes the alias with the given name and kind. Kind cannot be ANY.  If\nthe alias does not exist, NOT_FOUND is returned.  If the request provides\na revision ID and the alias does not refer to that revision, ABORTED is\nreturned.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}",
	//   "httpMethod": "DELETE",
	//   "id": "source.projects.repos.aliases.delete",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "kind",
	//     "name"
	//   ],
	//   "parameters": {
	//     "kind": {
	//       "description": "The kind of the alias to delete.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "MERCURIAL_BRANCH_DEPRECATED",
	//         "OTHER",
	//         "SPECIAL_DEPRECATED"
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The name of the alias to delete.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "revisionId": {
	//       "description": "If non-empty, must match the revision that the alias refers to.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.aliases.get":

type ProjectsReposAliasesGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	kind         string
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Returns information about an alias. Kind ANY returns a FIXED
// or
// MOVABLE alias, in that order, and ignores all other kinds.
func (r *ProjectsReposAliasesService) Get(projectId string, repoName string, kind string, name string) *ProjectsReposAliasesGetCall {
	c := &ProjectsReposAliasesGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.kind = kind
	c.name = name
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposAliasesGetCall) RepoIdUid(repoIdUid string) *ProjectsReposAliasesGetCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposAliasesGetCall) Fields(s ...googleapi.Field) *ProjectsReposAliasesGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposAliasesGetCall) IfNoneMatch(entityTag string) *ProjectsReposAliasesGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposAliasesGetCall) Context(ctx context.Context) *ProjectsReposAliasesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposAliasesGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"kind":      c.kind,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.aliases.get" call.
// Exactly one of *Alias or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Alias.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsReposAliasesGetCall) Do(opts ...googleapi.CallOption) (*Alias, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Alias{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns information about an alias. Kind ANY returns a FIXED or\nMOVABLE alias, in that order, and ignores all other kinds.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.aliases.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "kind",
	//     "name"
	//   ],
	//   "parameters": {
	//     "kind": {
	//       "description": "The kind of the alias.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "MERCURIAL_BRANCH_DEPRECATED",
	//         "OTHER",
	//         "SPECIAL_DEPRECATED"
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The alias name.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}",
	//   "response": {
	//     "$ref": "Alias"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.aliases.list":

type ProjectsReposAliasesListCall struct {
	s            *Service
	projectId    string
	repoName     string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Returns a list of aliases of the given kind. Kind ANY returns
// all aliases
// in the repo. The order in which the aliases are returned is
// undefined.
func (r *ProjectsReposAliasesService) List(projectId string, repoName string) *ProjectsReposAliasesListCall {
	c := &ProjectsReposAliasesListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	return c
}

// Kind sets the optional parameter "kind": Return only aliases of this
// kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "MERCURIAL_BRANCH_DEPRECATED"
//   "OTHER"
//   "SPECIAL_DEPRECATED"
func (c *ProjectsReposAliasesListCall) Kind(kind string) *ProjectsReposAliasesListCall {
	c.urlParams_.Set("kind", kind)
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposAliasesListCall) PageSize(pageSize int64) *ProjectsReposAliasesListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposAliasesListCall) PageToken(pageToken string) *ProjectsReposAliasesListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposAliasesListCall) RepoIdUid(repoIdUid string) *ProjectsReposAliasesListCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposAliasesListCall) Fields(s ...googleapi.Field) *ProjectsReposAliasesListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposAliasesListCall) IfNoneMatch(entityTag string) *ProjectsReposAliasesListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposAliasesListCall) Context(ctx context.Context) *ProjectsReposAliasesListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposAliasesListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/aliases")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.aliases.list" call.
// Exactly one of *ListAliasesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListAliasesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposAliasesListCall) Do(opts ...googleapi.CallOption) (*ListAliasesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListAliasesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns a list of aliases of the given kind. Kind ANY returns all aliases\nin the repo. The order in which the aliases are returned is undefined.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/aliases",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.aliases.list",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "kind": {
	//       "description": "Return only aliases of this kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "MERCURIAL_BRANCH_DEPRECATED",
	//         "OTHER",
	//         "SPECIAL_DEPRECATED"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/aliases",
	//   "response": {
	//     "$ref": "ListAliasesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposAliasesListCall) Pages(ctx context.Context, f func(*ListAliasesResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.aliases.listFiles":

type ProjectsReposAliasesListFilesCall struct {
	s            *Service
	projectId    string
	repoName     string
	kind         string
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// ListFiles: ListFiles returns a list of all files in a SourceContext.
// The
// information about each file includes its path and its hash.
// The result is ordered by path. Pagination is supported.
func (r *ProjectsReposAliasesService) ListFiles(projectId string, repoName string, kind string, name string) *ProjectsReposAliasesListFilesCall {
	c := &ProjectsReposAliasesListFilesCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.kind = kind
	c.name = name
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposAliasesListFilesCall) PageSize(pageSize int64) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposAliasesListFilesCall) PageToken(pageToken string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudRepoRevisionId sets the optional parameter
// "sourceContext.cloudRepo.revisionId": A revision ID.
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudRepoRevisionId(sourceContextCloudRepoRevisionId string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.revisionId", sourceContextCloudRepoRevisionId)
	return c
}

// SourceContextCloudWorkspaceSnapshotId sets the optional parameter
// "sourceContext.cloudWorkspace.snapshotId": The ID of the snapshot.
// An empty snapshot_id refers to the most recent snapshot.
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudWorkspaceSnapshotId(sourceContextCloudWorkspaceSnapshotId string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.snapshotId", sourceContextCloudWorkspaceSnapshotId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdName sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.name": The unique
// name of the workspace within the repo.  This is the name
// chosen by the client in the Source API's CreateWorkspace method.
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudWorkspaceWorkspaceIdName(sourceContextCloudWorkspaceWorkspaceIdName string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.name", sourceContextCloudWorkspaceWorkspaceIdName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.project
// Id": The ID of the project.
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoNam
// e": The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposAliasesListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposAliasesListFilesCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposAliasesListFilesCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposAliasesListFilesCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposAliasesListFilesCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposAliasesListFilesCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposAliasesListFilesCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposAliasesListFilesCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposAliasesListFilesCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposAliasesListFilesCall) Fields(s ...googleapi.Field) *ProjectsReposAliasesListFilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposAliasesListFilesCall) IfNoneMatch(entityTag string) *ProjectsReposAliasesListFilesCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposAliasesListFilesCall) Context(ctx context.Context) *ProjectsReposAliasesListFilesCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposAliasesListFilesCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}:listFiles")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"kind":      c.kind,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.aliases.listFiles" call.
// Exactly one of *ListFilesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListFilesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposAliasesListFilesCall) Do(opts ...googleapi.CallOption) (*ListFilesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListFilesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "ListFiles returns a list of all files in a SourceContext. The\ninformation about each file includes its path and its hash.\nThe result is ordered by path. Pagination is supported.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}:listFiles",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.aliases.listFiles",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "kind",
	//     "name"
	//   ],
	//   "parameters": {
	//     "kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The alias name.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.revisionId": {
	//       "description": "A revision ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}:listFiles",
	//   "response": {
	//     "$ref": "ListFilesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposAliasesListFilesCall) Pages(ctx context.Context, f func(*ListFilesResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.aliases.update":

type ProjectsReposAliasesUpdateCall struct {
	s          *Service
	projectId  string
	repoName   string
	aliasesId  string
	alias      *Alias
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Update: Updates the alias with the given name and kind. Kind cannot
// be ANY.  If
// the alias does not exist, NOT_FOUND is returned. If the request
// provides
// an old revision ID and the alias does not refer to that revision,
// ABORTED
// is returned.
func (r *ProjectsReposAliasesService) Update(projectId string, repoName string, aliasesId string, alias *Alias) *ProjectsReposAliasesUpdateCall {
	c := &ProjectsReposAliasesUpdateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.aliasesId = aliasesId
	c.alias = alias
	return c
}

// OldRevisionId sets the optional parameter "oldRevisionId": If
// non-empty, must match the revision that the alias refers to.
func (c *ProjectsReposAliasesUpdateCall) OldRevisionId(oldRevisionId string) *ProjectsReposAliasesUpdateCall {
	c.urlParams_.Set("oldRevisionId", oldRevisionId)
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposAliasesUpdateCall) RepoIdUid(repoIdUid string) *ProjectsReposAliasesUpdateCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposAliasesUpdateCall) Fields(s ...googleapi.Field) *ProjectsReposAliasesUpdateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposAliasesUpdateCall) Context(ctx context.Context) *ProjectsReposAliasesUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposAliasesUpdateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.alias)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/aliases/{aliasesId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"aliasesId": c.aliasesId,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.aliases.update" call.
// Exactly one of *Alias or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Alias.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsReposAliasesUpdateCall) Do(opts ...googleapi.CallOption) (*Alias, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Alias{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates the alias with the given name and kind. Kind cannot be ANY.  If\nthe alias does not exist, NOT_FOUND is returned. If the request provides\nan old revision ID and the alias does not refer to that revision, ABORTED\nis returned.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/aliases/{aliasesId}",
	//   "httpMethod": "PUT",
	//   "id": "source.projects.repos.aliases.update",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "aliasesId"
	//   ],
	//   "parameters": {
	//     "aliasesId": {
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "oldRevisionId": {
	//       "description": "If non-empty, must match the revision that the alias refers to.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/aliases/{aliasesId}",
	//   "request": {
	//     "$ref": "Alias"
	//   },
	//   "response": {
	//     "$ref": "Alias"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.aliases.files.get":

type ProjectsReposAliasesFilesGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	kind         string
	name         string
	path         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Read is given a SourceContext and path, and returns
// file or directory information about that path.
func (r *ProjectsReposAliasesFilesService) Get(projectId string, repoName string, kind string, name string, path string) *ProjectsReposAliasesFilesGetCall {
	c := &ProjectsReposAliasesFilesGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.kind = kind
	c.name = name
	c.path = path
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposAliasesFilesGetCall) PageSize(pageSize int64) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page, or if using start_index.
func (c *ProjectsReposAliasesFilesGetCall) PageToken(pageToken string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudRepoRevisionId sets the optional parameter
// "sourceContext.cloudRepo.revisionId": A revision ID.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudRepoRevisionId(sourceContextCloudRepoRevisionId string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.revisionId", sourceContextCloudRepoRevisionId)
	return c
}

// SourceContextCloudWorkspaceSnapshotId sets the optional parameter
// "sourceContext.cloudWorkspace.snapshotId": The ID of the snapshot.
// An empty snapshot_id refers to the most recent snapshot.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudWorkspaceSnapshotId(sourceContextCloudWorkspaceSnapshotId string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.snapshotId", sourceContextCloudWorkspaceSnapshotId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdName sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.name": The unique
// name of the workspace within the repo.  This is the name
// chosen by the client in the Source API's CreateWorkspace method.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdName(sourceContextCloudWorkspaceWorkspaceIdName string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.name", sourceContextCloudWorkspaceWorkspaceIdName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.project
// Id": The ID of the project.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoNam
// e": The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposAliasesFilesGetCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// StartPosition sets the optional parameter "startPosition": If path
// refers to a file, the position of the first byte of its contents
// to return. If path refers to a directory, the position of the first
// entry
// in the listing. If page_token is specified, this field is ignored.
func (c *ProjectsReposAliasesFilesGetCall) StartPosition(startPosition int64) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("startPosition", fmt.Sprint(startPosition))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposAliasesFilesGetCall) Fields(s ...googleapi.Field) *ProjectsReposAliasesFilesGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposAliasesFilesGetCall) IfNoneMatch(entityTag string) *ProjectsReposAliasesFilesGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposAliasesFilesGetCall) Context(ctx context.Context) *ProjectsReposAliasesFilesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposAliasesFilesGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}/files/{+path}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"kind":      c.kind,
		"name":      c.name,
		"path":      c.path,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.aliases.files.get" call.
// Exactly one of *ReadResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ReadResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposAliasesFilesGetCall) Do(opts ...googleapi.CallOption) (*ReadResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ReadResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Read is given a SourceContext and path, and returns\nfile or directory information about that path.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}/files/{filesId}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.aliases.files.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "kind",
	//     "name",
	//     "path"
	//   ],
	//   "parameters": {
	//     "kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The alias name.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page, or if using start_index.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "path": {
	//       "description": "Path to the file or directory from the root directory of the source\ncontext. It must not have leading or trailing slashes.",
	//       "location": "path",
	//       "pattern": "^.*$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.revisionId": {
	//       "description": "A revision ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "startPosition": {
	//       "description": "If path refers to a file, the position of the first byte of its contents\nto return. If path refers to a directory, the position of the first entry\nin the listing. If page_token is specified, this field is ignored.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/aliases/{kind}/{name}/files/{+path}",
	//   "response": {
	//     "$ref": "ReadResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposAliasesFilesGetCall) Pages(ctx context.Context, f func(*ReadResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.files.readFromWorkspaceOrAlias":

type ProjectsReposFilesReadFromWorkspaceOrAliasCall struct {
	s            *Service
	projectId    string
	repoName     string
	path         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// ReadFromWorkspaceOrAlias: ReadFromWorkspaceOrAlias performs a Read
// using either the most recent
// snapshot of the given workspace, if the workspace exists, or
// the
// revision referred to by the given alias if the workspace does not
// exist.
func (r *ProjectsReposFilesService) ReadFromWorkspaceOrAlias(projectId string, repoName string, path string) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c := &ProjectsReposFilesReadFromWorkspaceOrAliasCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.path = path
	return c
}

// Alias sets the optional parameter "alias": MOVABLE alias to read
// from, if the workspace doesn't exist.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) Alias(alias string) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.urlParams_.Set("alias", alias)
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) PageSize(pageSize int64) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) PageToken(pageToken string) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) RepoIdUid(repoIdUid string) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// StartPosition sets the optional parameter "startPosition": If path
// refers to a file, the position of the first byte of its contents
// to return. If path refers to a directory, the position of the first
// entry
// in the listing. If page_token is specified, this field is ignored.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) StartPosition(startPosition int64) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.urlParams_.Set("startPosition", fmt.Sprint(startPosition))
	return c
}

// WorkspaceName sets the optional parameter "workspaceName": Workspace
// to read from, if it exists.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) WorkspaceName(workspaceName string) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.urlParams_.Set("workspaceName", workspaceName)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) Fields(s ...googleapi.Field) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) IfNoneMatch(entityTag string) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) Context(ctx context.Context) *ProjectsReposFilesReadFromWorkspaceOrAliasCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/files/{+path}:readFromWorkspaceOrAlias")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"path":      c.path,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.files.readFromWorkspaceOrAlias" call.
// Exactly one of *ReadResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ReadResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) Do(opts ...googleapi.CallOption) (*ReadResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ReadResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "ReadFromWorkspaceOrAlias performs a Read using either the most recent\nsnapshot of the given workspace, if the workspace exists, or the\nrevision referred to by the given alias if the workspace does not exist.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/files/{filesId}:readFromWorkspaceOrAlias",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.files.readFromWorkspaceOrAlias",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "path"
	//   ],
	//   "parameters": {
	//     "alias": {
	//       "description": "MOVABLE alias to read from, if the workspace doesn't exist.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "path": {
	//       "description": "Path to the file or directory from the root directory of the source\ncontext. It must not have leading or trailing slashes.",
	//       "location": "path",
	//       "pattern": "^.*$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "startPosition": {
	//       "description": "If path refers to a file, the position of the first byte of its contents\nto return. If path refers to a directory, the position of the first entry\nin the listing. If page_token is specified, this field is ignored.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "workspaceName": {
	//       "description": "Workspace to read from, if it exists.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/files/{+path}:readFromWorkspaceOrAlias",
	//   "response": {
	//     "$ref": "ReadResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposFilesReadFromWorkspaceOrAliasCall) Pages(ctx context.Context, f func(*ReadResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.revisions.get":

type ProjectsReposRevisionsGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	revisionId   string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Retrieves revision metadata for a single revision.
func (r *ProjectsReposRevisionsService) Get(projectId string, repoName string, revisionId string) *ProjectsReposRevisionsGetCall {
	c := &ProjectsReposRevisionsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.revisionId = revisionId
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposRevisionsGetCall) RepoIdUid(repoIdUid string) *ProjectsReposRevisionsGetCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposRevisionsGetCall) Fields(s ...googleapi.Field) *ProjectsReposRevisionsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposRevisionsGetCall) IfNoneMatch(entityTag string) *ProjectsReposRevisionsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposRevisionsGetCall) Context(ctx context.Context) *ProjectsReposRevisionsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposRevisionsGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId":  c.projectId,
		"repoName":   c.repoName,
		"revisionId": c.revisionId,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.revisions.get" call.
// Exactly one of *Revision or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Revision.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposRevisionsGetCall) Do(opts ...googleapi.CallOption) (*Revision, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Revision{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves revision metadata for a single revision.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.revisions.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "revisionId"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "revisionId": {
	//       "description": "The ID of the revision.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}",
	//   "response": {
	//     "$ref": "Revision"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.revisions.getBatchGet":

type ProjectsReposRevisionsGetBatchGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// GetBatchGet: Retrieves revision metadata for several revisions at
// once. It returns an
// error if any retrieval fails.
func (r *ProjectsReposRevisionsService) GetBatchGet(projectId string, repoName string) *ProjectsReposRevisionsGetBatchGetCall {
	c := &ProjectsReposRevisionsGetBatchGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposRevisionsGetBatchGetCall) RepoIdUid(repoIdUid string) *ProjectsReposRevisionsGetBatchGetCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// RevisionIds sets the optional parameter "revisionIds": The revision
// IDs to retrieve.
func (c *ProjectsReposRevisionsGetBatchGetCall) RevisionIds(revisionIds ...string) *ProjectsReposRevisionsGetBatchGetCall {
	c.urlParams_.SetMulti("revisionIds", append([]string{}, revisionIds...))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposRevisionsGetBatchGetCall) Fields(s ...googleapi.Field) *ProjectsReposRevisionsGetBatchGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposRevisionsGetBatchGetCall) IfNoneMatch(entityTag string) *ProjectsReposRevisionsGetBatchGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposRevisionsGetBatchGetCall) Context(ctx context.Context) *ProjectsReposRevisionsGetBatchGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposRevisionsGetBatchGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/revisions:batchGet")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.revisions.getBatchGet" call.
// Exactly one of *GetRevisionsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetRevisionsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposRevisionsGetBatchGetCall) Do(opts ...googleapi.CallOption) (*GetRevisionsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &GetRevisionsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves revision metadata for several revisions at once. It returns an\nerror if any retrieval fails.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/revisions:batchGet",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.revisions.getBatchGet",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "revisionIds": {
	//       "description": "The revision IDs to retrieve.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/revisions:batchGet",
	//   "response": {
	//     "$ref": "GetRevisionsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.revisions.list":

type ProjectsReposRevisionsListCall struct {
	s            *Service
	projectId    string
	repoName     string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Retrieves all revisions topologically between the starts and
// ends.
// Uses the commit date to break ties in the topology (e.g. when a
// revision
// has two parents).
func (r *ProjectsReposRevisionsService) List(projectId string, repoName string) *ProjectsReposRevisionsListCall {
	c := &ProjectsReposRevisionsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	return c
}

// Ends sets the optional parameter "ends": Revision IDs (hexadecimal
// strings) that specify where the listing ends. If
// this field is present, the listing will contain only revisions that
// are
// topologically between starts and ends, inclusive.
func (c *ProjectsReposRevisionsListCall) Ends(ends ...string) *ProjectsReposRevisionsListCall {
	c.urlParams_.SetMulti("ends", append([]string{}, ends...))
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposRevisionsListCall) PageSize(pageSize int64) *ProjectsReposRevisionsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposRevisionsListCall) PageToken(pageToken string) *ProjectsReposRevisionsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Path sets the optional parameter "path": List only those revisions
// that modify path.
func (c *ProjectsReposRevisionsListCall) Path(path string) *ProjectsReposRevisionsListCall {
	c.urlParams_.Set("path", path)
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposRevisionsListCall) RepoIdUid(repoIdUid string) *ProjectsReposRevisionsListCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// Starts sets the optional parameter "starts": Revision IDs
// (hexadecimal strings) that specify where the listing
// begins. If empty, the repo heads (revisions with no children) are
// used.
func (c *ProjectsReposRevisionsListCall) Starts(starts ...string) *ProjectsReposRevisionsListCall {
	c.urlParams_.SetMulti("starts", append([]string{}, starts...))
	return c
}

// WalkDirection sets the optional parameter "walkDirection": The
// direction to walk the graph.
//
// Possible values:
//   "BACKWARD"
//   "FORWARD"
func (c *ProjectsReposRevisionsListCall) WalkDirection(walkDirection string) *ProjectsReposRevisionsListCall {
	c.urlParams_.Set("walkDirection", walkDirection)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposRevisionsListCall) Fields(s ...googleapi.Field) *ProjectsReposRevisionsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposRevisionsListCall) IfNoneMatch(entityTag string) *ProjectsReposRevisionsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposRevisionsListCall) Context(ctx context.Context) *ProjectsReposRevisionsListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposRevisionsListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/revisions")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.revisions.list" call.
// Exactly one of *ListRevisionsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListRevisionsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposRevisionsListCall) Do(opts ...googleapi.CallOption) (*ListRevisionsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListRevisionsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves all revisions topologically between the starts and ends.\nUses the commit date to break ties in the topology (e.g. when a revision\nhas two parents).",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/revisions",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.revisions.list",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "ends": {
	//       "description": "Revision IDs (hexadecimal strings) that specify where the listing ends. If\nthis field is present, the listing will contain only revisions that are\ntopologically between starts and ends, inclusive.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "path": {
	//       "description": "List only those revisions that modify path.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "starts": {
	//       "description": "Revision IDs (hexadecimal strings) that specify where the listing\nbegins. If empty, the repo heads (revisions with no children) are used.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     },
	//     "walkDirection": {
	//       "description": "The direction to walk the graph.",
	//       "enum": [
	//         "BACKWARD",
	//         "FORWARD"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/revisions",
	//   "response": {
	//     "$ref": "ListRevisionsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposRevisionsListCall) Pages(ctx context.Context, f func(*ListRevisionsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.revisions.listFiles":

type ProjectsReposRevisionsListFilesCall struct {
	s            *Service
	projectId    string
	repoName     string
	revisionId   string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// ListFiles: ListFiles returns a list of all files in a SourceContext.
// The
// information about each file includes its path and its hash.
// The result is ordered by path. Pagination is supported.
func (r *ProjectsReposRevisionsService) ListFiles(projectId string, repoName string, revisionId string) *ProjectsReposRevisionsListFilesCall {
	c := &ProjectsReposRevisionsListFilesCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.revisionId = revisionId
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposRevisionsListFilesCall) PageSize(pageSize int64) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposRevisionsListFilesCall) PageToken(pageToken string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasContextKind sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudRepoAliasContextKind(sourceContextCloudRepoAliasContextKind string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.kind", sourceContextCloudRepoAliasContextKind)
	return c
}

// SourceContextCloudRepoAliasContextName sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.name": The alias name.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudRepoAliasContextName(sourceContextCloudRepoAliasContextName string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.name", sourceContextCloudRepoAliasContextName)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudWorkspaceSnapshotId sets the optional parameter
// "sourceContext.cloudWorkspace.snapshotId": The ID of the snapshot.
// An empty snapshot_id refers to the most recent snapshot.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudWorkspaceSnapshotId(sourceContextCloudWorkspaceSnapshotId string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.snapshotId", sourceContextCloudWorkspaceSnapshotId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdName sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.name": The unique
// name of the workspace within the repo.  This is the name
// chosen by the client in the Source API's CreateWorkspace method.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudWorkspaceWorkspaceIdName(sourceContextCloudWorkspaceWorkspaceIdName string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.name", sourceContextCloudWorkspaceWorkspaceIdName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.project
// Id": The ID of the project.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoNam
// e": The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposRevisionsListFilesCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposRevisionsListFilesCall) Fields(s ...googleapi.Field) *ProjectsReposRevisionsListFilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposRevisionsListFilesCall) IfNoneMatch(entityTag string) *ProjectsReposRevisionsListFilesCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposRevisionsListFilesCall) Context(ctx context.Context) *ProjectsReposRevisionsListFilesCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposRevisionsListFilesCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}:listFiles")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId":  c.projectId,
		"repoName":   c.repoName,
		"revisionId": c.revisionId,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.revisions.listFiles" call.
// Exactly one of *ListFilesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListFilesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposRevisionsListFilesCall) Do(opts ...googleapi.CallOption) (*ListFilesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListFilesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "ListFiles returns a list of all files in a SourceContext. The\ninformation about each file includes its path and its hash.\nThe result is ordered by path. Pagination is supported.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}:listFiles",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.revisions.listFiles",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "revisionId"
	//   ],
	//   "parameters": {
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "revisionId": {
	//       "description": "A revision ID.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}:listFiles",
	//   "response": {
	//     "$ref": "ListFilesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposRevisionsListFilesCall) Pages(ctx context.Context, f func(*ListFilesResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.revisions.files.get":

type ProjectsReposRevisionsFilesGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	revisionId   string
	path         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Read is given a SourceContext and path, and returns
// file or directory information about that path.
func (r *ProjectsReposRevisionsFilesService) Get(projectId string, repoName string, revisionId string, path string) *ProjectsReposRevisionsFilesGetCall {
	c := &ProjectsReposRevisionsFilesGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.revisionId = revisionId
	c.path = path
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposRevisionsFilesGetCall) PageSize(pageSize int64) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page, or if using start_index.
func (c *ProjectsReposRevisionsFilesGetCall) PageToken(pageToken string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasContextKind sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudRepoAliasContextKind(sourceContextCloudRepoAliasContextKind string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.kind", sourceContextCloudRepoAliasContextKind)
	return c
}

// SourceContextCloudRepoAliasContextName sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.name": The alias name.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudRepoAliasContextName(sourceContextCloudRepoAliasContextName string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.name", sourceContextCloudRepoAliasContextName)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudWorkspaceSnapshotId sets the optional parameter
// "sourceContext.cloudWorkspace.snapshotId": The ID of the snapshot.
// An empty snapshot_id refers to the most recent snapshot.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudWorkspaceSnapshotId(sourceContextCloudWorkspaceSnapshotId string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.snapshotId", sourceContextCloudWorkspaceSnapshotId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdName sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.name": The unique
// name of the workspace within the repo.  This is the name
// chosen by the client in the Source API's CreateWorkspace method.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdName(sourceContextCloudWorkspaceWorkspaceIdName string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.name", sourceContextCloudWorkspaceWorkspaceIdName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.project
// Id": The ID of the project.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName
// sets the optional parameter
// "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoNam
// e": The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName(sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName", sourceContextCloudWorkspaceWorkspaceIdRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposRevisionsFilesGetCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// StartPosition sets the optional parameter "startPosition": If path
// refers to a file, the position of the first byte of its contents
// to return. If path refers to a directory, the position of the first
// entry
// in the listing. If page_token is specified, this field is ignored.
func (c *ProjectsReposRevisionsFilesGetCall) StartPosition(startPosition int64) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("startPosition", fmt.Sprint(startPosition))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposRevisionsFilesGetCall) Fields(s ...googleapi.Field) *ProjectsReposRevisionsFilesGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposRevisionsFilesGetCall) IfNoneMatch(entityTag string) *ProjectsReposRevisionsFilesGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposRevisionsFilesGetCall) Context(ctx context.Context) *ProjectsReposRevisionsFilesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposRevisionsFilesGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}/files/{+path}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId":  c.projectId,
		"repoName":   c.repoName,
		"revisionId": c.revisionId,
		"path":       c.path,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.revisions.files.get" call.
// Exactly one of *ReadResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ReadResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposRevisionsFilesGetCall) Do(opts ...googleapi.CallOption) (*ReadResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ReadResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Read is given a SourceContext and path, and returns\nfile or directory information about that path.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}/files/{filesId}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.revisions.files.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "revisionId",
	//     "path"
	//   ],
	//   "parameters": {
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page, or if using start_index.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "path": {
	//       "description": "Path to the file or directory from the root directory of the source\ncontext. It must not have leading or trailing slashes.",
	//       "location": "path",
	//       "pattern": "^.*$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "revisionId": {
	//       "description": "A revision ID.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "startPosition": {
	//       "description": "If path refers to a file, the position of the first byte of its contents\nto return. If path refers to a directory, the position of the first entry\nin the listing. If page_token is specified, this field is ignored.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/revisions/{revisionId}/files/{+path}",
	//   "response": {
	//     "$ref": "ReadResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposRevisionsFilesGetCall) Pages(ctx context.Context, f func(*ReadResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.workspaces.commitWorkspace":

type ProjectsReposWorkspacesCommitWorkspaceCall struct {
	s                      *Service
	projectId              string
	repoName               string
	name                   string
	commitworkspacerequest *CommitWorkspaceRequest
	urlParams_             gensupport.URLParams
	ctx_                   context.Context
}

// CommitWorkspace: Commits some or all of the modified files in a
// workspace. This creates a
// new revision in the repo with the workspace's contents. Returns
// ABORTED if the workspace ID
// in the request contains a snapshot ID and it is not the same as
// the
// workspace's current snapshot ID or if the workspace is
// simultaneously
// modified by another client.
func (r *ProjectsReposWorkspacesService) CommitWorkspace(projectId string, repoName string, name string, commitworkspacerequest *CommitWorkspaceRequest) *ProjectsReposWorkspacesCommitWorkspaceCall {
	c := &ProjectsReposWorkspacesCommitWorkspaceCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.commitworkspacerequest = commitworkspacerequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesCommitWorkspaceCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesCommitWorkspaceCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesCommitWorkspaceCall) Context(ctx context.Context) *ProjectsReposWorkspacesCommitWorkspaceCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesCommitWorkspaceCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.commitworkspacerequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:commitWorkspace")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.commitWorkspace" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesCommitWorkspaceCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Commits some or all of the modified files in a workspace. This creates a\nnew revision in the repo with the workspace's contents. Returns ABORTED if the workspace ID\nin the request contains a snapshot ID and it is not the same as the\nworkspace's current snapshot ID or if the workspace is simultaneously\nmodified by another client.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:commitWorkspace",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.workspaces.commitWorkspace",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:commitWorkspace",
	//   "request": {
	//     "$ref": "CommitWorkspaceRequest"
	//   },
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.create":

type ProjectsReposWorkspacesCreateCall struct {
	s                      *Service
	projectId              string
	repoName               string
	createworkspacerequest *CreateWorkspaceRequest
	urlParams_             gensupport.URLParams
	ctx_                   context.Context
}

// Create: Creates a workspace.
func (r *ProjectsReposWorkspacesService) Create(projectId string, repoName string, createworkspacerequest *CreateWorkspaceRequest) *ProjectsReposWorkspacesCreateCall {
	c := &ProjectsReposWorkspacesCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.createworkspacerequest = createworkspacerequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesCreateCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesCreateCall) Context(ctx context.Context) *ProjectsReposWorkspacesCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesCreateCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.createworkspacerequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.create" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesCreateCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates a workspace.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.workspaces.create",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces",
	//   "request": {
	//     "$ref": "CreateWorkspaceRequest"
	//   },
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.delete":

type ProjectsReposWorkspacesDeleteCall struct {
	s          *Service
	projectId  string
	repoName   string
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes a workspace. Uncommitted changes are lost. If the
// workspace does
// not exist, NOT_FOUND is returned. Returns ABORTED when the workspace
// is
// simultaneously modified by another client.
func (r *ProjectsReposWorkspacesService) Delete(projectId string, repoName string, name string) *ProjectsReposWorkspacesDeleteCall {
	c := &ProjectsReposWorkspacesDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	return c
}

// CurrentSnapshotId sets the optional parameter "currentSnapshotId": If
// non-empty, current_snapshot_id must refer to the most recent update
// to
// the workspace, or ABORTED is returned.
func (c *ProjectsReposWorkspacesDeleteCall) CurrentSnapshotId(currentSnapshotId string) *ProjectsReposWorkspacesDeleteCall {
	c.urlParams_.Set("currentSnapshotId", currentSnapshotId)
	return c
}

// WorkspaceIdRepoIdUid sets the optional parameter
// "workspaceId.repoId.uid": A server-assigned, globally unique
// identifier.
func (c *ProjectsReposWorkspacesDeleteCall) WorkspaceIdRepoIdUid(workspaceIdRepoIdUid string) *ProjectsReposWorkspacesDeleteCall {
	c.urlParams_.Set("workspaceId.repoId.uid", workspaceIdRepoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesDeleteCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesDeleteCall) Context(ctx context.Context) *ProjectsReposWorkspacesDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesDeleteCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsReposWorkspacesDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Empty{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Deletes a workspace. Uncommitted changes are lost. If the workspace does\nnot exist, NOT_FOUND is returned. Returns ABORTED when the workspace is\nsimultaneously modified by another client.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}",
	//   "httpMethod": "DELETE",
	//   "id": "source.projects.repos.workspaces.delete",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "currentSnapshotId": {
	//       "description": "If non-empty, current_snapshot_id must refer to the most recent update to\nthe workspace, or ABORTED is returned.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.get":

type ProjectsReposWorkspacesGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Returns workspace metadata.
func (r *ProjectsReposWorkspacesService) Get(projectId string, repoName string, name string) *ProjectsReposWorkspacesGetCall {
	c := &ProjectsReposWorkspacesGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	return c
}

// WorkspaceIdRepoIdUid sets the optional parameter
// "workspaceId.repoId.uid": A server-assigned, globally unique
// identifier.
func (c *ProjectsReposWorkspacesGetCall) WorkspaceIdRepoIdUid(workspaceIdRepoIdUid string) *ProjectsReposWorkspacesGetCall {
	c.urlParams_.Set("workspaceId.repoId.uid", workspaceIdRepoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesGetCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesGetCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesGetCall) Context(ctx context.Context) *ProjectsReposWorkspacesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.get" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesGetCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns workspace metadata.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}",
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.list":

type ProjectsReposWorkspacesListCall struct {
	s            *Service
	projectId    string
	repoName     string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Returns all workspaces belonging to a repo.
func (r *ProjectsReposWorkspacesService) List(projectId string, repoName string) *ProjectsReposWorkspacesListCall {
	c := &ProjectsReposWorkspacesListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	return c
}

// RepoIdUid sets the optional parameter "repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposWorkspacesListCall) RepoIdUid(repoIdUid string) *ProjectsReposWorkspacesListCall {
	c.urlParams_.Set("repoId.uid", repoIdUid)
	return c
}

// View sets the optional parameter "view": Specifies which parts of the
// Workspace resource should be returned in the
// response.
//
// Possible values:
//   "STANDARD"
//   "MINIMAL"
//   "FULL"
func (c *ProjectsReposWorkspacesListCall) View(view string) *ProjectsReposWorkspacesListCall {
	c.urlParams_.Set("view", view)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesListCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesListCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesListCall) Context(ctx context.Context) *ProjectsReposWorkspacesListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.list" call.
// Exactly one of *ListWorkspacesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListWorkspacesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesListCall) Do(opts ...googleapi.CallOption) (*ListWorkspacesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListWorkspacesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Returns all workspaces belonging to a repo.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.list",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "view": {
	//       "description": "Specifies which parts of the Workspace resource should be returned in the\nresponse.",
	//       "enum": [
	//         "STANDARD",
	//         "MINIMAL",
	//         "FULL"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces",
	//   "response": {
	//     "$ref": "ListWorkspacesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.listFiles":

type ProjectsReposWorkspacesListFilesCall struct {
	s            *Service
	projectId    string
	repoName     string
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// ListFiles: ListFiles returns a list of all files in a SourceContext.
// The
// information about each file includes its path and its hash.
// The result is ordered by path. Pagination is supported.
func (r *ProjectsReposWorkspacesService) ListFiles(projectId string, repoName string, name string) *ProjectsReposWorkspacesListFilesCall {
	c := &ProjectsReposWorkspacesListFilesCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposWorkspacesListFilesCall) PageSize(pageSize int64) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposWorkspacesListFilesCall) PageToken(pageToken string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasContextKind sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudRepoAliasContextKind(sourceContextCloudRepoAliasContextKind string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.kind", sourceContextCloudRepoAliasContextKind)
	return c
}

// SourceContextCloudRepoAliasContextName sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudRepoAliasContextName(sourceContextCloudRepoAliasContextName string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.name", sourceContextCloudRepoAliasContextName)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdProjectId sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.projectId":
// The ID of the project.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudRepoRepoIdProjectRepoIdProjectId(sourceContextCloudRepoRepoIdProjectRepoIdProjectId string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.projectId", sourceContextCloudRepoRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdRepoName sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.repoName":
// The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudRepoRepoIdProjectRepoIdRepoName(sourceContextCloudRepoRepoIdProjectRepoIdRepoName string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.repoName", sourceContextCloudRepoRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudRepoRevisionId sets the optional parameter
// "sourceContext.cloudRepo.revisionId": A revision ID.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudRepoRevisionId(sourceContextCloudRepoRevisionId string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.revisionId", sourceContextCloudRepoRevisionId)
	return c
}

// SourceContextCloudWorkspaceSnapshotId sets the optional parameter
// "sourceContext.cloudWorkspace.snapshotId": The ID of the snapshot.
// An empty snapshot_id refers to the most recent snapshot.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudWorkspaceSnapshotId(sourceContextCloudWorkspaceSnapshotId string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.snapshotId", sourceContextCloudWorkspaceSnapshotId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposWorkspacesListFilesCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesListFilesCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesListFilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesListFilesCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesListFilesCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesListFilesCall) Context(ctx context.Context) *ProjectsReposWorkspacesListFilesCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesListFilesCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:listFiles")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.listFiles" call.
// Exactly one of *ListFilesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListFilesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesListFilesCall) Do(opts ...googleapi.CallOption) (*ListFilesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListFilesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "ListFiles returns a list of all files in a SourceContext. The\ninformation about each file includes its path and its hash.\nThe result is ordered by path. Pagination is supported.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:listFiles",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.listFiles",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.revisionId": {
	//       "description": "A revision ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:listFiles",
	//   "response": {
	//     "$ref": "ListFilesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposWorkspacesListFilesCall) Pages(ctx context.Context, f func(*ListFilesResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.workspaces.modifyWorkspace":

type ProjectsReposWorkspacesModifyWorkspaceCall struct {
	s                      *Service
	projectId              string
	repoName               string
	name                   string
	modifyworkspacerequest *ModifyWorkspaceRequest
	urlParams_             gensupport.URLParams
	ctx_                   context.Context
}

// ModifyWorkspace: Applies an ordered sequence of file modification
// actions to a workspace.
// Returns ABORTED if current_snapshot_id in the request does not refer
// to
// the most recent update to the workspace or if the workspace
// is
// simultaneously modified by another client.
func (r *ProjectsReposWorkspacesService) ModifyWorkspace(projectId string, repoName string, name string, modifyworkspacerequest *ModifyWorkspaceRequest) *ProjectsReposWorkspacesModifyWorkspaceCall {
	c := &ProjectsReposWorkspacesModifyWorkspaceCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.modifyworkspacerequest = modifyworkspacerequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesModifyWorkspaceCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesModifyWorkspaceCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesModifyWorkspaceCall) Context(ctx context.Context) *ProjectsReposWorkspacesModifyWorkspaceCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesModifyWorkspaceCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.modifyworkspacerequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:modifyWorkspace")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.modifyWorkspace" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesModifyWorkspaceCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Applies an ordered sequence of file modification actions to a workspace.\nReturns ABORTED if current_snapshot_id in the request does not refer to\nthe most recent update to the workspace or if the workspace is\nsimultaneously modified by another client.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:modifyWorkspace",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.workspaces.modifyWorkspace",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:modifyWorkspace",
	//   "request": {
	//     "$ref": "ModifyWorkspaceRequest"
	//   },
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.refreshWorkspace":

type ProjectsReposWorkspacesRefreshWorkspaceCall struct {
	s                       *Service
	projectId               string
	repoName                string
	name                    string
	refreshworkspacerequest *RefreshWorkspaceRequest
	urlParams_              gensupport.URLParams
	ctx_                    context.Context
}

// RefreshWorkspace: Brings a workspace up to date by merging in the
// changes made between its
// baseline and the revision to which its alias currently
// refers.
// FAILED_PRECONDITION is returned if the alias refers to a revision
// that is
// not a descendant of the workspace baseline, or if the workspace has
// no
// baseline. Returns ABORTED when the workspace is simultaneously
// modified by
// another client.
//
// A refresh may involve merging files in the workspace with files in
// the
// current alias revision. If this merge results in conflicts, then
// the
// workspace is in a merge state: the merge_info field of Workspace will
// be
// populated, and conflicting files in the workspace will contain
// conflict
// markers.
func (r *ProjectsReposWorkspacesService) RefreshWorkspace(projectId string, repoName string, name string, refreshworkspacerequest *RefreshWorkspaceRequest) *ProjectsReposWorkspacesRefreshWorkspaceCall {
	c := &ProjectsReposWorkspacesRefreshWorkspaceCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.refreshworkspacerequest = refreshworkspacerequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesRefreshWorkspaceCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesRefreshWorkspaceCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesRefreshWorkspaceCall) Context(ctx context.Context) *ProjectsReposWorkspacesRefreshWorkspaceCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesRefreshWorkspaceCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.refreshworkspacerequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:refreshWorkspace")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.refreshWorkspace" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesRefreshWorkspaceCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Brings a workspace up to date by merging in the changes made between its\nbaseline and the revision to which its alias currently refers.\nFAILED_PRECONDITION is returned if the alias refers to a revision that is\nnot a descendant of the workspace baseline, or if the workspace has no\nbaseline. Returns ABORTED when the workspace is simultaneously modified by\nanother client.\n\nA refresh may involve merging files in the workspace with files in the\ncurrent alias revision. If this merge results in conflicts, then the\nworkspace is in a merge state: the merge_info field of Workspace will be\npopulated, and conflicting files in the workspace will contain conflict\nmarkers.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:refreshWorkspace",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.workspaces.refreshWorkspace",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:refreshWorkspace",
	//   "request": {
	//     "$ref": "RefreshWorkspaceRequest"
	//   },
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.resolveFiles":

type ProjectsReposWorkspacesResolveFilesCall struct {
	s                   *Service
	projectId           string
	repoName            string
	name                string
	resolvefilesrequest *ResolveFilesRequest
	urlParams_          gensupport.URLParams
	ctx_                context.Context
}

// ResolveFiles: Marks files modified as part of a merge as having been
// resolved. Returns
// ABORTED when the workspace is simultaneously modified by another
// client.
func (r *ProjectsReposWorkspacesService) ResolveFiles(projectId string, repoName string, name string, resolvefilesrequest *ResolveFilesRequest) *ProjectsReposWorkspacesResolveFilesCall {
	c := &ProjectsReposWorkspacesResolveFilesCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.resolvefilesrequest = resolvefilesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesResolveFilesCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesResolveFilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesResolveFilesCall) Context(ctx context.Context) *ProjectsReposWorkspacesResolveFilesCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesResolveFilesCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.resolvefilesrequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:resolveFiles")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.resolveFiles" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesResolveFilesCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Marks files modified as part of a merge as having been resolved. Returns\nABORTED when the workspace is simultaneously modified by another client.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:resolveFiles",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.workspaces.resolveFiles",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:resolveFiles",
	//   "request": {
	//     "$ref": "ResolveFilesRequest"
	//   },
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.revertRefresh":

type ProjectsReposWorkspacesRevertRefreshCall struct {
	s                    *Service
	projectId            string
	repoName             string
	name                 string
	revertrefreshrequest *RevertRefreshRequest
	urlParams_           gensupport.URLParams
	ctx_                 context.Context
}

// RevertRefresh: If a call to RefreshWorkspace results in conflicts,
// use RevertRefresh to
// restore the workspace to the state it was in before the refresh.
// Returns
// FAILED_PRECONDITION if not preceded by a call to RefreshWorkspace, or
// if
// there are no unresolved conflicts remaining. Returns ABORTED when
// the
// workspace is simultaneously modified by another client.
func (r *ProjectsReposWorkspacesService) RevertRefresh(projectId string, repoName string, name string, revertrefreshrequest *RevertRefreshRequest) *ProjectsReposWorkspacesRevertRefreshCall {
	c := &ProjectsReposWorkspacesRevertRefreshCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.revertrefreshrequest = revertrefreshrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesRevertRefreshCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesRevertRefreshCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesRevertRefreshCall) Context(ctx context.Context) *ProjectsReposWorkspacesRevertRefreshCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesRevertRefreshCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.revertrefreshrequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:revertRefresh")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.revertRefresh" call.
// Exactly one of *Workspace or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Workspace.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesRevertRefreshCall) Do(opts ...googleapi.CallOption) (*Workspace, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Workspace{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "If a call to RefreshWorkspace results in conflicts, use RevertRefresh to\nrestore the workspace to the state it was in before the refresh.  Returns\nFAILED_PRECONDITION if not preceded by a call to RefreshWorkspace, or if\nthere are no unresolved conflicts remaining. Returns ABORTED when the\nworkspace is simultaneously modified by another client.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:revertRefresh",
	//   "httpMethod": "POST",
	//   "id": "source.projects.repos.workspaces.revertRefresh",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}:revertRefresh",
	//   "request": {
	//     "$ref": "RevertRefreshRequest"
	//   },
	//   "response": {
	//     "$ref": "Workspace"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.files.get":

type ProjectsReposWorkspacesFilesGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	name         string
	path         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Read is given a SourceContext and path, and returns
// file or directory information about that path.
func (r *ProjectsReposWorkspacesFilesService) Get(projectId string, repoName string, name string, path string) *ProjectsReposWorkspacesFilesGetCall {
	c := &ProjectsReposWorkspacesFilesGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.path = path
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposWorkspacesFilesGetCall) PageSize(pageSize int64) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page, or if using start_index.
func (c *ProjectsReposWorkspacesFilesGetCall) PageToken(pageToken string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasContextKind sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudRepoAliasContextKind(sourceContextCloudRepoAliasContextKind string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.kind", sourceContextCloudRepoAliasContextKind)
	return c
}

// SourceContextCloudRepoAliasContextName sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudRepoAliasContextName(sourceContextCloudRepoAliasContextName string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.name", sourceContextCloudRepoAliasContextName)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdProjectId sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.projectId":
// The ID of the project.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudRepoRepoIdProjectRepoIdProjectId(sourceContextCloudRepoRepoIdProjectRepoIdProjectId string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.projectId", sourceContextCloudRepoRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdRepoName sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.repoName":
// The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudRepoRepoIdProjectRepoIdRepoName(sourceContextCloudRepoRepoIdProjectRepoIdRepoName string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.repoName", sourceContextCloudRepoRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudRepoRevisionId sets the optional parameter
// "sourceContext.cloudRepo.revisionId": A revision ID.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudRepoRevisionId(sourceContextCloudRepoRevisionId string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.revisionId", sourceContextCloudRepoRevisionId)
	return c
}

// SourceContextCloudWorkspaceSnapshotId sets the optional parameter
// "sourceContext.cloudWorkspace.snapshotId": The ID of the snapshot.
// An empty snapshot_id refers to the most recent snapshot.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudWorkspaceSnapshotId(sourceContextCloudWorkspaceSnapshotId string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.snapshotId", sourceContextCloudWorkspaceSnapshotId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposWorkspacesFilesGetCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// StartPosition sets the optional parameter "startPosition": If path
// refers to a file, the position of the first byte of its contents
// to return. If path refers to a directory, the position of the first
// entry
// in the listing. If page_token is specified, this field is ignored.
func (c *ProjectsReposWorkspacesFilesGetCall) StartPosition(startPosition int64) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("startPosition", fmt.Sprint(startPosition))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesFilesGetCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesFilesGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesFilesGetCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesFilesGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesFilesGetCall) Context(ctx context.Context) *ProjectsReposWorkspacesFilesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesFilesGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/files/{+path}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
		"path":      c.path,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.files.get" call.
// Exactly one of *ReadResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ReadResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesFilesGetCall) Do(opts ...googleapi.CallOption) (*ReadResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ReadResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Read is given a SourceContext and path, and returns\nfile or directory information about that path.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/files/{filesId}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.files.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name",
	//     "path"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page, or if using start_index.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "path": {
	//       "description": "Path to the file or directory from the root directory of the source\ncontext. It must not have leading or trailing slashes.",
	//       "location": "path",
	//       "pattern": "^.*$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.revisionId": {
	//       "description": "A revision ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "startPosition": {
	//       "description": "If path refers to a file, the position of the first byte of its contents\nto return. If path refers to a directory, the position of the first entry\nin the listing. If page_token is specified, this field is ignored.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/files/{+path}",
	//   "response": {
	//     "$ref": "ReadResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposWorkspacesFilesGetCall) Pages(ctx context.Context, f func(*ReadResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.workspaces.snapshots.get":

type ProjectsReposWorkspacesSnapshotsGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	name         string
	snapshotId   string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Gets a workspace snapshot.
func (r *ProjectsReposWorkspacesSnapshotsService) Get(projectId string, repoName string, name string, snapshotId string) *ProjectsReposWorkspacesSnapshotsGetCall {
	c := &ProjectsReposWorkspacesSnapshotsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.snapshotId = snapshotId
	return c
}

// WorkspaceIdRepoIdUid sets the optional parameter
// "workspaceId.repoId.uid": A server-assigned, globally unique
// identifier.
func (c *ProjectsReposWorkspacesSnapshotsGetCall) WorkspaceIdRepoIdUid(workspaceIdRepoIdUid string) *ProjectsReposWorkspacesSnapshotsGetCall {
	c.urlParams_.Set("workspaceId.repoId.uid", workspaceIdRepoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesSnapshotsGetCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesSnapshotsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesSnapshotsGetCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesSnapshotsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesSnapshotsGetCall) Context(ctx context.Context) *ProjectsReposWorkspacesSnapshotsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesSnapshotsGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId":  c.projectId,
		"repoName":   c.repoName,
		"name":       c.name,
		"snapshotId": c.snapshotId,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.snapshots.get" call.
// Exactly one of *Snapshot or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Snapshot.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesSnapshotsGetCall) Do(opts ...googleapi.CallOption) (*Snapshot, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Snapshot{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets a workspace snapshot.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.snapshots.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name",
	//     "snapshotId"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "snapshotId": {
	//       "description": "The ID of the snapshot to get. If empty, the most recent snapshot is\nretrieved.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}",
	//   "response": {
	//     "$ref": "Snapshot"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "source.projects.repos.workspaces.snapshots.list":

type ProjectsReposWorkspacesSnapshotsListCall struct {
	s            *Service
	projectId    string
	repoName     string
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists all the snapshots made to a workspace, sorted from most
// recent to
// least recent.
func (r *ProjectsReposWorkspacesSnapshotsService) List(projectId string, repoName string, name string) *ProjectsReposWorkspacesSnapshotsListCall {
	c := &ProjectsReposWorkspacesSnapshotsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposWorkspacesSnapshotsListCall) PageSize(pageSize int64) *ProjectsReposWorkspacesSnapshotsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposWorkspacesSnapshotsListCall) PageToken(pageToken string) *ProjectsReposWorkspacesSnapshotsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// WorkspaceIdRepoIdUid sets the optional parameter
// "workspaceId.repoId.uid": A server-assigned, globally unique
// identifier.
func (c *ProjectsReposWorkspacesSnapshotsListCall) WorkspaceIdRepoIdUid(workspaceIdRepoIdUid string) *ProjectsReposWorkspacesSnapshotsListCall {
	c.urlParams_.Set("workspaceId.repoId.uid", workspaceIdRepoIdUid)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesSnapshotsListCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesSnapshotsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesSnapshotsListCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesSnapshotsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesSnapshotsListCall) Context(ctx context.Context) *ProjectsReposWorkspacesSnapshotsListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesSnapshotsListCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"repoName":  c.repoName,
		"name":      c.name,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.snapshots.list" call.
// Exactly one of *ListSnapshotsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListSnapshotsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesSnapshotsListCall) Do(opts ...googleapi.CallOption) (*ListSnapshotsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListSnapshotsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists all the snapshots made to a workspace, sorted from most recent to\nleast recent.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.snapshots.list",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots",
	//   "response": {
	//     "$ref": "ListSnapshotsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposWorkspacesSnapshotsListCall) Pages(ctx context.Context, f func(*ListSnapshotsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.workspaces.snapshots.listFiles":

type ProjectsReposWorkspacesSnapshotsListFilesCall struct {
	s            *Service
	projectId    string
	repoName     string
	name         string
	snapshotId   string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// ListFiles: ListFiles returns a list of all files in a SourceContext.
// The
// information about each file includes its path and its hash.
// The result is ordered by path. Pagination is supported.
func (r *ProjectsReposWorkspacesSnapshotsService) ListFiles(projectId string, repoName string, name string, snapshotId string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c := &ProjectsReposWorkspacesSnapshotsListFilesCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.snapshotId = snapshotId
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) PageSize(pageSize int64) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) PageToken(pageToken string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasContextKind sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudRepoAliasContextKind(sourceContextCloudRepoAliasContextKind string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.kind", sourceContextCloudRepoAliasContextKind)
	return c
}

// SourceContextCloudRepoAliasContextName sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudRepoAliasContextName(sourceContextCloudRepoAliasContextName string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.name", sourceContextCloudRepoAliasContextName)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdProjectId sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.projectId":
// The ID of the project.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudRepoRepoIdProjectRepoIdProjectId(sourceContextCloudRepoRepoIdProjectRepoIdProjectId string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.projectId", sourceContextCloudRepoRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdRepoName sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.repoName":
// The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudRepoRepoIdProjectRepoIdRepoName(sourceContextCloudRepoRepoIdProjectRepoIdRepoName string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.repoName", sourceContextCloudRepoRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudRepoRevisionId sets the optional parameter
// "sourceContext.cloudRepo.revisionId": A revision ID.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudRepoRevisionId(sourceContextCloudRepoRevisionId string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudRepo.revisionId", sourceContextCloudRepoRevisionId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) Context(ctx context.Context) *ProjectsReposWorkspacesSnapshotsListFilesCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}:listFiles")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId":  c.projectId,
		"repoName":   c.repoName,
		"name":       c.name,
		"snapshotId": c.snapshotId,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.snapshots.listFiles" call.
// Exactly one of *ListFilesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListFilesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) Do(opts ...googleapi.CallOption) (*ListFilesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListFilesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "ListFiles returns a list of all files in a SourceContext. The\ninformation about each file includes its path and its hash.\nThe result is ordered by path. Pagination is supported.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}:listFiles",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.snapshots.listFiles",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name",
	//     "snapshotId"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.revisionId": {
	//       "description": "A revision ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}:listFiles",
	//   "response": {
	//     "$ref": "ListFilesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposWorkspacesSnapshotsListFilesCall) Pages(ctx context.Context, f func(*ListFilesResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.projects.repos.workspaces.snapshots.files.get":

type ProjectsReposWorkspacesSnapshotsFilesGetCall struct {
	s            *Service
	projectId    string
	repoName     string
	name         string
	snapshotId   string
	path         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Read is given a SourceContext and path, and returns
// file or directory information about that path.
func (r *ProjectsReposWorkspacesSnapshotsFilesService) Get(projectId string, repoName string, name string, snapshotId string, path string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c := &ProjectsReposWorkspacesSnapshotsFilesGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.repoName = repoName
	c.name = name
	c.snapshotId = snapshotId
	c.path = path
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of values to return.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) PageSize(pageSize int64) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The value of
// next_page_token from the previous call.
// Omit for the first page, or if using start_index.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) PageToken(pageToken string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// SourceContextCloudRepoAliasContextKind sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudRepoAliasContextKind(sourceContextCloudRepoAliasContextKind string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.kind", sourceContextCloudRepoAliasContextKind)
	return c
}

// SourceContextCloudRepoAliasContextName sets the optional parameter
// "sourceContext.cloudRepo.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudRepoAliasContextName(sourceContextCloudRepoAliasContextName string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasContext.name", sourceContextCloudRepoAliasContextName)
	return c
}

// SourceContextCloudRepoAliasName sets the optional parameter
// "sourceContext.cloudRepo.aliasName": The name of an alias (branch,
// tag, etc.).
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudRepoAliasName(sourceContextCloudRepoAliasName string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.aliasName", sourceContextCloudRepoAliasName)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdProjectId sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.projectId":
// The ID of the project.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudRepoRepoIdProjectRepoIdProjectId(sourceContextCloudRepoRepoIdProjectRepoIdProjectId string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.projectId", sourceContextCloudRepoRepoIdProjectRepoIdProjectId)
	return c
}

// SourceContextCloudRepoRepoIdProjectRepoIdRepoName sets the optional
// parameter "sourceContext.cloudRepo.repoId.projectRepoId.repoName":
// The name of the repo. Leave empty for the default repo.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudRepoRepoIdProjectRepoIdRepoName(sourceContextCloudRepoRepoIdProjectRepoIdRepoName string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.projectRepoId.repoName", sourceContextCloudRepoRepoIdProjectRepoIdRepoName)
	return c
}

// SourceContextCloudRepoRepoIdUid sets the optional parameter
// "sourceContext.cloudRepo.repoId.uid": A server-assigned, globally
// unique identifier.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudRepoRepoIdUid(sourceContextCloudRepoRepoIdUid string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.repoId.uid", sourceContextCloudRepoRepoIdUid)
	return c
}

// SourceContextCloudRepoRevisionId sets the optional parameter
// "sourceContext.cloudRepo.revisionId": A revision ID.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudRepoRevisionId(sourceContextCloudRepoRevisionId string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudRepo.revisionId", sourceContextCloudRepoRevisionId)
	return c
}

// SourceContextCloudWorkspaceWorkspaceIdRepoIdUid sets the optional
// parameter "sourceContext.cloudWorkspace.workspaceId.repoId.uid": A
// server-assigned, globally unique identifier.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextCloudWorkspaceWorkspaceIdRepoIdUid(sourceContextCloudWorkspaceWorkspaceIdRepoIdUid string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.cloudWorkspace.workspaceId.repoId.uid", sourceContextCloudWorkspaceWorkspaceIdRepoIdUid)
	return c
}

// SourceContextGerritAliasContextKind sets the optional parameter
// "sourceContext.gerrit.aliasContext.kind": The alias kind.
//
// Possible values:
//   "ANY"
//   "FIXED"
//   "MOVABLE"
//   "OTHER"
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGerritAliasContextKind(sourceContextGerritAliasContextKind string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.kind", sourceContextGerritAliasContextKind)
	return c
}

// SourceContextGerritAliasContextName sets the optional parameter
// "sourceContext.gerrit.aliasContext.name": The alias name.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGerritAliasContextName(sourceContextGerritAliasContextName string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasContext.name", sourceContextGerritAliasContextName)
	return c
}

// SourceContextGerritAliasName sets the optional parameter
// "sourceContext.gerrit.aliasName": The name of an alias (branch, tag,
// etc.).
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGerritAliasName(sourceContextGerritAliasName string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.aliasName", sourceContextGerritAliasName)
	return c
}

// SourceContextGerritGerritProject sets the optional parameter
// "sourceContext.gerrit.gerritProject": The full project name within
// the host. Projects may be nested, so
// "project/subproject" is a valid project name.
// The "repo name" is hostURI/project.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGerritGerritProject(sourceContextGerritGerritProject string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.gerritProject", sourceContextGerritGerritProject)
	return c
}

// SourceContextGerritHostUri sets the optional parameter
// "sourceContext.gerrit.hostUri": The URI of a running Gerrit instance.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGerritHostUri(sourceContextGerritHostUri string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.hostUri", sourceContextGerritHostUri)
	return c
}

// SourceContextGerritRevisionId sets the optional parameter
// "sourceContext.gerrit.revisionId": A revision (commit) ID.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGerritRevisionId(sourceContextGerritRevisionId string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.gerrit.revisionId", sourceContextGerritRevisionId)
	return c
}

// SourceContextGitRevisionId sets the optional parameter
// "sourceContext.git.revisionId": Git commit hash.
// required.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGitRevisionId(sourceContextGitRevisionId string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.git.revisionId", sourceContextGitRevisionId)
	return c
}

// SourceContextGitUrl sets the optional parameter
// "sourceContext.git.url": Git repository URL.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) SourceContextGitUrl(sourceContextGitUrl string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("sourceContext.git.url", sourceContextGitUrl)
	return c
}

// StartPosition sets the optional parameter "startPosition": If path
// refers to a file, the position of the first byte of its contents
// to return. If path refers to a directory, the position of the first
// entry
// in the listing. If page_token is specified, this field is ignored.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) StartPosition(startPosition int64) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("startPosition", fmt.Sprint(startPosition))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) Fields(s ...googleapi.Field) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) IfNoneMatch(entityTag string) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) Context(ctx context.Context) *ProjectsReposWorkspacesSnapshotsFilesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		reqHeaders.Set("If-None-Match", c.ifNoneMatch_)
	}
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}/files/{+path}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	req.Header = reqHeaders
	googleapi.Expand(req.URL, map[string]string{
		"projectId":  c.projectId,
		"repoName":   c.repoName,
		"name":       c.name,
		"snapshotId": c.snapshotId,
		"path":       c.path,
	})
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.projects.repos.workspaces.snapshots.files.get" call.
// Exactly one of *ReadResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ReadResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) Do(opts ...googleapi.CallOption) (*ReadResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ReadResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Read is given a SourceContext and path, and returns\nfile or directory information about that path.",
	//   "flatPath": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}/files/{filesId}",
	//   "httpMethod": "GET",
	//   "id": "source.projects.repos.workspaces.snapshots.files.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "repoName",
	//     "name",
	//     "snapshotId",
	//     "path"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The unique name of the workspace within the repo.  This is the name\nchosen by the client in the Source API's CreateWorkspace method.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of values to return.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageToken": {
	//       "description": "The value of next_page_token from the previous call.\nOmit for the first page, or if using start_index.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "path": {
	//       "description": "Path to the file or directory from the root directory of the source\ncontext. It must not have leading or trailing slashes.",
	//       "location": "path",
	//       "pattern": "^.*$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "The ID of the project.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "snapshotId": {
	//       "description": "The ID of the snapshot.\nAn empty snapshot_id refers to the most recent snapshot.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.projectId": {
	//       "description": "The ID of the project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.projectRepoId.repoName": {
	//       "description": "The name of the repo. Leave empty for the default repo.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudRepo.revisionId": {
	//       "description": "A revision ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.cloudWorkspace.workspaceId.repoId.uid": {
	//       "description": "A server-assigned, globally unique identifier.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.kind": {
	//       "description": "The alias kind.",
	//       "enum": [
	//         "ANY",
	//         "FIXED",
	//         "MOVABLE",
	//         "OTHER"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasContext.name": {
	//       "description": "The alias name.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.aliasName": {
	//       "description": "The name of an alias (branch, tag, etc.).",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.gerritProject": {
	//       "description": "The full project name within the host. Projects may be nested, so\n\"project/subproject\" is a valid project name.\nThe \"repo name\" is hostURI/project.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.hostUri": {
	//       "description": "The URI of a running Gerrit instance.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.gerrit.revisionId": {
	//       "description": "A revision (commit) ID.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.revisionId": {
	//       "description": "Git commit hash.\nrequired.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "sourceContext.git.url": {
	//       "description": "Git repository URL.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "startPosition": {
	//       "description": "If path refers to a file, the position of the first byte of its contents\nto return. If path refers to a directory, the position of the first entry\nin the listing. If page_token is specified, this field is ignored.",
	//       "format": "int64",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/repos/{repoName}/workspaces/{name}/snapshots/{snapshotId}/files/{+path}",
	//   "response": {
	//     "$ref": "ReadResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsReposWorkspacesSnapshotsFilesGetCall) Pages(ctx context.Context, f func(*ReadResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "source.listChangedFiles":

type V1ListChangedFilesCall struct {
	s                       *Service
	listchangedfilesrequest *ListChangedFilesRequest
	urlParams_              gensupport.URLParams
	ctx_                    context.Context
}

// ListChangedFiles: ListChangedFiles computes the files that have
// changed between two revisions
// or workspace snapshots in the same repo. It returns a list
// of
// ChangeFileInfos.
//
// ListChangedFiles does not perform copy/rename detection, so the
// from_path of
// ChangeFileInfo is unset. Examine the changed_files field of the
// Revision
// resource to determine copy/rename information.
//
// The result is ordered by path. Pagination is supported.
func (r *V1Service) ListChangedFiles(listchangedfilesrequest *ListChangedFilesRequest) *V1ListChangedFilesCall {
	c := &V1ListChangedFilesCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.listchangedfilesrequest = listchangedfilesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *V1ListChangedFilesCall) Fields(s ...googleapi.Field) *V1ListChangedFilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *V1ListChangedFilesCall) Context(ctx context.Context) *V1ListChangedFilesCall {
	c.ctx_ = ctx
	return c
}

func (c *V1ListChangedFilesCall) doRequest(alt string) (*http.Response, error) {
	reqHeaders := make(http.Header)
	reqHeaders.Set("User-Agent", c.s.userAgent())
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.listchangedfilesrequest)
	if err != nil {
		return nil, err
	}
	reqHeaders.Set("Content-Type", "application/json")
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1:listChangedFiles")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	req.Header = reqHeaders
	return gensupport.SendRequest(c.ctx_, c.s.client, req)
}

// Do executes the "source.listChangedFiles" call.
// Exactly one of *ListChangedFilesResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ListChangedFilesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *V1ListChangedFilesCall) Do(opts ...googleapi.CallOption) (*ListChangedFilesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListChangedFilesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	target := &ret
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "ListChangedFiles computes the files that have changed between two revisions\nor workspace snapshots in the same repo. It returns a list of\nChangeFileInfos.\n\nListChangedFiles does not perform copy/rename detection, so the from_path of\nChangeFileInfo is unset. Examine the changed_files field of the Revision\nresource to determine copy/rename information.\n\nThe result is ordered by path. Pagination is supported.",
	//   "flatPath": "v1:listChangedFiles",
	//   "httpMethod": "POST",
	//   "id": "source.listChangedFiles",
	//   "parameterOrder": [],
	//   "parameters": {},
	//   "path": "v1:listChangedFiles",
	//   "request": {
	//     "$ref": "ListChangedFilesRequest"
	//   },
	//   "response": {
	//     "$ref": "ListChangedFilesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}
