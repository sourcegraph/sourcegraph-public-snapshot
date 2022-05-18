package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// lintLoggingLibraries enforces that only usages of lib/log are added
func lintLoggingLibraries() lint.Runner {
	const header = "Logging library linter"

	var (
		bannedImports = []string{
			// No standard log library
			`"log"`,
			// No log15 - we only catch import changes for now, checking for 'log15.' is
			// too sensitive to just code moves.
			`"github.com/inconshreveable/log15"`,
			// No zap - we re-rexport everything via lib/log
			`"go.uber.org/zap"`,
			`"go.uber.org/zap/zapcore"`,
		}

		allowedFiles = []string{
			// Banned imports will match on the linter here
			"dev/sg/linters/liblog.go",
			// We re-export things here
			"lib/log",
			// We allow one usage of a direct zap import here
			"internal/observation/fields.go",
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
				if strings.Contains(l, banned) {
					return errors.Newf(`banned usage of '%s': use "github.com/sourcegraph/sourcegraph/lib/log" instead`,
						banned)
				}
			}
		}
		return nil
	}

	return func(ctx context.Context, state *repo.State) *lint.Report {
		diffs, err := state.GetDiff("**/*.go")
		if err != nil {
			return &lint.Report{
				Header: header,
				Err:    err,
			}
		}

		errs := diffs.IterateHunks(checkHunk)

		return &lint.Report{
			Header: header,
			Output: func() string {
				if errs != nil {
					return strings.TrimSpace(errs.Error()) +
						"\n\nLearn more about logging and why some libraries are banned: https://docs.sourcegraph.com/dev/how-to/add_logging"
				}
				return ""
			}(),
			Err: errs,
		}
	}
}
