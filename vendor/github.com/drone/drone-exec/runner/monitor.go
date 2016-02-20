package runner

import (
	"io"

	"github.com/drone/drone-exec/parser"
)

// A MonitorFunc returns a monitor for the step described by the
// parameters (section is "clone", "build", etc., and key is the key
// of the multi-build build step, deploy plugin, etc.). The node may
// be used as a unique identifier for the step.
//
// The func must never return nil.
type MonitorFunc func(section, key string, node parser.Node) Monitor

// A Monitor is provided by consumers of drone-exec to track the state
// and receive the logs for specific steps. It is scoped to a specific
// step by virtue of being returned by a MonitorFunc.
type Monitor interface {
	// Start is called just prior to the given step beginning
	// execution.
	Start()

	// Skip is called just prior to the given step being skipped.
	Skip()

	// End is called after the given step has ended. If ok is true,
	// the step succeeded; if ok is false, it failed.
	//
	// If allowFailure is true, then this build step is allowed to
	// fail without failing the whole build.
	End(ok, allowFailure bool)

	// Logger is called to retrieve the destinations to write logs to
	// during execution of the given step.
	Logger() (stdout, stderr io.Writer)
}
