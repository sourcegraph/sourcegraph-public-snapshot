package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/instancehealth"
)

// note that this file is called '_testcmd.go' because '_test.go' cannot be used

func init() {
	usage := `'src snapshot test' uses exported summary data to validate a restored and upgraded instance.

USAGE
	src login # site-admin authentication required
	src [-v] snapshot test [-summary-path="./src-snapshot-summary.json"]
`
	flagSet := flag.NewFlagSet("test", flag.ExitOnError)
	snapshotSummaryPath := flagSet.String("summary-path", "./src-snapshot-summary.json", "path to read snapshot summary from")
	since := flagSet.Duration("since", 1*time.Hour, "duration ago to look for healthcheck data")
	apiFlags := api.NewFlags(flagSet)

	snapshotCommands = append(snapshotCommands, &command{
		flagSet: flagSet,
		handler: func(args []string) error {
			if err := flagSet.Parse(args); err != nil {
				return err
			}
			out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
			client := cfg.apiClient(apiFlags, flagSet.Output())

			// Fetch health data
			instanceHealth, err := instancehealth.GetIndicators(context.Background(), client)
			if err != nil {
				return err
			}

			// Optionally validate snapshots
			if *snapshotSummaryPath != "" {
				f, err := os.OpenFile(*snapshotSummaryPath, os.O_RDONLY, os.ModePerm)
				if err != nil {
					return errors.Wrap(err, "open snapshot file")
				}
				var recordedSummary snapshotSummary
				if err := json.NewDecoder(f).Decode(&recordedSummary); err != nil {
					return errors.Wrap(err, "read snapshot file")
				}
				// Fetch new snapshot
				newSummary, err := fetchSnapshotSummary(context.Background(), client)
				if err != nil {
					return errors.Wrap(err, "get snapshot")
				}
				if err := compareSnapshotSummaries(out, recordedSummary, *newSummary); err != nil {
					return err
				}
			}

			// generate checks set
			checks := instancehealth.NewChecks(*since, *instanceHealth)

			// Run checks
			var validationErrors error
			for _, check := range checks {
				validationErrors = errors.Append(validationErrors, check(out))
			}
			if validationErrors != nil {
				out.WriteLine(output.Linef(output.EmojiFailure, output.StyleFailure,
					"Critical issues found: %s", err.Error()))
				return errors.New("validation failed")
			}
			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess,
				"No critical issues found!"))
			return nil
		},
		usageFunc: func() { fmt.Fprint(flag.CommandLine.Output(), usage) },
	})
}

func compareSnapshotSummaries(out *output.Output, recordedSummary, newSummary snapshotSummary) error {
	b := out.Block(output.Styled(output.StyleBold, "Snapshot contents"))
	defer b.Close()

	// Compare
	diff := cmp.Diff(recordedSummary, newSummary)
	if diff != "" {
		b.WriteLine(output.Line(output.EmojiFailure, output.StyleFailure, "Snapshot diff detected:"))
		b.WriteCode("diff", diff)
		return errors.New("snapshot mismatch")
	}
	b.WriteLine(output.Emoji(output.EmojiSuccess, "Snapshots match!"))
	return nil
}
