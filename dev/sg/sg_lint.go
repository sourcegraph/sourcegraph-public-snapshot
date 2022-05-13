package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/linters"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var lintGenerateAnnotations bool

var lintCommand = &cli.Command{
	Name:        "lint",
	ArgsUsage:   "[targets...]",
	Usage:       "Run all or specified linters on the codebase",
	Description: `Run all or specified linters on the codebase and display failures, if any. To run all checks, don't provide an argument.`,
	Category:    CategoryDev,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "annotations",
			Usage:       "Write helpful output to annotations directory",
			Destination: &lintGenerateAnnotations,
		},
	},
	Before: func(cmd *cli.Context) error {
		// If more than 1 target is requested, hijack subcommands by setting it to nil
		// so that the main lint command can handle it the run.
		if cmd.Args().Len() > 1 {
			cmd.App.Commands = nil
		}
		return nil
	},
	Action: func(cmd *cli.Context) error {
		var fns []lint.Runner
		targets := cmd.Args().Slice()

		if len(targets) == 0 {
			// If no args provided, run all
			for _, c := range linters.Targets {
				fns = append(fns, c.Linters...)
				targets = append(targets, c.Name)
			}
		} else {
			// Otherwise run requested set
			allLintTargetsMap := make(map[string][]lint.Runner, len(linters.Targets))
			for _, c := range linters.Targets {
				allLintTargetsMap[c.Name] = c.Linters
			}
			for _, t := range targets {
				runners, ok := allLintTargetsMap[t]
				if !ok {
					std.Out.WriteFailuref("unrecognized target %q provided", t)
					return flag.ErrHelp
				}
				fns = append(fns, runners...)
			}
		}

		std.Out.WriteNoticef("Running checks from targets: %s", strings.Join(targets, ", "))
		return runCheckScriptsAndReport(cmd.Context, cmd.App.Writer, fns...)
	},
	Subcommands: lintTargets(linters.Targets).Commands(),
}

// runCheckScriptsAndReport concurrently runs all fns and report as each check finishes. Returns an error
// if any of the fns fails.
func runCheckScriptsAndReport(ctx context.Context, dst io.Writer, fns ...lint.Runner) error {
	_, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	// Get currently checked out branch so linters can optimize
	branch, err := run.TrimResult(run.GitCmd("branch", "--show-current"))
	if err != nil {
		return err
	}
	repoState := &repo.State{Branch: branch}

	// We need the Verbose flag to print above the pending indicator.
	out := output.NewOutput(dst, output.OutputOpts{
		ForceColor: true,
		ForceTTY:   true,
		Verbose:    true,
	})

	// Spawn a goroutine for each check and increment count to report completion. We use
	// a single start time for the sake of simplicity.
	start := time.Now()
	var count int64
	total := len(fns)
	pending := out.Pending(output.Styledf(output.StylePending, "Running linters (done: 0/%d)", total))
	var wg sync.WaitGroup
	reportsCh := make(chan *lint.Report)
	wg.Add(total)
	for _, fn := range fns {
		go func(fn lint.Runner) {
			reportsCh <- fn(ctx, repoState)
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
		printLintReport(pending, start, report)
		pending.Updatef("Running linters (done: %d/%d)", count, total)
		if report.Err != nil {
			messages = append(messages, report.Header)
			hasErr = true
		}
	}

	pending.Complete(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "Done running linters."))

	// return the final error, if any
	if hasErr {
		return errors.Newf("failed linters: %s", strings.Join(messages, ", "))
	}
	return nil
}

func printLintReport(pending output.Pending, start time.Time, report *lint.Report) {
	msg := fmt.Sprintf("%s (%ds)", report.Header, time.Since(start)/time.Second)
	if report.Err != nil {
		pending.VerboseLine(output.Linef(output.EmojiFailure, output.StyleWarning, msg))
		pending.Verbose(report.Summary())

		if lintGenerateAnnotations {
			repoRoot, err := root.RepositoryRoot()
			if err != nil {
				return // do nothing
			}
			annotationPath := filepath.Join(repoRoot, "annotations")
			os.MkdirAll(annotationPath, os.ModePerm)
			if err := os.WriteFile(filepath.Join(annotationPath, report.Header), []byte(report.Summary()+"\n"), os.ModePerm); err != nil {
				return // do nothing
			}
		}
		return
	}

	pending.VerboseLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, msg))
	if verbose {
		pending.Verbose(report.Summary())
	}
}

type lintTargets []lint.Target

// Commands converts all lint targets to CLI commands
func (lt lintTargets) Commands() (cmds []*cli.Command) {
	for _, c := range lt {
		c := c // local reference
		cmds = append(cmds, &cli.Command{
			Name:  c.Name,
			Usage: c.Help,
			Action: func(cmd *cli.Context) error {
				if cmd.NArg() > 0 {
					std.Out.WriteFailuref("unrecognized argument %q provided", cmd.Args().First())
					return flag.ErrHelp
				}
				std.Out.WriteNoticef("Running checks from target: %s", c.Name)
				return runCheckScriptsAndReport(cmd.Context, cmd.App.Writer, c.Linters...)
			},
			// Completions to chain multiple commands
			BashComplete: completeOptions(func() (options []string) {
				for _, c := range lt {
					options = append(options, c.Name)
				}
				return options
			}),
		})
	}
	return cmds
}
