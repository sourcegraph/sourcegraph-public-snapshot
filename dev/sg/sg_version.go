package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	versionChangelogNext    bool
	versionChangelogEntries int

	versionCommand = &cli.Command{
		Name:     "version",
		Usage:    "View details for this installation of sg",
		Action:   execAdapter(versionExec),
		Category: CategoryUtil,
		Subcommands: []*cli.Command{
			{
				Name:    "changelog",
				Aliases: []string{"changes"},
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
						Destination: &versionChangelogEntries,
					},
				},
				Action: execAdapter(changelogExec),
			},
		},
	}
)

func versionExec(ctx context.Context, args []string) error {
	stdout.Out.Write(BuildCommit)
	return nil
}

func changelogExec(ctx context.Context, args []string) error {
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
		writeWarningLinef("Dev version detected - just showing recent changes.")
		title = "Recent sg changes"
	}

	gitLog := exec.Command("git", append(logArgs, "--", "./dev/sg")...)
	gitLog.Env = os.Environ()
	out, err := run.InRoot(gitLog)
	if err != nil {
		return err
	}

	block := stdout.Out.Block(output.Linef("", output.StyleSearchQuery, title))
	if len(out) == 0 {
		block.Write("No changes found.")
	} else {
		block.Write(out + "...")
	}
	block.Close()

	stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion,
		"Only showing %d entries - configure with 'sg version changelog -limit=50'", versionChangelogEntries))
	return nil
}

const sgOneLineCmd = `curl --proto '=https' --tlsv1.2 -sSLf https://install.sg.dev | sh`

func checkSgVersion(ctx context.Context) error {
	_, err := root.RepositoryRoot()
	if err != nil {
		// Ignore the error, because we only want to check the version if we're
		// in sourcegraph/sourcegraph
		return nil
	}

	if BuildCommit == "dev" {
		// If `sg` was built with a dirty `./dev/sg` directory it's a dev build
		// and we don't need to display this message.
		return nil
	}

	rev := strings.TrimPrefix(BuildCommit, "dev-")
	out, err := run.GitCmd("rev-list", fmt.Sprintf("%s..origin/main", rev), "./dev/sg")
	if err != nil {
		fmt.Printf("error getting new commits since %s in ./dev/sg: %s\n", rev, err)
		fmt.Printf("try reinstalling sg with `%s`.\n", sgOneLineCmd)
		os.Exit(1)
	}

	out = strings.TrimSpace(out)
	if out == "" {
		// No newer commits found. sg is up to date.
		return nil
	}

	if skipAutoUpdatesFlag {
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "╭──────────────────────────────────────────────────────────────────╮  "))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "│                                                                  │░░"))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "│ HEY! New version of sg available. Run 'sg update' to install it. │░░"))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "│       To see what's new, run 'sg version changelog -next'.       │░░"))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "│                                                                  │░░"))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "╰──────────────────────────────────────────────────────────────────╯░░"))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░"))
		return nil
	}

	stdout.Out.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "Auto updating sg ..."))
	newPath, err := updateToPrebuiltSG(ctx)
	if err != nil {
		return err
	}
	return syscall.Exec(newPath, os.Args, os.Environ())
}
