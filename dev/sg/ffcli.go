package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

// execAdapter is a compatibility layer for ffcli-style exec functions. Do not use if you
// are creating a new command! If you are updating an existing command, consider a
// migration to cli.ActionFunc.
func execAdapter(exec func(context.Context, []string) error) cli.ActionFunc {
	return func(cmd *cli.Context) error {
		return exec(cmd.Context, cmd.Args().Slice())
	}
}
