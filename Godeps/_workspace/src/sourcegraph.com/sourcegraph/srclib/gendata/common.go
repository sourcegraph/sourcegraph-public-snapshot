package gendata

import "fmt"

type GenDataOpt struct {
	Repo      string `short:"r" long:"repo" description:"repo to build" required:"yes"`
	CommitID  string `short:"c" long:"commit" description:"commit ID to build"`
	GenSource bool   `long:"gen-source" description:"whether to emit source files for the generated data"`
}

func (o *GenDataOpt) validate() error {
	if o.CommitID == "" && !o.GenSource {
		return fmt.Errorf("--commit must be non-empty or --gen-source must be true")
	}
	return nil
}

// GenDataCmd is a dummy command that serves as the parent of the
// data-generating subcommands.
type GenDataCmd struct{}

func (c *GenDataCmd) Execute(args []string) error { return nil }
