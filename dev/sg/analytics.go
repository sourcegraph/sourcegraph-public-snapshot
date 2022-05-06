package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
)

func addAnalyticsHooks(start time.Time, commandPath []string, commands []*cli.Command) {
	if len(commands) == 0 {
		return
	}
	for _, command := range commands {
		if len(command.Subcommands) > 0 {
			addAnalyticsHooks(start, append(commandPath, command.Name), command.Subcommands)
			continue
		}

		analyticsHook := makeAnalyticsHook(start, append(commandPath, command.Name))

		// This command has no subcommands, so we add analytics hook to indicate it has
		// been used.
		command.After = analyticsHook

		// Make sure analytics hook is called even on interrupts. Note that this only
		// works if you 'go build' sg, not if you 'go run'.
		var wrappedBeforeHook cli.BeforeFunc
		if command.Before == nil {
			wrappedBeforeHook = command.Before
		}
		command.Before = func(cmd *cli.Context) error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				analyticsHook(cmd)
				os.Exit(1)
			}()

			if wrappedBeforeHook != nil {
				return wrappedBeforeHook(cmd)
			}
			return nil
		}
	}
}

func makeAnalyticsHook(start time.Time, commandPath []string) func(cmd *cli.Context) error {
	return func(cmd *cli.Context) error {
		// Log an sg usage occurrence
		totalDuration := time.Since(start)
		analytics.LogDuration(cmd.Context, "sg_action", commandPath, totalDuration)

		// Persist all tracked to disk
		flagsUsed := cmd.FlagNames()
		if err := analytics.Persist(cmd.Context, strings.Join(commandPath, " "), flagsUsed); err != nil {
			writeSkippedLinef("failed to persist events: %s", err)
		}

		return nil
	}
}
