package linters

import (
	"context"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func lintGoDirectives(ctx context.Context, state *repo.State) *lint.Report {
	const header = "Lint Go directives"
	directivesRegexp := regexp.MustCompile("^// go:[a-z]+")

	diff, err := state.GetDiff("**/*.go")
	if err != nil {
		return &lint.Report{Header: header, Err: err}
	}

	errs := diff.IterateHunks(func(file string, hunk repo.DiffHunk) error {
		if directivesRegexp.MatchString(strings.Join(hunk.AddedLines, "\n")) {
			return errors.New("Go compiler directives must have no spaces between the // and 'go'")
		}
		return nil
	})

	return &lint.Report{
		Header: header,
		Err:    errs,
	}
}
