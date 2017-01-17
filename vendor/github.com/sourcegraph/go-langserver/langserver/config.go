package langserver

import "os"

var (
	// GOLSP_WARMUP_ON_INITIALIZE toggles if we typecheck the whole
	// workspace in the background on initialize. This trades off initial
	// CPU and memory to hide perceived latency of the first few
	// requests. If the LSP server is long lived the tradeoff is usually
	// worth it.
	envWarmupOnInitialize = os.Getenv("GOLSP_WARMUP_ON_INITIALIZE")
)
