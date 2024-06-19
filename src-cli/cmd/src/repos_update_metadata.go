package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  Update the metadata value for a metadata key on a repository:

    	$ src repos update-metadata -repo=repoID -key=my-key -value=new-value

  Omitting -value will set the value of the key to null.

  [DEPRECATED] Note that 'update-kvp' is deprecated and will be removed in future release. Use 'update-metadata' instead.
`

	flagSet := flag.NewFlagSet("update-metadata", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag     = flagSet.String("repo", "", `The ID of the repo with the metadata key to be updated (required if -repo-name is not specified)`)
		repoNameFlag = flagSet.String("repo-name", "", `The name of the repo to add the key-value pair metadata to (required if -repo is not specified)`)
		keyFlag      = flagSet.String("key", "", `The name of the metadata key to be updated (required)`)
		valueFlag    = flagSet.String("value", "", `The new metadata value of the metadata key to be set. Defaults to null.`)
		apiFlags     = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		keyFlag = nil
		valueFlag = nil
		flagSet.Visit(func(f *flag.Flag) {
			if f.Name == "key" {
				key := f.Value.String()
				keyFlag = &key
			}

			if f.Name == "value" {
				value := f.Value.String()
				valueFlag = &value
			}
		})
		if keyFlag == nil {
			return errors.New("error: key is required")
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())
		ctx := context.Background()
		repoID, err := getRepoIdOrError(ctx, client, repoFlag, repoNameFlag)
		if err != nil {
			return err
		}

		query := `mutation updateMetadata(
  $repo: ID!,
  $key: String!,
  $value: String,
) {
  updateRepoKeyValuePair(
    repo: $repo,
    key: $key,
    value: $value,
  ) {
    alwaysNil
  }
}`

		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repo":  *repoID,
			"key":   *keyFlag,
			"value": valueFlag,
		}).Do(ctx, nil); err != nil || !ok {
			return err
		}

		if valueFlag != nil {
			fmt.Printf("Value of key '%s' updated to '%v'\n", *keyFlag, *valueFlag)
		} else {
			fmt.Printf("Value of key '%s' updated to <nil>\n", *keyFlag)
		}
		return nil
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		aliases:   []string{"update-kvp"},
		handler:   handler,
		usageFunc: usageFunc,
	})
}
