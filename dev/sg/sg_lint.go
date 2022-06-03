package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/linters"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var lintGenerateAnnotations bool

var (
	lintFix = &cli.BoolFlag{
		Name:    "fix",
		Aliases: []string{"f"},
		Usage:   "Fix linters that can automatically be fixed",
	}
)

var helpFlagRegexp = regexp.MustCompile("-h|-help|--help")

var lintCommand = &cli.Command{
	Name:      "lint",
	ArgsUsage: "[targets...]",
	Usage:     "Run all or specified linters on the codebase",
	Description: `To run all checks, don't provide an argument. You can also provide multiple arguments to run linters for multiple targets.

Some targets have linters that can automatically be fixed, which can be enabled with 'sg lint -fix'.`,
	UsageText: `
# Run all possible checks
sg lint

# Run only go related checks
sg lint go

# Run only shell related checks
sg lint shell

# Run only client related checks
sg lint client

# List all available check groups
sg lint --help

# Automatically fix issues from linters that support it
sg lint -fix [targets...]
`,
	Category: CategoryDev,
	Flags: []cli.Flag{
		lintFix,
		&cli.BoolFlag{
			Name:        "annotations",
			Usage:       "Write helpful output to annotations directory",
			Destination: &lintGenerateAnnotations,
		},
	},
	Before: func(cmd *cli.Context) error {
		// If more than 1 target is requested and there's no help flag present, hijack
		// subcommands by setting it to nil so that the main lint command can handle it
		// the run.
		hasHelpFlag := helpFlagRegexp.MatchString(strings.Join(cmd.Args().Slice(), " "))
		if cmd.Args().Len() > 1 && !hasHelpFlag {
			cmd.App.Commands = nil
		}
		return nil
	},
	Action: func(cmd *cli.Context) error {
		var ls []lint.Linter
		targets := cmd.Args().Slice()

		if len(targets) == 0 {
			// If no args provided, run all
			for _, t := range linters.Targets {
				ls = append(ls, t.Linters...)
				targets = append(targets, t.Name)
			}
		} else {
			// Otherwise run requested set
			allLintTargetsMap := make(map[string][]lint.Linter, len(linters.Targets))
			for _, c := range linters.Targets {
				allLintTargetsMap[c.Name] = c.Linters
			}
			for _, t := range targets {
				targetLinters, ok := allLintTargetsMap[t]
				if !ok {
					std.Out.WriteFailuref("unrecognized target %q provided", t)
					return flag.ErrHelp
				}
				ls = append(ls, targetLinters...)
			}
		}

		fix := lintFix.Get(cmd)
		if fix {
			std.Out.WriteNoticef("Running checks and attempting to fix issues from targets: %s", strings.Join(targets, ", "))
		} else {
			std.Out.WriteNoticef("Running checks from targets: %s", strings.Join(targets, ", "))
		}
		return runCheckScriptsAndReport(cmd.Context, cmd.App.Writer, fix, ls...)
	},
	Subcommands: lintTargets(linters.Targets).Commands(),
}

type checkResult struct {
	fixable bool
	*lint.Report
}

// runCheckScriptsAndReport concurrently runs all fns and report as each check finishes. Returns an error
// if any of the fns fails.
func runCheckScriptsAndReport(ctx context.Context, dst io.Writer, fix bool, lns ...lint.Linter) error {
	_, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	// Get currently checked out ref and merge base so linters can optimize
	repoState, err := repo.GetState(ctx)
	if err != nil {
		return errors.Wrap(err, "repo.GetState")
	}

	// We need the Verbose flag to print above the pending indicator.
	out := std.NewOutput(dst, true)

	// Spawn a goroutine for each check and increment count to report completion.
	var count int64
	total := len(lns)
	pending := out.Pending(output.Styledf(output.StylePending, "Running linters (done: 0/%d)", total))
	var wg sync.WaitGroup
	resultsCh := make(chan *checkResult)
	wg.Add(total)

	// We use a single start time for the sake of simplicity.
	start := time.Now()

	// linterTimeout sets the very long time for a linter to run for. We definitely do not
	// want to allow linters to take any longer.
	linterTimeout := 5 * time.Minute
	runnerCtx, cancelRunners := context.WithTimeout(ctx, linterTimeout)
	for _, linter := range lns {
		go func(ln lint.Linter) {
			var res checkResult

			if fx, fixable := lint.Fixable(ln); fixable {
				var report *lint.Report
				if fix {
					report = fx.Fix(runnerCtx, repoState)
				} else {
					report = fx.Check(runnerCtx, repoState)
				}
				res = checkResult{fixable: true, Report: report}
			} else {
				res = checkResult{
					fixable: false,
					Report:  ln.Check(runnerCtx, repoState),
				}
			}

			resultsCh <- &res
			wg.Done()
		}(linter)
	}
	go func() {
		wg.Wait()
		close(resultsCh)
		cancelRunners()
	}()

	// consume check reports
	var hasErr bool
	var failedLinters []string
	var fixableErrors int
	var fixableSuccess int
	for result := range resultsCh {
		count++
		printLintReport(pending, start, result.Report)
		pending.Updatef("Running linters (done: %d/%d)", count, total)

		// Failed
		if result.Err != nil {
			failedLinters = append(failedLinters, result.Header)
			hasErr = true
			if result.fixable {
				fixableErrors++
			}
		}

		// Success!
		if result.fixable {
			fixableSuccess++
		}

		// Log analytics for each linter
		const eventName = "lint_runner"
		labels := []string{result.Header}
		if fix {
			labels = append(labels, "fix")
		}
		if runnerCtx.Err() == context.DeadlineExceeded {
			analytics.LogEvent(ctx, eventName, labels, start, "deadline exceeded")
		} else if result.Err != nil {
			analytics.LogEvent(ctx, eventName, labels, start, "failed")
		} else {
			analytics.LogEvent(ctx, eventName, labels, start, "succeeded")
		}
	}

	pending.Complete(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "Done running linters."))

	// Fix success!
	if fix {
		if fixableSuccess > 0 {
			out.WriteSuccessf("%d linters applied their fixes or had no changes to apply!", fixableSuccess)
		}
		if fixableErrors > 0 {
			out.WriteWarningf("%d linters failed to automatically fix issues.", fixableErrors)
		}
	} else if fixableErrors > 0 {
		// indicate some errors might be fixable!
		suggestion := "One or more of the failed linters can be fixed with `sg lint --fix`."
		out.WriteSuggestionf(suggestion)

		if lintGenerateAnnotations {
			writeAnnotation("Some linter issues are automatically fixable!", suggestion, true)
		}
	}

	// return the final error, if any
	if hasErr {
		return errors.Newf("failed linters: %s", strings.Join(failedLinters, ", "))
	}
	return nil
}

func printLintReport(pending output.Pending, start time.Time, report *lint.Report) {
	msg := fmt.Sprintf("%s (%ds)", report.Header, time.Since(start)/time.Second)
	if report.Err != nil {
		pending.VerboseLine(output.Linef(output.EmojiFailure, output.StyleWarning, msg))
		pending.Verbose(report.Summary())

		if lintGenerateAnnotations {
			writeAnnotation(report.Header, report.Summary(), false)
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
	for _, t := range lt {
		target := t // local reference

		var hasFixable bool
		for _, l := range target.Linters {
			if _, fixable := lint.Fixable(l); fixable {
				hasFixable = true
			}
		}

		var description string
		if hasFixable {
			description = fmt.Sprintf("\n\nThis target has linters that can be fixed with 'sg lint -fix %s'", target.Name)
		}

		cmds = append(cmds, &cli.Command{
			Name:        target.Name,
			Usage:       target.Help,
			Description: description,
			Action: func(cmd *cli.Context) error {
				if cmd.NArg() > 0 {
					std.Out.WriteFailuref("unrecognized argument %q provided", cmd.Args().First())
					return flag.ErrHelp
				}
				std.Out.WriteNoticef("Running checks from target: %s", target.Name)
				return runCheckScriptsAndReport(cmd.Context, cmd.App.Writer, lintFix.Get(cmd), target.Linters...)
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

func writeAnnotation(name string, content string, markdown bool) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return // do nothing
	}
	annotationPath := filepath.Join(repoRoot, "annotations")
	os.MkdirAll(annotationPath, os.ModePerm)
	file := filepath.Join(annotationPath, name)
	if markdown {
		file += ".md"
	}
	_ = os.WriteFile(file, []byte(content+"\n"), os.ModePerm)
}
