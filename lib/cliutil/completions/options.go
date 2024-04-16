package completions

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// defaultCompletions renders the default completions for a command.
func defaultCompletions() cli.BashCompleteFunc {
	return func(c *cli.Context) { cli.DefaultCompleteWithFlags(c.Command)(c) }
}

// CompleteArgs provides autocompletions based on the options returned by
// generateOptions. generateOptions is provided for every argument.
//
// generateOptions must not write to output, or reference any
// resources that are initialized elsewhere.
func CompleteArgs(generateOptions func() (options []string)) cli.BashCompleteFunc {
	return func(c *cli.Context) {
		CompletePositionalArgs(func(_ cli.Args) (options []string) {
			return generateOptions()
		})(c)
	}
}

// CompletePositionalArgs provides autocompletions for multiple positional
// arguments based on the options returned by the corresponding options generator.
// If there are more arguments than generators, no completion is provided.
//
// Each generator must not write to output, or reference any resources that are
// initialized elsewhere.
func CompletePositionalArgs(generators ...func(args cli.Args) (options []string)) cli.BashCompleteFunc {
	return func(c *cli.Context) {
		// Let default handler deal with flag completions if the latest argument
		// looks like the start of a flag
		if c.NArg() > 0 {
			if f := c.Args().Get(c.NArg() - 1); f == "-" || f == "--" {
				defaultCompletions()(c)
				return
			}
		}

		// More arguments than positional options generators, we have no more
		// completions to offer
		if len(generators) <= c.NArg() {
			return
		}

		// Generate the options for this posarg
		opts := generators[c.NArg()](c.Args())

		// If there are no options, render the previous options if we can, as
		// user may be typing the previous argument
		if len(opts) == 0 && c.NArg() >= 1 {
			opts = generators[c.NArg()-1](c.Args())
		}

		// Render the options
		for _, opt := range opts {
			fmt.Fprintf(c.App.Writer, "%s\n", opt)
		}
		// Also render default completions if there are no args yet
		if c.Args().Len() == 0 {
			defaultCompletions()(c)
		}
	}
}
