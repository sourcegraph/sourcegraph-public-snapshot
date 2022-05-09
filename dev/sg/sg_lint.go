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

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var lintGenerateAnnotations bool

var lintCommand = &cli.Command{
	Name:        "lint",
	ArgsUsage:   "[target]",
	Usage:       "Run all or specified linter on the codebase",
	Description: `Run all or specified linter on the codebase and display failures, if any. To run all checks, don't provide an argument.`,
	Category:    CategoryDev,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "annotations",
			Usage:       "Write helpful output to annotations directory",
			Destination: &lintGenerateAnnotations,
		},
	},
	Action: func(cmd *cli.Context) error {
		if cmd.NArg() > 0 {
			writeFailureLinef("unrecognized command %q provided", cmd.Args().First())
			return flag.ErrHelp
		}
		var fns []lint.Runner
		for _, c := range allLintTargets {
			fns = append(fns, c.Linters...)
		}
		return runCheckScriptsAndReport(cmd.Context, fns...)
	},
	Subcommands: allLintTargets.Commands(),
}

// runCheckScriptsAndReport concurrently runs all fns and report as each check finishes. Returns an error
// if any of the fns fails.
func runCheckScriptsAndReport(ctx context.Context, fns ...lint.Runner) error {
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
		printLintReport(pending, report)
		pending.Updatef("Running linters (done: %d/%d)", count, total)
		if report.Err != nil {
			messages = append(messages, report.Header)
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

func printLintReport(pending output.Pending, report *lint.Report) {
	msg := fmt.Sprintf("%s (%ds)", report.Header, report.Duration/time.Second)
	if report.Err != nil {
		pending.VerboseLine(output.Linef(output.EmojiFailure, output.StyleWarning, msg))
		if report.Output != "" {
			pending.Verbose(report.Output)
		} else {
			pending.Verbose(report.Err.Error())
		}

		if lintGenerateAnnotations {
			repoRoot, err := root.RepositoryRoot()
			if err != nil {
				return // do nothing
			}
			annotationPath := filepath.Join(repoRoot, "annotations")
			os.MkdirAll(annotationPath, os.ModePerm)
			if err := os.WriteFile(filepath.Join(annotationPath, report.Header), []byte(report.Output+"\n"), os.ModePerm); err != nil {
				return // do nothing
			}
		}
		return
	}
	pending.VerboseLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, msg))
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
					writeFailureLinef("unrecognized argument %q provided", cmd.Args().First())
					return flag.ErrHelp
				}
				return runCheckScriptsAndReport(cmd.Context, c.Linters...)
			},
		})
	}
	return cmds
}
