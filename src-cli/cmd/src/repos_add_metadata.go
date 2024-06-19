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

  Add a key-value pair metadata to a repository:

    	$ src repos add-metadata -repo=repoID -key=mykey -value=myvalue

  Omitting -value will create a tag (a key with a null value).

  [DEPRECATED] Note that 'add-kvp' is deprecated and will be removed in future release. Use 'add-metadata' instead.
`

	flagSet := flag.NewFlagSet("add-metadata", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag     = flagSet.String("repo", "", `The ID of the repo to add the key-value pair metadata to (required if -repo-name is not specified)`)
		repoNameFlag = flagSet.String("repo-name", "", `The name of the repo to add the key-value pair metadata to (required if -repo is not specified)`)
		keyFlag      = flagSet.String("key", "", `The name of the  metadata key to add (required)`)
		valueFlag    = flagSet.String("value", "", `The  metadata value associated with the  metadata key. Defaults to null.`)
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

		query := `mutation addRepoMetadata(
  $repo: ID!,
  $key: String!,
  $value: String,
) {
  addRepoKeyValuePair(
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
			fmt.Printf("Key-value pair metadata '%s:%v' created.\n", *keyFlag, *valueFlag)
		} else {
			fmt.Printf("Key-value pair metadata '%s:<nil>' created.\n", *keyFlag)
		}
		return nil
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		aliases:   []string{"add-kvp"},
		handler:   handler,
		usageFunc: usageFunc,
	})
}
