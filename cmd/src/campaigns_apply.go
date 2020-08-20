package main

import (
	"context"
	"flag"
	"fmt"
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

func init() {
	usage := `
'src campaigns apply' is used to create or update a campaign on a Sourcegraph
instance.

Usage:

    src campaigns apply -f FILE -namespace NAMESPACE [command options]

Examples:

    Create a campaign spec, but don't apply it:

        $ src campaigns apply -f campaign.spec.yaml -namespace myuser

    Create a campaign spec, and apply it to create or update a campaign:

        $ src campagins apply -f campaign.spec.yaml -namespace myorg -apply

`

	cacheDir := defaultCacheDir()

	flagSet := flag.NewFlagSet("apply", flag.ExitOnError)
	var (
		allowUnsupportedFlag = flagSet.Bool("allow-unsupported", false, "Allow unsupported code hosts.")
		applyFlag            = flagSet.Bool("apply", false, "Immediately apply the campaign spec to create or update a campaign.")
		cacheDirFlag         = flagSet.String("cache", cacheDir, "Directory for caching results.")
		clearCacheFlag       = flagSet.Bool("clear-cache", false, "If true, clears the cache and executes all steps anew.")
		fileFlag             = flagSet.String("f", "", "The campaign spec file to read.")
		keepFlag             = flagSet.Bool("keep-logs", false, "Retain logs after executing steps.")
		namespaceFlag        = flagSet.String("namespace", "", "The user or organization namespace to place the campaign within.")
		parallelismFlag      = flagSet.Int("j", 0, "The maximum number of parallel jobs. (Default: GOMAXPROCS.)")
		timeoutFlag          = flagSet.Duration("timeout", 60*time.Minute, "The maximum duration a single set of campaign steps can take.")
		apiFlags             = api.NewFlags(flagSet)
	)

	var (
		pendingColor = output.StylePending
		successColor = output.StyleSuccess
		successEmoji = output.EmojiSuccess
	)

	createPending := func(out *output.Output, message string) output.Pending {
		return out.Pending(output.Line("", pendingColor, message))
	}

	completePending := func(p output.Pending, message string) {
		p.Complete(output.Line(successEmoji, successColor, message))
	}

	doApply := func(out *output.Output) error {
		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		// Parse flags and build up our service options.
		var errs *multierror.Error
		svc := campaigns.NewService(&campaigns.ServiceOpts{
			AllowUnsupported: *allowUnsupportedFlag,
			Client:           client,
		})

		specFile, err := campaignsOpenFileFlag(fileFlag)
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			defer specFile.Close()
		}

		if namespaceFlag == nil || *namespaceFlag == "" {
			errs = multierror.Append(errs, &usageError{errors.New("a namespace must be provided with -namespace")})
		}

		opts := campaigns.ExecutorOpts{
			Cache:      svc.NewExecutionCache(*cacheDirFlag),
			ClearCache: *clearCacheFlag,
			KeepLogs:   *keepFlag,
			Timeout:    *timeoutFlag,
		}
		if parallelismFlag != nil || *parallelismFlag <= 0 {
			opts.Parallelism = runtime.GOMAXPROCS(0)
		} else {
			opts.Parallelism = *parallelismFlag
		}
		executor := svc.NewExecutor(opts, nil)

		if errs != nil {
			return errs
		}

		pending := createPending(out, "Parsing campaign spec")
		campaignSpec, err := svc.ParseCampaignSpec(specFile)
		if err != nil {
			return errors.Wrap(err, "parsing campaign spec")
		}

		if err := campaignsValidateSpec(out, campaignSpec); err != nil {
			return err
		}
		completePending(pending, "Parsing campaign spec")

		pending = createPending(out, "Resolving namespace")
		namespace, err := svc.ResolveNamespace(ctx, *namespaceFlag)
		if err != nil {
			return err
		}
		completePending(pending, "Resolving namespace")

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
			return err
		}
		if progress != nil {
			progress.Complete()
		}

		if logFiles := executor.LogFiles(); len(logFiles) > 0 && *keepFlag {
			block := out.Block(output.Line("", successColor, "Preserving log files:"))
			for _, file := range logFiles {
				block.Write(file)
			}
		}

		progress = out.Progress([]output.ProgressBar{
			{Label: "Sending changeset specs", Max: float64(len(specs))},
		}, nil)
		ids := make([]campaigns.ChangesetSpecID, len(specs))
		for i, spec := range specs {
			id, err := svc.CreateChangesetSpec(ctx, spec)
			if err != nil {
				return err
			}
			ids[i] = id
			progress.SetValue(0, float64(i+1))
		}
		progress.Complete()

		pending = createPending(out, "Creating campaign spec on Sourcegraph")
		id, url, err := svc.CreateCampaignSpec(ctx, namespace, campaignSpec, ids)
		if err != nil {
			return err
		}
		completePending(pending, "Creating campaign spec on Sourcegraph")

		if *applyFlag {
			pending := createPending(out, "Applying campaign spec")
			campaign, err := svc.ApplyCampaign(ctx, id)
			if err != nil {
				return err
			}
			completePending(pending, "Applying campaign spec")

			out.Write("")
			block := out.Block(output.Line(successEmoji, successColor, "Campaign applied!"))
			block.Write("To view the campaign, go to:")
			block.Writef("%s%s", cfg.Endpoint, campaign.URL)
		} else {
			out.Write("")
			block := out.Block(output.Line(successEmoji, successColor, "To preview or apply the campaign spec, go to:"))
			block.Writef("%s%s", cfg.Endpoint, url)
		}

		return nil
	}

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		if err := doApply(out); err != nil {
			out.Write("")
			block := out.Block(output.Line("❌", output.StyleWarning, "Error"))
			block.Write(err.Error())
		}

		return nil
	}

	campaignsCommands = append(campaignsCommands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

func defaultCacheDir() string {
	uc, err := os.UserCacheDir()
	if err != nil {
		return ""
	}

	return path.Join(uc, "sourcegraph", "campaigns")
}
