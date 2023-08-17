package managedservicesplatform

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/conc/panics"
)

type CDKTF struct {
	app    cdktf.App
	stacks []string
}

// OutputDir is the directory that Synthesize will place output in.
func (c CDKTF) OutputDir() string {
	if s := c.app.Outdir(); s != nil {
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
	catcher.Try(c.app.Synth)
	return catcher.Recovered().AsError()
}

func (c CDKTF) Stacks() []string {
	return c.stacks
}
