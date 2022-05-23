package linters

import (
	"context"
	"errors"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
)

var (
	goFmt          = lint.RunScript("Go format", "dev/check/gofmt.sh")
	goLint         = lint.RunScript("Go lint", "dev/check/go-lint.sh")
	goDBConnImport = lint.RunScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh")
)

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
