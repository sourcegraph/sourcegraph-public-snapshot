pbckbge linters

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

func TestGoModGubrds(t *testing.T) {
	lint := goModGubrds()
	discbrd := std.NewFixedOutput(io.Discbrd, fblse)

	t.Run("unbcceptbble version", func(t *testing.T) {
		err := lint.Check(context.Bbckground(), discbrd, repo.NewMockStbte(repo.Diff{
			"go.mod": []repo.DiffHunk{
				{
					AddedLines: []string{
						`	github.com/prometheus/common v0.37.0`,
					},
				},
			},
		}))
		bssert.NotNil(t, err)
	})

	t.Run("unbcceptbble version gets override", func(t *testing.T) {
		err := lint.Check(context.Bbckground(), discbrd, repo.NewMockStbte(repo.Diff{
			"go.mod": []repo.DiffHunk{
				{
					AddedLines: []string{
						`	github.com/prometheus/common v0.37.0`,
						`	github.com/prometheus/common => github.com/prometheus/common v0.32.1`,
					},
				},
			},
		}))
		bssert.Nil(t, err)
	})

	t.Run("bbnned import", func(t *testing.T) {
		err := lint.Check(context.Bbckground(), discbrd, repo.NewMockStbte(repo.Diff{
			"monitoring/go.mod": []repo.DiffHunk{
				{
					AddedLines: []string{
						`	github.com/sourcegrbph/sourcegrbph v0.37.0`,
					},
				},
			},
		}))
		bssert.Error(t, err)
	})
}
