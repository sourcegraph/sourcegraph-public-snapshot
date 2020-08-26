package main

import (
	"context"
	"flag"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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
	clearCache       bool
	file             string
	keep             bool
	namespace        string
	parallelism      int
	timeout          time.Duration
}

func newCampaignsApplyFlags(flagSet *flag.FlagSet, cacheDir string) *campaignsApplyFlags {
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
		"Directory for caching results.",
	)
	flagSet.BoolVar(
		&caf.clearCache, "clear-cache", false,
		"If true, clears the cache and executes all steps anew.",
	)
	flagSet.StringVar(
		&caf.file, "f", "",
		"The campaign spec file to read.",
	)
	flagSet.BoolVar(
		&caf.keep, "keep-logs", false,
		"Retain logs after executing steps.",
	)
	flagSet.StringVar(
		&caf.namespace, "namespace", "",
		"The user of organization namespace to place the campaign within.",
	)
	flagSet.IntVar(
		&caf.parallelism, "j", 0,
		"The maximum number of parallel jobs. (Default: GOMAXPROCS.)",
	)
	flagSet.DurationVar(
		&caf.timeout, "timeout", 60*time.Minute,
		"The maximum duration a single set of campaign steps can take.",
	)

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
	// Parse flags and build up our service options.
	var errs *multierror.Error

	specFile, err := campaignsOpenFileFlag(&flags.file)
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		defer specFile.Close()
	}

	if flags.namespace == "" {
		errs = multierror.Append(errs, &usageError{errors.New("a namespace must be provided with -namespace")})
	}

	opts := campaigns.ExecutorOpts{
		Cache:      svc.NewExecutionCache(flags.cacheDir),
		ClearCache: flags.clearCache,
		KeepLogs:   flags.keep,
		Timeout:    flags.timeout,
	}
	if flags.parallelism <= 0 {
		opts.Parallelism = runtime.GOMAXPROCS(0)
	} else {
		opts.Parallelism = flags.parallelism
	}
	executor := svc.NewExecutor(opts, nil)

	if errs != nil {
		return "", "", errs
	}

	pending := campaignsCreatePending(out, "Parsing campaign spec")
	campaignSpec, err := svc.ParseCampaignSpec(specFile)
	if err != nil {
		return "", "", errors.Wrap(err, "parsing campaign spec")
	}

	if err := campaignsValidateSpec(out, campaignSpec); err != nil {
		return "", "", err
	}
	campaignsCompletePending(pending, "Parsing campaign spec")

	pending = campaignsCreatePending(out, "Resolving namespace")
	namespace, err := svc.ResolveNamespace(ctx, flags.namespace)
	if err != nil {
		return "", "", err
	}
	campaignsCompletePending(pending, "Resolving namespace")

	var progress output.Progress
	specs, err := svc.ExecuteCampaignSpec(ctx, executor, campaignSpec, func(statuses []*campaigns.TaskStatus) {
		if progress == nil {
			progress = out.Progress([]output.ProgressBar{{
				Label: "Executing steps",
				Max:   float64(len(statuses)),
			}}, nil)
		}

		complete := 0
		for _, ts := range statuses {
			if !ts.FinishedAt.IsZero() {
				complete += 1
			}
		}
		progress.SetValue(0, float64(complete))
	})
	if err != nil {
		return "", "", err
	}
	if progress != nil {
		progress.Complete()
	}

	if logFiles := executor.LogFiles(); len(logFiles) > 0 && flags.keep {
		func() {
			block := out.Block(output.Line("", campaignsSuccessColor, "Preserving log files:"))
			defer block.Close()

			for _, file := range logFiles {
				block.Write(file)
			}
		}()
	}

	progress = out.Progress([]output.ProgressBar{
		{Label: "Sending changeset specs", Max: float64(len(specs))},
	}, nil)
	ids := make([]campaigns.ChangesetSpecID, len(specs))
	for i, spec := range specs {
		id, err := svc.CreateChangesetSpec(ctx, spec)
		if err != nil {
			return "", "", err
		}
		ids[i] = id
		progress.SetValue(0, float64(i+1))
	}
	progress.Complete()

	pending = campaignsCreatePending(out, "Creating campaign spec on Sourcegraph")
	id, url, err := svc.CreateCampaignSpec(ctx, namespace, campaignSpec, ids)
	if err != nil {
		return "", "", err
	}
	campaignsCompletePending(pending, "Creating campaign spec on Sourcegraph")

	return id, url, nil
}
