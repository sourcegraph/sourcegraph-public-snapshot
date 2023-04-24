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

  Delete a key-value pair metadata from a repository:

		$ src repos delete-metadata -repo=repoID -key=mykey

  [DEPRECATED] Note 'delete-kvp' is deprecated and will be removed in future release. Use 'delete-metadata' instead.
`

	flagSet := flag.NewFlagSet("delete-metadata", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag = flagSet.String("repo", "", `The ID of the repo with the key-value pair metadata to be deleted (required)`)
		keyFlag  = flagSet.String("key", "", `The name of the  metadata key to be deleted (required)`)
		apiFlags = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}
		if *repoFlag == "" {
			return errors.New("error: repo is required")
		}

		keyFlag = nil
		flagSet.Visit(func(f *flag.Flag) {
			if f.Name == "key" {
				key := f.Value.String()
				keyFlag = &key
			}
		})
		if keyFlag == nil {
			return errors.New("error: key is required")
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation deleteRepoMetadata(
  $repo: ID!,
  $key: String!,
) {
  deleteRepoKeyValuePair(
    repo: $repo,
    key: $key,
  ) {
    alwaysNil
  }
}`

		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repo": *repoFlag,
			"key":  *keyFlag,
		}).Do(context.Background(), nil); err != nil || !ok {
			return err
		}

		fmt.Printf("Key-value pair metadata with key '%s' deleted.\n", *keyFlag)
		return nil
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		aliases:   []string{"delete-kvp"},
		handler:   handler,
		usageFunc: usageFunc,
	})
}
