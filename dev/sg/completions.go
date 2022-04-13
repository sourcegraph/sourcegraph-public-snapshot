package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// completeOptions provides autocompletions based on the options returned by generateOptions
func completeOptions(generateOptions func() (options []string)) cli.BashCompleteFunc {
	return func(ctx *cli.Context) {
		for _, opt := range generateOptions() {
			fmt.Fprintf(ctx.App.Writer, "%s\n", opt)
		}
	}
}
