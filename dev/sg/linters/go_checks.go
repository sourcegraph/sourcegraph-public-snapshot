package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var goDBConnImport = runScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh")

func lintSGExit() *linter {
	return runCheck("Lint dev/sg exit signals", func(ctx context.Context, out *std.Output, s *repo.State) error {
		diff, err := s.GetDiff("dev/sg/***.go")
		if err != nil {
			return err
		}

		return diff.IterateHunks(func(file string, hunk repo.DiffHunk) error {
			if strings.HasPrefix(file, "dev/sg/interrupt") ||
				strings.HasSuffix(file, "_test.go") ||
				file == "dev/sg/linters/go_checks.go" {
				return nil
			}

			for _, added := range hunk.AddedLines {
				// Ignore comments
				if strings.HasPrefix(strings.TrimSpace(added), "//") {
					continue
				}

				if strings.Contains(added, "os.Exit") ||
					strings.Contains(added, "signal.Notify") ||
					strings.Contains(added, "logger.Fatal") ||
					strings.Contains(added, "log.Fatal") {
					return errors.New("do not use 'os.Exit' or 'signal.Notify' or fatal logging, since they break 'dev/sg/internal/interrupt'")
				}
			}

			return nil
		})
	})
}

// lintLoggingLibraries enforces that only usages of github.com/sourcegraph/log are added
func lintLoggingLibraries() *linter {
	return newUsageLinter("Logging libraries linter", usageLinterOptions{
		Target: "**/*.go",
		BannedUsages: []string{
			// No standard log library
			`"log"`,
			// No log15 - we only catch import changes for now, checking for 'log15.' is
			// too sensitive to just code moves.
			`"github.com/inconshreveable/log15"`,
			// No zap - we re-rexport everything via github.com/sourcegraph/log
			`"go.uber.org/zap"`,
			`"go.uber.org/zap/zapcore"`,
		},
		AllowedFiles: []string{
			// Let everything in dev use whatever they want
			"dev",
			// Banned imports will match on the linter here
			"dev/sg/linters",
			// We allow one usage of a direct zap import here
			"internal/observation/fields.go",
			// Inits old loggers
			"internal/logging/main.go",
			// Dependencies require direct usage of zap
			"cmd/frontend/internal/app/otlpadapter",
			// Legacy and special case handling of panics in background routines
			"lib/background/goroutine.go",
			// Need to make a logger shim for the OpenFGA server
			"lib/managedservicesplatform/iam/openfga_server.go",
		},
		ErrorFunc: func(bannedImport string) error {
			return errors.Newf(`banned usage of '%s': use "github.com/sourcegraph/log" instead`,
				bannedImport)
		},
		HelpText: "Learn more about logging and why some libraries are banned: https://docs-legacy.sourcegraph.com/dev/how-to/add_logging",
	})
}

// keep up to date with dev/linters/tracinglibraries/tracinglibraries.go
func lintTracingLibraries() *linter {
	return newUsageLinter("Tracing libraries linter", usageLinterOptions{
		Target: "**/*.go",
		BannedUsages: []string{
			// No OpenTracing
			`"github.com/opentracing/opentracing-go"`,
			// No OpenTracing util library
			`"github.com/sourcegraph/sourcegraph/internal/trace/ot"`,
		},
		AllowedFiles: []string{
			// Banned imports will match on the linter here
			"dev/sg/linters",
			// Adapters here
			"internal/tracer",
		},
		ErrorFunc: func(bannedImport string) error {
			return errors.Newf(`banned usage of '%s': use "go.opentelemetry.io/otel/trace" instead`,
				bannedImport)
		},
		HelpText: "OpenTracing interop with OpenTelemetry is set up, but the libraries are deprecated - use OpenTelemetry directly instead: https://go.opentelemetry.io/otel/trace",
	})
}
