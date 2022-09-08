package linters

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestLibLogLinter(t *testing.T) {
	lint := lintLoggingLibraries()
	discard := std.NewFixedOutput(io.Discard, false)

	t.Run("no false positives", func(t *testing.T) {
		err := lint.Check(context.Background(), discard, repo.NewMockState(repo.Diff{
			"cmd/foobar/command.go": []repo.DiffHunk{
				{
					AddedLines: []string{
						`args: []string{"log", "--name-status"}`,
						`// do not use "github.com/inconshreveable/log15"`,
					},
				},
			},
		}))
		assert.Nil(t, err)
	})

	t.Run("catch imports", func(t *testing.T) {
		err := lint.Check(context.Background(), discard, repo.NewMockState(repo.Diff{
			"cmd/foobar/command.go": []repo.DiffHunk{
				{
					AddedLines: []string{
						`import (`,
						`	"github.com/inconshreveable/log15"`,
						`)`,
					},
				},
			},
		}))
		assert.NotNil(t, err)
	})

	t.Run("allowlist", func(t *testing.T) {
		err := lint.Check(context.Background(), discard, repo.NewMockState(repo.Diff{
			"dev/foobar.go": []repo.DiffHunk{
				{
					AddedLines: []string{
						`log15.Info("hi!")`,
					},
				},
			},
		}))
		assert.Nil(t, err)
	})
}
