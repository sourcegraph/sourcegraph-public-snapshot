package linters

import (
	"context"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func lintGoDirectives() *linter {
	return runCheck("Lint Go directives", func(ctx context.Context, out *std.Output, state *repo.State) error {
		directivesRegexp := regexp.MustCompile("^// go:[a-z]+")

		diff, err := state.GetDiff("**/*.go")
		if err != nil {
			return err
		}

		return diff.IterateHunks(func(file string, hunk repo.DiffHunk) error {
			if directivesRegexp.MatchString(strings.Join(hunk.AddedLines, "\n")) {
				return errors.New("Go compiler directives must have no spaces between the // and 'go'")
			}
			return nil
		})
	})
}
