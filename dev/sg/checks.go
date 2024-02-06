package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var checks = map[string]check.CheckFunc{
	"postgres":      check.Any(checkSourcegraphDatabase, check.CheckPostgresConnection),
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
	"bazelisk":      check.WrapErrMessage(check.InPath("bazelisk"), "brew install bazelisk"),
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

func checkSourcegraphDatabase(ctx context.Context) error {
	// This check runs only in the `sourcegraph/sourcegraph` repository, so
	// we try to parse the globalConf and use its `Env` to configure the
	// Postgres connection.
	config, _ := getConfig()
	if config == nil {
		return errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
	}

	getEnv := func(key string) string {
		// First look into process env, emulating the logic in makeEnv used
		// in internal/run/run.go
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
		// Otherwise check in globalConf.Env
		return config.Env[key]
	}

	dsn := postgresdsn.New("", "", getEnv)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to Sourcegraph Postgres database at %s. Please check the settings in sg.config.yml (see https://docs.sourcegraph.com/dev/background-information/sg#changing-database-configuration)", dsn)
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx)
}
