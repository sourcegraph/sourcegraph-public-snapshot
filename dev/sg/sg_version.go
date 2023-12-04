package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
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
		Category: category.Util,
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
