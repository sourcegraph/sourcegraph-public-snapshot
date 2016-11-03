package langserver

import "os"

var (
	// WORKSPACE_REFERENCE_PARALLELISM controls the number of goroutines that
	// are used to handle workspace/reference requests. These typically are ran
	// in the background, not in response to a user request, and as such the
	// default is 1/4 the number of CPU. A minimum value of 1 is always enforced.
	envWorkspaceReferenceParallelism = os.Getenv("WORKSPACE_REFERENCE_PARALLELISM")
)
