package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := fmt.Sprintf(`'src snapshot summary' generates summary data for acceptance testing of a restored Sourcegraph instance with 'src snapshot test'.

USAGE
	src login # site-admin authentication required
	src [-v] snapshot summary

SUMMARY DATA
	Use '-dump-requests' to see what data is requested, or open the generated summary at %q.
`, srcSnapshotSummaryPath)
	flagSet := flag.NewFlagSet("summary", flag.ExitOnError)
	apiFlags := api.NewFlags(flagSet)

	snapshotCommands = append(snapshotCommands, &command{
		flagSet: flagSet,
		handler: func(args []string) error {
			if err := flagSet.Parse(args); err != nil {
				return err
			}
			out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})

			client := cfg.apiClient(apiFlags, flagSet.Output())

			snapshotResult, err := fetchSnapshotSummary(context.Background(), client)
			if err != nil {
				return err
			}

			_ = os.MkdirAll(srcSnapshotDir, os.ModePerm)

			f, err := os.OpenFile(srcSnapshotSummaryPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "open snapshot file")
			}
			enc := json.NewEncoder(f)
			enc.SetIndent("", "\t")
			if err := enc.Encode(snapshotResult); err != nil {
				return errors.Wrap(err, "write snapshot file")
			}

			out.WriteLine(output.Emojif(output.EmojiSuccess, "Summary snapshot data generated in %q!", srcSnapshotSummaryPath))
			return nil
		},
		usageFunc: func() { fmt.Fprint(flag.CommandLine.Output(), usage) },
	})
}

type snapshotSummary struct {
	ExternalServices struct {
		TotalCount int
		Nodes      []struct {
			Kind string
			ID   string
		}
	}
	Site struct {
		AuthProviders struct {
			TotalCount int
			Nodes      []struct {
				ServiceType string
				ServiceID   string
			}
		}
	}
}

func fetchSnapshotSummary(ctx context.Context, client api.Client) (*snapshotSummary, error) {
	var snapshotResult snapshotSummary
	ok, err := client.NewQuery(`
		query GenerateSnapshotAcceptanceData {
			externalServices {
				totalCount
				nodes {
					kind
					id
				}
			}
			site {
				authProviders {
					totalCount
					nodes {
						serviceType
						serviceID
					}
				}
			}
		}
	`).Do(ctx, &snapshotResult)
	if err != nil {
		return nil, errors.Wrap(err, "generate snapshot")
	} else if !ok {
		return nil, errors.New("received no data")
	}
	return &snapshotResult, nil
}
