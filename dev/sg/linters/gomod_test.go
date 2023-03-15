package linters

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestGoModGuards(t *testing.T) {
	lint := goModGuards()
	discard := std.NewFixedOutput(io.Discard, false)

	t.Run("unacceptable version", func(t *testing.T) {
		err := lint.Check(context.Background(), discard, repo.NewMockState(repo.Diff{
			"go.mod": []repo.DiffHunk{
				{
					AddedLines: []string{
						`	github.com/prometheus/common v0.37.0`,
					},
				},
			},
		}))
		assert.NotNil(t, err)
	})

	t.Run("unacceptable version gets override", func(t *testing.T) {
		err := lint.Check(context.Background(), discard, repo.NewMockState(repo.Diff{
			"go.mod": []repo.DiffHunk{
				{
					AddedLines: []string{
						`	github.com/prometheus/common v0.37.0`,
						`	github.com/prometheus/common => github.com/prometheus/common v0.32.1`,
					},
				},
			},
		}))
		assert.Nil(t, err)
	})

	t.Run("banned import", func(t *testing.T) {
		err := lint.Check(context.Background(), discard, repo.NewMockState(repo.Diff{
			"monitoring/go.mod": []repo.DiffHunk{
				{
					AddedLines: []string{
						`	github.com/sourcegraph/sourcegraph v0.37.0`,
					},
				},
			},
		}))
		assert.Error(t, err)
	})
}
