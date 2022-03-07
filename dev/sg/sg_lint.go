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

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	lintFlagSet             = flag.NewFlagSet("sg lint", flag.ExitOnError)
	lintGenerateAnnotations = lintFlagSet.Bool("annotations", false, "Write helpful output to annotations directory")

	lintShellFlagSet   = flag.NewFlagSet("sg lint shell", flag.ExitOnError)
	lintURLsFlagSet    = flag.NewFlagSet("sg lint urls", flag.ExitOnError)
	lintGoFlagSet      = flag.NewFlagSet("sg lint go", flag.ExitOnError)
	lintDocsiteFlagSet = flag.NewFlagSet("sg lint docsite", flag.ExitOnError)
	lintDockerFlagSet  = flag.NewFlagSet("sg lint docker", flag.ExitOnError)
	lintClientFlagSet  = flag.NewFlagSet("sg lint client", flag.ExitOnError)
)

var allLintTargets = lintTargets{
	{
		Name:      "urls",
		ShortHelp: "Check for broken urls in the codebase.",
		FlagSet:   lintURLsFlagSet,
		Linters: []lintFunc{
			runLintScript("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name:      "go",
		ShortHelp: "Check go code for linting errors, forbidden imports, generated files...",
		FlagSet:   lintGoFlagSet,
		Linters: []lintFunc{
			runLintScript("Go format", "dev/check/gofmt.sh"),
			runLintScript("Go generate", "dev/check/go-generate.sh"),
			runLintScript("Go lint", "dev/check/go-lint.sh"),
			runLintScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh"),
			runLintScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh"),
		},
	},
	{
		Name:      "docsite",
		ShortHelp: "Check the code powering docs.sourcegraph.com for broken links and linting errors.",
		FlagSet:   lintDocsiteFlagSet,
		Linters: []lintFunc{
			runLintScript("Docsite lint", "dev/check/docsite.sh"),
		},
	},
	{
		Name:      "docker",
		ShortHelp: "Check Dockerfiles for Sourcegraph best practices",
		FlagSet:   lintDockerFlagSet,
		Linters: []lintFunc{
			runLintScript("Docker lint", "dev/check/docker-lint.sh"),
			lintDockerfiles,
		},
	},
	{
		Name:      "client",
		ShortHelp: "Check client code for linting errors, forbidden imports, ...",
		FlagSet:   lintClientFlagSet,
		Linters: []lintFunc{
			runLintScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh"),
			runLintScript("Inline templates", "dev/check/template-inlines.sh"),
			runLintScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			runLintScript("SVG Compression", "dev/check/svgo.sh"),
		},
	},
	{
		Name:      "shell",
		ShortHelp: "Check shell code for linting errors, formatting, ...",
		FlagSet:   lintShellFlagSet,
		Linters: []lintFunc{
			runLintScript("Shell formatting", "dev/check/shfmt.sh"),
			runLintScript("Shell lint", "dev/check/shellcheck.sh"),
		},
	},
}

var lintCommand = &ffcli.Command{
	Name:       "lint",
	ShortUsage: "sg lint [target]",
	ShortHelp:  "Run all or specified linter on the codebase.",
	LongHelp:   `Run all or specified linter on the codebase and display failures, if any. To run all checks, don't provide an argument.`,
	FlagSet:    lintFlagSet,
	Exec: func(ctx context.Context, args []string) error {
		if len(args) > 0 {
			return errors.New("unrecognized command, please run 'sg lint --help' to list available linters")
		}
		var fns []lintFunc
		for _, c := range allLintTargets {
			fns = append(fns, c.Linters...)
		}
		return runCheckScriptsAndReport(fns...)(ctx, args)
	},
	Subcommands: allLintTargets.Commands(),
}

type lintFunc func(context.Context) *lintReport

// runCheckScriptsAndReport concurrently runs all fns and report as each check finishes. Returns an error
// if any of the fns fails.
func runCheckScriptsAndReport(fns ...lintFunc) func(context.Context, []string) error {
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
		pending := out.Pending(output.Linef("", output.StylePending, "Running linters (done: 0/%d)", total))
		var wg sync.WaitGroup
		reportsCh := make(chan *lintReport)
		wg.Add(total)
		for _, fn := range fns {
			go func(fn func(context.Context) *lintReport) {
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
			printLintReport(pending, report)
			pending.Updatef("Running linters (done: %d/%d)", count, total)
			if report.err != nil {
				messages = append(messages, report.header)
				hasErr = true
			}
		}
		pending.Complete(output.Linef("", output.StyleBold, "Done running linters."))

		// return the final error, if any
		if hasErr {
			return errors.Newf("failed linters: %s", strings.Join(messages, ", "))
		}
		return nil
	}
}

type lintReport struct {
	duration time.Duration
	header   string
	output   string
	err      error
}

func printLintReport(pending output.Pending, report *lintReport) {
	msg := fmt.Sprintf("%s (%ds)", report.header, report.duration/time.Second)
	if report.err != nil {
		pending.VerboseLine(output.Linef(output.EmojiFailure, output.StyleWarning, msg))
		pending.Verbose(report.output)
		if *lintGenerateAnnotations {
			repoRoot, err := root.RepositoryRoot()
			if err != nil {
				return // do nothing
			}
			annotationPath := filepath.Join(repoRoot, "annotations")
			os.MkdirAll(annotationPath, os.ModePerm)
			if err := os.WriteFile(filepath.Join(annotationPath, report.header), []byte(report.output+"\n"), os.ModePerm); err != nil {
				return // do nothing
			}
		}
		return
	}
	pending.VerboseLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, msg))
}

func runLintScript(header string, script string) lintFunc {
	return lintFunc(func(ctx context.Context) *lintReport {
		start := time.Now()
		out, err := run.BashInRoot(ctx, script, nil)
		return &lintReport{
			header:   header,
			duration: time.Since(start),
			output:   out,
			err:      err,
		}
	})
}

// lintTarget denotes a linter task that can be run by `sg lint`
type lintTarget struct {
	Name      string
	ShortHelp string
	FlagSet   *flag.FlagSet
	Linters   []lintFunc
}

type lintTargets []lintTarget

// Commands converts all lint targets to CLI commands
func (lt lintTargets) Commands() (cmds []*ffcli.Command) {
	for _, c := range lt {
		cmds = append(cmds, &ffcli.Command{
			Name:      c.Name,
			ShortHelp: c.ShortHelp,
			FlagSet:   lintURLsFlagSet,
			Exec:      runCheckScriptsAndReport(c.Linters...),
		})
	}
	return cmds
}

func lintDockerfiles(ctx context.Context) *lintReport {
	start := time.Now()
	var combinedErrors error
	for _, dir := range []string{
		"docker-images",
		// cmd dirs
		"cmd",
		"enterprise/cmd",
		"internal/cmd",
		// dev dirs
		"dev",
		"enterprise/dev",
	} {
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

				if err := docker.ProcessDockerfile(data, docker.CheckDockerfile(path)); err != nil {
					// track error but don't exit
					combinedErrors = errors.Append(combinedErrors, err)
				}

				return nil
			},
		); err != nil {
			combinedErrors = errors.Append(combinedErrors, err)
		}
	}
	return &lintReport{
		duration: time.Since(start),
		header:   "Sourcegraph Dockerfile linters",
		output: func() string {
			if combinedErrors != nil {
				return strings.TrimSpace(combinedErrors.Error())
			}
			return ""
		}(),
		err: combinedErrors,
	}
}
