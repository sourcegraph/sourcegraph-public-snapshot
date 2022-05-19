package main

import (
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
)

// addAnalyticsHooks wraps command actions with analytics hooks. We reconstruct commandPath
// ourselves because the library's state (and hence .FullName()) seems to get a bit funky.
func addAnalyticsHooks(start time.Time, commandPath []string, commands []*cli.Command) {
	for _, command := range commands {
		if len(command.Subcommands) > 0 {
			addAnalyticsHooks(start, append(commandPath, command.Name), command.Subcommands)
		}

		// No action to perform analytics on
		if command.Action == nil {
			continue
		}

		// Set up analytics hook for command
		analyticsHook := makeAnalyticsHook(start, append(commandPath, command.Name))

		// Wrap action with analytics
		wrappedAction := command.Action
		command.Action = func(cmd *cli.Context) error {
			// Make sure analytics hook gets called before exit
			interrupt.Register(func() { analyticsHook(cmd, "cancelled") })

			// Call the underlying action
			actionErr := wrappedAction(cmd)

			// Capture analytics post-run
			if actionErr != nil {
				analyticsHook(cmd, "error")
			} else {
				analyticsHook(cmd, "success")
			}

			return actionErr
		}
	}
}

func makeAnalyticsHook(start time.Time, commandPath []string) func(ctx *cli.Context, events ...string) {
	return func(cmd *cli.Context, events ...string) {
		// Log an sg usage occurrence
		analytics.LogEvent(cmd.Context, "sg_action", commandPath, start, events...)

		// Persist all tracked to disk
		flagsUsed := cmd.FlagNames()
		if err := analytics.Persist(cmd.Context, strings.Join(commandPath, " "), flagsUsed); err != nil {
			std.Out.WriteSkippedf("failed to persist events: %s", err)
		}
	}
}
