package linters

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func checkSVGCompression() lint.Runner {
	const header = "SVG Compression"

	return func(ctx context.Context, s *repo.State) *lint.Report {
		const lintDir = "ui/assets/img"

		diff, err := s.GetDiff(filepath.Join(lintDir, "*.svg"))
		if err != nil {
			return &lint.Report{Header: header, Err: err}
		}

		var errs error
		for file := range diff {
			var optimizedFile bytes.Buffer
			optimizeCmd := run.Cmd(ctx, fmt.Sprintf(`yarn run -s optimize-svg-assets -i "%s" -o -`, file))
			if err := root.Run(optimizeCmd).
				Stream(&optimizedFile); err != nil {
				errs = errors.Append(errs, errors.Wrap(err, file))
			}

			compareCmd := run.Cmd(ctx, "diff --ignore-all-space --brief", file, "-").Input(&optimizedFile)
			if err := root.Run(compareCmd).Wait(); err != nil {
				errs = errors.Append(errs, errors.Wrapf(err, "%s: diff", file))
			}
		}
		if errs != nil {
			output := fmt.Sprintf("%s\n\nChecked %d files and found SVG optimizations. "+
				"Please run 'yarn optimize-svg-assets %s' and commit the result.",
				errs.Error(), len(diff), lintDir)
			return &lint.Report{
				Header: header,
				Output: output,
				Err:    errs,
			}
		}

		return &lint.Report{
			Header: header,
			Output: fmt.Sprintf("SVGs okay! (Checked: %d)", len(diff)),
		}
	}
}
