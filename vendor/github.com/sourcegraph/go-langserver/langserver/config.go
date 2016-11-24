package langserver

import "os"

var (
	// GOLSP_WORKSPACE_REFERENCE_PARALLELISM controls the number of goroutines that
	// are used to handle workspace/reference requests. These typically are ran
	// in the background, not in response to a user request, and as such the
	// default is 1/4 the number of CPU. A minimum value of 1 is always enforced.
	envWorkspaceReferenceParallelism = os.Getenv("GOLSP_WORKSPACE_REFERENCE_PARALLELISM")

	// GOLSP_WARMUP_ON_INITIALIZE toggles if we typecheck the whole
	// workspace in the background on initialize. This trades off initial
	// CPU and memory to hide perceived latency of the first few
	// requests. If the LSP server is long lived the tradeoff is usually
	// worth it.
	envWarmupOnInitialize = os.Getenv("GOLSP_WARMUP_ON_INITIALIZE")
)
