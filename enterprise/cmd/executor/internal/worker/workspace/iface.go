package workspace

import "context"

type CloneOptions struct {
	ExecutorName   string
	EndpointURL    string
	GitServicePath string
	ExecutorToken  string
}

type Workspace interface {
	// Path represents the block device path when firecracker is enabled and the
	// directory when firecracker is disabled where the workspace is configured.
	Path() string
	// ScriptFilenames holds the ordered set of script filenames to be invoked.
	ScriptFilenames() []string
	// Remove cleans up the workspace post execution. If keep workspace is true,
	// the implementation will only clean up additional resources, while keeping
	// the workspace contents on disk for debugging purposes.
	Remove(ctx context.Context, keepWorkspace bool)
}
