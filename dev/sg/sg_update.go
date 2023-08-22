package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/download"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var updateCommand = &cli.Command{
	Name:  "update",
	Usage: "Update local sg installation",
	Description: `Update local sg installation with the latest changes. To see what's new, run:

    sg version changelog -next`,
	Category: category.Util,
	Action: func(cmd *cli.Context) error {
		p := std.Out.Pending(output.Styled(output.StylePending, "Downloading latest sg release..."))
		if _, err := updateToPrebuiltSG(cmd.Context); err != nil {
			p.Destroy()
			return err
		}
		p.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "sg has been updated!"))

		std.Out.Write("To see what's new, run 'sg version changelog'.")
		return nil
	},
}

// updateToPrebuiltSG downloads the latest release of sg prebuilt binaries and install it.
func updateToPrebuiltSG(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://github.com/sourcegraph/sg/releases/latest", nil)
	if err != nil {
		return false, err
	}
	// We use the RoundTripper to make an HTTP request without having to deal
	// with redirections.
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return false, errors.Wrap(err, "GitHub latest release")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return false, errors.Newf("GitHub latest release: unexpected status code %d", resp.StatusCode)
	}

	location := resp.Header.Get("location")
	if location == "" {
		return false, errors.New("GitHub latest release: empty location")
	}
	location = strings.ReplaceAll(location, "/tag/", "/download/")
	downloadURL := fmt.Sprintf("%s/sg_%s_%s", location, runtime.GOOS, runtime.GOARCH)

	currentExecPath, err := os.Executable()
	if err != nil {
		return false, err
	}
	return download.Executable(ctx, downloadURL, currentExecPath, false)
}

func checkSgVersionAndUpdate(ctx context.Context, out *std.Output, skipUpdate bool) error {
	ctx, span := analytics.StartSpan(ctx, "auto_update", "background",
		trace.WithAttributes(attribute.Bool("skipUpdate", skipUpdate)))
	defer span.End()

	if BuildCommit == "dev" {
		// If `sg` was built with a dirty `./dev/sg` directory it's a dev build
		// and we don't need to display this message.
		out.Verbose("Skipping update check on dev build")
		span.Skipped()
		return nil
	}

	_, err := root.RepositoryRoot()
	if err != nil {
		// Ignore the error, because we only want to check the version if we're
		// in sourcegraph/sourcegraph
		span.Skipped()
		return nil
	}

	rev := strings.TrimPrefix(BuildCommit, "dev-")

	// If the revision of sg is not found locally, the user has likely not run 'git fetch'
	// recently, and we can skip the version check for now.
	if !repo.HasCommit(ctx, rev) {
		out.VerboseLine(output.Styledf(output.StyleWarning,
			"current sg version %s not found locally - you may want to run 'git fetch origin main'.", rev))
		span.Skipped()
		return nil
	}

	// Check for new commits since the current build of 'sg'
	revList, err := run.GitCmd("rev-list", fmt.Sprintf("%s..origin/main", rev), "--", "./dev/sg")
	if err != nil {
		// Unexpected error occured
		span.RecordError("check_error", err)
		return err
	}
	revList = strings.TrimSpace(revList)
	if revList == "" {
		// No newer commits found. sg is up to date.
		span.AddEvent("already_up_to_date")
		span.Skipped()
		return nil
	}
	span.SetAttributes(attribute.String("rev-list", revList))

	if skipUpdate {
		out.WriteLine(output.Styled(output.StyleSearchMatch, "╭──────────────────────────────────────────────────────────────────╮  "))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│                                                                  │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│ HEY! New version of sg available. Run 'sg update' to install it. │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│       To see what's new, run 'sg version changelog -next'.       │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│                                                                  │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "╰──────────────────────────────────────────────────────────────────╯░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░"))

		span.Skipped()
		return nil
	}

	out.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "Auto updating sg ..."))
	updated, err := updateToPrebuiltSG(ctx)
	if err != nil {
		span.RecordError("failed", err)
		return errors.Newf("failed to install update: %s", err)
	}
	if !updated {
		span.Skipped("not_updated")
		return nil
	}

	out.WriteSuccessf("sg has been updated!")
	out.Write("To see what's new, run 'sg version changelog'.")
	span.Succeeded()
	return nil
}
