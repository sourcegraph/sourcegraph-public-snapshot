package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

type usageLinterOptions struct {
	// Target is a glob provided to find relevant diffs to check.
	Target string
	// BannedUsages is a list of disallowed strings.
	//
	// For a linter that disallows new imports, for example, you should provide fully
	// quoted imports paths for packages that are no longer allowed, i.e.:
	//
	//   []string{`"log"`, `"github.com/inconshreveable/log15"`}
	//
	// The created linter will check added hunks for these substrings.
	BannedUsages []string
	// AllowedFiles are filepaths where banned usages are allowed. Supports files and
	// directories.
	AllowedFiles []string
	// ErrorFunc is used to create an error when a banned usage is found.
	ErrorFunc func(bannedImport string) error
	// HelpText is shown when errors are found.
	HelpText string
}

// newUsageLinter is a helper that creates a linter that guards against *additions* that
// introduce usages banned strings.
func newUsageLinter(opts usageLinterOptions) *linter {
	// checkHunk returns an error if a banned library is used
	checkHunk := func(file string, hunk repo.DiffHunk) error {
		for _, allowed := range opts.AllowedFiles {
			if strings.HasPrefix(file, allowed) {
				return nil
			}
		}

		for _, l := range hunk.AddedLines {
			for _, banned := range opts.BannedUsages {
				if strings.TrimSpace(l) == banned {
					return opts.ErrorFunc(banned)
				}
			}
		}
		return nil
	}

	return runCheck("Logging library linter", func(ctx context.Context, out *std.Output, state *repo.State) error {
		diffs, err := state.GetDiff(opts.Target)
		if err != nil {
			return err
		}

		errs := diffs.IterateHunks(checkHunk)
		if errs != nil && opts.HelpText != "" {
			out.Write(opts.HelpText)
		}
		return errs
	})
}
