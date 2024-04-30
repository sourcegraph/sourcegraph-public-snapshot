package main

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var checks = map[string]check.CheckFunc{
	"postgres":      check.Any(check.SourcegraphDatabase(getConfig), check.PostgresConnection),
	"redis":         check.Redis,
	"caddy-trusted": check.Caddy,
	"asdf":          check.ASDF,
	"git":           check.Git,
	"pnpm":          check.PNPM,
	"go":            check.Go,
	"node":          check.Node,
	"rust":          check.Rust,
	"docker":        check.Docker,
	"ibazel":        check.WrapErrMessage(check.InPath("ibazel"), "brew install ibazel"),
	"bazelisk":      check.Bazelisk,
	"dev-private":   check.DevPrivate,
}

func runChecksWithName(ctx context.Context, names []string) error {
	funcs := make(map[string]check.CheckFunc, len(names))
	for _, name := range names {
		if c, ok := checks[name]; ok {
			funcs[name] = c
		} else {
			return errors.Newf("check %q not found", name)
		}
	}

	return runChecks(ctx, funcs)
}

func runChecks(ctx context.Context, checks map[string]check.CheckFunc) error {
	if len(checks) == 0 {
		return nil
	}

	std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleBold, "Running %d checks...", len(checks)))

	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Scripts used in various CheckFuncs are typically written with bash-compatible shells in mind.
	// Because of this, we throw a warning in non-compatible shells and ask that
	// users set up environments in both their shell and bash to avoid issues.
	if !usershell.IsSupportedShell(ctx) {
		shell := usershell.ShellType(ctx)
		std.Out.WriteWarningf("You're running on unsupported shell '%s'. "+
			"If you run into error, you may run 'SHELL=(which bash) sg setup' to setup your environment.",
			shell)
	}

	var failed []string

	for name, c := range checks {
		p := std.Out.Pending(output.Linef(output.EmojiLightbulb, output.StylePending, "Running check %q...", name))

		if err := c(ctx); err != nil {
			p.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Check %q failed with the following errors:", name))

			std.Out.WriteLine(output.Styledf(output.StyleWarning, "%s", err))

			failed = append(failed, name)
		} else {
			p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Check %q success!", name))
		}
	}

	if len(failed) == 0 {
		return nil
	}

	std.Out.Write("")
	std.Out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleBold, "The following checks failed:"))
	for _, name := range failed {
		std.Out.Writef("- %s", name)
	}

	std.Out.Write("")
	std.Out.WriteSuggestionf("Run 'sg setup' to make sure your system is setup correctly")
	std.Out.Write("")

	return errors.Newf("%d failed checks", len(failed))
}
