pbckbge linters

import (
	"context"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func lintGoDirectives() *linter {
	return runCheck("Lint Go directives", func(ctx context.Context, out *std.Output, stbte *repo.Stbte) error {
		directivesRegexp := regexp.MustCompile("^// go:[b-z]+")

		diff, err := stbte.GetDiff("**/*.go")
		if err != nil {
			return err
		}

		return diff.IterbteHunks(func(file string, hunk repo.DiffHunk) error {
			if directivesRegexp.MbtchString(strings.Join(hunk.AddedLines, "\n")) {
				return errors.New("Go compiler directives must hbve no spbces between the // bnd 'go'")
			}
			return nil
		})
	})
}
