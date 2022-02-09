package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	doctorFlagSet = flag.NewFlagSet("sg doctor", flag.ExitOnError)
	doctorCommand = &ffcli.Command{
		Name:       "doctor",
		ShortUsage: "sg doctor",
		ShortHelp:  "Run the checks defined in the sg config file.",
		LongHelp: `Run the checks defined in the sg config file to make sure your system is healthy.

See the "checks:" in the configuration file.`,
		FlagSet: doctorFlagSet,
		Exec:    doctorExec,
	}
)

var checkFuncs = map[string]dependencyCheck{
	"postgres": anyChecks(checkSourcegraphDatabase, checkPostgresConnection),
	"redis":    retryCheck(checkRedisConnection, 5, 500*time.Millisecond),
	"psql":     checkInPath("psql"),
	// TODO: get these versions from .tool-versions
	"git":    combineChecks(checkInPath("git"), checkGitVersion(">= 2.34.1")),
	"yarn":   combineChecks(checkInPath("yarn"), checkYarnVersion("~> 1.22.4")),
	"go":     combineChecks(checkInPath("go"), checkGoVersion("1.17.5")),
	"docker": wrapCheckErr(checkInPath("docker"), "if Docker is installed and the check fails, you might need to start Docker.app and restart terminal and 'sg setup'"),
}

type builtinCheck struct {
	Name        string
	Func        dependencyCheck
	FailMessage string
}

func doctorExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		stdout.Out.WriteLine(errLine)
		os.Exit(1)
	}

	var funcs []builtinCheck
	var checks []run.Check
	for _, c := range globalConf.Checks {
		if c.Cmd != "" {
			checks = append(checks, run.Check{
				Name:        c.Name,
				Cmd:         c.Cmd,
				FailMessage: c.FailMessage,
			})
		} else if fn, ok := checkFuncs[c.CheckFunc]; ok {
			funcs = append(funcs, builtinCheck{
				Name:        c.CheckFunc,
				Func:        fn,
				FailMessage: c.FailMessage,
			})
		}
	}

	_, err := run.Checks(ctx, globalConf.Env, checks...)
	if err != nil {
		return err
	}

	// No funcs, early exit
	if len(funcs) == 0 {
		return nil
	}

	ctx, err = usershell.Context(ctx)
	if err != nil {
		return err
	}

	var failedchecks []string
	for _, check := range funcs {
		// TODO: Formatting here is duplicated from run.Checks
		p := stdout.Out.Pending(output.Linef(output.EmojiLightbulb, output.StylePending, "Running check %q...", check.Name))

		if err := check.Func(ctx); err != nil {
			p.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Check %q failed: %s", check.Name, err))

			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "%s", check.FailMessage))

			failedchecks = append(failedchecks, check.Name)
		} else {
			p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Check %q success!", check.Name))
		}
	}

	if len(failedchecks) != 0 {
		return errors.Newf("failed checks: %s", strings.Join(failedchecks, ", "))
	}
	return nil
}
