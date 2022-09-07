package linters

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

var (
	goFmt          = runScript("Go format", "dev/check/gofmt.sh")
	goDBConnImport = runScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh")
)

func goLint() *linter {
	return runCheck("Go lint", func(ctx context.Context, out *std.Output, args *repo.State) error {
		return root.Run(run.Bash(ctx, "dev/check/go-lint.sh")).
			Map(func(ctx context.Context, line []byte, dst io.Writer) (int, error) {
				// Ignore go mod download stuff
				if bytes.HasPrefix(line, []byte("go: downloading ")) {
					return 0, nil
				}
				return dst.Write(line)
			}).
			StreamLines(out.Write)
	})
}

func lintSGExit() *linter {
	return runCheck("Lint dev/sg exit signals", func(ctx context.Context, out *std.Output, s *repo.State) error {
		diff, err := s.GetDiff("dev/sg/***.go")
		if err != nil {
			return err
		}

		return diff.IterateHunks(func(file string, hunk repo.DiffHunk) error {
			if strings.HasPrefix(file, "dev/sg/interrupt") ||
				strings.HasSuffix(file, "_test.go") ||
				file == "dev/sg/linters/go_checks.go" {
				return nil
			}

			for _, added := range hunk.AddedLines {
				// Ignore comments
				if strings.HasPrefix(strings.TrimSpace(added), "//") {
					continue
				}

				if strings.Contains(added, "os.Exit") ||
					strings.Contains(added, "signal.Notify") ||
					strings.Contains(added, "logger.Fatal") ||
					strings.Contains(added, "log.Fatal") {
					return errors.New("do not use 'os.Exit' or 'signal.Notify' or fatal logging, since they break 'dev/sg/internal/interrupt'")
				}
			}

			return nil
		})
	})
}
