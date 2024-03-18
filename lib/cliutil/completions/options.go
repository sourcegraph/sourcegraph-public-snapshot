package completions

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// defaultCompletions renders the default completions for a command.
func defaultCompletions() cli.ShellCompleteFunc {
	return func(ctx context.Context, cmd *cli.Command) { cli.DefaultCompleteWithFlags(cmd)(ctx, cmd) }
}

// CompleteArgs provides autocompletions based on the options returned by
// generateOptions. generateOptions is provided for every argument.
//
// generateOptions must not write to output, or reference any
// resources that are initialized elsewhere.
func CompleteArgs(generateOptions func() (options []string)) cli.ShellCompleteFunc {
	return func(ctx context.Context, cmd *cli.Command) {
		CompletePositionalArgs(func(_ cli.Args) (options []string) {
			return generateOptions()
		})(ctx, cmd)
	}
}

// CompletePositionalArgs provides autocompletions for multiple positional
// arguments based on the options returned by the corresponding options generator.
// If there are more arguments than generators, no completion is provided.
//
// Each generator must not write to output, or reference any resources that are
// initialized elsewhere.
func CompletePositionalArgs(generators ...func(args cli.Args) (options []string)) cli.ShellCompleteFunc {
	return func(ctx context.Context, cmd *cli.Command) {
		// Let default handler deal with flag completions if the latest argument
		// looks like the start of a flag
		if cmd.NArg() > 0 {
			if f := cmd.Args().Get(cmd.NArg() - 1); f == "-" || f == "--" {
				defaultCompletions()(ctx, cmd)
				return
			}
		}

		// More arguments than positional options generators, we have no more
		// completions to offer
		if len(generators) <= cmd.NArg() {
			return
		}

		// Generate the options for this posarg
		opts := generators[cmd.NArg()](cmd.Args())

		// If there are no options, render the previous options if we can, as
		// user may be typing the previous argument
		if len(opts) == 0 && cmd.NArg() >= 1 {
			opts = generators[cmd.NArg()-1](cmd.Args())
		}

		// Render the options
		for _, opt := range opts {
			fmt.Fprintf(cmd.Writer, "%s\n", opt)
		}
		// Also render default completions if there are no args yet
		if cmd.Args().Len() == 0 {
			defaultCompletions()(ctx, cmd)
		}
	}
}
