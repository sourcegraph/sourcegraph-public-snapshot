package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	versionChangelogNext    bool
	versionChangelogEntries int

	versionCommand = &cli.Command{
		Name:     "version",
		Usage:    "View details for this installation of sg",
		Action:   versionExec,
		Category: CategoryUtil,
		Subcommands: []*cli.Command{
			{
				Name:    "changelog",
				Aliases: []string{"c"},
				Usage:   "See what's changed in or since this version of sg",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "next",
						Usage:       "Show changelog for changes you would get if you upgrade.",
						Destination: &versionChangelogNext,
					},
					&cli.IntFlag{
						Name:        "limit",
						Usage:       "Number of changelog entries to show.",
						Value:       5,
						Destination: &versionChangelogEntries,
					},
				},
				Action: changelogExec,
			},
		},
	}
)

func versionExec(ctx *cli.Context) error {
	std.Out.Write(BuildCommit)
	return nil
}

func changelogExec(ctx *cli.Context) error {
	if _, err := run.GitCmd("fetch", "origin", "main"); err != nil {
		return errors.Newf("failed to update main: %s", err)
	}

	logArgs := []string{
		// Format nicely
		"log", "--pretty=%C(reset)%s %C(dim)%h by %an, %ar",
		"--color=always",
		// Filter out stuff we don't want
		"--no-merges",
		// Limit entries
		fmt.Sprintf("--max-count=%d", versionChangelogEntries),
	}
	var title string
	if BuildCommit != "dev" {
		current := strings.TrimPrefix(BuildCommit, "dev-")
		if versionChangelogNext {
			logArgs = append(logArgs, current+"..origin/main")
			title = fmt.Sprintf("Changes since sg release %s", BuildCommit)
		} else {
			logArgs = append(logArgs, current)
			title = fmt.Sprintf("Changes in sg release %s", BuildCommit)
		}
	} else {
		std.Out.WriteWarningf("Dev version detected - just showing recent changes.")
		title = "Recent sg changes"
	}

	gitLog := exec.Command("git", append(logArgs, "--", "./dev/sg")...)
	gitLog.Env = os.Environ()
	out, err := run.InRoot(gitLog)
	if err != nil {
		return err
	}

	block := std.Out.Block(output.Styled(output.StyleSearchQuery, title))
	if len(out) == 0 {
		block.Write("No changes found.")
	} else {
		block.Write(out + "...")
	}
	block.Close()

	std.Out.WriteLine(output.Styledf(output.StyleSuggestion,
		"Only showing %d entries - configure with 'sg version changelog -limit=50'", versionChangelogEntries))
	return nil
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
	revOut, err := run.GitCmd("rev-list", fmt.Sprintf("%s..origin/main", rev), "--", "./dev/sg")
	if err != nil {
		if strings.Contains(revOut, "bad revision") {
			// installed revision is not available locally, that is fine - we wait for the
			// user to eventually do a fetch
			return errors.New("current sg version not found - you may want to run 'git fetch origin main'.")
		}

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

	std.Out.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "Auto updating sg ..."))
	if _, err := updateToPrebuiltSG(ctx); err != nil {
		analytics.LogEvent(ctx, "auto_update", []string{"failed"}, start)
		return errors.Newf("failed to install update: %s", err)
	}
	out.WriteSuccessf("sg has been updated!")
	out.Write("To see what's new, run 'sg version changelog'.")

	analytics.LogEvent(ctx, "auto_update", []string{"updated"}, start)

	return nil
}
