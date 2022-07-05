package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
)

// addAnalyticsHooks wraps command actions with analytics hooks. We reconstruct commandPath
// ourselves because the library's state (and hence .FullName()) seems to get a bit funky.
//
// It also handles watching for panics and formatting them in a useful manner.
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
		command.Action = func(cmd *cli.Context) (actionErr error) {
			// Make sure analytics hook gets called before exit (interrupts or panics)
			interrupt.Register(func() { analyticsHook(cmd, nil, "cancelled") })
			defer func() {
				if p := recover(); p != nil {
					// Render a more elegant message
					std.Out.WriteWarningf("Encountered panic - please open an issue with the command output:\n\t%s",
						sgBugReportTemplate)
					message := fmt.Sprintf("%v:\n%s", p, getRelevantStack("addAnalyticsHooks"))
					actionErr = cli.NewExitError(message, 1)

					// Log event
					analyticsHook(cmd, actionErr, "panic")
				}
			}()

			// Call the underlying action
			actionErr = wrappedAction(cmd)

			// Capture analytics post-run
			if actionErr != nil {
				analyticsHook(cmd, actionErr, "error")
			} else {
				analyticsHook(cmd, actionErr, "success")
			}

			return actionErr
		}
	}
}

func makeAnalyticsHook(start time.Time, commandPath []string) func(ctx *cli.Context, err error, events ...string) {
	return func(cmd *cli.Context, err error, events ...string) {
		// Log an sg usage occurrence
		event := analytics.LogEvent(cmd.Context, "sg_action", commandPath, start, events...)
		if err != nil {
			event.Properties["error_details"] = err.Error()
		}

		// Persist all tracked to disk
		flagsUsed := cmd.FlagNames()
		if err := analytics.Persist(cmd.Context, strings.Join(commandPath, " "), flagsUsed); err != nil {
			std.Out.WriteSkippedf("failed to persist events: %s", err)
		}
	}
}

// getRelevantStack generates a stacktrace that encapsulates the relevant parts of a
// stacktrace for user-friendly reading.
func getRelevantStack(excludeFunctions ...string) string {
	callers := make([]uintptr, 32)
	n := runtime.Callers(3, callers) // recover -> getRelevantStack -> runtime.Callers
	frames := runtime.CallersFrames(callers[:n])

	var stack strings.Builder
	for {
		frame, next := frames.Next()

		var excludedFunction bool
		for _, e := range excludeFunctions {
			if strings.Contains(frame.Function, e) {
				excludedFunction = true
				break
			}
		}

		// Only include frames from sg and things that are not excluded.
		if !strings.Contains(frame.File, "dev/sg/") || excludedFunction {
			if !next {
				break
			}
			continue
		}

		stack.WriteString(frame.Function)
		stack.WriteByte('\n')
		stack.WriteString(fmt.Sprintf("\t%s:%d\n", frame.File, frame.Line))
		if !next {
			break
		}
	}

	return stack.String()
}
