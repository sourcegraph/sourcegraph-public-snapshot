package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns"
	"github.com/sourcegraph/src-cli/internal/output"
)

var (
	campaignsPendingColor = output.StylePending
	campaignsSuccessColor = output.StyleSuccess
	campaignsSuccessEmoji = output.EmojiSuccess
)

type campaignsApplyFlags struct {
	allowUnsupported bool
	api              *api.Flags
	apply            bool
	cacheDir         string
	tempDir          string
	clearCache       bool
	file             string
	keepLogs         bool
	namespace        string
	parallelism      int
	timeout          time.Duration
	workspace        string
	cleanArchives    bool
	skipErrors       bool
}

func newCampaignsApplyFlags(flagSet *flag.FlagSet, cacheDir, tempDir string) *campaignsApplyFlags {
	caf := &campaignsApplyFlags{
		api: api.NewFlags(flagSet),
	}

	flagSet.BoolVar(
		&caf.allowUnsupported, "allow-unsupported", false,
		"Allow unsupported code hosts.",
	)
	flagSet.BoolVar(
		&caf.apply, "apply", false,
		"Ignored.",
	)
	flagSet.StringVar(
		&caf.cacheDir, "cache", cacheDir,
		"Directory for caching results and repository archives.",
	)
	flagSet.BoolVar(
		&caf.clearCache, "clear-cache", false,
		"If true, clears the execution cache and executes all steps anew.",
	)
	flagSet.StringVar(
		&caf.tempDir, "tmp", tempDir,
		"Directory for storing temporary data, such as log files. Default is /tmp. Can also be set with environment variable SRC_CAMPAIGNS_TMP_DIR; if both are set, this flag will be used and not the environment variable.",
	)
	flagSet.StringVar(
		&caf.file, "f", "",
		"The campaign spec file to read.",
	)
	flagSet.BoolVar(
		&caf.keepLogs, "keep-logs", false,
		"Retain logs after executing steps.",
	)
	flagSet.StringVar(
		&caf.namespace, "namespace", "",
		"The user or organization namespace to place the campaign within. Default is the currently authenticated user.",
	)
	flagSet.StringVar(&caf.namespace, "n", "", "Alias for -namespace.")

	flagSet.IntVar(
		&caf.parallelism, "j", runtime.GOMAXPROCS(0),
		"The maximum number of parallel jobs. Default is GOMAXPROCS.",
	)
	flagSet.DurationVar(
		&caf.timeout, "timeout", 60*time.Minute,
		"The maximum duration a single set of campaign steps can take.",
	)
	flagSet.BoolVar(
		&caf.cleanArchives, "clean-archives", true,
		"If true, deletes downloaded repository archives after executing campaign steps.",
	)
	flagSet.BoolVar(
		&caf.skipErrors, "skip-errors", false,
		"If true, errors encountered while executing steps in a repository won't stop the execution of the campaign spec but only cause that repository to be skipped.",
	)

	flagSet.StringVar(
		&caf.workspace, "workspace", "auto",
		`Workspace mode to use ("auto", "bind", or "volume")`,
	)

	flagSet.BoolVar(verbose, "v", false, "print verbose output")

	return caf
}

func campaignsCreatePending(out *output.Output, message string) output.Pending {
	return out.Pending(output.Line("", campaignsPendingColor, message))
}

func campaignsCompletePending(p output.Pending, message string) {
	p.Complete(output.Line(campaignsSuccessEmoji, campaignsSuccessColor, message))
}

func campaignsDefaultCacheDir() string {
	uc, err := os.UserCacheDir()
	if err != nil {
		return ""
	}

	return path.Join(uc, "sourcegraph", "campaigns")
}

// campaignsDefaultTempDirPrefix returns the prefix to be passed to ioutil.TempFile. If the
// environment variable SRC_CAMPAIGNS_TMP_DIR is set, that is used as the
// prefix. Otherwise we use "/tmp".
func campaignsDefaultTempDirPrefix() string {
	p := os.Getenv("SRC_CAMPAIGNS_TMP_DIR")
	if p != "" {
		return p
	}
	// On macOS, we use an explicit prefix for our temp directories, because
	// otherwise Go would use $TMPDIR, which is set to `/var/folders` per
	// default on macOS. But Docker for Mac doesn't have `/var/folders` in its
	// default set of shared folders, but it does have `/tmp` in there.
	if runtime.GOOS == "darwin" {
		return "/tmp"

	}
	return os.TempDir()
}

func campaignsOpenFileFlag(flag *string) (io.ReadCloser, error) {
	if flag == nil || *flag == "" || *flag == "-" {
		return os.Stdin, nil
	}

	file, err := os.Open(*flag)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file %q", *flag)
	}
	return file, nil
}

// campaignsExecute performs all the steps required to upload the campaign spec
// to Sourcegraph, including execution as needed. The return values are the
// spec ID, spec URL, and error.
func campaignsExecute(ctx context.Context, out *output.Output, svc *campaigns.Service, flags *campaignsApplyFlags) (campaigns.CampaignSpecID, string, error) {
	if err := checkExecutable("git", "version"); err != nil {
		return "", "", err
	}

	if err := checkExecutable("docker", "version"); err != nil {
		return "", "", err
	}

	// Parse flags and build up our service and executor options.

	specFile, err := campaignsOpenFileFlag(&flags.file)
	if err != nil {
		return "", "", err
	}
	defer specFile.Close()

	pending := campaignsCreatePending(out, "Parsing campaign spec")
	campaignSpec, rawSpec, err := campaignsParseSpec(out, svc, specFile)
	if err != nil {
		return "", "", err
	}
	campaignsCompletePending(pending, "Parsing campaign spec")

	pending = campaignsCreatePending(out, "Resolving namespace")
	namespace, err := svc.ResolveNamespace(ctx, flags.namespace)
	if err != nil {
		return "", "", err
	}
	campaignsCompletePending(pending, "Resolving namespace")

	imageProgress := out.Progress([]output.ProgressBar{{
		Label: "Preparing container images",
		Max:   1.0,
	}}, nil)
	err = svc.SetDockerImages(ctx, campaignSpec, func(perc float64) {
		imageProgress.SetValue(0, perc)
	})
	if err != nil {
		return "", "", err
	}
	imageProgress.Complete()

	pending = campaignsCreatePending(out, "Resolving repositories")
	repos, err := svc.ResolveRepositories(ctx, campaignSpec)
	if err != nil {
		if repoSet, ok := err.(campaigns.UnsupportedRepoSet); ok {
			campaignsCompletePending(pending, "Resolved repositories")

			block := out.Block(output.Line(" ", output.StyleWarning, "Some repositories are hosted on unsupported code hosts and will be skipped. Use the -allow-unsupported flag to avoid skipping them."))
			for repo := range repoSet {
				block.Write(repo.Name)
			}
			block.Close()
		} else {
			return "", "", errors.Wrap(err, "resolving repositories")
		}
	} else {
		campaignsCompletePending(pending, fmt.Sprintf("Resolved %d repositories", len(repos)))
	}

	pending = campaignsCreatePending(out, "Determining workspaces")
	tasks, err := svc.BuildTasks(ctx, repos, campaignSpec)
	if err != nil {
		return "", "", errors.Wrap(err, "Calculating execution plan")
	}
	campaignsCompletePending(pending, fmt.Sprintf("Found %d workspaces", len(tasks)))

	pending = campaignsCreatePending(out, "Preparing workspaces")
	workspaceCreator := svc.NewWorkspaceCreator(ctx, flags.cacheDir, flags.tempDir, campaignSpec.Steps)
	pending.VerboseLine(output.Linef("🚧", output.StyleSuccess, "Workspace creator: %T", workspaceCreator))
	campaignsCompletePending(pending, "Prepared workspaces")

	opts := campaigns.ExecutorOpts{
		Cache:       svc.NewExecutionCache(flags.cacheDir),
		Creator:     workspaceCreator,
		RepoFetcher: svc.NewRepoFetcher(flags.cacheDir, flags.cleanArchives),
		ClearCache:  flags.clearCache,
		KeepLogs:    flags.keepLogs,
		Timeout:     flags.timeout,
		TempDir:     flags.tempDir,
		Parallelism: flags.parallelism,
	}

	p := newCampaignProgressPrinter(out, *verbose, flags.parallelism)
	specs, logFiles, err := svc.ExecuteCampaignSpec(ctx, opts, tasks, campaignSpec, p.PrintStatuses, flags.skipErrors)
	if err != nil && !flags.skipErrors {
		return "", "", err
	}
	p.Complete()
	if err != nil && flags.skipErrors {
		printExecutionError(out, err)
		out.WriteLine(output.Line(output.EmojiWarning, output.StyleWarning, "Skipping errors because -skip-errors was used."))
	}

	if len(logFiles) > 0 && flags.keepLogs {
		func() {
			block := out.Block(output.Line("", campaignsSuccessColor, "Preserving log files:"))
			defer block.Close()

			for _, file := range logFiles {
				block.Write(file)
			}
		}()
	}

	ids := make([]campaigns.ChangesetSpecID, len(specs))

	if len(specs) > 0 {
		var label string
		if len(specs) == 1 {
			label = "Sending changeset spec"
		} else {
			label = fmt.Sprintf("Sending %d changeset specs", len(specs))
		}

		progress := out.Progress([]output.ProgressBar{
			{Label: label, Max: float64(len(specs))},
		}, nil)

		for i, spec := range specs {
			id, err := svc.CreateChangesetSpec(ctx, spec)
			if err != nil {
				return "", "", err
			}
			ids[i] = id
			progress.SetValue(0, float64(i+1))
		}
		progress.Complete()
	} else {
		if len(repos) == 0 {
			out.WriteLine(output.Linef(output.EmojiWarning, output.StyleWarning, `No changeset specs created`))
		}
	}

	pending = campaignsCreatePending(out, "Creating campaign spec on Sourcegraph")
	id, url, err := svc.CreateCampaignSpec(ctx, namespace, rawSpec, ids)
	campaignsCompletePending(pending, "Creating campaign spec on Sourcegraph")
	if err != nil {
		return "", "", prettyPrintCampaignsUnlicensedError(out, err)
	}

	return id, url, nil
}

// campaignsParseSpec parses and validates the given campaign spec. If the spec
// has validation errors, the errors are output in a human readable form and an
// exitCodeError is returned.
func campaignsParseSpec(out *output.Output, svc *campaigns.Service, input io.ReadCloser) (*campaigns.CampaignSpec, string, error) {
	spec, raw, err := svc.ParseCampaignSpec(input)
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			block := out.Block(output.Line("\u274c", output.StyleWarning, "Campaign spec failed validation."))
			defer block.Close()

			for i, err := range merr.Errors {
				block.Writef("%d. %s", i+1, err)
			}

			return nil, "", &exitCodeError{
				error:    nil,
				exitCode: 2,
			}
		} else {
			// This shouldn't happen; let's just punt and let the normal
			// rendering occur.
			return nil, "", err
		}
	}

	return spec, raw, nil
}

// printExecutionError is used to print the possible error returned by
// campaignsExecute.
func printExecutionError(out *output.Output, err error) {
	// exitCodeError shouldn't generate any specific output, since it indicates
	// that this was done deeper in the call stack.
	if _, ok := err.(*exitCodeError); ok {
		return
	}

	out.Write("")

	writeErrs := func(errs []error) {
		var block *output.Block

		if len(errs) > 1 {
			block = out.Block(output.Linef(output.EmojiFailure, output.StyleWarning, "%d errors:", len(errs)))
		} else {
			block = out.Block(output.Line(output.EmojiFailure, output.StyleWarning, "Error:"))
		}

		for _, e := range errs {
			if taskErr, ok := e.(campaigns.TaskExecutionErr); ok {
				block.Write(formatTaskExecutionErr(taskErr))
			} else {
				if err == context.Canceled {
					block.Writef("%sAborting", output.StyleBold)
				} else {
					block.Writef("%s%s=%+v)", output.StyleBold, e.Error())
				}
			}
		}

		if block != nil {
			block.Close()
		}
	}

	switch err := err.(type) {
	case parallel.Errors, *multierror.Error, api.GraphQlErrors:
		writeErrs(flattenErrs(err))

	default:
		writeErrs([]error{err})
	}

	out.Write("")
	out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleSuggestion, "The troubleshooting documentation can help to narrow down the cause of the errors: https://docs.sourcegraph.com/campaigns/references/troubleshooting"))
}

func flattenErrs(err error) (result []error) {
	switch errs := err.(type) {
	case parallel.Errors:
		for _, e := range errs {
			result = append(result, flattenErrs(e)...)
		}

	case *multierror.Error:
		for _, e := range errs.Errors {
			result = append(result, flattenErrs(e)...)
		}

	case api.GraphQlErrors:
		for _, e := range errs {
			result = append(result, flattenErrs(e)...)
		}

	default:
		result = append(result, errs)
	}

	return result
}

func formatTaskExecutionErr(err campaigns.TaskExecutionErr) string {
	if ee, ok := errors.Cause(err).(*exec.ExitError); ok && ee.String() == "signal: killed" {
		return fmt.Sprintf(
			"%s%s%s: killed by interrupt signal",
			output.StyleBold,
			err.Repository,
			output.StyleReset,
		)
	}

	return fmt.Sprintf(
		"%s%s%s:\n%s\nLog: %s\n",
		output.StyleBold,
		err.Repository,
		output.StyleReset,
		err.Err,
		err.Logfile,
	)
}

// prettyPrintCampaignsUnlicensedError introspects the given error returned when
// creating a campaign spec and ascertains whether it's a licensing error. If it
// is, then a better message is output. Regardless, the return value of this
// function should be used to replace the original error passed in to ensure
// that the displayed output is sensible.
func prettyPrintCampaignsUnlicensedError(out *output.Output, err error) error {
	// Pull apart the error to see if it's a licensing error: if so, we should
	// display a friendlier and more actionable message than the usual GraphQL
	// error output.
	if gerrs, ok := err.(api.GraphQlErrors); ok {
		// A licensing error should be the sole error returned, so we'll only
		// pretty print if there's one error.
		if len(gerrs) == 1 {
			if code, cerr := gerrs[0].Code(); cerr != nil {
				// We got a malformed value in the error extensions; at this
				// point, there's not much sensible we can do. Let's log this in
				// verbose mode, but let the original error bubble up rather
				// than this one.
				out.Verbosef("Unexpected error parsing the GraphQL error: %v", cerr)
			} else if code == "ErrCampaignsUnlicensed" {
				// OK, let's print a better message, then return an
				// exitCodeError to suppress the normal automatic error block.
				// Note that we have hand wrapped the output at 80 (printable)
				// characters: having automatic wrapping some day would be nice,
				// but this should be sufficient for now.
				block := out.Block(output.Line("🪙", output.StyleWarning, "Campaigns is a paid feature of Sourcegraph. All users can create sample"))
				block.WriteLine(output.Linef("", output.StyleWarning, "campaigns with up to 5 changesets without a license. Contact Sourcegraph sales"))
				block.WriteLine(output.Linef("", output.StyleWarning, "at %shttps://about.sourcegraph.com/contact/sales/%s to obtain a trial license.", output.StyleSearchLink, output.StyleWarning))
				block.Write("")
				block.WriteLine(output.Linef("", output.StyleWarning, "To proceed with this campaign, you will need to create 5 or fewer changesets."))
				block.WriteLine(output.Linef("", output.StyleWarning, "To do so, you could try adding %scount:5%s to your %srepositoriesMatchingQuery%s search,", output.StyleSearchAlertProposedQuery, output.StyleWarning, output.StyleReset, output.StyleWarning))
				block.WriteLine(output.Linef("", output.StyleWarning, "or reduce the number of changesets in %simportChangesets%s.", output.StyleReset, output.StyleWarning))
				block.Close()
				return &exitCodeError{exitCode: graphqlErrorsExitCode}
			}
		}
	}

	// In all other cases, we'll just return the original error.
	return err
}

func sumDiffStats(fileDiffs []*diff.FileDiff) diff.Stat {
	sum := diff.Stat{}
	for _, fileDiff := range fileDiffs {
		stat := fileDiff.Stat()
		sum.Added += stat.Added
		sum.Changed += stat.Changed
		sum.Deleted += stat.Deleted
	}
	return sum
}

func diffStatDescription(fileDiffs []*diff.FileDiff) string {
	var plural string
	if len(fileDiffs) > 1 {
		plural = "s"
	}

	return fmt.Sprintf("%d file%s changed", len(fileDiffs), plural)
}

func diffStatDiagram(stat diff.Stat) string {
	const maxWidth = 20
	added := float64(stat.Added + stat.Changed)
	deleted := float64(stat.Deleted + stat.Changed)
	if total := added + deleted; total > maxWidth {
		x := float64(20) / total
		added *= x
		deleted *= x
	}
	return fmt.Sprintf("%s%s%s%s%s",
		output.StyleLinesAdded, strings.Repeat("+", int(added)),
		output.StyleLinesDeleted, strings.Repeat("-", int(deleted)),
		output.StyleReset,
	)
}

func checkExecutable(cmd string, args ...string) error {
	if err := exec.Command(cmd, args...).Run(); err != nil {
		return fmt.Errorf(
			"failed to execute \"%s %s\":\n\t%s\n\n'src campaigns' require %q to be available.",
			cmd,
			strings.Join(args, " "),
			err,
			cmd,
		)
	}
	return nil
}

func contextCancelOnInterrupt(parent context.Context) (context.Context, func()) {
	ctx, ctxCancel := context.WithCancel(parent)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case <-c:
			ctxCancel()
		case <-ctx.Done():
		}
	}()

	return ctx, func() {
		signal.Stop(c)
		ctxCancel()
	}
}
