package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type bzlgenTarget struct {
	order  int
	cmd    string
	args   []string
	env    []string
	protip string
}

var bzlgenTargets = map[string]bzlgenTarget{
	"builds": {
		order: 1,
		cmd:   "run",
		args:  []string{"//:gazelle"},
	},
	"godeps": {
		cmd:  "run",
		args: []string{"//:gazelle-update-repos"},
	},
	"rustdeps": {
		cmd:    "sync",
		args:   []string{"--only=crate_index"},
		env:    []string{"CARGO_BAZEL_REPIN=1"},
		protip: "run with CARGO_BAZEL_ISOLATED=0 for faster (but less sandboxed) repinning.",
	},
}

var bzlgenCommand = &cli.Command{
	Name:      "bzlgen",
	Usage:     "generates bazel build files that can be generated automatically",
	UsageText: "sg bzlgen [group...]",
	Category:  category.Dev,
	Before: func(ctx *cli.Context) error {
		for i := 0; i < ctx.NArg(); i++ {
			if _, ok := bzlgenTargets[ctx.Args().Get(i)]; !ok && ctx.Args().Get(i) != "all" {
				return errors.Errorf("group doesn't exist %q", ctx.Args().Get(i))
			}
		}
		return nil
	},
	Action: func(ctx *cli.Context) error {
		var groups []bzlgenTarget
		var groupNames []string
		if slices.Contains(ctx.Args().Slice(), "all") {
			groups = maps.Values(bzlgenTargets)
			groupNames = maps.Keys(bzlgenTargets)
		} else if ctx.NArg() == 0 {
			groups = []bzlgenTarget{bzlgenTargets["builds"]}
			groupNames = []string{"builds"}
		} else {
			for i := 0; i < ctx.NArg(); i++ {
				groups = append(groups, bzlgenTargets[ctx.Args().Get(i)])
				groupNames = append(groupNames, ctx.Args().Get(i))
			}
		}

		slices.SortFunc(groups, func(a, b bzlgenTarget) bool {
			return a.order < b.order
		})

		std.Out.WriteLine(output.Emojif(output.EmojiAsterisk, "Generating the following Bazel groups: %s", strings.Join(groupNames, ", ")))

		for _, c := range groups {
			root, err := root.RepositoryRoot()
			if err != nil {
				return err
			}

			std.Out.WriteNoticef("running command %q", strings.Join(append([]string{"bazel", c.cmd}, c.args...), " "))
			if c.protip != "" {
				std.Out.WriteLine(output.Emojif(output.EmojiLightbulb, "pro-tip: %s", c.protip))
			}

			args := append([]string{c.cmd, "--noshow_progress"}, c.args...)
			cmd := exec.CommandContext(ctx.Context, "bazel", args...)
			cmd.Dir = root
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = c.env
			cmd.Env = append(cmd.Env, os.Environ()...)

			err = cmd.Run()
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 110 {
				return nil
			} else if err != nil {
				return err
			}
		}
		return nil
	},
}
