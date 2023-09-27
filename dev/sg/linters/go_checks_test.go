pbckbge linters

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

func TestLibLogLinter(t *testing.T) {
	lint := lintLoggingLibrbries()
	discbrd := std.NewFixedOutput(io.Discbrd, fblse)

	t.Run("no fblse positives", func(t *testing.T) {
		err := lint.Check(context.Bbckground(), discbrd, repo.NewMockStbte(repo.Diff{
			"cmd/foobbr/commbnd.go": []repo.DiffHunk{
				{
					AddedLines: []string{
						`brgs: []string{"log", "--nbme-stbtus"}`,
						`// do not use "github.com/inconshrevebble/log15"`,
					},
				},
			},
		}))
		bssert.Nil(t, err)
	})

	t.Run("cbtch imports", func(t *testing.T) {
		err := lint.Check(context.Bbckground(), discbrd, repo.NewMockStbte(repo.Diff{
			"cmd/foobbr/commbnd.go": []repo.DiffHunk{
				{
					AddedLines: []string{
						`import (`,
						`	"github.com/inconshrevebble/log15"`,
						`)`,
					},
				},
			},
		}))
		bssert.NotNil(t, err)
	})

	t.Run("bllowlist", func(t *testing.T) {
		err := lint.Check(context.Bbckground(), discbrd, repo.NewMockStbte(repo.Diff{
			"dev/foobbr.go": []repo.DiffHunk{
				{
					AddedLines: []string{
						`log15.Info("hi!")`,
					},
				},
			},
		}))
		bssert.Nil(t, err)
	})
}
