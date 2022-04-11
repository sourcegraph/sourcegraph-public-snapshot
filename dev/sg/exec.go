package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

func execAdapter(exec func(context.Context, []string) error) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return exec(ctx.Context, ctx.Args().Tail())
	}
}
