package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
)

// addAnalyticsHooks wraps command actions with analytics hooks. We reconstruct commandPath
// ourselves because the library's state (and hence .FullName()) seems to get a bit funky.
//
// It also handles watching for panics and formatting them in a useful manner.
func addAnalyticsHooks(commandPath []string, commands []*cli.Command) {
	for _, command := range commands {
		fullCommandPath := append(commandPath, command.Name)
		if len(command.Subcommands) > 0 {
			addAnalyticsHooks(fullCommandPath, command.Subcommands)
		}

		// No action to perform analytics on
		if command.Action == nil {
			continue
		}

		// Set up analytics hook for command
		fullCommand := strings.Join(fullCommandPath, " ")

		// Wrap action with analytics
		wrappedAction := command.Action
		command.Action = func(cmd *cli.Context) (actionErr error) {
			var span *analytics.Span
			cmd.Context, span = analytics.StartSpan(cmd.Context, fullCommand, "action",
				trace.WithAttributes(
					attribute.StringSlice("flags", cmd.FlagNames()),
					attribute.Int("args", cmd.NArg()),
				))
			defer span.End()

			// Make sure analytics are persisted before exit (interrupts or panics)
			defer func() {
				if p := recover(); p != nil {
					// Render a more elegant message
					std.Out.WriteWarningf("Encountered panic - please open an issue with the command output:\n\t%s",
						sgBugReportTemplate)
					message := fmt.Sprintf("%v:\n%s", p, getRelevantStack("addAnalyticsHooks"))
					actionErr = cli.Exit(message, 1)

					// Log event
					span.RecordError("panic", actionErr)
				}
			}()
			interrupt.Register(func() {
				span.Cancelled()
				span.End()
			})

			// Call the underlying action
			actionErr = wrappedAction(cmd)

			// Capture analytics post-run
			if actionErr != nil {
				span.RecordError("error", actionErr)
			} else {
				span.Succeeded()
			}

			return actionErr
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
