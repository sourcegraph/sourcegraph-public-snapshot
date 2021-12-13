package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
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

	scriptChecks = map[string][]checkScriptFn{
		"urls": {
			runCheckScript("Broken urls", "dev/check/broken-urls.bash"),
		},
		"go": {
			runCheckScript("Go format", "dev/check/gofmt.sh"),
			runCheckScript("Go generate", "dev/check/go-generate.sh"),
			runCheckScript("Go lint", "dev/check/go-lint.sh"),
			runCheckScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh"),
			runCheckScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh"),
		},
		"docsite": {
			runCheckScript("Docsite lint", "dev/check/docsite.sh"),
		},
		"docker": {
			runCheckScript("Docker forbidden alpine base images", "dev/check/no-alpine-guard.sh"),
		},
		"client": {
			runCheckScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh"),
			runCheckScript("Inline templates", "dev/check/template-inlines.sh"),
			runCheckScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			runCheckScript("SVG Compression", "dev/check/svgo.sh"),
		},
		"shell": {
			runCheckScript("Shell formatting", "dev/check/shfmt.sh"),
			runCheckScript("Shell lint", "dev/check/shellcheck.sh"),
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
			var fns []checkScriptFn
			for _, scriptFns := range scriptChecks {
				fns = append(fns, scriptFns...)
			}
			return runCheckScriptsAndReport(fns...)(ctx, args)
		},
		Subcommands: []*ffcli.Command{
			{
				Name:      "urls",
				ShortHelp: "Check for broken urls in the codebase.",
				FlagSet:   checkURLsFlagSet,
				Exec:      runCheckScriptsAndReport(scriptChecks["urls"]...),
			},
			{
				Name:      "go",
				ShortHelp: "Check go code for linting errors, forbidden imports, generated files...",
				FlagSet:   checkGoFlagSet,
				Exec:      runCheckScriptsAndReport(scriptChecks["go"]...),
			},
			{
				Name:      "docsite",
				FlagSet:   checkDocsiteFlagSet,
				ShortHelp: "Check the code powering docs.sourcegraph.com for broken links and linting errors.",
				Exec:      runCheckScriptsAndReport(scriptChecks["docsite"]...),
			},
			{
				Name:      "docker",
				FlagSet:   checkDockerFlagSet,
				ShortHelp: "Check for forbidden docker base images",
				Exec:      runCheckScriptsAndReport(scriptChecks["docker"]...),
			},
			{
				Name:      "client",
				FlagSet:   checkClientFlagSet,
				ShortHelp: "Check client code for linting errors, forbidden imports, ...",
				Exec:      runCheckScriptsAndReport(scriptChecks["client"]...),
			},
			{
				Name:      "shell",
				FlagSet:   checkShellFlagSet,
				ShortHelp: "Check shell code for linting errors, formatting, ...",
				Exec:      runCheckScriptsAndReport(scriptChecks["shell"]...),
			},
		},
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

		// swawn a go routine for each check and increment count to report completion
		var count int64
		total := len(fns)
		pending := out.Pending(output.Linef("", output.StylePending, "Running checks (done: 0/%d)", total))
		var wg sync.WaitGroup
		reportsCh := make(chan *checkReport)
		wg.Add(total)
		for _, fn := range fns {
			go func(fn func(context.Context) *checkReport) {
				report := fn(ctx)
				atomic.AddInt64(&count, 1)
				reportsCh <- report
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
			pending.Destroy()
			printCheckReport(report)
			if count != int64(total) {
				pending = out.Pending(output.Linef("", output.StylePending, "Running checks (done: %d/%d)", count, total))
			}
			if report.err != nil {
				messages = append(messages, report.err.Error())
				hasErr = true
			}
		}
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

func printCheckReport(report *checkReport) {
	timing := fmt.Sprintf(" (%ds)", report.duration/time.Second)
	if report.err != nil {
		writeFailureLine(report.header + timing)
		out.Write(report.output)
		return
	}
	writeSuccessLine(report.header + timing)
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
