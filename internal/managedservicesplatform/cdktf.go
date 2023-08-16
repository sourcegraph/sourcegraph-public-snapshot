package managedservicesplatform

import (
	"github.com/sourcegraph/conc/panics"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
)

type CDKTF struct {
	stacks *stack.Set
}

// OutputDir is the directory that Synthesize will place output in.
func (c CDKTF) OutputDir() string {
	if s := c.stacks.App.Outdir(); s != nil {
		return *s
	}
	return ""
}

// Synthesize all resources to the output directory that was originally
// configured.
func (c CDKTF) Synthesize() error {
	// CDKTF is prone to panics for no good reason, so make a best-effort
	// attempt to capture them.
	var catcher panics.Catcher
	catcher.Try(c.stacks.App.Synth)
	return catcher.Recovered().AsError()
}

func (c CDKTF) Stacks() []string {
	var names []string
	for _, s := range c.stacks.GetStacks() {
		names = append(names, s.Name)
	}
	return names
}
