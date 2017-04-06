package zap

import (
	"encoding/json"
	"errors"
	"fmt"
)

// WorkspaceIdentifier identifies a workspace.
type WorkspaceIdentifier struct {
	Dir string `json:"dir"` // workspace directory (usually the top-level git repository directory)
}

// Ref returns a RefIdentifier for the named ref in the workspace.
func (w WorkspaceIdentifier) Ref(name string) RefIdentifier {
	CheckRefName(name)
	return RefIdentifier{Repo: w.Dir, Ref: name}
}

// TEMPORARY: Remove when we remove the CheckXyz panic helpers.
func (w WorkspaceIdentifier) RefNoCheck(name string) RefIdentifier {
	return RefIdentifier{Repo: w.Dir, Ref: name}
}

// Branch returns a RefIdentifier for the named branch in the workspace.
func (w WorkspaceIdentifier) Branch(branch BranchName) RefIdentifier {
	CheckBranchName(branch)
	return w.Ref(branch.Ref())
}

// WorkspaceAddParams contains parameters for the "workspace/add"
// request.
type WorkspaceAddParams struct {
	WorkspaceIdentifier
}

// WorkspaceAddResult is the result from the "workspace/add" request.
type WorkspaceAddResult struct{}

// WorkspaceRemoveParams contains parameters for the
// "workspace/remove" request.
type WorkspaceRemoveParams struct {
	WorkspaceIdentifier
}

// WorkspaceRemoveResult is the result from the "workspace/remove"
// request.
type WorkspaceRemoveResult struct{}

// WorkspaceConfigureParams contains parameters for the
// "workspace/configure" request.
type WorkspaceConfigureParams struct {
	WorkspaceIdentifier
	WorkspaceConfiguration
}

// WorkspaceConfiguration holds configuration settings for a
// workspace.
// TODO(nick): delete
type WorkspaceConfiguration struct {
}

// RepoConfiguration describes the configuration for a repository.
type RepoConfiguration struct {
	// TODO(nick): delete
	Workspace *WorkspaceConfiguration `json:"workspace"`

	// Remotes contains the remote repositories that this server is
	// derived from. Currently only 1 remote is allowed, and it must
	// be named "default".
	Remotes map[string]RepoRemoteConfiguration `json:"remotes"`
}

func (c RepoConfiguration) DefaultRemote() *RepoRemoteConfiguration {
	if len(c.Remotes) > 1 {
		panic("expected only one remote")
	}
	for _, value := range c.Remotes {
		return &value
	}
	return nil
}

// TODO(sqs): hack to "safely" determine a default upstream, until we
// have a full config for this.
func (c *RepoConfiguration) DefaultUpstream() (string, error) {
	if len(c.Remotes) == 1 {
		for k := range c.Remotes {
			return k, nil
		}
	}
	return "", errors.New("unable to determine branch's default upstream: more than 1 remote exists")
}

func (c RepoConfiguration) String() string {
	return fmt.Sprintf("workspace(%+v) remotes(%+v)", c.Workspace, c.Remotes)
}

// DeepCopy returns a deep copy of c.
func (c RepoConfiguration) DeepCopy() RepoConfiguration {
	tmp, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	var copy RepoConfiguration
	if err := json.Unmarshal(tmp, &copy); err != nil {
		panic(err)
	}
	return copy
}

// AuthSetParams contains parameters for the "auth/set" request.
type AuthSetParams struct {
	// The auth token.
	Token string `json:"token"`
}

// AuthSetResult is the result from the "auth/set" request.
type AuthSetResult struct{}

// AuthGetResult is the result from the "auth/get" request.
type AuthGetResult struct {
	// The auth token.
	Token string `json:"token"`
}

// WorkspaceConfigureResult is the result from the
// "workspace/configure" request.
type WorkspaceConfigureResult struct{}

// WorkspaceServerCapabilities describes the capabilities provided by
// a Zap workspace server (which is intended to run on a user's
// machine and serve as a proxy between their editor and the remote
// upstream Zap server).
type WorkspaceServerCapabilities struct{}

// WorkspaceStatusParams contains parameters for the
// "workspace/status" request.
type WorkspaceStatusParams struct {
	WorkspaceIdentifier
}

// WorkspaceStatusResult is the result from the "workspace/status"
// request.
type WorkspaceStatusResult struct {
	// WorkRef is the name of the work ref for the local server (e.g.,
	// "head/alice").
	WorkRef string `json:"workRef"`
}

// WorkspaceBranchCreateParams contains parameters for the
// "workspace/branch/create" request.
type WorkspaceBranchCreateParams struct {
	WorkspaceIdentifier // the workspace in which to run this operation

	// Branch is the name of the branch to create. If empty, it uses
	// the current Git branch name as the Zap branch name.
	Branch BranchName `json:"branch,omitempty"`

	// Overwrite is whether to overwrite the branch with the current
	// workspace state, deleting all of the branch's existing state.
	//
	// If Overwrite is true, the branch must exist. If Overwrite is
	// false, the branch must not exist.
	Overwrite bool `json:"overwrite,omitempty"`

	// BufferFiles is a map of buffer filename (e.g.,
	// "#mydir/myfile.txt") to contents.
	//
	// If BufferFiles is set (by an editor), it is used as the source
	// of truth for unsaved files in the editor. The workspace server
	// also stores unsaved files' contents, but the BufferFiles map
	// (if set) is used instead. This is useful because the desired
	// behavior when creating a branch from your editor is "keep the
	// current contents of all unsaved files in my editor"; if there
	// was a syncing error and the workspace server's unsaved file
	// contents had diverged from what was in your editor, you would
	// want not want creating a branch to use the (old) data from the
	// workspace server.
	//
	// If BufferFiles is nil, then the workspace server uses the
	// unsaved file contents it is aware of (not the actual contents
	// of unsaved files in your editor). This occurs when
	// workspace/branch/create is called from the CLI (`zap create`).
	//
	// BufferFiles doesn't contain the contents of *saved* files (only
	// unsaved, a.k.a. buffered, files). Those are supplied by the
	// local server when it receives the workspace/branch/create
	// request from the editor. The local server is responsible for
	// computing the diffs between each local file on disk and its
	// unsaved contents sent by the editor in this field.
	BufferFiles map[string]string `json:"bufferFiles"`
}

// WorkspaceBranchCreateResult is the result of the
// "workspace/branch/create" request.
type WorkspaceBranchCreateResult struct {
	Ref  string  `json:"ref"`  // the name of the ref that was created (e.g., "branch/foo")
	Data RefData `json:"data"` // the ref data of the newly created branch
}

// WorkspaceBranchSetParams contains parameters for the
// "workspace/branch/set" request.
type WorkspaceBranchSetParams struct {
	WorkspaceIdentifier // the workspace in which to run this operation

	// Branch is the name of the branch to create. If empty, it uses
	// the current Git branch name as the Zap branch name.
	Branch BranchName `json:"branch,omitempty"`
}

// WorkspaceBranchSetResult is the result of the
// "workspace/branch/set" request.
type WorkspaceBranchSetResult struct {
	Ref string `json:"ref"` // the name of the ref that was set to (e.g., "branch/foo")
}

// A WorkspaceStateConflictError occurs during a call to
// "workspace/branch/set" when the workspace's base Git commit and
// branch doesn't match that of the
// (WorkspaceBranchSetParams).Branch. It describes the differences.
type WorkspaceStateConflictError struct {
	Branch BranchName `json:"branch"` // the Zap branch name being set

	DirtyGitWorktree bool    `json:"dirtyGitWorktree"`
	RequiredBase     RefBase `json:"requiredBase"`
	CurrentBase      RefBase `json:"currentBase"`
}

func (e *WorkspaceStateConflictError) Error() string {
	var reasons []string
	if e.DirtyGitWorktree {
		reasons = append(reasons, "git worktree is dirty")
	}
	if e.RequiredBase.GitBase != e.CurrentBase.GitBase {
		reasons = append(reasons, fmt.Sprintf("git HEAD commit is %s but need %s", e.RequiredBase.GitBase, e.CurrentBase.GitBase))
	}
	if e.RequiredBase.GitBranch != e.CurrentBase.GitBranch {
		reasons = append(reasons, fmt.Sprintf("git HEAD branch is %s but need %s", e.RequiredBase.GitBranch, e.CurrentBase.GitBranch))
	}
	return fmt.Sprintf("workspace state conflict: %v", reasons)
}

// WorkspaceBranchCloseParams contains parameters for the
// "workspace/branch/close" request.
type WorkspaceBranchCloseParams struct {
	WorkspaceIdentifier // the workspace in which to run this operation
}

// WorkspaceWillSaveFileParams contains parameters for the
// "workspace/willSaveFile" request.
type WorkspaceWillSaveFileParams struct {
	URI string `json:"uri"` // URI of file that will be saved
}
