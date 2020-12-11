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
		&caf.parallelism, "j", 0,
		"The maximum number of parallel jobs. (Default: GOMAXPROCS.)",
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

	// Parse flags and build up our service options.
	var errs *multierror.Error

	specFile, err := campaignsOpenFileFlag(&flags.file)
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		defer specFile.Close()
	}

	opts := campaigns.ExecutorOpts{
		Cache:      svc.NewExecutionCache(flags.cacheDir),
		Creator:    svc.NewWorkspaceCreator(flags.cacheDir, flags.cleanArchives),
		ClearCache: flags.clearCache,
		KeepLogs:   flags.keepLogs,
		Timeout:    flags.timeout,
		TempDir:    flags.tempDir,
	}
	if flags.parallelism <= 0 {
		opts.Parallelism = runtime.GOMAXPROCS(0)
	} else {
		opts.Parallelism = flags.parallelism
	}
	executor := svc.NewExecutor(opts)

	if errs != nil {
		return "", "", errs
	}

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
		Max:   float64(len(campaignSpec.Steps)),
	}}, nil)
	err = svc.SetDockerImages(ctx, campaignSpec, func(step int) {
		imageProgress.SetValue(0, float64(step))
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
		campaignsCompletePending(pending, "Resolved repositories")
	}

	p := newCampaignProgressPrinter(out, *verbose, opts.Parallelism)
	specs, err := svc.ExecuteCampaignSpec(ctx, repos, executor, campaignSpec, p.PrintStatuses, flags.skipErrors)
	if err != nil && !flags.skipErrors {
		return "", "", err

	}
	p.Complete()
	if err != nil && flags.skipErrors {
		printExecutionError(out, err)
		out.WriteLine(output.Line(output.EmojiWarning, output.StyleWarning, "Skipping errors because -skip-errors was used."))
	}

	if logFiles := executor.LogFiles(); len(logFiles) > 0 && flags.keepLogs {
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
	if err != nil {
		return "", "", err
	}
	campaignsCompletePending(pending, "Creating campaign spec on Sourcegraph")

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
				block.Writef("%s%s", output.StyleBold, e.Error())
			}
		}

		if block != nil {
			block.Close()
		}
	}

	switch err := err.(type) {
	case parallel.Errors, *multierror.Error:
		writeErrs(flattenErrs(err))

	default:
		writeErrs([]error{err})
	}

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
