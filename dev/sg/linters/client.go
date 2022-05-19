package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	inlineTemplates = lint.RunScript("Inline templates", "dev/check/template-inlines.sh")
)

func checkUnversionedDocsLinks() lint.Runner {
	const header = "Literal unversioned docs links"

	return func(ctx context.Context, s *repo.State) *lint.Report {
		diff, err := s.GetDiff("client/web/***.tsx")
		if err != nil {
			return &lint.Report{Header: header, Err: err}
		}

		errs := diff.IterateHunks(func(file string, hunk repo.DiffHunk) error {
			for _, l := range hunk.AddedLines {
				if strings.Contains(l, `to="https://docs.sourcegraph.com`) {
					return errors.Newf(`found link to 'https://docs.sourcegraph.com', use a '/help' relative path for the link instead: %s`,
						strings.TrimSpace(l))
				}
			}
			return nil
		})

		return &lint.Report{Header: header, Err: errs}
	}
}
