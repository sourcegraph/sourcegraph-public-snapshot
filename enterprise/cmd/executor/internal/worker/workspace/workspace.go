package workspace

import "context"

// CloneOptions holds the options for cloning a workspace.
type CloneOptions struct {
	ExecutorName   string
	EndpointURL    string
	GitServicePath string
	ExecutorToken  string
}

// Workspace represents a workspace that can be used to execute a job.
type Workspace interface {
	// Path represents the block device path when firecracker is enabled and the
	// directory when firecracker is disabled where the workspace is configured.
	Path() string
	// WorkingDirectory returns the working directory where the repository, scripts, and supporting files are located.
	WorkingDirectory() string
	// ScriptFilenames holds the ordered set of script filenames to be invoked.
	ScriptFilenames() []string
	// Remove cleans up the workspace post execution. If keep workspace is true,
	// the implementation will only clean up additional resources, while keeping
	// the workspace contents on disk for debugging purposes.
	Remove(ctx context.Context, keepWorkspace bool)
}
