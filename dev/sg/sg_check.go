package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	checkFlagSet = flag.NewFlagSet("sg check", flag.ExitOnError)

	checkShellFlagSet   = flag.NewFlagSet("sg check shell", flag.ExitOnError)
	checkURLsFlagSet    = flag.NewFlagSet("sg check urls", flag.ExitOnError)
	checkGoFlagSet      = flag.NewFlagSet("sg check go", flag.ExitOnError)
	checkDocsiteFlagSet = flag.NewFlagSet("sg check docsite", flag.ExitOnError)
	checkDockerFlagSet  = flag.NewFlagSet("sg check docker", flag.ExitOnError)
	checkClientFlagSet  = flag.NewFlagSet("sg check client", flag.ExitOnError)

	allCheckTargets = checkTargets{
		{
			Name:      "urls",
			ShortHelp: "Check for broken urls in the codebase.",
			FlagSet:   checkURLsFlagSet,
			Checks: []checkScriptFn{
				runCheckScript("Broken urls", "dev/check/broken-urls.bash"),
			},
		},
		{
			Name:      "go",
			ShortHelp: "Check go code for linting errors, forbidden imports, generated files...",
			FlagSet:   checkGoFlagSet,
			Checks: []checkScriptFn{
				runCheckScript("Go format", "dev/check/gofmt.sh"),
				runCheckScript("Go generate", "dev/check/go-generate.sh"),
				runCheckScript("Go lint", "dev/check/go-lint.sh"),
				runCheckScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh"),
				runCheckScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh"),
			},
		},
		{
			Name:      "docsite",
			ShortHelp: "Check the code powering docs.sourcegraph.com for broken links and linting errors.",
			FlagSet:   checkDocsiteFlagSet,
			Checks: []checkScriptFn{
				runCheckScript("Docsite lint", "dev/check/docsite.sh"),
			},
		},
		{
			Name:      "docker",
			ShortHelp: "Check Dockerfiles for Sourcegraph best practices",
			FlagSet:   checkDockerFlagSet,
			Checks: []checkScriptFn{
				runCheckScript("Docker lint", "dev/check/docker-lint.sh"),
				runCheckScript("Docker forbidden alpine base images", "dev/check/no-alpine-guard.sh"),
				checkDockerfiles,
			},
		},
		{
			Name:      "client",
			ShortHelp: "Check client code for linting errors, forbidden imports, ...",
			FlagSet:   checkClientFlagSet,
			Checks: []checkScriptFn{
				runCheckScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh"),
				runCheckScript("Inline templates", "dev/check/template-inlines.sh"),
				runCheckScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
				runCheckScript("SVG Compression", "dev/check/svgo.sh"),
			},
		},
		{
			Name:      "shell",
			ShortHelp: "Check shell code for linting errors, formatting, ...",
			FlagSet:   checkShellFlagSet,
			Checks: []checkScriptFn{
				runCheckScript("Shell formatting", "dev/check/shfmt.sh"),
				runCheckScript("Shell lint", "dev/check/shellcheck.sh"),
			},
		},
	}
)

var (
	checkCommand = &ffcli.Command{
		Name:       "check",
		ShortUsage: "sg check [go|client|shell|docsite|docker|urls]",
		ShortHelp:  "Run all checks on the codebase.",
		LongHelp: `Run all checks on the codebase and display failures, if any.
Run sg check --help for a list of all checks.
`,
		FlagSet: checkFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				return errors.New("unrecognized command, please run sg check --help to list available checks")
			}
			var fns []checkScriptFn
			for _, c := range allCheckTargets {
				fns = append(fns, c.Checks...)
			}
			return runCheckScriptsAndReport(fns...)(ctx, args)
		},
		Subcommands: allCheckTargets.Commands(),
	}
)

type checkScriptFn func(context.Context) *checkReport

// runCheckScriptsAndReport concurrently runs all fns and report as each check finishes. Returns an error
// if any of the fns fails.
func runCheckScriptsAndReport(fns ...checkScriptFn) func(context.Context, []string) error {
	return func(ctx context.Context, _ []string) error {
		_, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		// We need the Verbose flag to print above the pending indicator.
		out := output.NewOutput(os.Stdout, output.OutputOpts{
			ForceColor: true,
			ForceTTY:   true,
			Verbose:    true,
		})

		// Spawn a goroutine for each check and increment count to report completion.
		var count int64
		total := len(fns)
		pending := out.Pending(output.Linef("", output.StylePending, "Running checks (done: 0/%d)", total))
		var wg sync.WaitGroup
		reportsCh := make(chan *checkReport)
		wg.Add(total)
		for _, fn := range fns {
			go func(fn func(context.Context) *checkReport) {
				reportsCh <- fn(ctx)
				wg.Done()
			}(fn)
		}
		go func() {
			wg.Wait()
			close(reportsCh)
		}()

		// consume check reports
		var hasErr bool
		var messages []string
		for report := range reportsCh {
			count++
			printCheckReport(pending, report)
			pending.Updatef("Running checks (done: %d/%d)", count, total)
			if report.err != nil {
				messages = append(messages, report.header)
				hasErr = true
			}
		}
		pending.Complete(output.Linef("", output.StyleBold, "Done running checks."))

		// return the final error, if any
		if hasErr {
			return errors.Newf("failed checks: %s", strings.Join(messages, ", "))
		}
		return nil
	}
}

type checkReport struct {
	duration time.Duration
	header   string
	output   string
	err      error
}

func printCheckReport(pending output.Pending, report *checkReport) {
	msg := fmt.Sprintf("%s (%ds)", report.header, report.duration/time.Second)
	if report.err != nil {
		pending.VerboseLine(output.Linef(output.EmojiFailure, output.StyleWarning, msg))
		pending.Verbose(report.output)
		return
	}
	pending.VerboseLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, msg))
}

func runCheckScript(header string, script string) checkScriptFn {
	return checkScriptFn(func(ctx context.Context) *checkReport {
		start := time.Now()
		out, err := run.BashInRoot(ctx, script, nil)
		return &checkReport{
			header:   header,
			duration: time.Since(start),
			output:   out,
			err:      err,
		}
	})
}

// checkTarget denotes a check that can be run by `sg check`
type checkTarget struct {
	Name      string
	ShortHelp string
	FlagSet   *flag.FlagSet
	Checks    []checkScriptFn
}

type checkTargets []checkTarget

// Commands converts all check targets to CLI commands
func (cs checkTargets) Commands() (cmds []*ffcli.Command) {
	for _, c := range cs {
		cmds = append(cmds, &ffcli.Command{
			Name:      c.Name,
			ShortHelp: c.ShortHelp,
			FlagSet:   checkURLsFlagSet,
			Exec:      runCheckScriptsAndReport(c.Checks...),
		})
	}
	return cmds
}

func checkDockerfiles(ctx context.Context) *checkReport {
	start := time.Now()
	var combinedErrors error
	for _, dir := range []string{"docker-images", "cmd", "enterprise/cmd"} {
		if err := filepath.Walk(dir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !strings.Contains(filepath.Base(path), "Dockerfile") {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				if err := docker.ProcessDockerfile(data, func(is []instructions.Stage) error {
					var errs error
					for _, i := range is {
						for _, c := range i.Commands {
							if err := docker.CheckCommand(c); err != nil {
								errs = errors.Append(errs, errors.Wrapf(err, "%s:%d", path, c.Location()[0].Start.Line))
							}
						}
					}
					return errs
				}); err != nil {
					// track error but don't exit
					combinedErrors = errors.Append(combinedErrors, err)
				}

				return nil
			},
		); err != nil {
			combinedErrors = errors.Append(combinedErrors, err)
		}
	}
	return &checkReport{
		duration: time.Since(start),
		header:   "Custom Dockerfile checks",
		output: func() string {
			if combinedErrors != nil {
				return strings.TrimSpace(combinedErrors.Error())
			}
			return ""
		}(),
		err: combinedErrors,
	}
}
