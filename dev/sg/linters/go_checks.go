package linters

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

var (
	goFmt          = lint.RunScript("Go format", "dev/check/gofmt.sh")
	goDBConnImport = lint.RunScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh")
)

func goLint() lint.Runner {
	return func(ctx context.Context, _ *repo.State) *lint.Report {
		var dst bytes.Buffer
		err := root.Run(run.Bash(ctx, "dev/check/go-lint.sh")).
			Map(func(ctx context.Context, line []byte, dst io.Writer) (int, error) {
				// Ignore go mod download stuff
				if bytes.HasPrefix(line, []byte("go: downloading ")) {
					return 0, nil
				}
				return dst.Write(line)
			}).
			Stream(&dst)

		return &lint.Report{
			Header: "Go lint",
			Output: dst.String(),
			Err:    err,
		}
	}
}

func lintSGExit() lint.Runner {
	const header = "Lint dev/sg exit signals"

	return func(ctx context.Context, s *repo.State) *lint.Report {
		diff, err := s.GetDiff("dev/sg/***.go")
		if err != nil {
			return &lint.Report{Header: header, Err: err}
		}

		mErr := diff.IterateHunks(func(file string, hunk repo.DiffHunk) error {
			if strings.HasPrefix(file, "dev/sg/interrupt") || file == "dev/sg/linters/go_checks.go" {
				return nil
			}

			for _, added := range hunk.AddedLines {
				// Ignore comments
				if strings.HasPrefix(strings.TrimSpace(added), "//") {
					continue
				}

				if strings.Contains(added, "os.Exit") || strings.Contains(added, "signal.Notify") {
					return errors.New("do not use 'os.Exit' or 'signal.Notify', use the 'dev/sg/internal/interrupt' package instead")
				}
			}

			return nil
		})

		return &lint.Report{Header: header, Err: mErr}
	}
}
