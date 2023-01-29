package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/batches/ui"
)

func init() {
	usage := `'src batch remote' runs a batch spec on the Sourcegraph instance.

Usage:

    src batch remote [-f FILE]
    src batch remote FILE

Examples:

    $ src batch remote -f batch.spec.yaml

`

	flagSet := flag.NewFlagSet("remote", flag.ExitOnError)
	flags := newBatchExecutionFlags(flagSet)

	var (
		fileFlag = flagSet.String("f", "", "The name of the batch spec file to run.")
	)

	handler := func(args []string) error {
		// Various bits of Batch Changes boilerplate.
		ctx := context.Background()

		if err := flagSet.Parse(args); err != nil {
			return err
		}

		file, err := getBatchSpecFile(flagSet, fileFlag)
		if err != nil {
			return err
		}

		svc := service.New(&service.Opts{
			Client: cfg.apiClient(flags.api, flagSet.Output()),
		})

		_, ffs, err := svc.DetermineLicenseAndFeatureFlags(ctx)
		if err != nil {
			return err
		}

		if err := validateSourcegraphVersionConstraint(ctx, ffs); err != nil {
			return err
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		ui := &ui.TUI{Out: out}

		// OK, now for the real stuff. We have to load in the batch spec, and we
		// may as well validate it at the same time so we don't even have to go to
		// the backend if it's invalid.
		ui.ParsingBatchSpec()
		spec, batchSpecDir, raw, err := parseBatchSpec(ctx, file, svc)
		if err != nil {
			ui.ParsingBatchSpecFailure(err)
			return err
		}
		ui.ParsingBatchSpecSuccess()

		// We're going to need the namespace ID, so let's figure that out.
		ui.ResolvingNamespace()
		namespace, err := svc.ResolveNamespace(ctx, flags.namespace)
		if err != nil {
			return err
		}
		ui.ResolvingNamespaceSuccess(namespace.ID)

		ui.SendingBatchChange()
		batchChangeID, batchChangeName, err := svc.UpsertBatchChange(ctx, spec.Name, namespace.ID)
		if err != nil {
			return err
		}
		ui.SendingBatchChangeSuccess()

		ui.SendingBatchSpec()
		batchSpecID, err := svc.CreateBatchSpecFromRaw(
			ctx,
			raw,
			namespace.ID,
			flags.allowIgnored,
			flags.allowUnsupported,
			flags.clearCache,
			batchChangeID,
		)
		if err != nil {
			return err
		}
		ui.SendingBatchSpecSuccess()

		hasWorkspaceFiles := false
		for _, step := range spec.Steps {
			if len(step.Mount) > 0 {
				hasWorkspaceFiles = true
				break
			}
		}
		if hasWorkspaceFiles {
			ui.UploadingWorkspaceFiles()
			if err = svc.UploadBatchSpecWorkspaceFiles(ctx, batchSpecDir, batchSpecID, spec.Steps); err != nil {
				return err
			}
			ui.UploadingWorkspaceFilesSuccess()
		}

		// Wait for the workspaces to be resolved.
		ui.ResolvingWorkspaces()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		var res *service.BatchSpecWorkspaceResolution
		for range ticker.C {
			res, err = svc.GetBatchSpecWorkspaceResolution(ctx, batchSpecID)
			if err != nil {
				return err
			}

			if res.State == "FAILED" {
				return errors.Newf("workspace resolution failed: %s", res.FailureMessage)
			} else if res.State == "COMPLETED" {
				break
			}
		}
		ui.ResolvingWorkspacesSuccess(res.Workspaces.TotalCount)

		// We have to enqueue this for execution with a separate operation.
		//
		// TODO: when the execute flag is wired up in the upsert mutation, just set
		// it there and remove this.
		ui.ExecutingBatchSpec()
		batchSpecID, err = svc.ExecuteBatchSpec(ctx, batchSpecID, flags.clearCache)
		if err != nil {
			return err
		}
		ui.ExecutingBatchSpecSuccess()

		executionURL := fmt.Sprintf(
			"%s/%s/batch-changes/%s/executions/%s",
			strings.TrimSuffix(cfg.Endpoint, "/"),
			strings.TrimPrefix(namespace.URL, "/"),
			batchChangeName,
			batchSpecID,
		)
		ui.RemoteSuccess(executionURL)

		return nil
	}

	batchCommands = append(batchCommands, &command{
		flagSet: flagSet,
		aliases: []string{},
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src batch %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}
