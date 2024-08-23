package jsii

import "github.com/aws/jsii-runtime-go/internal/kernel"

// Close finalizes the runtime process, signalling the end of the execution to
// the jsii kernel process, and waiting for graceful termination. The best
// practice is to defer call this at the beginning of the "main" function.
func Close() {
	kernel.CloseClient()
}
