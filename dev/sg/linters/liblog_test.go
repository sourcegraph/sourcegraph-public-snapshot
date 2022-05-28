package linters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
)

func TestLibLogLinter(t *testing.T) {
	lint := lintLoggingLibraries()

	t.Run("no false positives", func(t *testing.T) {
		report := lint(context.Background(), repo.NewMockState(repo.Diff{
			"cmd/foobar/command.go": []repo.DiffHunk{
				{
					AddedLines: []string{
						`args: []string{"log", "--name-status"}`,
						`// do not use "github.com/inconshreveable/log15"`,
					},
				},
			},
		}))
		assert.Nil(t, report.Err)
	})

	t.Run("catch imports", func(t *testing.T) {
		report := lint(context.Background(), repo.NewMockState(repo.Diff{
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
		assert.NotNil(t, report.Err)
	})
}
