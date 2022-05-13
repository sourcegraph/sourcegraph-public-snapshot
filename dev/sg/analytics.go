package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
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
			// Make sure analytics hook is called even on interrupts. Note that this only
			// works if you 'go build' sg, not if you 'go run'.
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-interrupt
				analyticsHook(cmd, "cancelled")
				os.Exit(1)
			}()

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
