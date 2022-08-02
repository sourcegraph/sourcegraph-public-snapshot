package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// lintLoggingLibraries enforces that only usages of github.com/sourcegraph/log are added
func lintLoggingLibraries() *linter {
	var (
		bannedImports = []string{
			// No standard log library
			`"log"`,
			// No log15 - we only catch import changes for now, checking for 'log15.' is
			// too sensitive to just code moves.
			`"github.com/inconshreveable/log15"`,
			// No zap - we re-rexport everything via github.com/sourcegraph/log
			`"go.uber.org/zap"`,
			`"go.uber.org/zap/zapcore"`,
		}

		allowedFiles = []string{
			// Let everything in dev use whatever they want
			"dev", "enterprise/dev",
			// Banned imports will match on the linter here
			"dev/sg/linters/liblog.go",
			// We allow one usage of a direct zap import here
			"internal/observation/fields.go",
			// Dependencies require direct usage of zap
			"cmd/frontend/internal/app/otlpadapter",
		}
	)

	// checkHunk returns an error if a banned library is used
	checkHunk := func(file string, hunk repo.DiffHunk) error {
		for _, allowed := range allowedFiles {
			if strings.HasPrefix(file, allowed) {
				return nil
			}
		}

		for _, l := range hunk.AddedLines {
			for _, banned := range bannedImports {
				if strings.TrimSpace(l) == banned {
					return errors.Newf(`banned usage of '%s': use "github.com/sourcegraph/log" instead`,
						banned)
				}
			}
		}
		return nil
	}

	return runCheck("Logging library linter", func(ctx context.Context, out *std.Output, state *repo.State) error {
		diffs, err := state.GetDiff("**/*.go")
		if err != nil {
			return err
		}

		errs := diffs.IterateHunks(checkHunk)
		if errs != nil {
			out.Write("Learn more about logging and why some libraries are banned: https://docs.sourcegraph.com/dev/how-to/add_logging")
		}
		return errs
	})
}
