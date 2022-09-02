package linters

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func checkSVGCompression() *linter {
	const header = "SVG Compression"

	return runCheck(header, func(ctx context.Context, out *std.Output, state *repo.State) error {
		const lintDir = "ui/assets/img"

		diff, err := state.GetDiff(filepath.Join(lintDir, "*.svg"))
		if err != nil {
			return err
		}

		var errs error
		for file := range diff {
			var optimizedFile bytes.Buffer
			optimizeCmd := run.Cmd(ctx, fmt.Sprintf(`yarn run optimize-svg-assets -i "%s" -o -`, file))
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
			out.Writef("Checked %d files and found SVG optimizations. "+
				"Please run 'yarn optimize-svg-assets %s' and commit the result.",
				len(diff), lintDir)
			return errs
		}

		out.Verbosef("SVGs okay! (Checked: %d)", len(diff))
		return nil
	})
}
