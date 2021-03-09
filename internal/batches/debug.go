package batches

import "github.com/sourcegraph/src-cli/internal/output"

// DebugOut can be used to print debug messages in development to the TUI.
// For that it needs to be set to an actual *output.Output.
var DebugOut output.Writer = output.NoopWriter{}
