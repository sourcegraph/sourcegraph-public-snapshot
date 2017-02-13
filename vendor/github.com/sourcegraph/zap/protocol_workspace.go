package zap

import "fmt"

// WorkspaceIdentifier identifies a workspace.
type WorkspaceIdentifier struct {
	Dir string `json:"dir"` // workspace directory (usually the top-level git repository directory)
}

// Ref returns a RefIdentifier for the named ref in the workspace.
func (w WorkspaceIdentifier) Ref(name string) RefIdentifier {
	return RefIdentifier{Repo: w.Dir, Ref: name}
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
type WorkspaceConfiguration struct {
	// SyncTree is whether to sync the contents of this workspace
	// directory (and its subdirectories) with the upstream in the
	// background. If true, the tree will be synced, even if there is
	// no client currently watching the workspace.
	SyncTree bool `json:"syncTree"`
}

// RepoConfiguration describes the configuration for a repository.
type RepoConfiguration struct {
	Workspace *WorkspaceConfiguration            `json:"workspace"`
	Remotes   map[string]RepoRemoteConfiguration `json:"remotes"`
	Refs      map[string]RefConfiguration        `json:"refs"`
}

func (c RepoConfiguration) String() string {
	return fmt.Sprintf("workspace(%+v) remotes(%+v) refs(%+v)", c.Workspace, c.Remotes, c.Refs)
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

// WorkspaceCheckoutParams contains parameters for the
// "workspace/checkout" request.
type WorkspaceCheckoutParams struct {
	WorkspaceIdentifier // the workspace in which to check out the ref

	Ref string `json:"ref"` // the ref to check out
}

// WorkspaceResetParams contains parameters for the
// "workspace/reset" request.
type WorkspaceResetParams struct {
	WorkspaceIdentifier // the workspace to reset

	// BufferFiles is a map of buffer filename (e.g.,
	// "#mydir/myfile.txt") to contents. It contains the contents of
	// all unsaved files in the editor when the editor sent the
	// workspace/reset command.
	//
	// Files doesn't contain the contents of saved files. Those are
	// supplied by the local server when it receives the
	// workspace/reset request from the editor. The local server is
	// responsible for computing the diffs between each local file on
	// disk and its unsaved contents sent by the editor in this field.
	BufferFiles map[string]string `json:"bufferFiles"`

	// Ref is the current ref name.
	//
	// It is used to check consistency, to ensure the reset is
	// performed on the correct ref. If it does not match the current
	// ref name, the request fails; it doesn't check out the specified
	// ref. To check out a different ref and reset, you must send
	// workspace/checkout followed by workspace/reset.
	Ref string `json:"ref"`
}

// WorkspaceWillSaveFileParams contains parameters for the
// "workspace/willSaveFile" request.
type WorkspaceWillSaveFileParams struct {
	URI string `json:"uri"` // URI of file that will be saved
}
