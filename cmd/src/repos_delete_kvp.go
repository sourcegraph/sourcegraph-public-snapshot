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

  Delete a key-value pair from a repository:

    	$ src repos delete-kvp -repo=repoID -key=mykey

`

	flagSet := flag.NewFlagSet("delete-kvp", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag = flagSet.String("repo", "", `The ID of the repo with the key-value pair to be deleted (required)`)
		keyFlag  = flagSet.String("key", "", `The name of the key to be deleted (required)`)
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

		query := `mutation deleteKVP(
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

		fmt.Printf("Key-value pair with key '%s' deleted.\n", *keyFlag)
		return nil
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
