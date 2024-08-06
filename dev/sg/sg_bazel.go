package main

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type tipFn func() string

type bzlgenTarget struct {
	order  int
	cmd    string
	args   []string
	env    []string
	protip tipFn
}

func (t bzlgenTarget) showProTip(out *std.Output) {
	tip := ""
	if t.protip != nil {
		tip = t.protip()
	}

	if tip != "" {
		out.WriteLine(output.Emojif(output.EmojiLightbulb, "pro-tip: %s", tip))
	}
}

var bzlgenTargets = map[string]bzlgenTarget{
	"builds": {
		order: 1,
		cmd:   "run",
		args:  []string{"//:configure"},
	},
	"godeps": {
		cmd:  "run",
		args: []string{"//:gazelle-update-repos"},
	},
	"rustdeps": {
		cmd:  "sync",
		args: []string{"--only=crate_index"},
		env:  []string{"CARGO_BAZEL_REPIN=1"},
		protip: func() string {
			if os.Getenv("CARGO_BAZEL_ISOLATED") != "0" {
				return "run with CARGO_BAZEL_ISOLATED=0 for faster (but less sandboxed) repinning."
			}
			return ""
		},
	},
}

var bazelCommand = &cli.Command{
	Name:            "bazel",
	Aliases:         []string{"bz"},
	SkipFlagParsing: true,
	HideHelpCommand: true,
	Usage:           "Proxies the bazel CLI with custom commands for local dev convenience",
	Category:        category.Dev,
	Action: func(cctx *cli.Context) error {
		if slices.Equal(cctx.Args().Slice(), []string{"help"}) || slices.Equal(cctx.Args().Slice(), []string{"--help"}) || slices.Equal(cctx.Args().Slice(), []string{"-h"}) {
			fmt.Println("Additional commands from sg:")
			fmt.Println("  configure           Wrappers around some commands to generate various files required by Bazel")
			fmt.Println("Additional flags from sg:")
			fmt.Println("  --disable-remote-cache           Disable use of the remote cache for local env.")
		}

		// Walk the args, looking for our custom flag to disable the remote cache.
		// If we find it, we take not of it, but do not append it to the final args
		// that will be passed to the bazel command. Everything else is passed as-is.
		var disableRemoteCache bool
		args := make([]string, 0, len(cctx.Args().Slice()))
		for _, arg := range cctx.Args().Slice() {
			switch arg {
			case "--disable-remote-cache":
				disableRemoteCache = true
			case "--disable-remote-cache=true":
				disableRemoteCache = true
			case "--disable-remote-cache=false":
				disableRemoteCache = false
			default:
				args = append(args, arg)
			}
		}

		// If we end up running `sg bazel` in CI, we don't want to use the remote cache for local environment,
		// so we force disable the flag explicilty.
		if os.Getenv("CI") == "true" || os.Getenv("BUILDKITE") == "true" {
			disableRemoteCache = true
		}

		if !disableRemoteCache {
			rootDir, err := root.RepositoryRoot()
			if err != nil {
				return errors.Wrap(err, "getting repository root")
			}
			newArgs := make([]string, 0, len(args)+1)
			// Bazelrc flags must be added before the actual command (build, run, test ...)
			newArgs = append(newArgs, fmt.Sprintf("--bazelrc=%s", filepath.Join(rootDir, ".aspect/bazelrc/local.bazelrc")))
			newArgs = append(newArgs, args...)
			args = newArgs
		}

		cmd := exec.CommandContext(cctx.Context, "bazel", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	},
	Subcommands: []*cli.Command{
		{
			Name:  "configure",
			Usage: "Wrappers around some commands to generate various files required by Bazel",
			UsageText: `sg bazel configure [category...]

Available categories:
	- builds: updates BUILD.bazel files for Go & Typescript targets.
	- godeps: updates the bazel Go dependency targets based on go.mod changes.
	- rustdeps: updates the cargo bazel lockfile.
	- all: catch-all for all of the above

If no categories are referenced, then 'builds' is assumed as the default.`,
			BashComplete: completions.CompleteArgs(func() (options []string) {
				return append(maps.Keys(bzlgenTargets), "all")
			}),
			Before: func(ctx *cli.Context) error {
				for _, arg := range ctx.Args().Slice() {
					if _, ok := bzlgenTargets[arg]; !ok && arg != "all" {
						cli.HandleExitCoder(errors.Errorf("category doesn't exist %q, run `sg bazel configure --help` for full info.", arg))
						cli.ShowSubcommandHelpAndExit(ctx, 1)
						return nil
					}
				}
				return nil
			},
			Action: func(ctx *cli.Context) error {
				var categories []bzlgenTarget
				var categoryNames []string
				if slices.Contains(ctx.Args().Slice(), "all") {
					categories = maps.Values(bzlgenTargets)
					categoryNames = maps.Keys(bzlgenTargets)
				} else if ctx.NArg() == 0 {
					categories = []bzlgenTarget{bzlgenTargets["builds"]}
					categoryNames = []string{"builds"}
				} else {
					for i := range ctx.NArg() {
						categories = append(categories, bzlgenTargets[ctx.Args().Get(i)])
						categoryNames = append(categoryNames, ctx.Args().Get(i))
					}
				}

				slices.SortFunc(categories, func(a, b bzlgenTarget) int {
					return cmp.Compare(a.order, b.order)
				})

				std.Out.WriteLine(output.Emojif(output.EmojiAsterisk, "Invoking the following Bazel generating categories: %s", strings.Join(categoryNames, ", ")))

				for _, c := range categories {
					std.Out.WriteNoticef("running command %q", strings.Join(append([]string{"bazel", c.cmd}, c.args...), " "))
					c.showProTip(std.Out)

					args := append([]string{c.cmd, "--noshow_progress"}, c.args...)
					cmd := exec.CommandContext(ctx.Context, "bazel", args...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Env = c.env
					cmd.Env = append(cmd.Env, os.Environ()...)

					err := cmd.Run()
					var exitErr *exec.ExitError
					if errors.As(err, &exitErr) && exitErr.ExitCode() == 110 {
						return nil
					} else if err != nil {
						return err
					}
				}
				return nil
			},
		},
	},
}
