package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/download"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var updateCommand = &cli.Command{
	Name:  "update",
	Usage: "Update local sg installation",
	Description: `Update local sg installation with the latest changes. To see what's new, run:

    sg version changelog -next`,
	Category: CategoryUtil,
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
func updateToPrebuiltSG(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://github.com/sourcegraph/sg/releases/latest", nil)
	if err != nil {
		return "", err
	}
	// We use the RountTripper to make an HTTP request without having to deal
	// with redirections.
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("location")
	if location == "" {
		return "", errors.New("GitHub latest release: empty location")
	}
	location = strings.ReplaceAll(location, "/tag/", "/download/")
	downloadURL := fmt.Sprintf("%s/sg_%s_%s", location, runtime.GOOS, runtime.GOARCH)

	currentExecPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if err := download.Executable(ctx, downloadURL, currentExecPath); err != nil {
		return "", err
	}
	return currentExecPath, nil
}

func checkSgVersionAndUpdate(ctx context.Context, out *std.Output, skipUpdate bool) error {
	start := time.Now()

	if BuildCommit == "dev" {
		// If `sg` was built with a dirty `./dev/sg` directory it's a dev build
		// and we don't need to display this message.
		out.Verbose("Skipping update check on dev build")
		return nil
	}

	_, err := root.RepositoryRoot()
	if err != nil {
		// Ignore the error, because we only want to check the version if we're
		// in sourcegraph/sourcegraph
		return nil
	}

	rev := strings.TrimPrefix(BuildCommit, "dev-")

	// If the revision of sg is not found locally, the user has likely not run 'git fetch'
	// recently, and we can skip the version check for now.
	if !repo.HasCommit(ctx, rev) {
		out.VerboseLine(output.Styledf(output.StyleWarning,
			"current sg version %s not found locally - you may want to run 'git fetch origin main'.", rev))
		return nil
	}

	// Check for new commits since the current build of 'sg'
	revOut, err := run.GitCmd("rev-list", fmt.Sprintf("%s..origin/main", rev), "--", "./dev/sg")
	if err != nil {
		// Unexpected error occured
		analytics.LogEvent(ctx, "auto_update", []string{"check_error"}, start)
		return err
	}
	revOut = strings.TrimSpace(revOut)
	if revOut == "" {
		// No newer commits found. sg is up to date.
		return nil
	}

	if skipUpdate {
		out.WriteLine(output.Styled(output.StyleSearchMatch, "╭──────────────────────────────────────────────────────────────────╮  "))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│                                                                  │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│ HEY! New version of sg available. Run 'sg update' to install it. │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│       To see what's new, run 'sg version changelog -next'.       │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "│                                                                  │░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "╰──────────────────────────────────────────────────────────────────╯░░"))
		out.WriteLine(output.Styled(output.StyleSearchMatch, "  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░"))

		analytics.LogEvent(ctx, "auto_update", []string{"skipped"}, start)
		return nil
	}

	out.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "Auto updating sg ..."))
	if _, err := updateToPrebuiltSG(ctx); err != nil {
		analytics.LogEvent(ctx, "auto_update", []string{"failed"}, start)
		return errors.Newf("failed to install update: %s", err)
	}
	out.WriteSuccessf("sg has been updated!")
	out.Write("To see what's new, run 'sg version changelog'.")

	analytics.LogEvent(ctx, "auto_update", []string{"updated"}, start)

	return nil
}
