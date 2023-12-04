package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	inlineTemplates = runScript("Inline templates", "dev/check/template-inlines.sh")
)

func checkUnversionedDocsLinks() *linter {
	return runCheck("Literal unversioned docs links", func(ctx context.Context, out *std.Output, state *repo.State) error {
		diff, err := state.GetDiff("client/web/***.tsx")
		if err != nil {
			return err
		}

		return diff.IterateHunks(func(file string, hunk repo.DiffHunk) error {
			// Ignore Cody app directory since docs links don't work
			// with /help route there
			if strings.HasPrefix(file, "client/web/src/enterprise/app") {
				return nil
			}
			for _, l := range hunk.AddedLines {
				if strings.Contains(l, `to="https://docs.sourcegraph.com`) {
					return errors.Newf(`found link to 'https://docs.sourcegraph.com', use a '/help' relative path for the link instead: %s`,
						strings.TrimSpace(l))
				}
			}
			return nil
		})
	})
}
