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

	var errs error
	for file, hunks := range diff {
		for _, hunk := range hunks {
			if directivesRegexp.MatchString(strings.Join(hunk.AddedLines, "\n")) {
				errs = errors.Append(errors.Newf(`%s:%d: Go compiler directives must have no spaces between the // and 'go'`,
					file, hunk.StartLine))
			}
		}
	}

	return &lint.Report{
		Header: header,
		Err:    errs,
	}
}
