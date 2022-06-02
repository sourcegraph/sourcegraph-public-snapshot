package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// loggingLibraryLinter enforces that only usages of lib/log are added
type loggingLibraryLinter struct {
	bannedImports []string
	allowedFiles  []string
}

func newLoggingLibraryLinter() lint.Linter {
	return &loggingLibraryLinter{
		bannedImports: []string{
			// No standard log library
			`"log"`,
			// No log15 - we only catch import changes for now, checking for 'log15.' is
			// too sensitive to just code moves.
			`"github.com/inconshreveable/log15"`,
			// No zap - we re-rexport everything via lib/log
			`"go.uber.org/zap"`,
			`"go.uber.org/zap/zapcore"`,
		},
		allowedFiles: []string{
			// Banned imports will match on the linter here
			"dev/sg/linters/liblog.go",
			// We re-export things here
			"lib/log",
			// We allow one usage of a direct zap import here
			"internal/observation/fields.go",
		},
	}
}

func (l *loggingLibraryLinter) Check(ctx context.Context, state *repo.State) *lint.Report {
	const header = "Logging library linter"

	diffs, err := state.GetDiff("**/*.go")
	if err != nil {
		return &lint.Report{
			Header: header,
			Err:    err,
		}
	}

	errs := diffs.IterateHunks(l.checkHunk)

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

// checkHunk returns an error if a banned library is used
func (l *loggingLibraryLinter) checkHunk(file string, hunk repo.DiffHunk) error {
	for _, allowed := range l.allowedFiles {
		if strings.HasPrefix(file, allowed) {
			return nil
		}
	}

	for _, line := range hunk.AddedLines {
		for _, banned := range l.bannedImports {
			if strings.TrimSpace(line) == banned {
				return errors.Newf(`banned usage of '%s': use "github.com/sourcegraph/sourcegraph/lib/log" instead`,
					banned)
			}
		}
	}
	return nil
}
